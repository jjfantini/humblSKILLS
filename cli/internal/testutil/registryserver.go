package testutil

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

// RegistryServer is an httptest.Server that serves a configurable
// registry.json. Tarballs are not served here (the production Fetcher
// hardcodes codeload.github.com); use SeedTarball to pre-populate the
// on-disk tarball cache instead.
type RegistryServer struct {
	t   testing.TB
	srv *httptest.Server

	mu       sync.Mutex
	body     []byte
	status   int
	requests int
}

// NewRegistryServer returns a started RegistryServer serving reg as the
// initial payload. Status defaults to 200. The server is registered
// with t.Cleanup and will be closed automatically.
func NewRegistryServer(t testing.TB, reg *registry.Registry) *RegistryServer {
	t.Helper()
	rs := &RegistryServer{t: t, status: http.StatusOK}
	rs.SetRegistry(reg)
	rs.srv = httptest.NewServer(http.HandlerFunc(rs.handle))
	t.Cleanup(rs.srv.Close)
	return rs
}

// URL returns the base URL callers should pass as --registry or
// HUMBLSKILLS_REGISTRY. It always maps to the registry.json endpoint.
func (rs *RegistryServer) URL() string { return rs.srv.URL + "/registry.json" }

// SetRegistry replaces the payload served on subsequent requests.
func (rs *RegistryServer) SetRegistry(reg *registry.Registry) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if reg == nil {
		rs.body = nil
		return
	}
	if reg.SchemaVersion == 0 {
		reg.SchemaVersion = registry.SchemaVersion
	}
	body, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		rs.t.Fatalf("marshal registry: %v", err)
	}
	rs.body = body
}

// SetStatus sets the HTTP status returned on subsequent requests. Used
// to simulate network failures (404/500) and cache fall-back behaviour.
func (rs *RegistryServer) SetStatus(code int) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.status = code
}

// Requests returns the number of requests served so far.
func (rs *RegistryServer) Requests() int {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.requests
}

// SetRawBody is an escape hatch for tests that want to exercise the
// parser with deliberately malformed JSON or a mismatched
// schema_version. Bypasses SetRegistry's marshalling.
func (rs *RegistryServer) SetRawBody(body []byte) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.body = body
}

func (rs *RegistryServer) handle(w http.ResponseWriter, r *http.Request) {
	rs.mu.Lock()
	status := rs.status
	body := append([]byte(nil), rs.body...)
	rs.requests++
	rs.mu.Unlock()

	if !strings.HasSuffix(r.URL.Path, "/registry.json") {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// ---- tarball seeding -------------------------------------------------------

// SkillTree is the in-memory layout for one skill. Keys are
// repo-relative paths *under* skillPath (e.g. "SKILL.md",
// "data/notes.md" — not "skills/foo/SKILL.md").
type SkillTree = map[string]string

// TarballSpec describes what SeedTarball should write.
type TarballSpec struct {
	// Owner / Name / SHA match registry.Source. The tarball is written
	// at cacheDir/tars/{owner}-{name}-{sha}.tar.gz.
	Owner, Name, SHA string

	// RepoPrefix is the top-level directory GitHub prepends to codeload
	// tarballs, e.g. "jjfantini-humblSKILLS-abc1234". If empty, a
	// deterministic value derived from Owner/Name/SHA is used.
	RepoPrefix string

	// Skills maps a repo-relative skill directory (e.g. "skills/foo")
	// to the files inside that directory. Extra files outside skill
	// roots can be added via ExtraFiles.
	Skills map[string]SkillTree

	// ExtraFiles writes additional repo-relative files into the
	// tarball, e.g. "README.md", "registry.json". Useful for
	// build-registry tests.
	ExtraFiles map[string]string
}

// SeedTarball writes a gzipped tarball mimicking a codeload archive
// into the fetch cache at cacheDir/tars/. Returns the tarball path.
func SeedTarball(t testing.TB, cacheDir string, spec TarballSpec) string {
	t.Helper()
	if spec.Owner == "" || spec.Name == "" || spec.SHA == "" {
		t.Fatalf("seed tarball: owner/name/sha required")
	}
	prefix := spec.RepoPrefix
	if prefix == "" {
		short := spec.SHA
		if len(short) > 7 {
			short = short[:7]
		}
		prefix = spec.Owner + "-" + spec.Name + "-" + short
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	addDir := func(name string) {
		if err := tw.WriteHeader(&tar.Header{
			Name: name + "/", Typeflag: tar.TypeDir, Mode: 0o755,
		}); err != nil {
			t.Fatalf("tar dir %s: %v", name, err)
		}
	}
	addFile := func(name, body string) {
		if err := tw.WriteHeader(&tar.Header{
			Name: name, Typeflag: tar.TypeReg, Mode: 0o644, Size: int64(len(body)),
		}); err != nil {
			t.Fatalf("tar hdr %s: %v", name, err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatalf("tar body %s: %v", name, err)
		}
	}

	addDir(prefix)

	// Emit in sorted order so tarball bytes are deterministic across
	// runs (useful for golden-file comparisons of DirSHA-adjacent code).
	var skillPaths []string
	for sp := range spec.Skills {
		skillPaths = append(skillPaths, sp)
	}
	sort.Strings(skillPaths)
	for _, sp := range skillPaths {
		var files []string
		for f := range spec.Skills[sp] {
			files = append(files, f)
		}
		sort.Strings(files)
		for _, f := range files {
			addFile(prefix+"/"+sp+"/"+f, spec.Skills[sp][f])
		}
	}

	var extras []string
	for f := range spec.ExtraFiles {
		extras = append(extras, f)
	}
	sort.Strings(extras)
	for _, f := range extras {
		addFile(prefix+"/"+f, spec.ExtraFiles[f])
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gz close: %v", err)
	}

	dir := filepath.Join(cacheDir, "tars")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	fname := spec.Owner + "-" + spec.Name + "-" + spec.SHA + ".tar.gz"
	tarPath := filepath.Join(dir, fname)
	if err := os.WriteFile(tarPath, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write %s: %v", tarPath, err)
	}
	return tarPath
}
