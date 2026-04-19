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
				return false, "all_of failed at: " + describeRule(r)
			}
		}
		return true, "all_of rules matched"
	case len(d.AnyOf) > 0:
		for _, r := range d.AnyOf {
			if evalRule(r) {
				return true, "matched: " + describeRule(r)
			}
		}
		return false, "no any_of rule matched"
	default:
		return false, "no detect rules declared"
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

func describeRule(r DetectRule) string {
	switch {
	case r.PathExists != "" && r.Env != "":
		return fmt.Sprintf("path_exists=%s env=%s", r.PathExists, r.Env)
	case r.PathExists != "":
		return "path_exists=" + r.PathExists
	case r.Env != "":
		return "env=" + r.Env
	default:
		return "<empty rule>"
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
