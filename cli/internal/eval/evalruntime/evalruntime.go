// Package evalruntime ties the six runners into one Registry. Placed in
// its own package so the runner/ subtree has no dependency on secrets
// (keeps runner.Runner reusable by third parties) while still giving the
// CLI a one-stop import for "give me every runner wired up".
package evalruntime

import (
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner/anthropicapi"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner/claudecode"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner/codex"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner/cursor"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner/mock"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner/openaiapi"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
)

// DefaultRegistry returns the canonical six-runner registry ordered as
// the plan specifies: CLI runners first (lowest friction) then API
// runners then mock. Auto-detect walks in order.
func DefaultRegistry(store secrets.Store) *runner.Registry {
	return runner.NewRegistry(
		claudecode.New(),
		cursor.New(),
		codex.New(),
		anthropicapi.New(store),
		openaiapi.New(store),
		mock.New(),
	)
}
