package skillset

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "humblskills.json")

	s := New()
	s.Add("beta", "2.0.0")
	s.Add("alpha", "1.0.0")
	if err := Save(path, s); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != SchemaVersion {
		t.Errorf("schema = %d", got.SchemaVersion)
	}
	// Save sorts, so alpha comes first.
	if len(got.Skills) != 2 || got.Skills[0].Name != "alpha" || got.Skills[1].Name != "beta" {
		t.Fatalf("unexpected skills: %+v", got.Skills)
	}
	if got.Skills[0].Version != "1.0.0" {
		t.Errorf("alpha version = %q", got.Skills[0].Version)
	}
}

func TestAdd_DedupesLastWins(t *testing.T) {
	s := New()
	s.Add("foo", "1.0.0")
	s.Add("foo", "1.0.1")
	s.Add("  ", "x") // blank name ignored
	if len(s.Skills) != 1 {
		t.Fatalf("skills = %d, want 1", len(s.Skills))
	}
	if s.Skills[0].Version != "1.0.1" {
		t.Errorf("version = %q, want 1.0.1", s.Skills[0].Version)
	}
}

func TestLoad_DefaultsSchemaZero(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.json")
	if err := os.WriteFile(path, []byte(`{"skills":[{"name":"foo"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("minimal file should load: %v", err)
	}
	if got.SchemaVersion != SchemaVersion || len(got.Skills) != 1 {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestValidate_Errors(t *testing.T) {
	cases := map[string]*Set{
		"bad schema": {SchemaVersion: 999, Skills: []Skill{{Name: "a"}}},
		"empty name": {SchemaVersion: SchemaVersion, Skills: []Skill{{Name: "  "}}},
		"duplicate":  {SchemaVersion: SchemaVersion, Skills: []Skill{{Name: "a"}, {Name: "a"}}},
	}
	for name, s := range cases {
		if err := s.Validate(); err == nil {
			t.Errorf("%s: expected validation error", name)
		}
	}
}

func TestNames(t *testing.T) {
	s := New()
	s.Add("a", "")
	s.Add("b", "")
	names := s.Names()
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("names = %v", names)
	}
}

func TestIsRemote(t *testing.T) {
	cases := map[string]bool{
		"http://example.com/s.json":  true,
		"https://example.com/s.json": true,
		"file:///tmp/s.json":         false,
		"./humblskills.json":         false,
		"/abs/path/s.json":           false,
	}
	for src, want := range cases {
		if got := isRemote(src); got != want {
			t.Errorf("isRemote(%q) = %v, want %v", src, got, want)
		}
	}
}

func TestFileURLPath(t *testing.T) {
	cases := map[string]string{
		"file:///tmp/s.json":     "/tmp/s.json",
		"file:///abs/humbl.json": "/abs/humbl.json",
		"file://relative/s.json": "/s.json", // host="relative", path="/s.json"
		"file://" + "/only/path": "/only/path",
	}
	for in, want := range cases {
		if got := fileURLPath(in); got != want {
			t.Errorf("fileURLPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLoadFrom_LocalPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "humblskills.json")
	if err := os.WriteFile(path, []byte(`{"skills":[{"name":"foo"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom(local): %v", err)
	}
	if len(got.Skills) != 1 || got.Skills[0].Name != "foo" {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestLoadFrom_FileURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "humblskills.json")
	if err := os.WriteFile(path, []byte(`{"skills":[{"name":"bar"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFrom("file://" + path)
	if err != nil {
		t.Fatalf("LoadFrom(file url): %v", err)
	}
	if len(got.Skills) != 1 || got.Skills[0].Name != "bar" {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestLoadFrom_RemoteHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"schema_version":1,"skills":[{"name":"remote-skill","version":"1.2.3"}]}`))
	}))
	defer srv.Close()

	got, err := LoadFrom(srv.URL + "/humblskills.json")
	if err != nil {
		t.Fatalf("LoadFrom(http): %v", err)
	}
	if len(got.Skills) != 1 || got.Skills[0].Name != "remote-skill" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestLoadFrom_RemoteNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer srv.Close()

	if _, err := LoadFrom(srv.URL + "/missing.json"); err == nil {
		t.Fatal("expected error for HTTP 404")
	}
}

func TestLoadFrom_RemoteBadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	if _, err := LoadFrom(srv.URL + "/bad.json"); err == nil {
		t.Fatal("expected parse error for invalid JSON")
	}
}

func TestLoadFrom_RemoteValidatesSchema(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"schema_version":999,"skills":[{"name":"x"}]}`))
	}))
	defer srv.Close()

	if _, err := LoadFrom(srv.URL + "/s.json"); err == nil {
		t.Fatal("expected validation error for bad schema_version")
	}
}
