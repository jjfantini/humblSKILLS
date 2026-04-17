package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func testRegistryBody(t *testing.T) []byte {
	t.Helper()
	r := Registry{
		SchemaVersion: SchemaVersion,
		Skills: []Skill{
			{Name: "a", Version: "0.1.0", Path: "skills/a"},
		},
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestLoad_FromNetworkAndCache(t *testing.T) {
	body := testRegistryBody(t)

	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Write(body)
	}))
	defer srv.Close()

	cache := t.TempDir()
	f := NewFetcher(srv.URL, cache)

	_, src, err := f.Load()
	if err != nil {
		t.Fatal(err)
	}
	if src != OriginNetwork {
		t.Errorf("first load source: got %q, want network", src)
	}

	_, src, err = f.Load()
	if err != nil {
		t.Fatal(err)
	}
	if src != OriginCache {
		t.Errorf("second load source: got %q, want cache", src)
	}
	if atomic.LoadInt32(&hits) != 1 {
		t.Errorf("expected 1 HTTP hit, got %d", hits)
	}
}

func TestLoad_CacheMiss_WhenTTLExpired(t *testing.T) {
	body := testRegistryBody(t)

	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Write(body)
	}))
	defer srv.Close()

	cache := t.TempDir()
	clock := time.Now()
	f := NewFetcher(srv.URL, cache)
	f.TTL = time.Minute
	f.Now = func() time.Time { return clock }

	_, _, _ = f.Load()
	// Advance past TTL.
	clock = clock.Add(2 * time.Minute)
	_, src, err := f.Load()
	if err != nil {
		t.Fatal(err)
	}
	if src != OriginNetwork {
		t.Errorf("stale cache should go to network, got %q", src)
	}
	if atomic.LoadInt32(&hits) != 2 {
		t.Errorf("expected 2 HTTP hits, got %d", hits)
	}
}

func TestLoad_CacheInvalidatedWhenURLChanges(t *testing.T) {
	body := testRegistryBody(t)
	srvA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.Write(body) }))
	defer srvA.Close()
	srvB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.Write(body) }))
	defer srvB.Close()

	cache := t.TempDir()

	f := NewFetcher(srvA.URL, cache)
	if _, _, err := f.Load(); err != nil {
		t.Fatal(err)
	}

	f2 := NewFetcher(srvB.URL, cache)
	_, src, err := f2.Load()
	if err != nil {
		t.Fatal(err)
	}
	if src != OriginNetwork {
		t.Errorf("URL change should invalidate cache, got source %q", src)
	}
}

func TestRefresh_IgnoresTTL(t *testing.T) {
	body := testRegistryBody(t)
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Write(body)
	}))
	defer srv.Close()

	cache := t.TempDir()
	f := NewFetcher(srv.URL, cache)
	_, _, _ = f.Load()
	_, src, err := f.Refresh()
	if err != nil {
		t.Fatal(err)
	}
	if src != OriginNetwork {
		t.Errorf("refresh should be network, got %q", src)
	}
	if atomic.LoadInt32(&hits) != 2 {
		t.Errorf("expected 2 hits, got %d", hits)
	}
}

func TestLoad_FileURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	if err := os.WriteFile(path, testRegistryBody(t), 0o644); err != nil {
		t.Fatal(err)
	}

	f := NewFetcher("file://"+path, t.TempDir())
	r, src, err := f.Load()
	if err != nil {
		t.Fatal(err)
	}
	if src != OriginFile {
		t.Errorf("got source %q", src)
	}
	if len(r.Skills) != 1 || r.Skills[0].Name != "a" {
		t.Errorf("got %+v", r.Skills)
	}
}

func TestLoad_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer srv.Close()

	f := NewFetcher(srv.URL, t.TempDir())
	if _, _, err := f.Load(); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_RejectsWrongSchemaVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"schema_version": 999, "skills": []}`))
	}))
	defer srv.Close()

	f := NewFetcher(srv.URL, t.TempDir())
	if _, _, err := f.Load(); err == nil {
		t.Fatal("expected schema_version error")
	}
}

func TestInspect_NoCacheYet(t *testing.T) {
	f := NewFetcher("https://example.invalid/registry.json", t.TempDir())
	info := f.Inspect()
	if info.Exists {
		t.Errorf("expected Exists=false, got %+v", info)
	}
}
