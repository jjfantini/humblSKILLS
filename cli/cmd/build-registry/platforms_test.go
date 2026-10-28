package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePlatformSkill(t *testing.T, skillsDir, name, extraMeta string) {
	t.Helper()
	d := filepath.Join(skillsDir, name)
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nname: " + name + "\ndescription: " + name + " description line long enough\n" +
		"version: 1.0.0\nmetadata:\n  category: development\n" + extraMeta + "---\n\nBody.\n"
	if err := os.WriteFile(filepath.Join(d, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeBase(t *testing.T, dir, skillsJSON string) string {
	t.Helper()
	p := filepath.Join(dir, "base-registry.json")
	if err := os.WriteFile(p, []byte(`{"schema_version":1,"skills":[`+skillsJSON+`]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

var testAdapters = map[string]struct{}{"claude-code": {}, "claude-desktop": {}, "cursor": {}, "codex": {}}

func TestCheckPlatformsKept(t *testing.T) {
	cases := []struct {
		name    string
		meta    string // metadata lines for skill "foo"
		base    string // base registry skills array body
		wantErr string // substring; empty means pass
	}{
		{
			name:    "dropped platform fails",
			meta:    "  platforms: [claude-code, cursor]\n",
			base:    `{"name":"foo","platforms":["claude-code","claude-desktop","cursor"]}`,
			wantErr: "foo drops claude-desktop",
		},
		{
			name: "same or added platforms pass",
			meta: "  platforms: [claude-code, cursor, codex]\n",
			base: `{"name":"foo","platforms":["claude-code","cursor"]}`,
		},
		{
			name: "new skill passes",
			meta: "  platforms: [claude-code]\n",
			base: `{"name":"other","platforms":["cursor"]}`,
		},
		{
			name: "empty now means every adapter",
			meta: "",
			base: `{"name":"foo","platforms":["claude-code"]}`,
		},
		{
			name:    "narrowing from every adapter fails",
			meta:    "  platforms: [claude-code, cursor, codex]\n",
			base:    `{"name":"foo"}`,
			wantErr: "foo drops claude-desktop",
		},
		{
			name: "rename matches previous name",
			meta: "  previous_names: [old-foo]\n  platforms: [claude-code, cursor]\n",
			base: `{"name":"old-foo","platforms":["claude-code","cursor"]}`,
		},
		{
			name: "platform without an adapter is ignored",
			meta: "  platforms: [claude-code]\n",
			base: `{"name":"foo","platforms":["claude-code","retired-agent"]}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			skillsDir := filepath.Join(root, "skills")
			writePlatformSkill(t, skillsDir, "foo", tc.meta)
			err := checkPlatformsKept(skillsDir, writeBase(t, root, tc.base), testAdapters)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("want pass, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestCheckPlatformsKept_MissingBase(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	writePlatformSkill(t, skillsDir, "foo", "")
	if err := checkPlatformsKept(skillsDir, filepath.Join(root, "nope.json"), testAdapters); err == nil {
		t.Fatal("want error for a missing base file")
	}
}
