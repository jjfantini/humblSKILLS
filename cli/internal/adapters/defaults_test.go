package adapters

import (
	"reflect"
	"testing"
)

var testAdapters = []Adapter{
	{Name: "claude-code"},
	{Name: "cursor"},
}

func TestPreferredDefaults_BothDetected_ClaudeCodeOnly(t *testing.T) {
	got := PreferredDefaults(testAdapters, map[string]bool{
		"claude-code": true,
		"cursor":      true,
	}, nil)
	want := []string{"claude-code"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPreferredDefaults_OnlyClaudeDetected(t *testing.T) {
	got := PreferredDefaults(testAdapters, map[string]bool{
		"claude-code": true,
	}, nil)
	want := []string{"claude-code"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPreferredDefaults_OnlyCursorDetected(t *testing.T) {
	got := PreferredDefaults(testAdapters, map[string]bool{
		"cursor": true,
	}, nil)
	want := []string{"cursor"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPreferredDefaults_NoneDetected(t *testing.T) {
	got := PreferredDefaults(testAdapters, map[string]bool{}, nil)
	if len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

func TestPreferredDefaults_ProfileWins(t *testing.T) {
	got := PreferredDefaults(testAdapters, map[string]bool{
		"claude-code": true,
		"cursor":      true,
	}, []string{"cursor"})
	want := []string{"cursor"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPreferredDefaults_ProfileDropsUnknown(t *testing.T) {
	got := PreferredDefaults(testAdapters, map[string]bool{
		"claude-code": true,
	}, []string{"cursor", "ghost"})
	want := []string{"cursor"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPreferredDefaults_CascadeOnlyDropsCursor(t *testing.T) {
	// A hypothetical third adapter should survive the dedup cascade —
	// the rule is specifically about claude/cursor duplication.
	adapterList := []Adapter{
		{Name: "claude-code"},
		{Name: "cursor"},
		{Name: "other"},
	}
	got := PreferredDefaults(adapterList, map[string]bool{
		"claude-code": true,
		"cursor":      true,
		"other":       true,
	}, nil)
	want := []string{"claude-code", "other"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
