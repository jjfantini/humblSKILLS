#!/usr/bin/env python3
"""Engine behind scripts/promote.sh.

Parses SKILL.md frontmatter, fills and validates the taxonomy the CLI
organizes skills by (category, role, tags), runs the promotion quality gate,
stages a clean copy into a registry checkout, and renders the commit message
and pull-request body.

Standard library only: it has to run on a stock macOS python3, which ships
without PyYAML. That is why the frontmatter parser below implements just the
YAML subset SKILL.md files use - block and flow mappings and sequences, plain,
quoted, and block (| >) scalars - and rejects the constructs the CLI's YAML
parser would also reject, such as an unquoted value containing ": ".

promote.sh is the only supported caller; the subcommands are its plumbing.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import shutil
import stat
import subprocess
import sys
import tempfile
import textwrap
from pathlib import Path

# Closed sets the CLI validates against: frontmatter.Categories and
# frontmatter.Roles in cli/internal/frontmatter/validate.go, and the adapter
# names in cli/internal/adapters/*.yaml. Duplicated here because this script
# runs on machines that only have the humblskills binary.
# cli/cmd/build-registry/promote_sync_test.go fails the humblSKILLS build if
# these drift.
CATEGORIES = ("development", "design", "writing", "meta")
ROLES = ("fde", "ds", "sdr")
PLATFORMS = ("claude-code", "claude-desktop", "codex", "cursor", "pi")

CATEGORY_HINTS = {
    "development": "git/workflow tooling, integrations, APIs, data and infra work",
    "design": "frontend, UI/UX, creative and media generation",
    "writing": "content, copy, and editing",
    "meta": "skill authoring, project onboarding, humblSKILLS-about-humblSKILLS",
}
ROLE_HINTS = {
    "fde": "forward-deployed engineer",
    "ds": "data scientist",
    "sdr": "sales development representative",
}

TARGETS = {
    "humblskills": {
        "label": "humblSKILLS",
        "repo": "jjfantini/humblSKILLS",
        "visibility": "public",
        "base": "develop",
        "role_required": False,
    },
    "happyskills": {
        "label": "happySKILLS",
        "repo": "jenningsfantini-happyrobot/happySKILLS",
        "visibility": "private",
        # Empty means: develop when the repo has one, else its default branch.
        "base": "",
        "role_required": True,
    },
}
TARGET_ALIASES = {
    "humblskills": "humblskills",
    "public": "humblskills",
    "happyskills": "happyskills",
    "happyrobot": "happyskills",
    "private": "happyskills",
}

# Every path a smart skill's brain lives in. Each must be preserved, or
# `humblskills update` overwrites what users accumulated with the registry copy.
BRAIN_PRESERVE = (
    "references/raw/",
    "references/wiki/",
    "references/decisions.md",
    "references/log.md",
    "references/patterns.md",
)
SMART_FILES = (
    "references/_index.md",
    "references/_brain.md",
    "references/_template.md",
    "references/patterns.md",
    "references/decisions.md",
    "references/log.md",
    "scripts/lint.sh",
)
SMART_DIRS = ("references/raw", "references/wiki")
TOP_LEVEL_KEYS = ("name", "description", "license", "compatibility", "allowed-tools", "metadata", "upstream")
LEGACY_TOP_LEVEL = ("version", "tags", "platforms", "requires", "preserve")

# Never shipped: OS and editor droppings, caches, and eval scratch space.
JUNK_DIRS = {".git", "__pycache__", ".pytest_cache", ".mypy_cache", ".ruff_cache", ".eval-workspace", ".idea", ".vscode"}
JUNK_FILES = {".DS_Store", "Thumbs.db", "desktop.ini", "Icon\r"}
JUNK_SUFFIXES = (".pyc", ".pyo", ".swp", ".swo", "~")
# Installed dependencies are never source; shipping them bloats every install.
DEPENDENCY_DIRS = {"node_modules", ".venv", "venv", "site-packages"}
SENSITIVE_NAMES = {".netrc", ".npmrc", ".pypirc", "id_rsa", "id_dsa", "id_ecdsa", "id_ed25519"}
SENSITIVE_SUFFIXES = (".p12", ".pfx", ".jks", ".keystore")
SAFE_ENV_SUFFIXES = (".example", ".sample", ".template")

# install downloads the whole repository tarball for every skill, so every byte
# a skill adds is paid by every install of every skill.
MiB = 1024 * 1024
FILE_WARN, FILE_FAIL = 1 * MiB, 5 * MiB
TOTAL_WARN, TOTAL_FAIL = 5 * MiB, 15 * MiB
SCAN_LIMIT = 2 * MiB

SECRET_PATTERNS = (
    ("private key", re.compile(r"-----BEGIN (?:[A-Z0-9]+ )*PRIVATE KEY(?: BLOCK)?-----")),
    ("AWS access key", re.compile(r"\b(?:AKIA|ASIA)[0-9A-Z]{16}\b")),
    ("GitHub token", re.compile(r"\b(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9]{36,}\b")),
    ("GitHub token", re.compile(r"\bgithub_pat_[A-Za-z0-9_]{50,}")),
    ("Slack token", re.compile(r"\bxox[abposr]-[0-9A-Za-z-]{10,}")),
    ("Slack webhook", re.compile(r"https://hooks\.slack\.com/services/T[0-9A-Z]{6,}/B[0-9A-Z]{6,}/[0-9A-Za-z]{16,}")),
    ("Anthropic API key", re.compile(r"\bsk-ant-[A-Za-z0-9_-]{20,}")),
    ("OpenAI API key", re.compile(r"\bsk-(?:proj|svcacct|admin)-[A-Za-z0-9_-]{20,}|\bsk-[A-Za-z0-9]{40,}\b")),
    ("Google API key", re.compile(r"\bAIza[0-9A-Za-z_-]{35}\b")),
    ("live API secret", re.compile(r"\b(?:sk|rk)_live_[0-9A-Za-z]{16,}")),
)
# A PEM header alone (docs, redaction fixtures) is not a key; key material is.
PEM_BODY = re.compile(r"^\s*(?:[A-Za-z0-9+/=]{40,}|Proc-Type:)")
JWT_PATTERN = re.compile(r"\beyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}")
PLACEHOLDER_HINTS = ("example", "xxxx", "your", "redacted", "placeholder", "dummy", "fake", "sample")
HOME_PATH = re.compile(r"(?:/Users|/home)/([A-Za-z0-9._-]+)/")
HOME_PATH_OK = {"you", "user", "username", "me", "name", "runner", "ubuntu", "example", "alice", "bob", "dev", "test", "demo"}
EMAIL = re.compile(r"[A-Za-z0-9._%+-]+@([A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,})")
EMAIL_OK_DOMAINS = ("example.com", "example.org", "example.net", "users.noreply.github.com", "anthropic.com")
INTERNAL_MENTION = re.compile(r"happy[\s_-]?robot", re.IGNORECASE)

NAME_RE = re.compile(r"^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$")
TAG_RE = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
SEMVER_RE = re.compile(
    r"^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)"
    r"(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$"
)
# "Use when ...", "Use it for ...", "Trigger on ..." - the clause that tells the
# router when to pick the skill. Every published description has one.
TRIGGER_RE = re.compile(
    r"\b(?:use (?:this skill |this |it |the skill )?(?:when|for|to|if|whenever|on)\b|use it\b|use this\b"
    r"|trigger(?:s|ed)? (?:on|when|by|for)\b|invoke (?:when|for)\b|activate (?:when|for)\b)",
    re.IGNORECASE,
)
NEGATIVE_TRIGGER_RE = re.compile(r"\b(?:do not use|don't use|not for\b|never use|instead\b)", re.IGNORECASE)
NAMING_RE = re.compile(r"(?:^|-)smart-[a-z0-9]")
PLACEHOLDER_BODY_RE = re.compile(r"<!--\s*TODO|\bTODO(?::|\s+short label\b)")
ANSI_RE = re.compile(r"\x1b\[[0-9;]*m")
LOG_ENTRY_RE = re.compile(r"^\[(?:INGEST|QUERY|LINT)\b", re.MULTILINE)
ENTRY_HEADING_RE = re.compile(r"^### \d{4}-\d{2}-\d{2} \|", re.MULTILINE)


class FrontmatterError(ValueError):
    pass


class UsageError(Exception):
    pass


# ---------------------------------------------------------------------------
# Frontmatter: the YAML subset SKILL.md uses
# ---------------------------------------------------------------------------

_KEY_RE = re.compile(r"^([A-Za-z0-9_][A-Za-z0-9_.\-]*)[ \t]*:(?:[ \t]+(.*)|[ \t]*)$")


def split_frontmatter(text: str):
    """Split SKILL.md into (frontmatter lines, body) the way the CLI does: an
    optional BOM and leading blank lines, an opening '---' line, and a closing
    line that is exactly '---' (trailing whitespace allowed)."""
    if text.startswith("\ufeff"):
        text = text[1:]
    stripped = text.lstrip(" \t\r\n")
    if not stripped.startswith("---"):
        raise FrontmatterError("missing leading '---' frontmatter delimiter")
    rest = stripped[3:].lstrip(" \t\r")
    if not rest.startswith("\n"):
        raise FrontmatterError("opening '---' must be followed by a newline")
    lines = rest[1:].split("\n")
    for i, line in enumerate(lines):
        if line.rstrip(" \t\r") == "---":
            return [ln.rstrip("\r") for ln in lines[:i]], "\n".join(lines[i + 1:]).lstrip(" \t\r\n")
    raise FrontmatterError("missing closing '---' frontmatter delimiter")


def parse_frontmatter(fm_lines):
    lines = list(fm_lines)
    value, i = _parse_mapping(lines, 0, 0)
    j = _next_content(lines, i)
    if j is not None:
        raise FrontmatterError(f"frontmatter line {j + 2}: could not parse {lines[j].strip()!r}")
    return value


def read_skill_md(path: Path):
    text = path.read_text(encoding="utf-8")
    fm_lines, body = split_frontmatter(text)
    return parse_frontmatter(fm_lines), body, text


def _indent(line: str) -> int:
    return len(line) - len(line.lstrip(" "))


def _skippable(line: str) -> bool:
    s = line.strip()
    return s == "" or s.startswith("#")


def _next_content(lines, i):
    while i < len(lines):
        if not _skippable(lines[i]):
            return i
        i += 1
    return None


def _is_seq_item(text: str) -> bool:
    return text == "-" or text.startswith("- ")


def _strip_comment(s: str) -> str:
    for k, ch in enumerate(s):
        if ch == "#" and (k == 0 or s[k - 1] in " \t"):
            return s[:k]
    return s


def _parse_mapping(lines, i, indent):
    out = {}
    while True:
        j = _next_content(lines, i)
        if j is None:
            return out, len(lines)
        line = lines[j]
        ind = _indent(line)
        if ind < indent:
            return out, j
        if ind > indent:
            raise FrontmatterError(f"frontmatter line {j + 2}: unexpected indentation")
        text = line[ind:]
        if _is_seq_item(text):
            return out, j
        m = _KEY_RE.match(text)
        if not m:
            raise FrontmatterError(f"frontmatter line {j + 2}: expected 'key: value', got {text.strip()!r}")
        key = m.group(1)
        if key in out:
            raise FrontmatterError(f"frontmatter line {j + 2}: duplicate key {key!r}")
        out[key], i = _parse_value(lines, j + 1, indent, m.group(2) or "", j)


def _parse_value(lines, i, indent, rest, key_line):
    head = _strip_comment(rest).strip()
    if head == "":
        j = _next_content(lines, i)
        if j is None:
            return None, len(lines)
        nind = _indent(lines[j])
        ntext = lines[j][nind:]
        if nind > indent:
            if _is_seq_item(ntext):
                return _parse_sequence(lines, j, nind)
            return _parse_mapping(lines, j, nind)
        if nind == indent and _is_seq_item(ntext):
            return _parse_sequence(lines, j, nind)
        return None, i
    if head[0] in "|>":
        return _parse_block_scalar(lines, i, indent, head, key_line)
    return _parse_inline(lines, i, indent, rest.strip(), key_line)


def _parse_sequence(lines, i, indent):
    out = []
    while True:
        j = _next_content(lines, i)
        if j is None:
            return out, len(lines)
        line = lines[j]
        ind = _indent(line)
        if ind > indent:
            raise FrontmatterError(f"frontmatter line {j + 2}: unexpected indentation")
        text = line[ind:]
        if ind < indent or not _is_seq_item(text):
            return out, j
        rest = text[1:].lstrip(" ")
        if rest == "" or rest.startswith("#"):
            value, i = _parse_value(lines, j + 1, indent, "", j)
            out.append(value)
            continue
        m = _KEY_RE.match(rest)
        if m and rest[0] not in "\"'[{":
            # "- key: value" opens a mapping item whose keys align with `key`.
            item_indent = ind + (len(text) - len(rest))
            lines[j] = " " * item_indent + rest
            value, i = _parse_mapping(lines, j, item_indent)
            out.append(value)
            continue
        value, i = _parse_inline(lines, j + 1, indent, rest, j)
        out.append(value)


def _parse_block_scalar(lines, i, indent, header, key_line):
    style, chomp, explicit = header[0], "clip", None
    for ch in header[1:]:
        if ch == "-":
            chomp = "strip"
        elif ch == "+":
            chomp = "keep"
        elif ch.isdigit() and ch != "0":
            explicit = int(ch)
        else:
            raise FrontmatterError(f"frontmatter line {key_line + 2}: bad block scalar header {header!r}")
    block = []
    j = i
    while j < len(lines):
        line = lines[j]
        if line.strip() == "":
            block.append("")
        elif _indent(line) <= indent:
            break
        else:
            block.append(line)
        j += 1
    trailing = 0
    while block and block[-1] == "":
        block.pop()
        trailing += 1
    if not block:
        return "", j
    first = next(ln for ln in block if ln)
    content_indent = indent + explicit if explicit else _indent(first)
    rows = []
    for ln in block:
        if ln and _indent(ln) < content_indent:
            raise FrontmatterError(f"frontmatter line {key_line + 2}: block scalar line is under-indented")
        rows.append(ln[content_indent:] if ln else "")
    text = "\n".join(rows) if style == "|" else _fold(rows)
    if chomp == "strip":
        return text, j
    if chomp == "keep":
        return text + "\n" * (1 + trailing), j
    return text + "\n", j


def _fold(rows):
    out = ""
    for k, row in enumerate(rows):
        if k == 0:
            out = row
            continue
        prev = rows[k - 1]
        if row == "":
            out += "\n"
        elif prev == "":
            out += row
        elif row[:1] in (" ", "\t") or prev[:1] in (" ", "\t"):
            out += "\n" + row
        else:
            out += " " + row
    return out


def _parse_inline(lines, i, indent, text, key_line):
    first = text[0]
    if first in "\"'":
        buf = text
        while True:
            value, end = _scan_quoted(buf, first)
            if end is not None:
                if _strip_comment(buf[end:]).strip():
                    raise FrontmatterError(f"frontmatter line {key_line + 2}: unexpected text after a quoted value")
                return value, i
            if i >= len(lines):
                raise FrontmatterError(f"frontmatter line {key_line + 2}: unterminated quoted value")
            buf += "\n" + lines[i]
            i += 1
    if first in "[{":
        buf = _strip_comment(text).strip()
        while not _flow_closed(buf):
            if i >= len(lines):
                raise FrontmatterError(f"frontmatter line {key_line + 2}: unterminated {first} collection")
            buf += " " + _strip_comment(lines[i]).strip()
            i += 1
        if first == "[":
            return _parse_flow_seq(buf, key_line), i
        return buf, i
    value = _plain(text, key_line)
    while i < len(lines):
        line = lines[i]
        if line.strip() == "":
            k = i
            while k < len(lines) and lines[k].strip() == "":
                k += 1
            if k < len(lines) and _indent(lines[k]) > indent and not lines[k].lstrip().startswith("#"):
                value += "\n" * (k - i)
                i = k
                continue
            break
        if _indent(line) <= indent or line.lstrip().startswith("#"):
            break
        part = _plain(line.strip(), i)
        value += part if value.endswith("\n") else " " + part
        i += 1
    return value, i


def _plain(text, line_no):
    t = _strip_comment(text).rstrip()
    if t[:1] in tuple("!&*@`%|>") or t.startswith(("? ", "- ")):
        raise FrontmatterError(f"frontmatter line {line_no + 2}: value starts with {t[:1]!r} - wrap it in quotes")
    if ": " in t or t.endswith(":"):
        raise FrontmatterError(
            f"frontmatter line {line_no + 2}: unquoted value has a ':' followed by a space or line end - wrap it in quotes"
        )
    return t


def _scan_quoted(buf, quote):
    out = []
    k = 1
    while k < len(buf):
        ch = buf[k]
        if quote == "'":
            if ch == "'":
                if buf[k + 1:k + 2] == "'":
                    out.append("'")
                    k += 2
                    continue
                return _fold_quoted("".join(out)), k + 1
            out.append(ch)
            k += 1
            continue
        if ch == "\\" and k + 1 < len(buf):
            nxt = buf[k + 1]
            simple = {"n": "\n", "t": "\t", '"': '"', "\\": "\\", "/": "/", "0": "\0", "r": "\r", " ": " "}
            if nxt in simple:
                out.append(simple[nxt])
                k += 2
                continue
            if nxt == "u" and re.match(r"[0-9a-fA-F]{4}", buf[k + 2:k + 6]):
                out.append(chr(int(buf[k + 2:k + 6], 16)))
                k += 6
                continue
            if nxt == "\n":
                k += 2
                while k < len(buf) and buf[k] in " \t":
                    k += 1
                continue
            out.append(nxt)
            k += 2
            continue
        if ch == '"':
            return _fold_quoted("".join(out)), k + 1
        out.append(ch)
        k += 1
    return None, None


def _fold_quoted(s):
    if "\n" not in s:
        return s
    parts = [p.strip(" \t") for p in s.split("\n")]
    out = parts[0]
    for p in parts[1:]:
        out += "\n" if p == "" else (" " + p if not out.endswith("\n") else p)
    return out


def _flow_closed(buf):
    depth, quote = 0, None
    for ch in buf:
        if quote:
            if ch == quote:
                quote = None
            continue
        if ch in "\"'":
            quote = ch
        elif ch in "[{":
            depth += 1
        elif ch in "]}":
            depth -= 1
    return depth == 0 and quote is None


def _parse_flow_seq(text, key_line):
    inner = text.strip()
    if not inner.endswith("]"):
        raise FrontmatterError(f"frontmatter line {key_line + 2}: unexpected text after ]")
    inner = inner[1:-1]
    items, buf, quote, depth = [], "", None, 0
    for ch in inner:
        if quote:
            buf += ch
            if ch == quote:
                quote = None
            continue
        if ch in "\"'":
            quote = ch
        elif ch in "[{":
            depth += 1
        elif ch in "]}":
            depth -= 1
        if ch == "," and depth == 0:
            items.append(buf)
            buf = ""
            continue
        buf += ch
    items.append(buf)
    out = []
    for raw in items:
        raw = raw.strip()
        if raw == "":
            continue
        if raw[0] in "\"'":
            value, end = _scan_quoted(raw, raw[0])
            if end is None or raw[end:].strip():
                raise FrontmatterError(f"frontmatter line {key_line + 2}: bad quoted item {raw!r}")
            out.append(value)
        elif raw[0] in "[{":
            out.append(raw)
        else:
            if ": " in raw:
                raise FrontmatterError(f"frontmatter line {key_line + 2}: unquoted item {raw!r} contains ': '")
            out.append(raw)
    return out


def as_list(value):
    if value is None:
        return []
    if isinstance(value, list):
        return [str(v) for v in value if v is not None]
    return [str(value)]


def as_str(value):
    if value is None:
        return ""
    if isinstance(value, (dict, list)):
        return ""
    return str(value)


def is_placeholder(value) -> bool:
    if isinstance(value, list):
        return any(is_placeholder(v) for v in value)
    return as_str(value).strip().upper().startswith("TODO")


# ---------------------------------------------------------------------------
# Taxonomy write-back: surgical edits under metadata:, everything else intact
# ---------------------------------------------------------------------------


class _MetadataEditor:
    def __init__(self, fm_lines):
        self.lines = list(fm_lines)
        self._locate()

    def _locate(self):
        meta = None
        for k, line in enumerate(self.lines):
            if re.match(r"^metadata[ \t]*:", line):
                if not re.match(r"^metadata[ \t]*:[ \t]*(#.*)?$", line):
                    raise FrontmatterError("metadata: must be a block mapping for promote.sh to edit it")
                meta = k
                break
        if meta is None:
            while self.lines and self.lines[-1].strip() == "":
                self.lines.pop()
            self.lines.append("metadata:")
            meta = len(self.lines) - 1
        self.meta = meta
        self.end = len(self.lines)
        for k in range(meta + 1, len(self.lines)):
            line = self.lines[k]
            if line.strip() and not line.lstrip().startswith("#") and _indent(line) == 0:
                self.end = k
                break
        self.child = "  "
        for k in range(meta + 1, self.end):
            line = self.lines[k]
            if line.strip() and not line.lstrip().startswith("#"):
                self.child = " " * _indent(line)
                break

    def _find(self, key):
        pat = re.compile(r"^" + re.escape(self.child) + re.escape(key) + r"[ \t]*:")
        for k in range(self.meta + 1, self.end):
            line = self.lines[k]
            if pat.match(line) and _indent(line) == len(self.child):
                return k
        return None

    def _extent(self, k):
        n = len(self.child)
        last, e = k, k + 1
        while e < self.end:
            line = self.lines[e]
            if _skippable(line):
                e += 1
                continue
            ind = _indent(line)
            if ind > n or (ind == n and _is_seq_item(line[n:])):
                last = e
                e += 1
                continue
            break
        return last + 1

    def set(self, key, value, after=()):
        line = f"{self.child}{key}: {value}"
        k = self._find(key)
        if k is not None:
            self.lines[k:self._extent(k)] = [line]
        else:
            pos = None
            for anchor in after:
                ak = self._find(anchor)
                if ak is not None:
                    pos = self._extent(ak)
                    break
            self.lines.insert(self.meta + 1 if pos is None else pos, line)
        self._locate()

    def remove(self, key):
        k = self._find(key)
        if k is not None:
            del self.lines[k:self._extent(k)]
            self._locate()


def normalize_tag(tag: str) -> str:
    t = str(tag).strip().strip("#").strip().lower()
    t = re.sub(r"[\s_]+", "-", t)
    return re.sub(r"-{2,}", "-", t).strip("-")


def merge_tags(existing, new_csv: str):
    out = []
    for raw in list(existing) + new_csv.split(","):
        t = normalize_tag(raw)
        if t and t != "todo" and t not in out:
            out.append(t)
    return out


def apply_taxonomy(skill_md: Path, category=None, role=None, tags=None):
    """Write category/role/tags into SKILL.md. Returns [(field, old, new)]."""
    raw = skill_md.read_text(encoding="utf-8")
    bom = "\ufeff" if raw.startswith("\ufeff") else ""
    lines = raw[len(bom):].split("\n")
    if not lines or lines[0].rstrip(" \t\r") != "---":
        raise FrontmatterError("SKILL.md must start with a '---' line for promote.sh to edit it")
    end = next((k for k in range(1, len(lines)) if lines[k].rstrip(" \t\r") == "---"), None)
    if end is None:
        raise FrontmatterError("missing closing '---' frontmatter delimiter")
    fm_lines = lines[1:end]
    meta = parse_frontmatter(fm_lines).get("metadata") or {}
    if not isinstance(meta, dict):
        raise FrontmatterError("metadata: must be a mapping")

    editor = _MetadataEditor(fm_lines)
    changes = []
    if category is not None and meta.get("category") != category:
        editor.set("category", category, after=("version", "author"))
        changes.append(("category", meta.get("category"), category))
    if role is not None:
        old = meta.get("role")
        if role == "none":
            if old is not None:
                editor.remove("role")
                changes.append(("role", old, None))
        elif old != role:
            editor.set("role", role, after=("category", "version", "author"))
            changes.append(("role", old, role))
    if tags is not None:
        old = as_list(meta.get("tags"))
        merged = merge_tags(old, tags)
        if merged != old:
            editor.set("tags", "[" + ", ".join(merged) + "]", after=("role", "category", "version", "author"))
            changes.append(("tags", old, merged))
    if not changes:
        return []

    new_meta = parse_frontmatter(editor.lines).get("metadata") or {}
    for field, _, new in changes:
        got = as_list(new_meta.get(field)) if field == "tags" else new_meta.get(field)
        if got != new:
            raise FrontmatterError(f"could not write {field} into SKILL.md safely; edit it by hand")
    out = bom + "\n".join([lines[0]] + editor.lines + lines[end:])
    tmp = skill_md.with_name(skill_md.name + ".promote-tmp")
    tmp.write_text(out, encoding="utf-8")
    os.chmod(tmp, stat.S_IMODE(os.stat(skill_md).st_mode))
    os.replace(tmp, skill_md)
    return changes


# ---------------------------------------------------------------------------
# Report
# ---------------------------------------------------------------------------

_COLOR = sys.stdout.isatty() and not os.environ.get("NO_COLOR")
_LEVELS = {"fail": ("31", "FAIL"), "warn": ("33", "WARN"), "info": ("36", "INFO"), "pass": ("32", "PASS")}


def _paint(code, s):
    return f"\033[{code}m{s}\033[0m" if _COLOR else s


def section(title):
    print(_paint("1", title))


class Report:
    def __init__(self, path=None):
        self.path = Path(path) if path else None
        self.data = {"results": []}
        if self.path and self.path.exists():
            self.data = json.loads(self.path.read_text(encoding="utf-8"))

    def add(self, level, check, msg, author_only=False):
        """author_only marks advice about the local copy that the PR already
        fixes; it is printed for the author but left out of the PR body."""
        self.data["results"].append({"level": level, "check": check, "msg": msg, "author_only": author_only})
        code, label = _LEVELS[level]
        print(f"  {_paint(code, '[' + label + ']')} {check}: {msg}")

    def count(self, level, since=0):
        return sum(1 for r in self.data["results"][since:] if r["level"] == level)

    def failed_since(self, mark):
        return self.count("fail", mark) > 0

    def mark(self):
        return len(self.data["results"])

    def save(self):
        if self.path:
            self.path.write_text(json.dumps(self.data, indent=2) + "\n", encoding="utf-8")


def tilde(path) -> str:
    p = str(path)
    # Resolved paths use the physical home; macOS homes and temp dirs often
    # sit behind a symlink (/var -> /private/var).
    for home in (str(Path.home()), str(Path.home().resolve())):
        if p == home:
            return "~"
        if p.startswith(home + os.sep):
            return "~" + p[len(home):]
    return p


# ---------------------------------------------------------------------------
# Files
# ---------------------------------------------------------------------------


def is_junk_file(name: str) -> bool:
    return name in JUNK_FILES or name.startswith("._") or name.endswith(JUNK_SUFFIXES)


def walk_skill(root: Path):
    """Yield (rel posix path, Path, kind) for every entry promote cares about.
    kind: file | symlink | junk | depdir. Junk and dependency dirs are not
    descended into."""
    for cur, dirs, files in os.walk(root):
        base = Path(cur)
        keep = []
        for d in sorted(dirs):
            p = base / d
            rel = p.relative_to(root).as_posix()
            if p.is_symlink():
                yield rel, p, "symlink"
            elif d in JUNK_DIRS:
                yield rel + "/", p, "junk"
            elif d in DEPENDENCY_DIRS:
                yield rel + "/", p, "depdir"
            else:
                keep.append(d)
        dirs[:] = keep
        for f in sorted(files):
            p = base / f
            rel = p.relative_to(root).as_posix()
            if p.is_symlink():
                yield rel, p, "symlink"
            elif is_junk_file(f):
                yield rel, p, "junk"
            else:
                yield rel, p, "file"


def read_text_file(path: Path):
    try:
        size = path.stat().st_size
    except OSError:
        return None
    if size > SCAN_LIMIT:
        return None
    try:
        data = path.read_bytes()
    except OSError:
        return None
    if b"\0" in data[:8192]:
        return None
    return data.decode("utf-8", errors="replace")


def stage(src: Path, dest: Path):
    """Copy a skill into a registry checkout: junk dropped, symlinks refused,
    modes normalized to what git can record (0644 / 0755)."""
    if dest.exists():
        shutil.rmtree(dest)
    dest.mkdir(parents=True)
    os.chmod(dest, 0o755)
    excluded = []
    for rel, p, kind in walk_skill(src):
        if kind == "symlink":
            raise UsageError(f"{rel} is a symlink; skills must be plain files (install rejects symlinks)")
        if kind in ("junk", "depdir"):
            excluded.append(rel)
            continue
        target = dest / rel
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(p, target)
        mode = 0o755 if os.stat(p).st_mode & stat.S_IXUSR else 0o644
        os.chmod(target, mode)
    for cur, dirs, files in os.walk(dest, topdown=False):
        if cur != str(dest) and not os.listdir(cur):
            os.rmdir(cur)
        else:
            os.chmod(cur, 0o755)
    return excluded


# ---------------------------------------------------------------------------
# Checks
# ---------------------------------------------------------------------------


def _strings(value, path=""):
    if isinstance(value, dict):
        for k, v in value.items():
            yield from _strings(v, f"{path}.{k}" if path else str(k))
    elif isinstance(value, list):
        for v in value:
            yield from _strings(v, path)
    elif value is not None:
        yield path, str(value)


def _first_sentence(description: str) -> str:
    text = " ".join(description.split())
    cuts = [text.find(m) for m in (" Use when", " Use this", " Use it", " Use for", " Do NOT", " Trigger")]
    m = re.search(r"\.\s", text)
    if m:
        cuts.append(m.start() + 1)
    cuts = [c for c in cuts if c > 0]
    return text[: min(cuts)].rstrip() if cuts else text


def semver_key(v: str):
    m = SEMVER_RE.match(v)
    if not m:
        return None
    return (int(m.group(1)), int(m.group(2)), int(m.group(3)), 0 if m.group(4) else 1, m.group(4) or "")


def parse_dep(raw: str):
    if raw == "":
        raise ValueError("empty dep")
    if "@" not in raw:
        return raw, "", ""
    name, spec = raw.split("@", 1)
    if not name:
        raise ValueError("dep name is empty")
    if not spec:
        raise ValueError("dep version spec is empty after '@'")
    op, ver = (">=", spec[2:].strip()) if spec.startswith(">=") else ("==", spec)
    if not SEMVER_RE.match(ver):
        raise ValueError(f"version {ver!r} is not valid semver (want MAJOR.MINOR.PATCH)")
    return name, op, ver


def preserve_errors(entries):
    errs, seen = [], []
    for raw in entries:
        t = str(raw).strip()
        if not t:
            errs.append(f"preserve entry {raw!r} is empty")
            continue
        if t.startswith("/") or re.match(r"^[A-Za-z]:", t):
            errs.append(f"preserve entry {raw!r} must be a relative path")
            continue
        if ".." in t.split("/"):
            errs.append(f"preserve entry {raw!r} must not contain '..'")
            continue
        trailing = t.endswith("/")
        clean = os.path.normpath(t[2:] if t.startswith("./") else t).replace(os.sep, "/")
        if clean == ".":
            errs.append(f"preserve entry {raw!r} normalizes to an empty path")
            continue
        clean += "/" if trailing else ""
        if clean in seen:
            errs.append(f"preserve entry {raw!r} is a duplicate")
            continue
        for other in seen:
            a, b = other.rstrip("/"), clean.rstrip("/")
            if a == b or a.startswith(b + "/") or b.startswith(a + "/"):
                errs.append(f"preserve entries {other!r} and {clean!r} overlap")
        seen.append(clean)
    return errs, seen


def gather_stats(skill_dir: Path, text: str):
    def count_files(rel, pattern="*"):
        d = skill_dir / rel
        if not d.is_dir():
            return 0
        return sum(1 for p in d.rglob(pattern) if p.is_file() and p.name != ".gitkeep" and not is_junk_file(p.name))

    def read(rel):
        p = skill_dir / rel
        return p.read_text(encoding="utf-8", errors="replace") if p.is_file() else ""

    return {
        "skill_md_lines": len(text.splitlines()),
        "wiki_concepts": count_files("references/wiki", "*.md"),
        "raw_files": count_files("references/raw"),
        "scripts": count_files("scripts"),
        "log_entries": len(LOG_ENTRY_RE.findall(read("references/log.md"))),
        "decisions": len(ENTRY_HEADING_RE.findall(read("references/decisions.md"))),
        "patterns": len(ENTRY_HEADING_RE.findall(read("references/patterns.md"))),
        "has_evals": (skill_dir / "evals" / "scenarios.json").is_file(),
    }


def check_offline(skill_dir: Path, target: str, rep: Report, allow_internal=False, role_decided=False):
    tgt = TARGETS[target]
    mark = rep.mark()
    skill_md = skill_dir / "SKILL.md"
    if not skill_md.is_file():
        rep.add("fail", "skill", f"no SKILL.md in {tilde(skill_dir)}")
        return None
    try:
        fm, body, text = read_skill_md(skill_md)
    except UnicodeDecodeError:
        rep.add("fail", "frontmatter", "SKILL.md is not valid UTF-8")
        return None
    except FrontmatterError as e:
        rep.add("fail", "frontmatter", str(e))
        return None
    if not isinstance(fm, dict):
        rep.add("fail", "frontmatter", "frontmatter must be a mapping")
        return None
    meta = fm.get("metadata")
    if meta is None:
        meta = {}
    if not isinstance(meta, dict):
        rep.add("fail", "frontmatter", "metadata: must be a mapping")
        return None

    name = as_str(fm.get("name")).strip()
    description = as_str(fm.get("description"))
    category = as_str(meta.get("category")).strip()
    role = meta.get("role")
    tags = as_list(meta.get("tags"))
    platforms = as_list(meta.get("platforms"))
    skill = {
        "name": name,
        "dir": tilde(skill_dir),
        "description": description.strip(),
        "summary": _first_sentence(description),
        "version": as_str(meta.get("version")).strip(),
        "author": as_str(meta.get("author")).strip(),
        "license": as_str(fm.get("license")).strip(),
        "category": category,
        "role": as_str(role).strip(),
        "tags": tags,
        "platforms": platforms,
        "requires": as_list(meta.get("requires")),
        "previous_names": as_list(meta.get("previous_names")),
        "upstream": bool(fm.get("upstream")),
    }
    rep.data["skill"] = skill
    rep.data["target"] = {"key": target, **{k: tgt[k] for k in ("label", "repo", "visibility")}}

    section("Frontmatter")
    fm_mark = rep.mark()
    # The CLI decodes these into typed fields; a wrong shape fails the whole
    # registry build, not just this skill.
    for key in ("name", "description", "license", "compatibility", "allowed-tools"):
        if isinstance(fm.get(key), (dict, list)):
            rep.add("fail", "frontmatter", f"{key} must be a single value")
    for key in ("author", "version", "category", "role"):
        if isinstance(meta.get(key), (dict, list)):
            rep.add("fail", "frontmatter", f"metadata.{key} must be a single value")
    for key in ("tags", "platforms", "requires", "preserve", "previous_names"):
        if meta.get(key) is not None and not isinstance(meta.get(key), list):
            rep.add("fail", "frontmatter", f"metadata.{key} must be a list, e.g. [a, b]")
    # name
    if not name:
        rep.add("fail", "name", "name is required")
    else:
        if name != skill_dir.name:
            rep.add("fail", "name", f"name {name!r} must match its directory {skill_dir.name!r}")
        if len(name) > 64:
            rep.add("fail", "name", f"name is {len(name)} characters; the limit is 64")
        if not NAME_RE.match(name):
            rep.add("fail", "name", f"{name!r} must be kebab-case: lowercase letters, digits, single hyphens, starting with a letter")
        lowered = name.lower()
        if lowered.startswith(("claude", "anthropic")):
            rep.add("fail", "name", "names starting with 'claude' or 'anthropic' are reserved and rejected on upload")
        elif "claude" in lowered or "anthropic" in lowered:
            rep.add("warn", "name", "'claude'/'anthropic' in a name can be rejected by claude.ai uploads")
    # description
    if not description.strip():
        rep.add("fail", "description", "description is required")
    elif is_placeholder(description):
        rep.add("fail", "description", "description is still the scaffold TODO")
    else:
        if len(description.strip()) > 1024:
            rep.add("fail", "description", f"description is {len(description.strip())} characters; the limit is 1024")
        if not TRIGGER_RE.search(description):
            rep.add("fail", "description", 'no trigger clause - add "Use when ..." with the phrases users actually type')
        if not NEGATIVE_TRIGGER_RE.search(description):
            rep.add("warn", "description", 'no negative trigger - add "Do NOT use for ..." so it stops firing on neighbours')
    # other top-level fields
    if not skill["license"]:
        rep.add("fail", "license", "license is required (first-party skills use MIT)")
    compat = fm.get("compatibility")
    if compat is not None:
        c = as_str(compat).strip()
        if is_placeholder(c):
            rep.add("fail", "compatibility", "still the scaffold TODO - fill it in or delete the line")
        elif not c:
            rep.add("fail", "compatibility", "compatibility is empty - fill it in or delete the line")
        elif len(c) > 500:
            rep.add("fail", "compatibility", f"compatibility is {len(c)} characters; the limit is 500")
    if is_placeholder(fm.get("allowed-tools")):
        rep.add("fail", "allowed-tools", "still the scaffold TODO - fill it in or delete the line")
    for key in LEGACY_TOP_LEVEL:
        if key in fm:
            rep.add("fail", "frontmatter", f"top-level {key}: belongs under metadata:")
    unknown = [k for k in fm if k not in TOP_LEVEL_KEYS and k not in LEGACY_TOP_LEVEL]
    if unknown:
        rep.add("warn", "frontmatter", f"non-standard top-level key(s) {', '.join(unknown)} - agentskills.io only defines name, description, license, compatibility, allowed-tools, metadata")
    bracketed = sorted({path for path, s in _strings(fm) if ("<" in s or ">" in s) and not is_placeholder(s)})
    if bracketed:
        rep.add("fail", "frontmatter", f"'<' or '>' in {', '.join(bracketed)} - frontmatter loads into the system prompt and angle brackets are rejected")
    # metadata core
    if not skill["author"]:
        rep.add("fail", "author", "metadata.author is required - reviewers need to know who maintains it")
    elif is_placeholder(skill["author"]):
        rep.add("fail", "author", "metadata.author is still the scaffold TODO")
    if not skill["version"]:
        rep.add("fail", "version", "metadata.version is required (MAJOR.MINOR.PATCH, e.g. 0.1.0)")
    elif not SEMVER_RE.match(skill["version"]):
        rep.add("fail", "version", f"{skill['version']!r} is not MAJOR.MINOR.PATCH semver")
    for p in platforms:
        if p not in PLATFORMS:
            rep.add("fail", "platforms", f"unknown platform {p!r} (known: {', '.join(PLATFORMS)})")
    perrs, preserved = preserve_errors(as_list(meta.get("preserve")))
    for e in perrs:
        rep.add("fail", "preserve", e)
    for raw in skill["requires"]:
        try:
            dep, _, _ = parse_dep(raw)
            if dep == name:
                rep.add("fail", "requires", f"a skill cannot require itself ({raw!r})")
        except ValueError as e:
            rep.add("fail", "requires", f"invalid dep {raw!r}: {e}")
    if fm.get("upstream"):
        up = fm["upstream"] if isinstance(fm["upstream"], dict) else {}
        pres = as_str(up.get("preserved")).strip()
        if not pres or not (skill_dir / pres).exists():
            rep.add("fail", "upstream", "upstream.preserved must point at the verbatim upstream copy inside the skill")
        else:
            rep.add("info", "upstream", f"mirrored skill (upstream {as_str(up.get('source')) or as_str(up.get('name'))}) - the reviewer checks provenance and license")
    if not rep.failed_since(fm_mark) and rep.count("warn", fm_mark) == 0:
        rep.add("pass", "frontmatter", "name, description, license, author, version, platforms, preserve")

    section("Taxonomy")
    tx_mark = rep.mark()
    if not category:
        rep.add("fail", "category", f"metadata.category is required - pass --category {'|'.join(CATEGORIES)}")
    elif is_placeholder(category):
        rep.add("fail", "category", f"still the scaffold TODO - pass --category {'|'.join(CATEGORIES)}")
    elif category not in CATEGORIES:
        rep.add("fail", "category", f"{category!r} is not one of {', '.join(CATEGORIES)}")
    role_s = as_str(role).strip()
    if role is None or role_s == "":
        if tgt["role_required"]:
            rep.add("fail", "role", f"{tgt['label']} organizes skills by role - pass --role {'|'.join(ROLES)}")
        elif role_decided:
            rep.add("info", "role", "unscoped by choice - listed for every role")
        else:
            rep.add("warn", "role", f"unscoped - pass --role {'|'.join(ROLES)} if it serves one role, or --role none to confirm it is general-purpose")
    elif is_placeholder(role_s):
        rep.add("fail", "role", f"still the scaffold TODO - pass --role {'|'.join(ROLES)}" + ("" if tgt["role_required"] else " or none"))
    elif role_s not in ROLES:
        rep.add("fail", "role", f"{role_s!r} is not one of {', '.join(ROLES)}")
    if not tags:
        rep.add("fail", "tags", "metadata.tags is required - pass --tags a,b,c (search matches tags)")
    elif is_placeholder(tags):
        rep.add("fail", "tags", "still the scaffold TODO - pass --tags a,b,c")
    else:
        bad = [t for t in tags if not TAG_RE.match(t)]
        if bad:
            rep.add("fail", "tags", f"tags must be kebab-case: {', '.join(repr(t) for t in bad)}")
        elif len(tags) < 3:
            rep.add("warn", "tags", f"only {len(tags)} tag(s) - 3+ keywords make it findable in `humblskills search`")
    if not rep.failed_since(tx_mark):
        where = " › ".join(x for x in (tgt["label"], category, role_s or None, name) if x)
        rep.add("pass", "taxonomy", f"{where}  (tags: {', '.join(tags)})")
        rep.data["placement"] = where

    section("Content")
    ct_mark = rep.mark()
    n_lines = len(text.splitlines())
    if n_lines > 500:
        rep.add("fail", "skill-md", f"SKILL.md is {n_lines} lines; move detail into references/ (limit 500, aim for under 200)")
    elif n_lines > 200:
        rep.add("warn", "skill-md", f"SKILL.md is {n_lines} lines; humblSKILLS keeps the router under ~200 and moves detail into references/")
    body_scan = re.sub(r"```.*?```", "", body, flags=re.DOTALL)
    body_scan = re.sub(r"`[^`\n]*`", "", body_scan)
    leftover = [ln.strip() for ln in body_scan.splitlines() if PLACEHOLDER_BODY_RE.search(ln)]
    if leftover:
        rep.add("fail", "placeholders", f"{len(leftover)} scaffold TODO line(s) left in SKILL.md, e.g. {leftover[0][:70]!r}")
    if not re.search(r"^##\s+Brain Protocol", body, re.MULTILINE):
        rep.add("fail", "skill-md", "no '## Brain Protocol' section - smart skills read their brain before acting")
    if not re.search(r"^##\s+Examples?\b", body, re.MULTILINE | re.IGNORECASE):
        rep.add("warn", "skill-md", "no '## Examples' section - two User says / Actions / Result examples sharpen triggering")
    if target == "humblskills" and name and not skill["upstream"]:
        if name.startswith("use-"):
            rep.add("warn", "naming", "drop the use- prefix: every skill is used, so it distinguishes nothing")
        elif not NAMING_RE.search(name):
            rep.add("warn", "naming", "first-party skills are named [<verb>-]smart-<noun> (smart-commit, create-smart-html)")
    if not rep.failed_since(ct_mark) and rep.count("warn", ct_mark) == 0:
        rep.add("pass", "content", f"{n_lines} lines, Brain Protocol and Examples present, no scaffold TODOs")

    section("Structure")
    st_mark = rep.mark()
    missing = [rel for rel in SMART_FILES if not (skill_dir / rel).is_file()]
    missing += [rel + "/" for rel in SMART_DIRS if not (skill_dir / rel).is_dir()]
    if missing:
        hint = ""
        if "scripts/lint.sh" in missing:
            hint = " (copy the linter: cp <smart-skill>/scripts/lint.sh <skill>/scripts/)"
        rep.add("fail", "structure", f"not a smart skill yet - missing {', '.join(missing)}{hint}; see smart-skill's create/migrate workflows")
    index = skill_dir / "references" / "_index.md"
    if index.is_file():
        itext = index.read_text(encoding="utf-8", errors="replace")
        if "<!-- GENERATED:START -->" not in itext or "<!-- GENERATED:END -->" not in itext:
            rep.add("fail", "structure", "references/_index.md lost its GENERATED markers - restore them so lint.sh can regenerate it")
    missing_preserve = [p for p in BRAIN_PRESERVE if p not in preserved]
    if missing_preserve:
        rep.add("fail", "preserve", f"metadata.preserve must list {', '.join(missing_preserve)} - otherwise `humblskills update` overwrites users' brain data")
    stats = gather_stats(skill_dir, text)
    rep.data["stats"] = stats
    if stats["wiki_concepts"] == 0 and (skill_dir / "references" / "wiki").is_dir():
        rep.add("warn", "structure", "references/wiki/ has no concepts yet - a smart skill with an empty wiki is a flat skill with extra files")
    if not rep.failed_since(st_mark) and rep.count("warn", st_mark) == 0:
        rep.add("pass", "structure", f"smart-skill layout: {stats['wiki_concepts']} wiki concept(s), {stats['raw_files']} raw source(s), brain paths preserved")

    section("Files")
    file_checks(skill_dir, target, rep, allow_internal)

    if not stats["has_evals"]:
        rep.add("info", "evals", f"no evals/scenarios.json - optional: `humblskills eval init {name or '<skill>'}` scaffolds one")
    rep.add(
        "info",
        "brain",
        f"ships to every user: log.md ({stats['log_entries']} entries), decisions.md ({stats['decisions']}), "
        f"patterns.md ({stats['patterns']}), raw/ ({stats['raw_files']} files) - scrub customer data and personal notes",
    )

    section("Lint")
    run_lint(skill_dir, in_place=False, rep=rep)
    return {"fm": fm, "meta": meta, "skill": skill, "failed": rep.failed_since(mark)}


def file_checks(skill_dir: Path, target: str, rep: Report, allow_internal: bool):
    mark = rep.mark()
    junk, total, n_files = [], 0, 0
    secrets, jwts, homes, internal, emails = [], [], [], [], []
    for rel, p, kind in walk_skill(skill_dir):
        if kind == "symlink":
            rep.add("fail", "files", f"{rel} is a symlink - registry hashing and install reject symlinks; copy the file in")
            continue
        if kind == "junk":
            junk.append(rel)
            continue
        if kind == "depdir":
            rep.add("fail", "files", f"{rel} is an installed dependency tree - ship a manifest and install at runtime instead")
            continue
        name = p.name
        if (name == ".env" or name.startswith(".env.")) and not name.endswith(SAFE_ENV_SUFFIXES):
            rep.add("fail", "secrets", f"{rel} looks like an environment file - never commit it")
        elif name in SENSITIVE_NAMES or name.endswith(SENSITIVE_SUFFIXES):
            rep.add("fail", "secrets", f"{rel} looks like a credential file - never commit it")
        size = p.stat().st_size
        total += size
        n_files += 1
        if size > FILE_FAIL:
            rep.add("fail", "size", f"{rel} is {size / MiB:.1f} MiB (limit {FILE_FAIL // MiB} MiB) - every install downloads the whole repo")
        elif size > FILE_WARN:
            rep.add("warn", "size", f"{rel} is {size / MiB:.1f} MiB - every install downloads the whole repo")
        text = read_text_file(p)
        if text is None:
            continue
        lines = text.splitlines()
        for idx, line in enumerate(lines):
            lineno = idx + 1
            for kind_name, pat in SECRET_PATTERNS:
                for m in pat.finditer(line):
                    if any(h in m.group(0).lower() for h in PLACEHOLDER_HINTS):
                        continue
                    if kind_name == "private key" and not (idx + 1 < len(lines) and PEM_BODY.match(lines[idx + 1])):
                        continue
                    secrets.append(f"{rel}:{lineno} ({kind_name})")
            if JWT_PATTERN.search(line):
                jwts.append(f"{rel}:{lineno}")
            for m in HOME_PATH.finditer(line):
                if m.group(1).lower() not in HOME_PATH_OK and not m.group(1).startswith(("$", "<")):
                    homes.append(f"{rel}:{lineno}")
            for m in EMAIL.finditer(line):
                domain = m.group(1).lower()
                if not domain.endswith(EMAIL_OK_DOMAINS) and "noreply" not in m.group(0).lower():
                    emails.append(f"{rel}:{lineno}")
            if target == "humblskills" and INTERNAL_MENTION.search(line) and not re.match(r"\s*author\s*:", line):
                internal.append(f"{rel}:{lineno}")
    for loc in secrets:
        rep.add("fail", "secrets", f"possible credential at {loc} - remove it and rotate it if it was ever real")
    if jwts:
        rep.add("warn", "secrets", f"JWT-shaped string at {', '.join(jwts[:3])} - make sure it is not a live token")
    if homes:
        rep.add("warn", "files", f"absolute home-directory path at {', '.join(homes[:3])} - not portable, and it exposes a username")
    if emails:
        rep.add("warn", "privacy", f"email address at {', '.join(emails[:3])}" + (f" (+{len(emails) - 3} more)" if len(emails) > 3 else "") + " - make sure none belongs to a customer or colleague")
    if internal:
        msg = f"mentions HappyRobot at {', '.join(internal[:3])}" + (f" (+{len(internal) - 3} more)" if len(internal) > 3 else "")
        if allow_internal:
            rep.add("warn", "public", msg + " - allowed by --allow-internal")
        else:
            rep.add("fail", "public", msg + " - HappyRobot-specific skills go to happySKILLS (--to happyskills); if the mention is public-safe, pass --allow-internal")
    if total > TOTAL_FAIL:
        rep.add("fail", "size", f"skill totals {total / MiB:.1f} MiB (limit {TOTAL_FAIL // MiB} MiB)")
    elif total > TOTAL_WARN:
        rep.add("warn", "size", f"skill totals {total / MiB:.1f} MiB")
    if junk:
        rep.add("info", "files", f"left out of the PR: {', '.join(junk[:5])}" + (" ..." if len(junk) > 5 else ""))
    if not rep.failed_since(mark) and rep.count("warn", mark) == 0:
        rep.add("pass", "files", f"{n_files} file(s), {total / 1024:.0f} KiB; no symlinks, credentials, or oversized files")


def run_lint(skill_dir: Path, in_place: bool, rep: Report, timeout=180):
    lint = skill_dir / "scripts" / "lint.sh"
    if not lint.is_file():
        rep.add("fail", "lint", "scripts/lint.sh is missing (see Structure)")
        return
    if shutil.which("bash") is None:
        rep.add("fail", "lint", "bash is required to run scripts/lint.sh")
        return
    work, tmp = skill_dir, None
    if not in_place:
        tmp = tempfile.mkdtemp(prefix="promote-lint-")
        work = Path(tmp) / skill_dir.name
        shutil.copytree(skill_dir, work, symlinks=True, ignore=shutil.ignore_patterns(*JUNK_DIRS, *DEPENDENCY_DIRS))
    index = work / "references" / "_index.md"
    log = work / "references" / "log.md"
    before = index.read_text(encoding="utf-8", errors="replace") if index.is_file() else None
    # lint.sh appends a dated [LINT] entry to log.md. The promoted copy keeps the
    # author's log byte-for-byte, so a re-run on another day is not a change.
    log_bytes = log.read_bytes() if log.is_file() else None
    try:
        proc = subprocess.run(
            ["bash", str(work / "scripts" / "lint.sh"), str(work)],
            capture_output=True, text=True, timeout=timeout, env={**os.environ, "NO_COLOR": "1"},
        )
    except subprocess.TimeoutExpired:
        rep.add("fail", "lint", f"scripts/lint.sh did not finish within {timeout}s")
        return
    finally:
        after = index.read_text(encoding="utf-8", errors="replace") if index.is_file() else None
        if log_bytes is not None:
            log.write_bytes(log_bytes)
        if tmp:
            shutil.rmtree(tmp, ignore_errors=True)
    out = ANSI_RE.sub("", proc.stdout + proc.stderr)
    if proc.returncode != 0:
        fails = [ln.strip().replace("[FAIL] ", "") for ln in out.splitlines() if "[FAIL]" in ln]
        detail = "; ".join(fails[:4]) or (out.strip().splitlines() or ["no output"])[-1]
        rep.add("fail", "lint", f"scripts/lint.sh exited {proc.returncode}: {detail}")
        return
    soft = re.search(r"soft findings:\s*(\d+)", out)
    rep.add("pass", "lint", "scripts/lint.sh exits 0" + (f" ({soft.group(1)} soft finding(s))" if soft else ""))
    if in_place:
        if before != after:
            rep.add("info", "lint", "regenerated references/_index.md in the promoted copy", author_only=True)
    elif before is not None and before != after:
        rep.add(
            "warn", "lint",
            "references/_index.md is stale - run `bash scripts/lint.sh` in the skill (the promoted copy is regenerated either way)",
            author_only=True,
        )


def published_skills(skills_root: Path):
    out = {}
    for d in sorted(skills_root.iterdir()):
        sm = d / "SKILL.md"
        if not d.is_dir() or d.name.startswith(".") or not sm.is_file():
            continue
        try:
            fm, _, _ = read_skill_md(sm)
        except (FrontmatterError, UnicodeDecodeError):
            out[d.name] = {"version": "", "previous_names": [], "dir": d.name}
            continue
        meta = fm.get("metadata") if isinstance(fm.get("metadata"), dict) else {}
        out[as_str(fm.get("name")) or d.name] = {
            "version": as_str(meta.get("version")),
            "previous_names": as_list(meta.get("previous_names")),
            "dir": d.name,
        }
    return out


def repo_platforms(repo_dir: Path):
    adapters = repo_dir / "cli" / "internal" / "adapters"
    if not adapters.is_dir():
        return None
    names = []
    for y in sorted(adapters.glob("*.yaml")):
        m = re.search(r"^name:\s*['\"]?([A-Za-z0-9_-]+)", y.read_text(encoding="utf-8"), re.MULTILINE)
        if m:
            names.append(m.group(1))
    return names or None


def check_repo(skill_dir: Path, target: str, repo_dir: Path, base: str, rep: Report):
    tgt = TARGETS[target]
    mark = rep.mark()
    fm, _, _ = read_skill_md(skill_dir / "SKILL.md")
    meta = fm.get("metadata") if isinstance(fm.get("metadata"), dict) else {}
    name = as_str(fm.get("name"))
    where = f"{tgt['label']}@{base}" if base else tgt["label"]
    skills_root = repo_dir / "skills"
    if not skills_root.is_dir():
        rep.add("fail", "registry", f"{tilde(repo_dir)} has no skills/ directory - not a skill registry checkout")
        return False
    published = published_skills(skills_root)
    if (skills_root / name).exists() or name in published:
        rep.add(
            "fail", "net-new",
            f"{name} already exists in {where}. promote.sh only adds net-new skills; merging changes or "
            "learnings into an existing skill is a separate path that is not supported yet",
        )
    for other, info in published.items():
        if name in info["previous_names"]:
            rep.add("fail", "net-new", f"{name} is a former name of {other} (previous_names) - `humblskills update` would read it as that rename; pick another name")
    own_prev = as_list(meta.get("previous_names"))
    for p in own_prev:
        if p in published:
            rep.add("fail", "previous_names", f"lists {p}, which {where} still publishes - a net-new skill cannot claim it")
    if own_prev and not rep.failed_since(mark):
        rep.add("warn", "previous_names", f"{', '.join(own_prev)} were never published in {where}; drop metadata.previous_names unless you mean it")
    for raw in as_list(meta.get("requires")):
        try:
            dep, op, ver = parse_dep(raw)
        except ValueError:
            continue
        if dep not in published:
            rep.add("fail", "requires", f"{dep} is not published in {where}")
        elif op:
            have = semver_key(published[dep]["version"])
            want = semver_key(ver)
            ok = have is not None and want is not None and (have == want if op == "==" else have >= want)
            if not ok:
                rep.add("fail", "requires", f"{raw} is not satisfied by {dep} {published[dep]['version']}")
    known = repo_platforms(repo_dir)
    if known:
        for p in as_list(meta.get("platforms")):
            if p not in known:
                rep.add("fail", "platforms", f"{p!r} has no adapter in {where} (known: {', '.join(known)})")
    if not rep.failed_since(mark):
        rep.add("pass", "net-new", f"{name} is new to {where} ({len(published)} skills published); requires and platforms resolve")
    return not rep.failed_since(mark)


# ---------------------------------------------------------------------------
# Rendering
# ---------------------------------------------------------------------------


def _wrap(text, width=72):
    if not text.strip():
        return ""
    return "\n".join(textwrap.wrap(" ".join(text.split()), width=width, break_on_hyphens=False, break_long_words=False))


def commit_title(name: str, mode: str) -> str:
    if mode == "update":
        return f"chore(skills): sync {name} from its local copy"
    title = f"feat(skills): add {name} skill"
    return title if len(title) <= 72 else f"feat(skills): add {name}"


def render_commit(data, mode):
    s, t = data["skill"], data["target"]
    title = commit_title(s["name"], mode)
    if mode == "update":
        body = "Re-promoted with smart-skill/scripts/promote.sh to carry review changes from the author's local copy."
        return f"{title}\n\n{_wrap(body)}\n"
    taxonomy = f"Category: {s['category']} · Role: {s['role'] or 'unscoped'} · Tags: {', '.join(s['tags'])}"
    tail = f"Promoted to {t['label']} with smart-skill/scripts/promote.sh."
    return f"{title}\n\n{_wrap(s['summary'])}\n\n{_wrap(taxonomy)}\n{_wrap(tail)}\n"


def render_pr(data, base, mode):
    s, t, st = data["skill"], data["target"], data.get("stats", {})
    results = [r for r in data.get("results", []) if not r.get("author_only")]
    n_pass = sum(1 for r in results if r["level"] == "pass")
    warns = [r for r in results if r["level"] == "warn"]
    infos = [r for r in results if r["level"] == "info"]
    role = f"`{s['role']}` ({ROLE_HINTS.get(s['role'], '')})" if s["role"] else "unscoped - listed for every role"
    platforms = ", ".join(f"`{p}`" for p in s["platforms"]) or "every detected platform"
    search = f"humblskills search --category={s['category']}" + (f" --role={s['role']}" if s["role"] else "")
    lines = [
        f"Adds **`{s['name']}`** to {t['label']}: {s['summary']}",
        "",
        "> Opened by `smart-skill/scripts/promote.sh` - the net-new skill promotion path. "
        "Re-running the same command from the author's machine updates this PR.",
        "",
        "## Where it lands in the CLI",
        "",
        "| | |",
        "|---|---|",
        f"| Registry | {t['label']} (`{t['repo']}`), base `{base}` |",
        f"| Category | `{s['category']}` ({CATEGORY_HINTS.get(s['category'], '')}) |",
        f"| Role | {role} |",
        f"| Tags | {', '.join(f'`{x}`' for x in s['tags'])} |",
        f"| Version | `{s['version']}` |",
        f"| Author | {s['author']} |",
        f"| Platforms | {platforms} |",
        "",
        f"Browser placement: `{data.get('placement', '')}` · find it with `{search}`",
        "",
        "## Description",
        "",
        "\n".join("> " + ln if ln else ">" for ln in s["description"].splitlines()),
        "",
        "## Contents",
        "",
        f"- `SKILL.md`: {st.get('skill_md_lines', '?')} lines · wiki concepts: {st.get('wiki_concepts', 0)} · "
        f"raw sources: {st.get('raw_files', 0)} · scripts: {st.get('scripts', 0)} · evals: {'yes' if st.get('has_evals') else 'no'}",
        f"- Brain shipped to every user: `log.md` ({st.get('log_entries', 0)} entries), `decisions.md` "
        f"({st.get('decisions', 0)}), `patterns.md` ({st.get('patterns', 0)}), `raw/` ({st.get('raw_files', 0)} files)",
        "",
        "## Quality gate",
        "",
        f"{n_pass} checks passed, {len(warns)} warning(s), 0 failures (`promote.sh` refuses to open a PR with failures).",
    ]
    if warns:
        lines += ["", "Warnings:", ""] + [f"- **{r['check']}**: {r['msg']}" for r in warns]
    if infos:
        lines += ["", "Notes:", ""] + [f"- {r['check']}: {r['msg']}" for r in infos]
    lines += [
        "",
        "## Reviewer checklist",
        "",
        "- [ ] The description triggers on what users actually type, and says when *not* to use it",
        "- [ ] Category, role, and tags put it where people will look",
        "- [ ] Brain files and `raw/` hold no customer data, credentials, or personal notes",
        "- [ ] Not a near-duplicate of an existing skill",
    ]
    if t["key"] == "humblskills":
        lines += [
            "- [ ] Nothing HappyRobot-internal (internal skills belong in happySKILLS)",
            "",
            "`registry.json` is not part of this PR: the Registry workflow regenerates it, and repins "
            "`source.sha`, once this merges into `develop`. Merge with a merge commit (`gh pr merge --merge`), never squash.",
        ]
    else:
        lines += [
            "",
            "`registry.json` is not part of this PR - regenerate it with this repo's registry build after merge.",
        ]
    if mode == "update":
        lines += ["", "_Updated by a re-run of `promote.sh`._"]
    return "\n".join(lines) + "\n"


# ---------------------------------------------------------------------------
# Lookup
# ---------------------------------------------------------------------------


def candidate_dirs(name: str):
    """Where humblskills and the agents keep skills, canonical store first."""
    home, cwd = Path.home(), Path.cwd()
    if os.environ.get("XDG_DATA_HOME"):
        data_home = Path(os.environ["XDG_DATA_HOME"])
    elif sys.platform == "darwin":
        data_home = home / "Library" / "Application Support"
    else:
        data_home = home / ".local" / "share"
    out = [
        home / ".humblskills" / "skills" / name,
        cwd / ".humblskills" / "skills" / name,
        data_home / "humblskills" / "skills" / name,
    ]
    out += [
        home / ".claude" / "skills" / name,
        home / ".cursor" / "skills" / name,
        home / ".agents" / "skills" / name,
        home / ".pi" / "agent" / "skills" / name,
        cwd / ".claude" / "skills" / name,
        cwd / ".cursor" / "skills" / name,
        cwd / ".agents" / "skills" / name,
        cwd / ".pi" / "skills" / name,
    ]
    return out


def resolve_skill(arg: str) -> Path:
    looks_like_path = "/" in arg or arg.startswith((".", "~")) or Path(arg).is_dir()
    if looks_like_path:
        p = Path(os.path.expanduser(arg)).resolve()
        if not (p / "SKILL.md").is_file():
            raise UsageError(f"{arg} is not a skill directory (no SKILL.md)")
        return p
    if not NAME_RE.match(arg):
        raise UsageError(f"{arg!r} is neither a skill name nor a path")
    found = []
    for c in candidate_dirs(arg):
        if (c / "SKILL.md").is_file():
            r = c.resolve()
            if r not in found:
                found.append(r)
    if not found:
        searched = "\n  ".join(tilde(c) for c in candidate_dirs(arg))
        raise UsageError(f"no skill named {arg} - looked in:\n  {searched}\npass the skill's directory instead")
    if len(found) > 1:
        listed = "\n  ".join(tilde(f) for f in found)
        raise UsageError(f"{arg} exists in more than one place:\n  {listed}\npass the directory you mean")
    return found[0]


def slug_from_registry_url(url: str):
    m = re.match(r"^https://raw\.githubusercontent\.com/([^/]+)/([^/]+)/", url or "")
    return f"{m.group(1)}/{m.group(2)}" if m else None


def discover_repo(target: str, profile_path: str):
    """The happySKILLS repo the user already reads from, when their humblskills
    profile has it as a named registry."""
    if target != "happyskills":
        return None
    try:
        data = json.loads(Path(profile_path).read_text(encoding="utf-8"))
    except (OSError, ValueError):
        return None
    entries = list(data.get("registries") or [])
    if data.get("registry"):
        entries.append({"name": "", "url": data["registry"]})
    for entry in entries:
        slug = slug_from_registry_url(entry.get("url", ""))
        if slug and slug.split("/")[1].lower() == "happyskills":
            return slug
    return None


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------


def resolve_target(raw: str) -> str:
    key = TARGET_ALIASES.get((raw or "").strip().lower())
    if not key:
        raise UsageError(f"unknown target {raw!r} - use humblskills (public) or happyskills (HappyRobot-internal)")
    return key


def cmd_target_info(a):
    key = resolve_target(a.to)
    tgt = TARGETS[key]
    profile = a.profile or os.environ.get("HUMBLSKILLS_PROFILE") or str(Path.home() / ".humblskills" / "profile.json")
    env_override = os.environ.get("HAPPYSKILLS_REPO") if key == "happyskills" else None
    repo, source = tgt["repo"], "default"
    discovered = discover_repo(key, profile)
    if a.repo:
        repo, source = a.repo, "--repo"
    elif env_override:
        repo, source = env_override, "HAPPYSKILLS_REPO"
    elif discovered:
        repo, source = discovered, f"registry in {tilde(profile)}"
    if not re.match(r"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$", repo):
        raise UsageError(f"repo {repo!r} must look like owner/name")
    for k, v in (("key", key), ("label", tgt["label"]), ("repo", repo), ("repo_source", source),
                 ("visibility", tgt["visibility"]), ("base", tgt["base"])):
        print(f"{k}={v}")


def cmd_resolve(a):
    print(resolve_skill(a.skill))


def cmd_taxonomy(a):
    for value, allowed, flag in ((a.category, CATEGORIES, "--category"), (a.role, ROLES + ("none",), "--role")):
        if value is not None and value not in allowed:
            raise UsageError(f"{flag} {value!r} must be one of {', '.join(allowed)}")
    if a.tags is not None and not merge_tags([], a.tags):
        raise UsageError("--tags needs at least one tag, e.g. --tags git,review")
    def show(v):
        if v in (None, "", []):
            return "(unset)"
        return ", ".join(v) if isinstance(v, list) else str(v)

    for field, old, new in apply_taxonomy(Path(a.skill_dir) / "SKILL.md", a.category, a.role, a.tags):
        print(f"  set {field}: {show(old)} -> {show(new)}")


def cmd_check(a):
    rep = Report(a.report)
    key = resolve_target(a.to)
    if a.repo_dir:
        section("Registry")
        ok = check_repo(Path(a.skill_dir), key, Path(a.repo_dir), a.base or "", rep)
        rep.save()
        return 0 if ok else 1
    res = check_offline(Path(a.skill_dir).resolve(), key, rep, a.allow_internal, a.role_decided)
    rep.save()
    print()
    section(f"Gate: {summarize(rep.data)}")
    return 1 if res is None or res["failed"] else 0


def cmd_stage(a):
    excluded = stage(Path(a.skill_dir), Path(a.dest))
    for rel in excluded:
        print(rel)


def cmd_lint(a):
    rep = Report(a.report)
    mark = rep.mark()
    run_lint(Path(a.skill_dir), in_place=True, rep=rep)
    rep.save()
    return 1 if rep.failed_since(mark) else 0


def cmd_record(a):
    rep = Report(a.report)
    rep.add(a.level, a.check, a.msg)
    if a.placement:
        rep.data["placement"] = a.placement
    rep.save()


def cmd_placement(a):
    data = json.loads(Path(a.registry).read_text(encoding="utf-8"))
    entry = next((s for s in data.get("skills", []) if s.get("name") == a.name), None)
    if entry is None:
        raise UsageError(f"{a.name} is missing from the generated registry")
    parts = [a.label, entry.get("category", ""), entry.get("role", ""), entry["name"]]
    print(" › ".join(p for p in parts if p))


def summarize(data) -> str:
    results = data.get("results", [])

    def n(level):
        return sum(1 for r in results if r["level"] == level)

    return f"{n('pass')} checks passed, {n('warn')} warning(s), {n('fail')} failure(s)"


def cmd_render(a):
    data = json.loads(Path(a.report).read_text(encoding="utf-8"))
    if a.kind == "commit":
        sys.stdout.write(render_commit(data, a.mode))
    elif a.kind == "title":
        print(commit_title(data["skill"]["name"], a.mode))
    elif a.kind == "summary":
        print(summarize(data))
    else:
        sys.stdout.write(render_pr(data, a.base, a.mode))


def main(argv=None):
    p = argparse.ArgumentParser(prog="promote.py", description="plumbing for scripts/promote.sh")
    sub = p.add_subparsers(dest="cmd", required=True)

    s = sub.add_parser("target-info")
    s.add_argument("--to", required=True)
    s.add_argument("--repo")
    s.add_argument("--profile")
    s.set_defaults(fn=cmd_target_info)

    s = sub.add_parser("resolve")
    s.add_argument("skill")
    s.set_defaults(fn=cmd_resolve)

    s = sub.add_parser("taxonomy")
    s.add_argument("skill_dir")
    s.add_argument("--category")
    s.add_argument("--role")
    s.add_argument("--tags")
    s.set_defaults(fn=cmd_taxonomy)

    s = sub.add_parser("check")
    s.add_argument("skill_dir")
    s.add_argument("--to", required=True)
    s.add_argument("--report")
    s.add_argument("--repo-dir")
    s.add_argument("--base")
    s.add_argument("--allow-internal", action="store_true")
    s.add_argument("--role-decided", action="store_true")
    s.set_defaults(fn=cmd_check)

    s = sub.add_parser("stage")
    s.add_argument("skill_dir")
    s.add_argument("dest")
    s.set_defaults(fn=cmd_stage)

    s = sub.add_parser("lint")
    s.add_argument("skill_dir")
    s.add_argument("--report")
    s.set_defaults(fn=cmd_lint)

    s = sub.add_parser("record")
    s.add_argument("--report", required=True)
    s.add_argument("--level", required=True, choices=sorted(_LEVELS))
    s.add_argument("--check", required=True)
    s.add_argument("--msg", required=True)
    s.add_argument("--placement")
    s.set_defaults(fn=cmd_record)

    s = sub.add_parser("placement")
    s.add_argument("registry")
    s.add_argument("name")
    s.add_argument("--label", required=True)
    s.set_defaults(fn=cmd_placement)

    s = sub.add_parser("render")
    s.add_argument("--report", required=True)
    s.add_argument("--kind", required=True, choices=("commit", "title", "pr", "summary"))
    s.add_argument("--mode", default="new", choices=("new", "update"))
    s.add_argument("--base", default="")
    s.set_defaults(fn=cmd_render)

    a = p.parse_args(argv)
    try:
        return a.fn(a) or 0
    except UsageError as e:
        print(f"promote: {e}", file=sys.stderr)
        return 2
    except FrontmatterError as e:
        print(f"promote: SKILL.md: {e}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    sys.exit(main())
