package selfupdate

import "golang.org/x/mod/semver"

// Normalize ensures v has the "v" prefix golang.org/x/mod/semver requires,
// without double-prefixing versions that already have one.
func Normalize(v string) string {
	if v == "" || v[0] == 'v' || v[0] == 'V' {
		return v
	}
	return "v" + v
}

// Compare reports the relative order of current vs. latest: negative when
// current < latest (an upgrade is available), zero when equal, positive
// when current is newer (e.g. a pre-release build). semver.Compare already
// treats a non-semver string (e.g. "dev", or a dirty local build's
// commit-ish version) as older than any valid version, which is exactly the
// "always offer to upgrade a dev build" behaviour we want.
func Compare(current, latest string) int {
	return semver.Compare(Normalize(current), Normalize(latest))
}

// IsUpgradeAvailable reports whether latest is strictly newer than current.
func IsUpgradeAvailable(current, latest string) bool {
	return Compare(current, latest) < 0
}
