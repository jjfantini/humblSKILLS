package ui

import (
	"errors"
	"testing"
)

func TestPrompter_ConfirmYesAlwaysTrue(t *testing.T) {
	p := &Prompter{Yes: true}
	got, err := p.Confirm("whatever", false)
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Errorf("Yes prompter must return true regardless of default")
	}
}

func TestPrompter_ConfirmNonInteractiveReturnsDefault(t *testing.T) {
	p := &Prompter{Interactive: false}
	got, err := p.Confirm("whatever", true)
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Errorf("non-interactive should fall back to default=true")
	}
}

func TestPrompter_MultiSelectYesReturnsAll(t *testing.T) {
	p := &Prompter{Yes: true}
	opts := []MultiSelectOption{
		{Label: "a", Value: "a"},
		{Label: "b", Value: "b"},
	}
	got, err := p.MultiSelect("t", opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("expected all values, got %v", got)
	}
}

func TestPrompter_MultiSelectNonInteractiveErrors(t *testing.T) {
	p := &Prompter{Interactive: false}
	_, err := p.MultiSelect("t", []MultiSelectOption{{Value: "a"}})
	if !errors.Is(err, ErrNonInteractive) {
		t.Errorf("expected ErrNonInteractive, got %v", err)
	}
}

func TestPrompter_MultiSelectEmpty(t *testing.T) {
	p := &Prompter{Interactive: false}
	got, err := p.MultiSelect("t", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}
