package ui

import "github.com/charmbracelet/huh"

// HuhTheme returns a *huh.Theme aligned with the humblskills palette. The base
// is huh's Catppuccin theme (which the CLI was already using) with a handful of
// Focused-state overrides so prompts share the same brand purple, muted crumb
// colour, and accent as the rest of the CLI.
func HuhTheme(t *Theme) *huh.Theme {
	if t == nil {
		t = DefaultTheme()
	}
	h := huh.ThemeCatppuccin()
	p := t.Palette

	// Titles + selection cues pick up the brand accent.
	h.Focused.Title = h.Focused.Title.Foreground(p.Brand).Bold(true)
	h.Focused.SelectSelector = h.Focused.SelectSelector.Foreground(p.Brand)
	h.Focused.MultiSelectSelector = h.Focused.MultiSelectSelector.Foreground(p.Brand)
	h.Focused.SelectedPrefix = h.Focused.SelectedPrefix.Foreground(p.Name)
	h.Focused.SelectedOption = h.Focused.SelectedOption.Foreground(p.Name)
	h.Focused.NextIndicator = h.Focused.NextIndicator.Foreground(p.Brand)
	h.Focused.PrevIndicator = h.Focused.PrevIndicator.Foreground(p.Brand)

	// Descriptions / blurred text use our muted crumb colour.
	h.Focused.Description = h.Focused.Description.Foreground(p.Muted)
	h.Blurred.Description = h.Blurred.Description.Foreground(p.Muted)

	// Error indicator uses our error red so prompts agree with Printer.
	h.Focused.ErrorIndicator = h.Focused.ErrorIndicator.Foreground(p.Error)
	h.Focused.ErrorMessage = h.Focused.ErrorMessage.Foreground(p.Error)

	return h
}
