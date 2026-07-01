package adapters

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DetectResult reports the outcome of evaluating one adapter's detect rules
// against the current environment.
type DetectResult struct {
	Name     string
	Detected bool
	Reason   string
}

// Detect evaluates every adapter's detect rules and returns one result per
// adapter, in the same order.
func Detect(adapters []Adapter) []DetectResult {
	out := make([]DetectResult, 0, len(adapters))
	for _, a := range adapters {
		ok, reason := evalRules(a.Detect)
		out = append(out, DetectResult{Name: a.Name, Detected: ok, Reason: reason})
	}
	return out
}

func evalRules(d DetectRules) (bool, string) {
	switch {
	case len(d.AllOf) > 0:
		for _, r := range d.AllOf {
			if !evalRule(r) {
				return false, describeMiss(r)
			}
		}
		return true, "all checks passed"
	case len(d.AnyOf) > 0:
		for _, r := range d.AnyOf {
			if evalRule(r) {
				return true, describeMatch(r)
			}
		}
		return false, "no matching paths or env vars found"
	default:
		return false, "no detection rules configured"
	}
}

func evalRule(r DetectRule) bool {
	if r.PathExists != "" {
		p := ExpandPath(r.PathExists)
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	if r.Env != "" {
		if _, ok := os.LookupEnv(r.Env); ok {
			return true
		}
	}
	return false
}

// describeMatch phrases why a rule matched, in plain language suitable for the
// doctor "reason" row (e.g. "found ~/.claude", "$CURSOR_TRACE_ID is set").
func describeMatch(r DetectRule) string {
	switch {
	case r.PathExists != "" && r.Env != "":
		return fmt.Sprintf("found %s or $%s set", r.PathExists, r.Env)
	case r.PathExists != "":
		return "found " + r.PathExists
	case r.Env != "":
		return "$" + r.Env + " is set"
	default:
		return "matched an empty rule"
	}
}

// describeMiss phrases why a rule did not match, mirroring describeMatch (e.g.
// "~/.cursor not found", "$CURSOR_TRACE_ID not set").
func describeMiss(r DetectRule) string {
	switch {
	case r.PathExists != "" && r.Env != "":
		return fmt.Sprintf("%s not found and $%s not set", r.PathExists, r.Env)
	case r.PathExists != "":
		return r.PathExists + " not found"
	case r.Env != "":
		return "$" + r.Env + " not set"
	default:
		return "empty detection rule"
	}
}

// ExpandPath resolves a leading ~ and expands $VAR / ${VAR} references. It
// returns the input unchanged if no home directory is available.
func ExpandPath(p string) string {
	if p == "" {
		return p
	}
	if strings.HasPrefix(p, "~/") || p == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			if p == "~" {
				p = home
			} else {
				p = filepath.Join(home, p[2:])
			}
		}
	}
	return os.ExpandEnv(p)
}
