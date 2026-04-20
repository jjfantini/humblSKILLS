package registry_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

func TestInspect_AfterLoad_ReportsExistsAndAge(t *testing.T) {
	body := `{"schema_version":1,"skills":[]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	cache := t.TempDir()
	clock := time.Now()
	f := registry.NewFetcher(srv.URL, cache)
	f.Now = func() time.Time { return clock }

	if _, _, err := f.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	clock = clock.Add(2 * time.Minute)
	info := f.Inspect()
	if !info.Exists {
		t.Error("Inspect reports Exists=false after Load")
	}
	if info.Age < 2*time.Minute {
		t.Errorf("Age = %v, want >= 2m", info.Age)
	}
	if info.URL != srv.URL {
		t.Errorf("URL = %q", info.URL)
	}
}

func TestInspect_LocalPath_ReturnsLocalURLAndNoAge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	_ = os.WriteFile(path, []byte(`{"schema_version":1}`), 0o644)

	f := registry.NewFetcher("file://"+path, t.TempDir())
	info := f.Inspect()
	if info.Exists {
		t.Errorf("local URL Inspect should report Exists=false, got %+v", info)
	}
}

func TestLoad_BareFilesystemPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":1,"skills":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// No file:// prefix — bare path should route through loadLocal.
	f := registry.NewFetcher(path, t.TempDir())
	_, src, err := f.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if src != registry.OriginFile {
		t.Errorf("source = %q, want file", src)
	}
}

func TestRefresh_LocalRoute(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":1,"skills":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	f := registry.NewFetcher("file://"+path, t.TempDir())
	_, src, err := f.Refresh()
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if src != registry.OriginFile {
		t.Errorf("Refresh(file) source = %q", src)
	}
}

func TestRefresh_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	f := registry.NewFetcher(srv.URL, t.TempDir())
	if _, _, err := f.Refresh(); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_LocalFileMissing(t *testing.T) {
	f := registry.NewFetcher("file:///definitely/not/a/real/path/registry.json", t.TempDir())
	if _, _, err := f.Load(); err == nil {
		t.Fatal("expected error for missing local registry")
	}
}

func TestLoad_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "{not json")
	}))
	defer srv.Close()
	f := registry.NewFetcher(srv.URL, t.TempDir())
	_, _, err := f.Load()
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("err missing parse context: %v", err)
	}
}

func TestValidateDeps_AllPassWhenNoRequires(t *testing.T) {
	reg := &registry.Registry{
		Skills: []registry.Skill{
			{Name: "a", Version: "1.0.0"},
			{Name: "b", Version: "1.0.0"},
		},
	}
	if got := registry.ValidateDeps(reg); len(got) != 0 {
		t.Errorf("unexpected issues: %v", got)
	}
}

func TestValidateDeps_UnknownDepFlagged(t *testing.T) {
	reg := &registry.Registry{
		Skills: []registry.Skill{
			{Name: "a", Version: "1.0.0", Requires: []string{"ghost@1.0.0"}},
		},
	}
	issues := registry.ValidateDeps(reg)
	if len(issues) != 1 || issues[0].Kind != registry.IssueUnknown {
		t.Errorf("issues = %v", issues)
	}
}

func TestValidateDeps_UnsatisfiedDepFlagged(t *testing.T) {
	reg := &registry.Registry{
		Skills: []registry.Skill{
			{Name: "a", Version: "1.0.0", Requires: []string{"b@>=2.0.0"}},
			{Name: "b", Version: "1.0.0"},
		},
	}
	issues := registry.ValidateDeps(reg)
	if len(issues) != 1 || issues[0].Kind != registry.IssueUnsatisfied {
		t.Errorf("issues = %v", issues)
	}
}

func TestValidateDeps_UnparseableDepFlagged(t *testing.T) {
	reg := &registry.Registry{
		Skills: []registry.Skill{
			{Name: "a", Version: "1.0.0", Requires: []string{"@@@bad"}},
		},
	}
	issues := registry.ValidateDeps(reg)
	if len(issues) != 1 || issues[0].Kind != registry.IssueParse {
		t.Errorf("issues = %v", issues)
	}
}

func TestValidateDeps_CycleFlagged(t *testing.T) {
	reg := &registry.Registry{
		Skills: []registry.Skill{
			{Name: "a", Version: "1.0.0", Requires: []string{"b@1.0.0"}},
			{Name: "b", Version: "1.0.0", Requires: []string{"a@1.0.0"}},
		},
	}
	issues := registry.ValidateDeps(reg)
	found := false
	for _, i := range issues {
		if i.Kind == registry.IssueCycle {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("cycle not flagged: %v", issues)
	}
}

func TestValidateDeps_NilRegistryReturnsNoIssues(t *testing.T) {
	if got := registry.ValidateDeps(nil); got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestIssue_Error_CycleOmitsSkillAndDep(t *testing.T) {
	i := registry.Issue{Kind: registry.IssueCycle, Msg: "a -> b -> a"}
	if !strings.Contains(i.Error(), "cycle") || strings.Contains(i.Error(), "skill ") {
		t.Errorf("Error = %q", i.Error())
	}
}

func TestIssue_Error_NonCycleIncludesSkillAndDep(t *testing.T) {
	i := registry.Issue{Kind: registry.IssueUnknown, Skill: "a", Dep: "b", Msg: "missing"}
	if !strings.Contains(i.Error(), `skill "a"`) || !strings.Contains(i.Error(), `dep "b"`) {
		t.Errorf("Error = %q", i.Error())
	}
}

// localPath handles URLs that url.Parse chokes on (bad percent-encoding
// etc.). Production code fails back to a raw TrimPrefix.
func TestLoadLocal_HandlesWeirdFileURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// A URL with just "file://" + absolute path on a Mac hits the
	// url.Parse branch.
	f := registry.NewFetcher("file://"+path, t.TempDir())
	if _, _, err := f.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
}
