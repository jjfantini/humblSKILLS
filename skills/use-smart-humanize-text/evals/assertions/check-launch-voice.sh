#!/usr/bin/env bash
# Internal indie-launch-copy-iteration voice checker.
# Emits JSON {violations, count, rules_checked, rules_violated}.
# Exits 0 iff count == 0; 1 otherwise; 2 on missing file.
#
# Checks TWO axes of learning (mirrors check-brand-voice.sh shape so the
# ceiling/retention assertions in scenarios.json reuse the same awk parser):
#   - BAD patterns to reject (9 cliche substrings, case-insensitive)
#   - GOOD patterns to reinforce (4 required structural moves)
# Used by harness assertions, not staged to the agent, so rule text here
# cannot leak to the model.

set -euo pipefail

file="${1:-}"
if [ -z "$file" ] || [ ! -f "$file" ]; then
  printf '{"error":"file not found: %s","violations":[],"count":0,"rules_checked":13,"rules_violated":[]}\n' "$file"
  exit 2
fi

violations=()
rule_hits_b=(0 0 0 0 0 0 0 0 0 0)   # index 1..9 used
rule_hits_g=(0 0 0 0 0)              # index 1..4 used

# ---------------------------------------------------------------------------
# BAD PATTERNS (reject)
# ---------------------------------------------------------------------------
# Each rule is a grep -iE pattern. First hit per rule is reported as the
# offender so the JSON stays compact even if the draft uses the word twice.

check_ban() {
  local rn="$1" label="$2" pattern="$3"
  if grep -qiE "$pattern" "$file"; then
    local offender
    offender=$(grep -oiE "$pattern" "$file" | head -1)
    violations+=("b${rn} cliche: found '${offender}' (must avoid ${label})")
    rule_hits_b[$rn]=1
  fi
}

check_ban 1 "'powerful'"        '\bpowerful\b'
check_ban 2 "'seamless'"        '\bseamless(ly)?\b'
check_ban 3 "'leverage'"        '\bleverag(e|es|ed|ing)\b'
check_ban 4 "'unleash'"         '\bunleash(es|ed|ing)?\b'
check_ban 5 "'intuitive'"       '\bintuitive(ly)?\b'
check_ban 6 "'effortless'"      '\beffortless(ly)?\b'
check_ban 7 "'revolutionary'"   '\brevolution(ary|ize|izes|ized|izing)\b'
check_ban 8 "'game-changer'"    '\bgame[ -]?chang(er|ers|ing)\b'
check_ban 9 "'cutting-edge'"    '\bcutting[ -]edge\b'

# ---------------------------------------------------------------------------
# GOOD PATTERNS (reinforce) -- absence == violation
# ---------------------------------------------------------------------------
# These detect that the draft included the positive voice move at all. They
# are intentionally generous: any credible token of the shape counts. A draft
# that repeats one move three times still only satisfies that one rule.

# g1: names a specific audience via "for <role-plural>" — optionally with
# one adjective in front (e.g. "for solo game devs", "for freelance
# writers", "for researchers"). Allow-list of plural role nouns keeps
# casual "for example" or "for later" from counting as an audience.
g1_audience_re='(^|[^A-Za-z])[Ff]or ([a-z]+[ -]){0,3}(designers|developers|devs|engineers|writers|podcasters|readers|founders|hackers|makers|streamers|students|teachers|researchers|artists|musicians|gamers|editors|tinkerers|hobbyists|marketers|listeners|organizers|organisers|folks|people|teams|freelancers)\b'
if grep -qE "$g1_audience_re" "$file"; then
  rule_hits_g[1]=0
else
  violations+=("g1 voice: missing named audience (e.g. 'for solo game devs', 'for freelance writers')")
  rule_hits_g[1]=1
fi

# g2: at least one concrete number with a unit. Accepts durations (s/min/hr),
# file sizes (KB/MB/GB), prices ($9, $9/month), percentages, and plain
# count+unit combos like "12 shortcuts" or "4 tabs".
g2_number_re='(\$[0-9]+(\.[0-9]+)?(/[A-Za-z]+)?|\b[0-9]+\s*(second|seconds|sec|minute|minutes|min|hour|hours|hr|day|days|week|weeks|KB|kb|MB|mb|GB|gb|px|%|percent|shortcuts?|tabs?|bookmarks?|clicks?|keystrokes?|sprites?|posts?|episodes?|tracks?|items?|lines?|files?|notes?)\b)'
if grep -qE "$g2_number_re" "$file"; then
  rule_hits_g[2]=0
else
  violations+=("g2 voice: missing concrete number with unit (e.g. '2-minute setup', '\$9/month', '30 shortcuts')")
  rule_hits_g[2]=1
fi

# g3: first-person sentence by the maker. Covers the standard openers Liana
# uses. Apostrophe-agnostic so curly quotes do not defeat it.
g3_first_person_re="(^|[^A-Za-z])I (built|made|wrote|created|shipped|designed|coded|hacked|launched|started|use|keep|prefer|'m|’m|am)\b"
if grep -qE "$g3_first_person_re" "$file"; then
  rule_hits_g[3]=0
else
  violations+=("g3 voice: missing first-person sentence (e.g. 'I built this because...')")
  rule_hits_g[3]=1
fi

# g4: names a limitation / missing feature. Honest-about-limits is a signature
# Liana move. Generous enough to catch any credible acknowledgement of an
# unfinished surface (contractions with or without "does", multi-word "no X
# Y yet", preamble words, "coming soon" / "on the roadmap" phrasings) so the
# detector measures "did the draft admit a limit?" rather than "did the
# draft use the exact keyword 'caveat'?".
g4_limit_re="(does not|doesn['’]t|don['’]t|can not|can['’]t|won['’]t|not yet|not ready|not there|not available|not supported|no [a-z]+(([ -][a-z]+){0,2}) yet|coming soon|on the roadmap|known (limitation|gap)|limitation:|caveat:|trade[ -]?off:|gap:|rough edges?|missing)"
if grep -qiE "$g4_limit_re" "$file"; then
  rule_hits_g[4]=0
else
  violations+=("g4 voice: missing named limitation (e.g. 'no mobile yet', 'doesn''t do X')")
  rule_hits_g[4]=1
fi

# ---------------------------------------------------------------------------
# Emit JSON
# ---------------------------------------------------------------------------
printf '{"violations":['
for i in "${!violations[@]}"; do
  [ "$i" -gt 0 ] && printf ','
  v=${violations[$i]//\\/\\\\}
  v=${v//\"/\\\"}
  printf '"%s"' "$v"
done
printf '],"count":%d,"rules_checked":13,"rules_violated":[' "${#violations[@]}"
first=1
for rn in 1 2 3 4 5 6 7 8 9; do
  if [ "${rule_hits_b[$rn]:-0}" = "1" ]; then
    [ "$first" -eq 0 ] && printf ','
    printf '"b%d"' "$rn"
    first=0
  fi
done
for rn in 1 2 3 4; do
  if [ "${rule_hits_g[$rn]:-0}" = "1" ]; then
    [ "$first" -eq 0 ] && printf ','
    printf '"g%d"' "$rn"
    first=0
  fi
done
printf ']}\n'

[ "${#violations[@]}" -eq 0 ]
