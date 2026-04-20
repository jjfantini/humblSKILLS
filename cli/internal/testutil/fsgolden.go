package testutil

import (
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// updateGolden is flipped by `go test -update`. When true, failing
// AssertGoldenDir / AssertGoldenFile calls rewrite the golden data
// rather than failing. The flag is registered once at package init.
var updateGolden = flag.Bool("update", false, "update testutil golden files instead of comparing")

// UpdateRequested reports whether -update was passed. Useful when
// individual test packages want to key custom regeneration logic off
// the same flag.
func UpdateRequested() bool { return updateGolden != nil && *updateGolden }

// SnapshotDir walks root and returns a deterministic map from
// relative slash-paths to file contents. Directories are represented
// as keys ending in "/" with empty value. Symlinks fail the test
// (the production code refuses to emit them).
func SnapshotDir(t testing.TB, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(root, func(p string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if fi.Mode()&os.ModeSymlink != 0 {
			t.Fatalf("snapshot: unexpected symlink at %s", p)
		}
		if fi.IsDir() {
			out[rel+"/"] = ""
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		out[rel] = string(data)
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", root, err)
	}
	return out
}

// AssertGoldenDir diffs the on-disk tree at actualDir against the
// golden tree at goldenDir. When `-update` is set, goldenDir is
// rewritten from actualDir. Otherwise any mismatch fails the test
// with a minimal diff pointing at the first offending path.
func AssertGoldenDir(t testing.TB, actualDir, goldenDir string) {
	t.Helper()
	actual := SnapshotDir(t, actualDir)

	if UpdateRequested() {
		if err := os.RemoveAll(goldenDir); err != nil {
			t.Fatalf("clean golden %s: %v", goldenDir, err)
		}
		for rel, body := range actual {
			dst := filepath.Join(goldenDir, filepath.FromSlash(rel))
			if strings.HasSuffix(rel, "/") {
				if err := os.MkdirAll(dst, 0o755); err != nil {
					t.Fatalf("mkdir %s: %v", dst, err)
				}
				continue
			}
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				t.Fatalf("mkdir %s: %v", filepath.Dir(dst), err)
			}
			if err := os.WriteFile(dst, []byte(body), 0o644); err != nil {
				t.Fatalf("write %s: %v", dst, err)
			}
		}
		return
	}

	expected := SnapshotDir(t, goldenDir)
	diff := diffSnapshots(expected, actual)
	if diff != "" {
		t.Fatalf("%s does not match %s:\n%s\n(rerun with -update to regenerate)", actualDir, goldenDir, diff)
	}
}

// AssertGoldenBytes compares data against the contents of goldenPath.
// With -update, goldenPath is rewritten.
func AssertGoldenBytes(t testing.TB, data []byte, goldenPath string) {
	t.Helper()
	if UpdateRequested() {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(goldenPath), err)
		}
		if err := os.WriteFile(goldenPath, data, 0o644); err != nil {
			t.Fatalf("write %s: %v", goldenPath, err)
		}
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v (rerun with -update to create)", goldenPath, err)
	}
	if string(want) != string(data) {
		t.Fatalf("%s mismatch\n--- want ---\n%s\n--- got ---\n%s",
			goldenPath, truncate(string(want)), truncate(string(data)))
	}
}

// diffSnapshots returns "" when the maps are equal, else a short
// description of the first differing key (missing on either side or
// mismatching body).
func diffSnapshots(want, got map[string]string) string {
	var keys []string
	seen := map[string]struct{}{}
	for k := range want {
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			keys = append(keys, k)
		}
	}
	for k := range got {
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		w, okW := want[k]
		g, okG := got[k]
		switch {
		case !okW:
			b.WriteString("  + unexpected: " + k + "\n")
		case !okG:
			b.WriteString("  - missing:    " + k + "\n")
		case w != g:
			b.WriteString("  ~ differs:    " + k + "\n")
			b.WriteString("    want: " + truncate(w) + "\n")
			b.WriteString("    got:  " + truncate(g) + "\n")
		}
	}
	return b.String()
}

func truncate(s string) string {
	const max = 200
	s = strings.ReplaceAll(s, "\n", "\\n")
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}
