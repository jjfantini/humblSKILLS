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
	}, nil, false)
	want := []string{"claude-code"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPreferredDefaults_OnlyClaudeDetected(t *testing.T) {
	got := PreferredDefaults(testAdapters, map[string]bool{
		"claude-code": true,
	}, nil, false)
	want := []string{"claude-code"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPreferredDefaults_OnlyCursorDetected(t *testing.T) {
	got := PreferredDefaults(testAdapters, map[string]bool{
		"cursor": true,
	}, nil, false)
	want := []string{"cursor"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPreferredDefaults_NoneDetected(t *testing.T) {
	got := PreferredDefaults(testAdapters, map[string]bool{}, nil, false)
	if len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

func TestPreferredDefaults_ProfileWins(t *testing.T) {
	got := PreferredDefaults(testAdapters, map[string]bool{
		"claude-code": true,
		"cursor":      true,
	}, []string{"cursor"}, false)
	want := []string{"cursor"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPreferredDefaults_ProfileWins_EvenWhenGlobal(t *testing.T) {
	got := PreferredDefaults(testAdapters, map[string]bool{
		"claude-code": true,
		"cursor":      true,
	}, []string{"cursor"}, true)
	want := []string{"cursor"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPreferredDefaults_ProfileDropsUnknown(t *testing.T) {
	got := PreferredDefaults(testAdapters, map[string]bool{
		"claude-code": true,
	}, []string{"cursor", "ghost"}, false)
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
	}, nil, false)
	want := []string{"claude-code", "other"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPreferredDefaults_Global_KeepsBothClaudeAndCursor(t *testing.T) {
	// Global humblskills scope always symlinks every detected platform — no
	// claude/cursor dedup heuristic applies.
	got := PreferredDefaults(testAdapters, map[string]bool{
		"claude-code": true,
		"cursor":      true,
	}, nil, true)
	want := []string{"claude-code", "cursor"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPreferredDefaults_Global_NoneDetected(t *testing.T) {
	got := PreferredDefaults(testAdapters, map[string]bool{}, nil, true)
	if len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}
