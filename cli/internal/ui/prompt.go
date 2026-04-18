package ui

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

// ErrNonInteractive signals that a prompt was required but the caller is not
// running on a TTY and did not pass --yes. Returned by MultiSelect /
// SingleSelect so callers can surface a helpful error.
var ErrNonInteractive = errors.New("non-interactive session: pass --yes or provide explicit args")

// Prompter asks the user questions. Behaviour degrades predictably when the
// process is not running interactively: Confirm falls back to dflt, and
// selection prompts return ErrNonInteractive so the caller can fail loudly
// rather than silently picking for the user.
type Prompter struct {
	Yes         bool
	Interactive bool
	Out         *os.File // used for TTY detection; defaults to stderr
}

// NewPrompter returns a Prompter configured from the caller's context. yes
// forces auto-accept; interactive=false disables prompting entirely.
func NewPrompter(yes bool) *Prompter {
	out := os.Stderr
	p := &Prompter{Yes: yes, Out: out}
	if yes {
		return p
	}
	if term.IsTerminal(int(out.Fd())) && term.IsTerminal(int(os.Stdin.Fd())) {
		p.Interactive = true
	}
	return p
}

// theme is the shared huh theme for every interactive prompt. Catppuccin
// gives the CLI a consistent, modern look and respects NO_COLOR via huh's
// internal colour handling.
func theme() *huh.Theme { return huh.ThemeCatppuccin() }

// Confirm asks a yes/no question. When non-interactive, returns dflt.
func (p *Prompter) Confirm(title string, dflt bool) (bool, error) {
	if p.Yes {
		return true, nil
	}
	if !p.Interactive {
		return dflt, nil
	}
	v := dflt
	err := huh.NewConfirm().
		Title(title).
		Affirmative("Yes").
		Negative("No").
		Value(&v).
		WithTheme(theme()).
		Run()
	if err != nil {
		return dflt, err
	}
	return v, nil
}

// MultiSelectOption is one entry shown in a MultiSelect.
type MultiSelectOption struct {
	Label string
	Value string
	// Selected preselects this option.
	Selected bool
}

// MultiSelect shows a checkbox list and returns the selected Values. When
// --yes is set, every option is returned. When non-interactive and --yes is
// NOT set, returns ErrNonInteractive.
func (p *Prompter) MultiSelect(title string, options []MultiSelectOption) ([]string, error) {
	if len(options) == 0 {
		return nil, nil
	}
	if p.Yes {
		out := make([]string, 0, len(options))
		for _, o := range options {
			out = append(out, o.Value)
		}
		return out, nil
	}
	if !p.Interactive {
		return nil, ErrNonInteractive
	}

	opts := make([]huh.Option[string], 0, len(options))
	for _, o := range options {
		label := o.Label
		if label == "" {
			label = o.Value
		}
		opts = append(opts, huh.NewOption(label, o.Value).Selected(o.Selected))
	}
	var selected []string
	err := huh.NewMultiSelect[string]().
		Title(title).
		Options(opts...).
		Value(&selected).
		WithTheme(theme()).
		Run()
	if err != nil {
		return nil, fmt.Errorf("prompt: %w", err)
	}
	return selected, nil
}

// SelectOption is one entry shown in a Select.
type SelectOption struct {
	// Label is the visible text. Falls back to Value when empty.
	Label string
	// Value is returned from Select when this entry is picked.
	Value string
}

// Select shows a single-select picker with type-to-filter. It returns the
// chosen option's Value. Requires an interactive TTY: returns
// ErrNonInteractive (or a usage-style error via --yes) when a selection can't
// be made for the user.
func (p *Prompter) Select(title, description string, options []SelectOption) (string, error) {
	if len(options) == 0 {
		return "", nil
	}
	if p.Yes {
		// --yes can't pick for the user — a single-select has no safe default.
		return "", ErrNonInteractive
	}
	if !p.Interactive {
		return "", ErrNonInteractive
	}

	opts := make([]huh.Option[string], 0, len(options))
	for _, o := range options {
		label := o.Label
		if label == "" {
			label = o.Value
		}
		opts = append(opts, huh.NewOption(label, o.Value))
	}
	var picked string
	sel := huh.NewSelect[string]().
		Title(title).
		Options(opts...).
		Filtering(true).
		Height(12).
		Value(&picked)
	if description != "" {
		sel = sel.Description(description)
	}
	if err := sel.WithTheme(theme()).Run(); err != nil {
		return "", fmt.Errorf("prompt: %w", err)
	}
	return picked, nil
}
