#!/usr/bin/env bash
# scan-secrets.sh - refuse to ship a handoff doc that leaks a credential.
#
# Usage:
#   bash scripts/scan-secrets.sh <file> [file...]
#   bash scripts/scan-secrets.sh --strict <file>   # WARN findings also fail
#
# Run this on the drafted handoff doc BEFORE telling the user it is ready.
#
# Output is masked: matches are replaced with <REDACTED:name> so the scan
# report itself never becomes a second copy of the secret.
#
# Exit codes:
#   0  no FAIL findings (WARN findings may exist - read them)
#   1  at least one FAIL finding (or any WARN with --strict)
#   2  invocation error

set -uo pipefail

STRICT=false
FILES=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --strict) STRICT=true; shift ;;
    --help|-h) sed -n '2,15p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
    *) FILES+=("$1"); shift ;;
  esac
done

if [[ ${#FILES[@]} -eq 0 ]]; then
  echo "usage: scan-secrets.sh [--strict] <file> [file...]" >&2
  exit 2
fi

# name|severity|extended-regex   (bash 3.2: no associative arrays)
PATTERNS=(
  'anthropic-key|FAIL|sk-ant-[A-Za-z0-9_-]{16,}'
  'openai-key|FAIL|sk-[A-Za-z0-9]{20,}'
  'uploadthing-key|FAIL|sk_live_[A-Za-z0-9]{16,}'
  'github-token|FAIL|gh[pousr]_[A-Za-z0-9]{20,}'
  'github-pat|FAIL|github_pat_[A-Za-z0-9_]{20,}'
  'aws-access-key|FAIL|(AKIA|ASIA)[0-9A-Z]{16}'
  'slack-token|FAIL|xox[baprs]-[A-Za-z0-9-]{10,}'
  'stripe-key|FAIL|[rsp]k_(live|test)_[A-Za-z0-9]{16,}'
  'google-api-key|FAIL|AIza[0-9A-Za-z_-]{35}'
  'jwt|FAIL|eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{6,}'
  'private-key-block|FAIL|-----BEGIN [A-Z ]*PRIVATE KEY-----'
  'bearer-header|FAIL|[Bb]earer [A-Za-z0-9._~+/-]{20,}'
  'basic-auth-url|FAIL|[a-z][a-z0-9+.-]*://[^/:@[:space:]]+:[^/@[:space:]]+@'
  'inline-credential|WARN|(password|passwd|secret|token|api[_-]?key|access[_-]?key|client[_-]?secret)["'"'"']?[[:space:]]*[:=][[:space:]]*["'"'"'][^"'"'"']{8,}'
  'email-address|WARN|[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}'
  'phone-number|WARN|\+?[0-9]{1,2}[ .-]?\(?[0-9]{3}\)?[ .-][0-9]{3}[ .-][0-9]{4}'
)

# Lines matching any of these are placeholders / safe references, not leaks.
SAFE='(op://|REDACTED|<[A-Za-z_-]+>|\$\{?[A-Z_]{3,}|xxx+|\*\*\*+|\.\.\.|EXAMPLE|noreply@|example\.(com|org)|security find-generic-password|op read|op run)'

fail_count=0
warn_count=0

for f in "${FILES[@]}"; do
  if [[ ! -f "$f" ]]; then
    echo "scan-secrets: no such file: $f" >&2
    exit 2
  fi

  for entry in "${PATTERNS[@]}"; do
    name="${entry%%|*}"
    rest="${entry#*|}"
    sev="${rest%%|*}"
    re="${rest#*|}"

    hits="$(grep -nE -e "$re" "$f" 2>/dev/null | grep -vE -e "$SAFE" || true)"
    [[ -z "$hits" ]] && continue

    while IFS= read -r line; do
      [[ -z "$line" ]] && continue
      lineno="${line%%:*}"
      body="${line#*:}"

      # Mask by literal substitution of each match. sed is unusable here: the
      # patterns contain '|', '/' and '#', so every delimiter choice collides
      # with an alternation and produces "parentheses not balanced".
      masked="$body"
      while IFS= read -r m; do
        [[ -z "$m" ]] && continue
        masked="${masked//"$m"/<REDACTED:$name>}"
      done <<< "$(printf '%s' "$body" | grep -oE -e "$re" 2>/dev/null || true)"

      # Belt and braces: if masking did not take, suppress the line rather
      # than let the scan report become a second copy of the secret.
      if printf '%s' "$masked" | grep -qE -e "$re" 2>/dev/null; then
        masked="[line body suppressed - masking failed]"
      fi

      printf '%s:%s  [%s] %s\n    %s\n' "$f" "$lineno" "$sev" "$name" "$masked"
      if [[ "$sev" == FAIL ]]; then
        fail_count=$((fail_count + 1))
      else
        warn_count=$((warn_count + 1))
      fi
    done <<< "$hits"
  done
done

echo
echo "scan-secrets: $fail_count FAIL, $warn_count WARN across ${#FILES[@]} file(s)"

if [[ $fail_count -gt 0 ]]; then
  cat <<'EOF'

FAIL means a live-shaped credential is in the document. Do NOT hand it off.
Replace the value with a retrieval instruction instead:
  op read "op://<vault>/<item>/<field>"
  security find-generic-password -s <service> -a <account> -w
See references/wiki/handoff/security/secret-handling.md.
EOF
  exit 1
fi

if [[ $warn_count -gt 0 && "$STRICT" == true ]]; then
  echo "--strict: treating WARN findings as failures."
  exit 1
fi

exit 0
