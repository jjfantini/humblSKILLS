#!/usr/bin/env bash
# Internal brand-voice compliance checker. Emits JSON {violations, count, rules_violated}.
# Exits 0 if count == 0, 1 otherwise. Used by harness assertions, not by the agent.

set -euo pipefail

file="${1:-}"
if [ -z "$file" ] || [ ! -f "$file" ]; then
  printf '{"error":"file not found: %s","violations":[],"count":0,"rules_checked":10}\n' "$file"
  exit 2
fi

violations=()
rule_hits=(0 0 0 0 0 0 0 0 0 0 0)

if grep -qE '\b(Arc Factor|ARCFACTOR|Arcfactor|arcfactor)\b' "$file"; then
  offender=$(grep -oE '\b(Arc Factor|ARCFACTOR|Arcfactor|arcfactor)\b' "$file" | head -1)
  violations+=("rule-1 brand: found '${offender}' (must be 'ArcFactor')")
  rule_hits[1]=1
fi

if grep -qE '\b(Pulse Index|pulseindex|PULSEINDEX|Pulseindex)\b' "$file"; then
  offender=$(grep -oE '\b(Pulse Index|pulseindex|PULSEINDEX|Pulseindex)\b' "$file" | head -1)
  violations+=("rule-2 product: found '${offender}' (must be 'PulseIndex')")
  rule_hits[2]=1
fi

if ! grep -qE 'Dr\. Maira Ostrowski' "$file"; then
  violations+=("rule-3 founder: missing 'Dr. Maira Ostrowski' canonical form")
  rule_hits[3]=1
fi
for bad in 'Mira Ostrowski' 'Myra Ostrowski' 'Maria Ostrowski' 'Maira Ostrowsky' 'Maira Ostrovski'; do
  if grep -qE "\b${bad}\b" "$file"; then
    violations+=("rule-3 founder: misspelling '${bad}' (must be 'Maira Ostrowski')")
    rule_hits[3]=1
  fi
done

if grep -qE '(^|[^A-Za-z])\$[0-9]' "$file"; then
  violations+=("rule-4 currency: found '\$'-prefixed amount (must be 'CAD <amount>')")
  rule_hits[4]=1
fi
if grep -qE '\b(USD|C\$)\b' "$file"; then
  offender=$(grep -oE '\b(USD|C\$)\b' "$file" | head -1)
  violations+=("rule-4 currency: found '${offender}' (must be 'CAD <amount>')")
  rule_hits[4]=1
fi

months='January|February|March|April|May|June|July|August|September|October|November|December|Jan|Feb|Mar|Apr|Jun|Jul|Aug|Sep|Sept|Oct|Nov|Dec'
if grep -qE "\b(${months}) [0-9]{1,2}(,|\b)" "$file"; then
  offender=$(grep -oE "\b(${months}) [0-9]{1,2}(,|\b)" "$file" | head -1)
  violations+=("rule-5 date: found '${offender}' (must be ISO YYYY-MM-DD)")
  rule_hits[5]=1
fi
if grep -qE '\b[0-9]{1,2}/[0-9]{1,2}/[0-9]{2,4}\b' "$file"; then
  offender=$(grep -oE '\b[0-9]{1,2}/[0-9]{1,2}/[0-9]{2,4}\b' "$file" | head -1)
  violations+=("rule-5 date: found slash-format '${offender}' (must be ISO YYYY-MM-DD)")
  rule_hits[5]=1
fi

if grep -qiE '\bclients?\b' "$file"; then
  violations+=("rule-6 terminology: found 'client' (must be 'customer')")
  rule_hits[6]=1
fi
if grep -qiE '\busers?\b' "$file"; then
  violations+=("rule-6 terminology: found 'user' (must be 'customer')")
  rule_hits[6]=1
fi

last_line=$(awk 'NF {last=$0} END {print last}' "$file")
last_line="${last_line%"${last_line##*[![:space:]]}"}"
if [ "$last_line" != "— ArcFactor Capital" ]; then
  violations+=("rule-7 closing: last line is '${last_line}' (must be exactly '— ArcFactor Capital')")
  rule_hits[7]=1
fi

if grep -qE '[0-9]%' "$file"; then
  violations+=("rule-8 percent: found '%' symbol (must be spelled 'percent')")
  rule_hits[8]=1
fi

for comp in 'Bridgewater' 'AQR' 'Renaissance' 'Two Sigma' 'D. E. Shaw' 'Citadel' 'Millennium' 'Point72' 'Man Group'; do
  if grep -qE "\b${comp}\b" "$file"; then
    violations+=("rule-9 competitor-name: found '${comp}' (must say 'a major incumbent')")
    rule_hits[9]=1
  fi
done

if grep -qE '(^|[^A-Za-z])AI([^A-Za-z]|$)' "$file"; then
  violations+=("rule-10 abbreviation: found bare 'AI' token (must be 'machine learning' or 'quantitative models')")
  rule_hits[10]=1
fi

printf '{"violations":['
for i in "${!violations[@]}"; do
  [ "$i" -gt 0 ] && printf ','
  v=${violations[$i]//\\/\\\\}
  v=${v//\"/\\\"}
  printf '"%s"' "$v"
done
printf '],"count":%d,"rules_checked":10,"rules_violated":[' "${#violations[@]}"
first=1
for rn in 1 2 3 4 5 6 7 8 9 10; do
  if [ "${rule_hits[$rn]:-0}" = "1" ]; then
    [ "$first" -eq 0 ] && printf ','
    printf '%d' "$rn"
    first=0
  fi
done
printf ']}\n'

[ "${#violations[@]}" -eq 0 ]
