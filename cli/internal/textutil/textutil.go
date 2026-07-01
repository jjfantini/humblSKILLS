// Package textutil holds tiny string/formatting helpers shared across the CLI
// and TUI. They previously existed as near-identical private copies in several
// packages (plural/pluralDash/plural2, firstNonEmpty/firstNonEmptyStr); this
// package is the single source of truth.
package textutil

import "strings"

// Plural returns the English plural suffix for a count: "" for exactly one,
// "s" otherwise. Use as fmt.Sprintf("%d skill%s", n, textutil.Plural(n)).
func Plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// FirstNonEmpty returns the first argument that is not the empty string, or ""
// if every argument is empty. Whitespace-only strings count as non-empty.
func FirstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}
	return ""
}

// FirstNonBlank returns the first argument that contains a non-whitespace
// character, or "" if none do. Unlike FirstNonEmpty, a value that is only
// whitespace is skipped. The returned value is the original (untrimmed) string.
func FirstNonBlank(vs ...string) string {
	for _, v := range vs {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
