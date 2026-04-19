package adapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandPath_Home(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	if got := ExpandPath("~/foo"); got != filepath.Join(home, "foo") {
		t.Errorf("~/foo: got %q", got)
	}
	if got := ExpandPath("~"); got != home {
		t.Errorf("~: got %q", got)
	}
}

func TestExpandPath_Env(t *testing.T) {
	t.Setenv("HUMBLSKILLS_TEST_DIR", "/tmp/ok")
	if got := ExpandPath("$HUMBLSKILLS_TEST_DIR/x"); got != "/tmp/ok/x" {
		t.Errorf("got %q", got)
	}
	if got := ExpandPath("${HUMBLSKILLS_TEST_DIR}/x"); got != "/tmp/ok/x" {
		t.Errorf("braced: got %q", got)
	}
}

func TestDetect_PathExists(t *testing.T) {
	dir := t.TempDir()
	adapter := Adapter{
		Name: "fake",
		Detect: DetectRules{
			AnyOf: []DetectRule{{PathExists: dir}},
		},
	}
	got := Detect([]Adapter{adapter})
	if len(got) != 1 || !got[0].Detected {
		t.Fatalf("expected detection, got %+v", got)
	}
	if !strings.Contains(got[0].Reason, dir) {
		t.Errorf("reason should mention matched path: %q", got[0].Reason)
	}
}

func TestDetect_PathMissing(t *testing.T) {
	adapter := Adapter{
		Name: "fake",
		Detect: DetectRules{
			AnyOf: []DetectRule{{PathExists: "/nonexistent/path/humblskills-test"}},
		},
	}
	got := Detect([]Adapter{adapter})
	if got[0].Detected {
		t.Errorf("expected not detected")
	}
}

func TestDetect_EnvVar(t *testing.T) {
	t.Setenv("HUMBLSKILLS_TEST_FLAG", "1")
	adapter := Adapter{
		Name:   "fake",
		Detect: DetectRules{AnyOf: []DetectRule{{Env: "HUMBLSKILLS_TEST_FLAG"}}},
	}
	got := Detect([]Adapter{adapter})
	if !got[0].Detected {
		t.Errorf("expected detection via env")
	}
}

func TestDetect_AllOf(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HUMBLSKILLS_TEST_ALL", "1")

	both := Adapter{
		Name: "both",
		Detect: DetectRules{AllOf: []DetectRule{
			{PathExists: dir},
			{Env: "HUMBLSKILLS_TEST_ALL"},
		}},
	}
	miss := Adapter{
		Name: "miss",
		Detect: DetectRules{AllOf: []DetectRule{
			{PathExists: dir},
			{Env: "HUMBLSKILLS_DEFINITELY_UNSET"},
		}},
	}
	got := Detect([]Adapter{both, miss})
	if !got[0].Detected {
		t.Errorf("all_of should match: %+v", got[0])
	}
	if got[1].Detected {
		t.Errorf("all_of should fail: %+v", got[1])
	}
}

func TestDetect_NoRules(t *testing.T) {
	a := Adapter{Name: "empty"}
	got := Detect([]Adapter{a})
	if got[0].Detected {
		t.Errorf("empty rules should not match")
	}
}
