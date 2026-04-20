// Package scenarios parses and validates per-skill eval specs.
//
// Two file formats are supported:
//
//  1. evals.json (Anthropic-standard). One "eval" entry per test case with a
//     single prompt and assertions. Backward-compatible: we lift each entry
//     into a one-session Scenario under the default "generic" family.
//  2. scenarios.json (humblSKILLS extension). Ordered sessions, retention
//     checks, transfer links between scenarios - the longitudinal arm.
//
// Assertion kinds, matching the plan:
//   - llm                  : default, sent to grader LLM
//   - path_exists:<rel>    : path exists under OutputDir (or cwd)
//   - exec:<cmd>           : shell command exits 0
//   - script:<rel>         : runs evals/assertions/<rel>
//   - regex:<rel>:<pattern>: file matches regex
//   - json_valid:<rel>     : file parses as JSON
package scenarios

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SchemaVersion is the current scenarios.json schema.
const SchemaVersion = 1

// Configuration IDs (arms).
const (
	ArmSmartSkill = "smart_skill"
	ArmFlatSkill  = "flat_skill"
	ArmNoSkill    = "no_skill"
)

// File is the top-level scenarios.json (or lifted evals.json) document.
type File struct {
	SkillName            string     `json:"skill_name"`
	SchemaVersion        int        `json:"schema_version"`
	Configurations       []string   `json:"configurations,omitempty"`
	RunsPerConfiguration int        `json:"runs_per_configuration,omitempty"`
	Scenarios            []Scenario `json:"scenarios"`
}

// Scenario is an ordered sequence of sessions plus family / transfer metadata.
type Scenario struct {
	ID           string    `json:"id"`
	Family       string    `json:"family,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
	TransferFrom []string  `json:"transfer_from,omitempty"`
	Sessions     []Session `json:"sessions"`
}

// Session is one prompt in a scenario. The `n` index is 1-based.
type Session struct {
	N              int           `json:"n"`
	Prompt         string        `json:"prompt"`
	Files          []string      `json:"files,omitempty"`
	RetentionCheck string        `json:"retention_check,omitempty"`
	Assertions     []Assertion   `json:"assertions,omitempty"`
	ExpectedOutput string        `json:"expected_output,omitempty"`
	Timeout        time.Duration `json:"-"` // parsed from TimeoutSec
	TimeoutSec     int           `json:"timeout_seconds,omitempty"`
}

// Assertion is one verifiable statement about a session's output. The Check
// string is the kind + optional argument, e.g. "llm", "path_exists:foo/bar",
// "regex:out.md:^Hello".
type Assertion struct {
	Text  string `json:"text"`
	Check string `json:"check,omitempty"`
}

// Load reads a scenarios.json or falls back to evals.json in the same
// directory, returning a validated File. Missing files return an error so
// callers can present "no evals configured" in the TUI.
func Load(evalsDir string) (*File, error) {
	sp := filepath.Join(evalsDir, "scenarios.json")
	ep := filepath.Join(evalsDir, "evals.json")
	if _, err := os.Stat(sp); err == nil {
		return loadScenarios(sp)
	}
	if _, err := os.Stat(ep); err == nil {
		return loadEvalsJSON(ep)
	}
	return nil, fmt.Errorf("no evals configured at %s (expected scenarios.json or evals.json)", evalsDir)
}

// LoadFromSkill finds the evals directory under a skill path and calls Load.
func LoadFromSkill(skillDir string) (*File, error) {
	return Load(filepath.Join(skillDir, "evals"))
}

func loadScenarios(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scenarios.json: %w", err)
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse scenarios.json: %w", err)
	}
	applyDefaults(&f)
	if err := Validate(&f); err != nil {
		return nil, err
	}
	return &f, nil
}

// LegacyEval matches the Anthropic-standard evals.json shape.
type LegacyEval struct {
	ID             int         `json:"id"`
	Prompt         string      `json:"prompt"`
	ExpectedOutput string      `json:"expected_output"`
	Files          []string    `json:"files,omitempty"`
	Assertions     []Assertion `json:"assertions,omitempty"`
}

type legacyFile struct {
	SkillName string       `json:"skill_name"`
	Evals     []LegacyEval `json:"evals"`
}

func loadEvalsJSON(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read evals.json: %w", err)
	}
	var lf legacyFile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("parse evals.json: %w", err)
	}
	f := &File{
		SkillName:     lf.SkillName,
		SchemaVersion: SchemaVersion,
	}
	for _, e := range lf.Evals {
		f.Scenarios = append(f.Scenarios, Scenario{
			ID:     fmt.Sprintf("eval-%d", e.ID),
			Family: "generic",
			Sessions: []Session{{
				N:              1,
				Prompt:         e.Prompt,
				Files:          e.Files,
				ExpectedOutput: e.ExpectedOutput,
				Assertions:     e.Assertions,
			}},
		})
	}
	applyDefaults(f)
	if err := Validate(f); err != nil {
		return nil, err
	}
	return f, nil
}

func applyDefaults(f *File) {
	if f.SchemaVersion == 0 {
		f.SchemaVersion = SchemaVersion
	}
	if len(f.Configurations) == 0 {
		f.Configurations = []string{ArmSmartSkill, ArmFlatSkill, ArmNoSkill}
	}
	if f.RunsPerConfiguration <= 0 {
		f.RunsPerConfiguration = 1
	}
	for si := range f.Scenarios {
		s := &f.Scenarios[si]
		if s.Family == "" {
			s.Family = "generic"
		}
		for ii := range s.Sessions {
			sess := &s.Sessions[ii]
			if sess.TimeoutSec > 0 {
				sess.Timeout = time.Duration(sess.TimeoutSec) * time.Second
			}
			for ai := range sess.Assertions {
				a := &sess.Assertions[ai]
				if a.Check == "" {
					a.Check = "llm"
				}
			}
		}
	}
}

// Validate catches shape errors the parser alone won't: duplicate IDs,
// sessions with zero N or missing prompts, unknown configuration values,
// out-of-range session numbering.
func Validate(f *File) error {
	if f == nil {
		return errors.New("nil file")
	}
	if f.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %d (expected %d)", f.SchemaVersion, SchemaVersion)
	}
	if f.SkillName == "" {
		return errors.New("skill_name is required")
	}
	known := map[string]bool{
		ArmSmartSkill: true, ArmFlatSkill: true, ArmNoSkill: true,
	}
	for _, c := range f.Configurations {
		if !known[c] {
			return fmt.Errorf("unknown configuration %q (want one of smart_skill / flat_skill / no_skill)", c)
		}
	}
	ids := map[string]bool{}
	for i, s := range f.Scenarios {
		if s.ID == "" {
			return fmt.Errorf("scenarios[%d]: id is required", i)
		}
		if ids[s.ID] {
			return fmt.Errorf("duplicate scenario id %q", s.ID)
		}
		ids[s.ID] = true
		if len(s.Sessions) == 0 {
			return fmt.Errorf("scenario %s: at least one session required", s.ID)
		}
		for j, sess := range s.Sessions {
			want := j + 1
			if sess.N == 0 {
				s.Sessions[j].N = want
			} else if sess.N != want {
				return fmt.Errorf("scenario %s: session[%d].n=%d (expected %d - sessions must be 1..K in order)", s.ID, j, sess.N, want)
			}
			if strings.TrimSpace(sess.Prompt) == "" {
				return fmt.Errorf("scenario %s: session %d prompt is empty", s.ID, sess.N)
			}
			for _, a := range sess.Assertions {
				if a.Text == "" {
					return fmt.Errorf("scenario %s: session %d has an assertion with empty text", s.ID, sess.N)
				}
				if err := validateCheck(a.Check); err != nil {
					return fmt.Errorf("scenario %s: session %d: %w", s.ID, sess.N, err)
				}
			}
		}
	}
	// Validate transfer_from references point at known scenario IDs.
	for _, s := range f.Scenarios {
		for _, t := range s.TransferFrom {
			if !ids[t] {
				return fmt.Errorf("scenario %s: transfer_from %q is not a known scenario id", s.ID, t)
			}
		}
	}
	return nil
}

// CheckKind is the prefix before the first colon in an assertion's Check.
type CheckKind string

const (
	CheckLLM        CheckKind = "llm"
	CheckPathExists CheckKind = "path_exists"
	CheckExec       CheckKind = "exec"
	CheckScript     CheckKind = "script"
	CheckRegex      CheckKind = "regex"
	CheckJSONValid  CheckKind = "json_valid"
)

// ParseCheck splits an assertion Check into kind + argument. Returns
// CheckLLM with empty arg for the default.
func ParseCheck(check string) (CheckKind, string) {
	if check == "" {
		return CheckLLM, ""
	}
	i := strings.Index(check, ":")
	if i < 0 {
		return CheckKind(check), ""
	}
	return CheckKind(check[:i]), check[i+1:]
}

func validateCheck(check string) error {
	kind, _ := ParseCheck(check)
	switch kind {
	case CheckLLM, CheckPathExists, CheckExec, CheckScript, CheckRegex, CheckJSONValid:
		return nil
	default:
		return fmt.Errorf("unknown assertion kind %q", kind)
	}
}

// TotalSessions sums session counts across all scenarios - used by the TUI
// to size progress bars and by the harness to pre-allocate work units.
func (f *File) TotalSessions() int {
	n := 0
	for _, s := range f.Scenarios {
		n += len(s.Sessions)
	}
	return n
}

// FindScenario returns the scenario with the given ID or nil.
func (f *File) FindScenario(id string) *Scenario {
	for i := range f.Scenarios {
		if f.Scenarios[i].ID == id {
			return &f.Scenarios[i]
		}
	}
	return nil
}
