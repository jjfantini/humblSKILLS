#!/bin/bash
# scan.sh - list App Store review CANDIDATES in an iOS / React Native / Expo repo.
#
# Usage: scan.sh [repo-dir]          (default: current directory)
#
# Prints one tab-separated line per candidate:
#   SEVERITY<TAB>GUIDELINE<TAB>file:line<TAB>check-id<TAB>matched text
# A candidate is NOT a finding. Verify each one with
# references/wiki/review/workflow/evidence-gate.md before reporting it.
# Exit 0 always: it lists, it does not gate. Pipe to sort to group by severity.
# Portable to macOS /bin/bash 3.2 and BSD grep. No dependencies.
set -u

ROOT="${1:-.}"
[ -d "$ROOT" ] || { echo "scan.sh: not a directory: $ROOT" >&2; exit 2; }
cd "$ROOT" || exit 2

EXCL="--exclude-dir=node_modules --exclude-dir=Pods --exclude-dir=.git --exclude-dir=build \
 --exclude-dir=DerivedData --exclude-dir=.expo --exclude-dir=dist --exclude-dir=vendor \
 --exclude-dir=Carthage --exclude-dir=__tests__ --exclude=package-lock.json --exclude=yarn.lock \
 --exclude=pnpm-lock.yaml --exclude=Podfile.lock"
INCL="--include=*.swift --include=*.m --include=*.mm --include=*.h --include=*.ts --include=*.tsx \
 --include=*.js --include=*.jsx --include=*.plist --include=*.json --include=*.strings \
 --include=*.xcstrings --include=*.entitlements"

# g [grep flags] ERE : search the source tree, prints file:line:text
g() { # shellcheck disable=SC2086
  grep -rnIE $EXCL $INCL "$@" . 2>/dev/null | sed 's|^\./||'; }
has() { [ -n "$(g -l "$@" | head -1)" ]; }

emit() { # sev guideline id  (reads file:line:text on stdin)
  while IFS= read -r hit; do
    loc="$(printf '%s' "$hit" | cut -d: -f1-2)"
    txt="$(printf '%s' "$hit" | cut -d: -f3- | sed 's/^[[:space:]]*//' | cut -c1-140)"
    printf '%s\t%s\t%s\t%s\t%s\n' "$1" "$2" "$loc" "$3" "$txt"
  done
}
whole() { printf '%s\t%s\t(repo)\t%s\t%s\n' "$1" "$2" "$3" "$4"; }

# ---- line patterns: id|severity|guideline|ERE --------------------------------
while IFS='|' read -r id sev gl re; do
  case "$id" in ''|\#*) continue ;; esac
  g "$re" | emit "$sev" "$gl" "$id"
done <<'EOF'
private-api|CRITICAL|2.5.1|NSSelectorFromString\(@?"_|dlopen\(|dlsym\(
hardcoded-secret|CRITICAL|1.6, 5.1|sk_live_[A-Za-z0-9]|AKIA[0-9A-Z]{16}|-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY
dynamic-code|CRITICAL|2.5.2|(^|[^A-Za-z0-9_.])eval\(|new Function\(
external-checkout|CRITICAL|3.1.1, 3.1.1(a), 3.1.3|checkout\.stripe\.com|buy\.stripe\.com|stripe\.com/checkout|paypal\.com/checkout|openStripeCheckout
ota-update|HIGH|2.5.2|codePush|code-push|Updates\.fetchUpdateAsync|Updates\.reloadAsync
secret-in-log|HIGH|1.6, 5.1|(console\.(log|info|debug|warn)|NSLog\(|print\(|os_log|debugPrint\().*([Tt]oken|[Pp]assword|[Ss]ecret|[Aa]pi[Kk]ey)
placeholder-content|HIGH|2.1(a)|[Ll]orem ipsum|[Cc]oming [Ss]oon|[Uu]nder [Cc]onstruction
precheck-terms|HIGH|1.2|\[(is)?([Aa]gree|[Aa]ccept|[Tt]erms|[Cc]onsent)[A-Za-z]*, *set[A-Za-z]*\] *= *useState\(true\)
pre-permission-wording|MEDIUM|5.1.1(iv)|(title=|Button\(|setTitle\()["'](Allow|Enable) (Location|Camera|Notifications|Tracking|Photos|Microphone|Contacts|Bluetooth|Health)
custom-review-prompt|MEDIUM|5.6.1|[Rr]ate (us|this app|our app)|5[ -]stars|[Ff]ive stars
android-reference|MEDIUM|2.3.10|[Aa]vailable on Android|Google Play|Play Store|Android version
pricing-word|MEDIUM|2.3.7|\((Free|FREE)\)|(Free|FREE)!
deprecated-uiwebview|HIGH|beyond-reference (ITMS-90809)|UIWebView
EOF

# ---- vague or empty purpose strings (5.1.1(ii)) ------------------------------
find . \( -name node_modules -o -name Pods -o -name .git -o -name build \) -prune -o \
     \( -name 'Info.plist' -o -name 'app.json' -o -name 'app.config.json' \) -type f -print 2>/dev/null |
while IFS= read -r f; do
  f="${f#./}"
  awk -v F="$f" '
    function chk(key, val, n) {
      if (length(val) < 25 || tolower(val) ~ /(needs?|requires?|wants?)( [a-z]+)? (access|permission)|^\$\(/)
        printf "MEDIUM\t5.1.1(ii)\t%s:%d\tvague-purpose-string\t%s = \"%s\"\n", F, n, key, val
    }
    /UsageDescription<\/key>/ { k=$0; gsub(/.*<key>|<\/key>.*/, "", k); want=1; next }
    want && /<string>/   { v=$0; gsub(/.*<string>|<\/string>.*/, "", v); chk(k, v, NR); want=0; next }
    want && /<string\/>/ { chk(k, "", NR); want=0; next }
    /"NS[A-Za-z]*UsageDescription"[[:space:]]*:/ {
      k=$0; sub(/^[^"]*"/, "", k); sub(/".*/, "", k)
      v=$0; sub(/^[^:]*:[[:space:]]*"/, "", v); sub(/".*/, "", v); chk(k, v, NR) }
  ' "$f"
done

# ---- presence checks: feature present while its counterpart is missing -------
ios=0
[ -n "$(find . -maxdepth 4 \( -name '*.xcodeproj' -o -name 'Info.plist' -o -name 'app.json' \) \
        -not -path '*/node_modules/*' 2>/dev/null | head -1)" ] && ios=1

if has 'GoogleMobileAds|AppLovin|FBAudienceNetwork|react-native-google-mobile-ads|AppsFlyer|react-native-appsflyer|ASIdentifierManager|advertisingIdentifier' \
   && ! has 'requestTrackingAuthorization|expo-tracking-transparency|react-native-tracking-transparency|requestTrackingPermissionsAsync'; then
  whole HIGH "5.1.2" att-missing "ad/attribution SDK or IDFA use found, no ATT request found; verify whether Apple-defined tracking happens"
fi
if has 'requestTrackingAuthorization|requestTrackingPermissionsAsync|expo-tracking-transparency' \
   && ! has 'NSUserTrackingUsageDescription|userTrackingPermission'; then
  whole HIGH "5.1.1(ii), 5.1.2" att-no-purpose-string "ATT requested but no NSUserTrackingUsageDescription found"
fi
if has 'createUser|createUserWithEmailAndPassword|signUp\(|signUpWith|createAccount|registerUser' \
   && ! has '[Dd]elete[A-Za-z]*[Aa]ccount|deleteUser|account[-_ ]?deletion|delete-account'; then
  whole HIGH "5.1.1(v)" account-deletion-missing "account creation found, no in-app deletion entry point found"
fi
if has 'GIDSignIn|GoogleSignin|react-native-google-signin|FBSDKLoginKit|LoginManager\.logIn|react-native-fbsdk' \
   && ! has 'ASAuthorizationAppleID|SignInWithAppleButton|expo-apple-authentication|react-native-apple-authentication|appleAuth'; then
  whole HIGH "4.8" login-alternative-missing "third-party social login found, no Sign in with Apple or other 4.8-compliant option found"
fi
if has 'SKPaymentQueue|Product\.purchase|Transaction\.updates|react-native-iap|expo-in-app-purchases|react-native-purchases|Purchases\.purchase' \
   && ! has -i 'restore|AppStore\.sync'; then
  whole HIGH "3.1.1" restore-purchases-missing "in-app purchase code found, no restore mechanism found"
fi
if [ "$ios" = 1 ] && ! has -i 'privacy[-_ ]?policy'; then
  whole HIGH "5.1.1(i)" privacy-policy-link-missing "no privacy policy link found in app source"
fi
if [ "$ios" = 1 ] && ! has 'privacyManifests' \
   && [ -z "$(find . -name PrivacyInfo.xcprivacy -not -path '*/node_modules/*' -not -path '*/Pods/*' 2>/dev/null | head -1)" ]; then
  whole MEDIUM "beyond-reference (ITMS-91053)" privacy-manifest-missing "no PrivacyInfo.xcprivacy or Expo privacyManifests found"
fi
if has 'react-native-webview|WKWebView'; then
  n="$(grep -rlIE $EXCL --include=*.swift --include=*.tsx --include=*.jsx . 2>/dev/null | wc -l | tr -d ' ')"
  [ "$n" -le 3 ] && whole MEDIUM "4.2" webview-wrapper "WebView with only $n UI source file(s); check for minimum native functionality"
fi
exit 0
