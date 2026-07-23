package tui

import (
	"os"
	"testing"
)

// RouterEnabled precedence: the env var (when set) always wins — "1" on,
// anything else off — and the profile preference only applies when it's unset.
func TestRouterEnabled_Precedence(t *testing.T) {
	restore := routerPref
	defer SetRouterPreference(restore)

	cases := []struct {
		name string
		env  *string // nil = unset
		pref bool
		want bool
	}{
		{"unset env, pref off", nil, false, false},
		{"unset env, pref on", nil, true, true},
		{"env 1 overrides pref off", strPtr("1"), false, true},
		{"env 0 overrides pref on", strPtr("0"), true, false},
		{"env empty overrides pref on", strPtr(""), true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("HUMBLSKILLS_TUI_ROUTER", "") // registers restore of the real value
			if tc.env != nil {
				os.Setenv("HUMBLSKILLS_TUI_ROUTER", *tc.env)
			} else {
				os.Unsetenv("HUMBLSKILLS_TUI_ROUTER")
			}
			SetRouterPreference(tc.pref)
			if got := RouterEnabled(); got != tc.want {
				t.Errorf("RouterEnabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

func strPtr(s string) *string { return &s }
