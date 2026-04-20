package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

func printer(opts ui.Options) (*ui.Printer, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errW := &bytes.Buffer{}
	opts.Out = out
	opts.Err = errW
	opts.NoColor = true
	return ui.New(opts), out, errW
}

func TestPrintln_NormalEmitsLine(t *testing.T) {
	p, out, _ := printer(ui.Options{Level: ui.LevelNormal})
	p.Println("hi")
	if got := strings.TrimSpace(out.String()); got != "hi" {
		t.Errorf("got %q", got)
	}
}

func TestPrintln_QuietSuppressed(t *testing.T) {
	p, out, _ := printer(ui.Options{Level: ui.LevelQuiet})
	p.Println("hidden")
	if out.Len() != 0 {
		t.Errorf("expected empty, got %q", out.String())
	}
}

func TestPrintln_JSONSuppressed(t *testing.T) {
	p, out, _ := printer(ui.Options{JSON: true})
	p.Println("hidden")
	if out.Len() != 0 {
		t.Errorf("expected empty in json mode, got %q", out.String())
	}
}

func TestHeader_NormalPrints(t *testing.T) {
	p, out, _ := printer(ui.Options{Level: ui.LevelNormal})
	p.Header("install")
	s := out.String()
	if !strings.Contains(s, "humblskills") || !strings.Contains(s, "install") {
		t.Errorf("header missing pieces: %q", s)
	}
}

func TestHeader_QuietSuppressed(t *testing.T) {
	p, out, _ := printer(ui.Options{Level: ui.LevelQuiet})
	p.Header("anything")
	if out.Len() != 0 {
		t.Errorf("header should be silent in quiet")
	}
}

func TestSection_NormalPrintsLabel(t *testing.T) {
	p, out, _ := printer(ui.Options{Level: ui.LevelNormal})
	p.Section("Adapters:")
	if !strings.Contains(out.String(), "Adapters:") {
		t.Errorf("section label missing: %q", out.String())
	}
}

func TestSection_JSONSuppressed(t *testing.T) {
	p, out, _ := printer(ui.Options{JSON: true})
	p.Section("x")
	if out.Len() != 0 {
		t.Errorf("section should be silent in json mode")
	}
}

func TestIsJSON_FlagRoundTrips(t *testing.T) {
	p, _, _ := printer(ui.Options{JSON: true})
	if !p.IsJSON() {
		t.Error("IsJSON = false after JSON:true")
	}
	p2, _, _ := printer(ui.Options{JSON: false})
	if p2.IsJSON() {
		t.Error("IsJSON = true after JSON:false")
	}
}

func TestOutErrAccessors_ReturnInjectedWriters(t *testing.T) {
	out := &bytes.Buffer{}
	errW := &bytes.Buffer{}
	p := ui.New(ui.Options{Out: out, Err: errW})
	if p.Out() != out {
		t.Error("Out() did not return injected writer")
	}
	if p.Err() != errW {
		t.Error("Err() did not return injected writer")
	}
}

func TestTheme_Accessor_Works(t *testing.T) {
	p, _, _ := printer(ui.Options{})
	if p.Theme() == nil {
		t.Error("Theme() returned nil")
	}
}

func TestNewPrompter_YesBypassesTTYDetection(t *testing.T) {
	p := ui.NewPrompter(true)
	if !p.Yes {
		t.Error("Yes not set")
	}
	// With Yes, Confirm returns true regardless of dflt.
	got, err := p.Confirm("proceed?", false)
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if !got {
		t.Error("Yes mode should return true from Confirm")
	}
}

func TestNewPrompter_NonInteractiveConfirmReturnsDefault(t *testing.T) {
	// NewPrompter in tests runs without a TTY, so Interactive=false.
	p := ui.NewPrompter(false)
	if p.Interactive {
		// Running under a TTY (unusual for CI but possible in local dev) —
		// this test's assumption breaks; skip rather than flake.
		t.Skip("stderr reports TTY; test can't exercise non-interactive path")
	}
	v, err := p.Confirm("q", true)
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if !v {
		t.Error("non-interactive Confirm should return dflt=true")
	}
	v, _ = p.Confirm("q", false)
	if v {
		t.Error("non-interactive Confirm should return dflt=false")
	}
}

func TestMultiSelect_EmptyOptionsNoOp(t *testing.T) {
	p := ui.NewPrompter(true)
	got, err := p.MultiSelect("pick", nil)
	if err != nil {
		t.Fatalf("MultiSelect: %v", err)
	}
	if got != nil {
		t.Errorf("empty options should yield nil, got %v", got)
	}
}

func TestMultiSelect_YesReturnsAllValues(t *testing.T) {
	p := ui.NewPrompter(true)
	opts := []ui.MultiSelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
	}
	got, err := p.MultiSelect("pick", opts)
	if err != nil {
		t.Fatalf("MultiSelect: %v", err)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("got %v", got)
	}
}

func TestMultiSelect_NonInteractiveReturnsErrNonInteractive(t *testing.T) {
	p := ui.NewPrompter(false)
	if p.Interactive {
		t.Skip("stderr reports TTY; test can't exercise non-interactive path")
	}
	_, err := p.MultiSelect("pick", []ui.MultiSelectOption{{Value: "a"}})
	if err != ui.ErrNonInteractive {
		t.Errorf("err = %v, want ErrNonInteractive", err)
	}
}

func TestSelect_EmptyOptionsReturnsEmpty(t *testing.T) {
	p := ui.NewPrompter(true)
	got, err := p.Select("pick", "", nil)
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if got != "" {
		t.Errorf("got %q", got)
	}
}

func TestSelect_YesCantPickForUser(t *testing.T) {
	p := ui.NewPrompter(true)
	_, err := p.Select("pick", "", []ui.SelectOption{{Value: "a"}})
	if err != ui.ErrNonInteractive {
		t.Errorf("err = %v, want ErrNonInteractive", err)
	}
}

func TestSelect_NonInteractiveReturnsErrNonInteractive(t *testing.T) {
	p := ui.NewPrompter(false)
	if p.Interactive {
		t.Skip("stderr reports TTY; test can't exercise non-interactive path")
	}
	_, err := p.Select("pick", "d", []ui.SelectOption{{Value: "a"}})
	if err != ui.ErrNonInteractive {
		t.Errorf("err = %v, want ErrNonInteractive", err)
	}
}

func TestSecret_NonInteractiveReturnsErrNonInteractive(t *testing.T) {
	p := ui.NewPrompter(false)
	if p.Interactive {
		t.Skip("stderr reports TTY; test can't exercise non-interactive path")
	}
	_, err := p.Secret("enter key")
	if err != ui.ErrNonInteractive {
		t.Errorf("err = %v, want ErrNonInteractive", err)
	}
}

func TestDefaultTheme_ReturnsUsableTheme(t *testing.T) {
	th := ui.DefaultTheme()
	if th == nil {
		t.Fatal("DefaultTheme returned nil")
	}
	// A rendered string must not be empty.
	if got := th.Brand.Render("x"); got == "" {
		t.Error("Brand render produced empty string")
	}
}

func TestHuhTheme_DerivedFromPalette(t *testing.T) {
	th := ui.HuhTheme(ui.DefaultTheme())
	if th == nil {
		t.Fatal("HuhTheme returned nil")
	}
}
