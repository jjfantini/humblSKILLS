package textutil

import "testing"

func TestPlural(t *testing.T) {
	cases := map[int]string{0: "s", 1: "", 2: "s", -1: "s", 100: "s"}
	for n, want := range cases {
		if got := Plural(n); got != want {
			t.Errorf("Plural(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := FirstNonEmpty("", "", "a", "b"); got != "a" {
		t.Errorf("got %q, want a", got)
	}
	if got := FirstNonEmpty("", ""); got != "" {
		t.Errorf("got %q, want empty", got)
	}
	// Whitespace counts as non-empty for FirstNonEmpty.
	if got := FirstNonEmpty("", "  ", "x"); got != "  " {
		t.Errorf("got %q, want two spaces", got)
	}
}

func TestFirstNonBlank(t *testing.T) {
	// Whitespace-only values are skipped; the original (untrimmed) value wins.
	if got := FirstNonBlank("", "  ", "  hi ", "x"); got != "  hi " {
		t.Errorf("got %q, want '  hi '", got)
	}
	if got := FirstNonBlank("", "   "); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}
