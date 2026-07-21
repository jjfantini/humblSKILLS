// Package fetch downloads skill source from GitHub as pinned tarballs and
// extracts individual skill directories into staging paths.
package fetch

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// DefaultHTTPTimeout is the HTTP request timeout for tarball downloads.
const DefaultHTTPTimeout = 60 * time.Second

// Fetcher downloads GitHub tarballs and caches them on disk, keyed by SHA so
// the cache entry is immutable and can be reused indefinitely.
type Fetcher struct {
	CacheDir string
	HTTP     *http.Client
	// Token, when non-empty, is sent as a Bearer Authorization header so skill
	// content can be pulled from a private repo's codeload endpoint.
	Token string
}

// NewFetcher returns a Fetcher with sensible defaults. cacheDir should be the
// humblskills-specific cache directory (e.g. $XDG_CACHE_HOME/humblskills).
func NewFetcher(cacheDir string) *Fetcher {
	return &Fetcher{
		CacheDir: cacheDir,
		HTTP:     &http.Client{Timeout: DefaultHTTPTimeout},
	}
}

// Fetch ensures the tarball for repo@sha is present on disk and returns its
// path. repo may be fully qualified ("github.com/owner/name") or bare
// ("owner/name"). Only GitHub is supported in v0.1.
func (f *Fetcher) Fetch(repo, sha string) (string, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return "", err
	}
	if sha == "" {
		return "", errors.New("fetch: empty sha")
	}

	dest := f.tarPath(owner, name, sha)
	if _, err := os.Stat(dest); err == nil {
		return dest, nil
	}

	url := fmt.Sprintf("https://codeload.github.com/%s/%s/tar.gz/%s", owner, name, sha)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", fmt.Errorf("create tar cache dir: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "humblskills-cli")
	if f.Token != "" {
		req.Header.Set("Authorization", "Bearer "+f.Token)
	}

	resp, err := f.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetch %s: HTTP %d", url, resp.StatusCode)
	}

	tmp := dest + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return "", fmt.Errorf("create tmp tar: %w", err)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		_ = out.Close()
		_ = os.Remove(tmp)
		return "", fmt.Errorf("write tmp tar: %w", err)
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("finalise tar: %w", err)
	}
	return dest, nil
}

// Extract unpacks files under skillPath (repo-relative, e.g. "skills/foo")
// from the gzipped tarball at tarPath into destDir. The top-level directory
// added by GitHub ("{owner}-{repo}-{short_sha}/") is stripped automatically.
// Symlinks and entries escaping destDir are rejected.
func Extract(tarPath, skillPath, destDir string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("open tarball: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip open: %w", err)
	}
	defer gz.Close()

	skillPath = path.Clean(strings.TrimSuffix(skillPath, "/"))
	if skillPath == "" || skillPath == "." {
		return errors.New("extract: empty skill path")
	}

	destDirAbs, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("abs dest: %w", err)
	}
	if err := os.MkdirAll(destDirAbs, 0o755); err != nil {
		return fmt.Errorf("create dest: %w", err)
	}

	tr := tar.NewReader(gz)
	prefix := ""
	matched := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read: %w", err)
		}
		// Skip PAX extended headers, which don't carry real files but do
		// sometimes appear first in GitHub-generated tarballs and would
		// otherwise poison our top-level-prefix detection.
		if hdr.Typeflag == tar.TypeXGlobalHeader || hdr.Typeflag == tar.TypeXHeader {
			continue
		}
		name := path.Clean(hdr.Name)
		if name == "." || strings.HasPrefix(name, "..") {
			return fmt.Errorf("refusing unsafe tar entry %q", hdr.Name)
		}
		if prefix == "" {
			i := strings.IndexByte(name, '/')
			if i < 0 {
				prefix = name
			} else {
				prefix = name[:i]
			}
		}
		target := prefix + "/" + skillPath
		rel, ok := trimPrefix(name, target)
		if !ok {
			continue
		}

		outPath := filepath.Join(destDirAbs, filepath.FromSlash(rel))
		if !strings.HasPrefix(outPath, destDirAbs+string(filepath.Separator)) && outPath != destDirAbs {
			return fmt.Errorf("refusing path traversal: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(outPath, os.FileMode(hdr.Mode)&0o777|0o700); err != nil {
				return fmt.Errorf("mkdir %s: %w", outPath, err)
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
				return fmt.Errorf("mkdir parent %s: %w", outPath, err)
			}
			if err := writeFile(outPath, tr, os.FileMode(hdr.Mode)&0o777); err != nil {
				return err
			}
			matched++
		case tar.TypeSymlink, tar.TypeLink:
			return fmt.Errorf("symlink/hardlink not supported in skill tree: %s", hdr.Name)
		default:
			// Skip other entry types (char device, fifo, ...).
		}
	}
	if matched == 0 {
		return fmt.Errorf("no files found under %q in tarball", skillPath)
	}
	return nil
}

// trimPrefix returns the portion of name after prefix + "/", or reports false
// if name isn't under prefix. prefix is treated as a directory path.
func trimPrefix(name, prefix string) (string, bool) {
	if name == prefix {
		return "", false // directory entry for the skill root itself
	}
	if strings.HasPrefix(name, prefix+"/") {
		return name[len(prefix)+1:], true
	}
	return "", false
}

func writeFile(path string, r io.Reader, mode os.FileMode) error {
	if mode == 0 {
		mode = 0o644
	}
	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	if _, err := io.Copy(out, r); err != nil {
		_ = out.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	return out.Close()
}

func (f *Fetcher) tarPath(owner, name, sha string) string {
	fname := fmt.Sprintf("%s-%s-%s.tar.gz", owner, name, sha)
	return filepath.Join(f.CacheDir, "tars", fname)
}

func splitRepo(repo string) (owner, name string, err error) {
	r := strings.TrimSpace(repo)
	r = strings.TrimPrefix(r, "https://")
	r = strings.TrimPrefix(r, "http://")
	r = strings.TrimPrefix(r, "github.com/")
	r = strings.TrimSuffix(r, ".git")
	parts := strings.Split(r, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unsupported repo %q (expected github.com/owner/name)", repo)
	}
	return parts[0], parts[1], nil
}
