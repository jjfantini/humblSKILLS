package tui

import (
	"os"
	"testing"
)

// The router is always on except for the HUMBLSKILLS_TUI_ROUTER emergency
// escape hatch: any value other than "1" turns it off.
func TestRouterEnabled(t *testing.T) {
	cases := []struct {
		name string
		env  *string // nil = unset
		want bool
	}{
		{"unset defaults on", nil, true},
		{"env 1 on", strPtr("1"), true},
		{"env 0 off", strPtr("0"), false},
		{"env empty off", strPtr(""), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("HUMBLSKILLS_TUI_ROUTER", "") // registers restore of the real value
			if tc.env != nil {
				os.Setenv("HUMBLSKILLS_TUI_ROUTER", *tc.env)
			} else {
				os.Unsetenv("HUMBLSKILLS_TUI_ROUTER")
			}
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
	restore := sessionWanted
	defer func() { sessionWanted = restore }()
	t.Setenv("HUMBLSKILLS_TUI_ROUTER", "")
	os.Unsetenv("HUMBLSKILLS_TUI_ROUTER")

	sessionWanted = false
	BeginSession()
	if !sessionWanted {
		t.Error("BeginSession did not engage the router")
	}

	os.Setenv("HUMBLSKILLS_TUI_ROUTER", "0")
	BeginSession()
	if sessionWanted {
		t.Error("BeginSession engaged the router despite the env escape hatch")
	}
}

func strPtr(s string) *string { return &s }
