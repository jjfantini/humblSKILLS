package fetch

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

// makeTarball writes a gzipped tarball that mimics the layout of a
// GitHub-generated archive: a single top-level directory containing the repo
// contents.
func makeTarball(t *testing.T, prefix string, files map[string]string) string {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	addDir := func(name string) {
		if err := tw.WriteHeader(&tar.Header{
			Name:     name + "/",
			Typeflag: tar.TypeDir,
			Mode:     0o755,
		}); err != nil {
			t.Fatalf("tar dir %s: %v", name, err)
		}
	}
	addDir(prefix)
	for p, body := range files {
		full := prefix + "/" + p
		if err := tw.WriteHeader(&tar.Header{
			Name:     full,
			Typeflag: tar.TypeReg,
			Mode:     0o644,
			Size:     int64(len(body)),
		}); err != nil {
			t.Fatalf("tar hdr %s: %v", p, err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatalf("tar body %s: %v", p, err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "archive.tar.gz")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtract_BasicSkill(t *testing.T) {
	tarPath := makeTarball(t, "jj-humblSKILLS-abc1234", map[string]string{
		"README.md":                           "repo readme",
		"skills/foo/SKILL.md":                 "# foo",
		"skills/foo/nested/helper.py":         "print('hi')",
		"skills/bar/SKILL.md":                 "# bar",
	})

	dest := filepath.Join(t.TempDir(), "out")
	if err := Extract(tarPath, "skills/foo", dest); err != nil {
		t.Fatalf("extract: %v", err)
	}

	skill, err := os.ReadFile(filepath.Join(dest, "SKILL.md"))
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}
	if string(skill) != "# foo" {
		t.Errorf("got %q", skill)
	}
	helper, err := os.ReadFile(filepath.Join(dest, "nested", "helper.py"))
	if err != nil {
		t.Fatalf("read helper: %v", err)
	}
	if string(helper) != "print('hi')" {
		t.Errorf("got %q", helper)
	}

	// bar shouldn't leak into foo's dest.
	if _, err := os.Stat(filepath.Join(dest, "..", "bar")); err == nil {
		t.Error("bar leaked outside dest")
	}
}

func TestExtract_SkipsPaxGlobalHeader(t *testing.T) {
	// GitHub's codeload tarballs prefix the archive with a pax_global_header
	// entry. The extractor must not treat it as the top-level directory.
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	_ = tw.WriteHeader(&tar.Header{Name: "pax_global_header", Typeflag: tar.TypeXGlobalHeader, Size: 0})
	_ = tw.WriteHeader(&tar.Header{Name: "repo-abc/", Typeflag: tar.TypeDir, Mode: 0o755})
	body := "# hi\n"
	_ = tw.WriteHeader(&tar.Header{Name: "repo-abc/skills/foo/SKILL.md", Typeflag: tar.TypeReg, Mode: 0o644, Size: int64(len(body))})
	_, _ = tw.Write([]byte(body))
	_ = tw.Close()
	_ = gz.Close()

	tarPath := filepath.Join(t.TempDir(), "with-pax.tar.gz")
	if err := os.WriteFile(tarPath, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "out")
	if err := Extract(tarPath, "skills/foo", dest); err != nil {
		t.Fatalf("extract: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dest, "SKILL.md"))
	if err != nil || string(b) != body {
		t.Errorf("read SKILL.md: %v %q", err, b)
	}
}

func TestExtract_UnknownSkill(t *testing.T) {
	tarPath := makeTarball(t, "x-repo-sha", map[string]string{
		"skills/foo/SKILL.md": "# foo",
	})
	err := Extract(tarPath, "skills/ghost", filepath.Join(t.TempDir(), "out"))
	if err == nil {
		t.Fatal("expected error for missing skill")
	}
}

func TestExtract_RejectsSymlink(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	_ = tw.WriteHeader(&tar.Header{Name: "pfx/", Typeflag: tar.TypeDir, Mode: 0o755})
	_ = tw.WriteHeader(&tar.Header{
		Name:     "pfx/skills/foo/link",
		Typeflag: tar.TypeSymlink,
		Linkname: "../../../etc/passwd",
	})
	_ = tw.Close()
	_ = gz.Close()

	tarPath := filepath.Join(t.TempDir(), "evil.tar.gz")
	if err := os.WriteFile(tarPath, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Extract(tarPath, "skills/foo", filepath.Join(t.TempDir(), "out")); err == nil {
		t.Fatal("expected symlink rejection")
	}
}

func TestSplitRepo(t *testing.T) {
	cases := map[string][2]string{
		"github.com/jjfantini/humblSKILLS":         {"jjfantini", "humblSKILLS"},
		"jjfantini/humblSKILLS":                    {"jjfantini", "humblSKILLS"},
		"https://github.com/jjfantini/humblSKILLS": {"jjfantini", "humblSKILLS"},
		"github.com/jjfantini/humblSKILLS.git":     {"jjfantini", "humblSKILLS"},
	}
	for in, want := range cases {
		o, n, err := splitRepo(in)
		if err != nil {
			t.Errorf("%s: %v", in, err)
			continue
		}
		if o != want[0] || n != want[1] {
			t.Errorf("%s: got %s/%s", in, o, n)
		}
	}
	if _, _, err := splitRepo("invalid"); err == nil {
		t.Error("expected error for bad repo")
	}
}
