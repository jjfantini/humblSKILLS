package ui

import (
	"bytes"
	"strings"
	"testing"
)

func newTestPrinter(opts Options) (*Printer, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	err := &bytes.Buffer{}
	opts.Out = out
	opts.Err = err
	opts.NoColor = true // deterministic output in tests
	return New(opts), out, err
}

func TestInfoNormal(t *testing.T) {
	p, out, _ := newTestPrinter(Options{Level: LevelNormal})
	p.Info("hello %s", "world")
	if !strings.Contains(out.String(), "hello world") {
		t.Errorf("got %q", out.String())
	}
}

func TestQuietSilencesInfo(t *testing.T) {
	p, out, _ := newTestPrinter(Options{Level: LevelQuiet})
	p.Info("hidden")
	p.Success("hidden")
	p.Warn("hidden")
	p.Detail("hidden")
	if out.Len() != 0 {
		t.Errorf("expected no stdout, got %q", out.String())
	}
}

func TestQuietStillEmitsError(t *testing.T) {
	p, _, errW := newTestPrinter(Options{Level: LevelQuiet})
	p.Error("boom")
	if !strings.Contains(errW.String(), "boom") {
		t.Errorf("got %q", errW.String())
	}
}

func TestVerboseEnablesDetail(t *testing.T) {
	p, out, _ := newTestPrinter(Options{Level: LevelVerbose})
	p.Detail("deep")
	if !strings.Contains(out.String(), "deep") {
		t.Errorf("got %q", out.String())
	}
}

func TestNormalSuppressesDetail(t *testing.T) {
	p, out, _ := newTestPrinter(Options{Level: LevelNormal})
	p.Detail("deep")
	if out.Len() != 0 {
		t.Errorf("expected empty out, got %q", out.String())
	}
}

func TestJSONModeSuppressesHuman(t *testing.T) {
	p, out, errW := newTestPrinter(Options{Level: LevelNormal, JSON: true})
	p.Info("hi")
	p.Success("done")
	p.Warn("eh")
	p.Detail("meh")
	if out.Len() != 0 {
		t.Errorf("expected empty stdout in JSON mode, got %q", out.String())
	}
	if errW.Len() != 0 {
		t.Errorf("expected empty stderr, got %q", errW.String())
	}
	// Errors still flow.
	p.Error("boom")
	if !strings.Contains(errW.String(), "boom") {
		t.Errorf("error should still print in JSON mode")
	}
}

func TestJSONEncodesDocument(t *testing.T) {
	p, out, _ := newTestPrinter(Options{JSON: true})
	if err := p.JSON(map[string]any{"k": "v"}); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, `"k": "v"`) {
		t.Errorf("got %q", got)
	}
}

func TestCaptureWriters_CollectsOutAndErrThenRestores(t *testing.T) {
	p, out, errW := newTestPrinter(Options{Level: LevelNormal})

	restore := p.CaptureWriters()
	p.Info("captured info")
	p.Warn("captured warn")
	p.Error("captured error")
	if out.Len() != 0 || errW.Len() != 0 {
		t.Fatalf("expected nothing written to the original writers while captured, got out=%q err=%q", out.String(), errW.String())
	}

	captured := restore()
	for _, want := range []string{"captured info", "captured warn", "captured error"} {
		if !strings.Contains(captured, want) {
			t.Errorf("captured output missing %q:\n%s", want, captured)
		}
	}

	// Restored writers should receive output normally again.
	p.Info("after restore")
	if !strings.Contains(out.String(), "after restore") {
		t.Errorf("expected writes after restore to reach the original writer, got %q", out.String())
	}
}

func TestNoColorStripsANSI(t *testing.T) {
	// No-color printer produces plain text.
	p, out, _ := newTestPrinter(Options{Level: LevelNormal})
	p.Success("done")
	s := out.String()
	if strings.Contains(s, "\x1b[") {
		t.Errorf("unexpected ANSI in no-color output: %q", s)
	}
}
