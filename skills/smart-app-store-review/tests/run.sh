#!/bin/bash
# tests/run.sh - hermetic check for scripts/scan.sh. Run with macOS /bin/bash 3.2:
#   /bin/bash tests/run.sh
set -u
HERE="$(cd "$(dirname "$0")/.." && pwd)"
SCAN="$HERE/scripts/scan.sh"
T="$(mktemp -d)"; trap 'rm -rf "$T"' EXIT
fail=0
want()   { grep -q "	$2	" "$1" || { echo "FAIL: expected $2"; fail=1; }; }
unwant() { if grep -q "	$2	" "$1"; then echo "FAIL: unexpected $2"; fail=1; fi; }

# --- a non-compliant Expo app -----------------------------------------------
mkdir -p "$T/bad/src" "$T/bad/node_modules/x"
cat > "$T/bad/app.json" <<'EOF'
{ "expo": { "ios": { "infoPlist": {
  "NSCameraUsageDescription": "This app needs camera access" } } } }
EOF
cat > "$T/bad/src/App.tsx" <<'EOF'
import { GoogleSignin } from '@react-native-google-signin/google-signin';
import { BannerAd } from 'react-native-google-mobile-ads';
const API_KEY = 'sk_live_abc123';
const [agreed, setAgreed] = useState(true);
console.log('Access token:', accessToken);
Linking.openURL('https://checkout.stripe.com/pay/x');
eval(downloadedCode);
const t = "Also available on Android";
export const signUp = (e, p) => auth.createUserWithEmailAndPassword(e, p);
<Button title="Enable Location" onPress={requestLocation} />
EOF
echo "eval(x); sk_live_ignored" > "$T/bad/node_modules/x/index.js"
/bin/bash "$SCAN" "$T/bad" > "$T/bad.out"
for id in hardcoded-secret precheck-terms secret-in-log external-checkout dynamic-code \
          android-reference account-deletion-missing login-alternative-missing att-missing \
          vague-purpose-string pre-permission-wording privacy-policy-link-missing privacy-manifest-missing; do
  want "$T/bad.out" "$id"
done
if grep -q node_modules "$T/bad.out"; then echo "FAIL: scanned node_modules"; fail=1; fi

# --- a compliant native app ---------------------------------------------------
mkdir -p "$T/good/App.xcodeproj" "$T/good/App"
cat > "$T/good/App/Info.plist" <<'EOF'
<key>NSCameraUsageDescription</key>
<string>Scan a receipt so the app can fill in the expense amount for you.</string>
EOF
cat > "$T/good/App/PrivacyInfo.xcprivacy" <<'EOF'
<plist></plist>
EOF
cat > "$T/good/App/Auth.swift" <<'EOF'
import AuthenticationServices
let req = ASAuthorizationAppleIDProvider().createRequest()
GIDSignIn.sharedInstance.signIn(withPresenting: vc)
func createAccount() {}
func deleteAccount() {}
let privacyPolicyURL = URL(string: "https://example.com/privacy")!
EOF
/bin/bash "$SCAN" "$T/good" > "$T/good.out"
for id in vague-purpose-string login-alternative-missing account-deletion-missing \
          privacy-policy-link-missing privacy-manifest-missing hardcoded-secret; do
  unwant "$T/good.out" "$id"
done

[ "$fail" = 0 ] && echo "PASS: scan.sh ($(wc -l < "$T/bad.out" | tr -d ' ') candidates on bad fixture, $(wc -l < "$T/good.out" | tr -d ' ') on good)"
exit "$fail"
