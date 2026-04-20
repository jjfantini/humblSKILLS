// Package env loads .env files for CLI startup. Kept dead-simple: walk up
// from CWD looking for a .env (stop at the git root or filesystem root),
// parse KEY=VALUE lines, and set only env vars that are NOT already
// present in the real environment. Env always wins.
//
// Silent by default so users who drop a .env in their repo root don't
// see noise; the verbose path emits a single "loaded N key(s) from
// <path>" line so operators can audit what came from disk vs env.
package env

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// LoadResult reports what LoadDotEnv did.
type LoadResult struct {
	Path   string   // absolute path of the .env that was loaded, or ""
	Loaded []string // keys that were set (env was not already set)
	Kept   []string // keys in the file but left alone (already set in env)
}

// LoadDotEnv walks up from startDir to find .env. Returns a zero result if
// none is found - not an error. The caller decides whether to log.
func LoadDotEnv(startDir string) (LoadResult, error) {
	if startDir == "" {
		d, err := os.Getwd()
		if err != nil {
			return LoadResult{}, err
		}
		startDir = d
	}
	path, err := findDotEnv(startDir)
	if err != nil || path == "" {
		return LoadResult{}, err
	}
	pairs, err := parse(path)
	if err != nil {
		return LoadResult{Path: path}, err
	}
	res := LoadResult{Path: path}
	for k, v := range pairs {
		if _, present := os.LookupEnv(k); present {
			res.Kept = append(res.Kept, k)
			continue
		}
		if err := os.Setenv(k, v); err != nil {
			return res, fmt.Errorf("set %s: %w", k, err)
		}
		res.Loaded = append(res.Loaded, k)
	}
	return res, nil
}

// findDotEnv walks up until it finds a .env or hits a git boundary /
// filesystem root. Returns "" with nil err when nothing found.
func findDotEnv(startDir string) (string, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for dir := abs; ; {
		candidate := filepath.Join(dir, ".env")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		// Stop at git root if we pass one - common ergonomic boundary.
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return "", nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

// parse reads a .env file into a map. Rules:
// - blank lines and `#` comments are skipped
// - lines with no `=` are skipped
// - leading `export ` is stripped
// - surrounding single or double quotes on the value are stripped
func parse(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]string{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		i := strings.Index(line, "=")
		if i <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:i])
		val := strings.TrimSpace(line[i+1:])
		val = strings.TrimSuffix(strings.TrimPrefix(val, `"`), `"`)
		val = strings.TrimSuffix(strings.TrimPrefix(val, "'"), "'")
		if key != "" {
			out[key] = val
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
