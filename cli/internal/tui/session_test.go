package tui

import (
	"os"
	"testing"
)

// RouterEnabled precedence: the env var (when set) always wins — "1" on,
// anything else off — then the profile preference, then the default (on).
func TestRouterEnabled_Precedence(t *testing.T) {
	restore := routerPref
	defer SetRouterPreference(restore)

	on, off := true, false
	cases := []struct {
		name string
		env  *string // nil = unset
		pref *bool   // nil = unset
		want bool
	}{
		{"all unset defaults on", nil, nil, true},
		{"pref off", nil, &off, false},
		{"pref on", nil, &on, true},
		{"env 1 overrides pref off", strPtr("1"), &off, true},
		{"env 0 overrides pref on", strPtr("0"), &on, false},
		{"env 0 overrides default", strPtr("0"), nil, false},
		{"env empty overrides default", strPtr(""), nil, false},
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

// The session router only engages once BeginSession opts the process in —
// one-shot commands never call it, so Run keeps a per-screen program and
// their post-screen terminal output stays visible.
func TestBeginSession_GatesRouter(t *testing.T) {
	restorePref := routerPref
	restoreWanted := sessionWanted
	defer func() {
		routerPref = restorePref
		sessionWanted = restoreWanted
	}()
	t.Setenv("HUMBLSKILLS_TUI_ROUTER", "")
	os.Unsetenv("HUMBLSKILLS_TUI_ROUTER")

	sessionWanted = false
	SetRouterPreference(nil)
	if sessionWanted {
		t.Fatal("sessionWanted true before BeginSession")
	}
	BeginSession()
	if !sessionWanted {
		t.Error("BeginSession did not engage the router with the default-on setting")
	}

	off := false
	SetRouterPreference(&off)
	BeginSession()
	if sessionWanted {
		t.Error("BeginSession engaged the router despite tui_router off")
	}
}

func strPtr(s string) *string { return &s }
