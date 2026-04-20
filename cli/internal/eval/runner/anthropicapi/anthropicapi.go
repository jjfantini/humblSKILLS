// Package anthropicapi is the pure-Go runner that speaks directly to the
// Anthropic Messages API. Implements the minimal Read/Write/Bash/Glob/Grep
// tool loop sandboxed to a scratch directory.
//
// Gated by secrets.GetAPIKey("anthropic") - that store resolves env >
// keyring > file, so users point an API key in whichever way is most
// convenient for their environment.
package anthropicapi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner/toolbox"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
)

// DefaultModel is the recommended Opus family model for evals.
const DefaultModel = "claude-opus-4-5"

// Pricing for the default model. Best-effort, overwritten if the caller
// supplies their own via Request.Model.
var defaultPricing = &runner.Pricing{
	PromptUSDPerMtok:     15,
	CompletionUSDPerMtok: 75,
}

// Runner is the anthropic-api backend.
type Runner struct {
	Store secrets.Store
}

// New returns a Runner. Caller passes the secrets store (the CLI always
// hands over the same one; tests can substitute an in-memory stub).
func New(store secrets.Store) *Runner { return &Runner{Store: store} }

func (r *Runner) Name() string { return "anthropic-api" }

func (r *Runner) Capabilities() runner.Capabilities {
	return runner.Capabilities{
		SupportsTools:    []string{"Read", "Write", "Bash", "Glob", "Grep"},
		SupportsParallel: true,
		DefaultModel:     DefaultModel,
		Pricing:          defaultPricing,
	}
}

func (r *Runner) DoctorCheck(ctx context.Context) runner.DoctorCheck {
	if r.Store == nil {
		return runner.DoctorCheck{
			Available: false,
			Reason:    "secrets store not configured",
			Fix:       "run `humblskills eval set-key anthropic`",
			RequiresKey: "anthropic",
		}
	}
	_, src, err := r.Store.Get("anthropic")
	if err != nil || src == secrets.SourceAbsent {
		return runner.DoctorCheck{
			Available:   false,
			Reason:      "ANTHROPIC_API_KEY not set",
			Fix:         "export ANTHROPIC_API_KEY=... or run `humblskills eval set-key anthropic`",
			RequiresKey: "anthropic",
		}
	}
	return runner.DoctorCheck{
		Available:   true,
		Version:     "anthropic-sdk-go v1.37.0",
		RequiresKey: "anthropic",
	}
}

// Execute runs the tool loop until the model emits end_turn. Up to 16
// tool-use rounds are allowed per request to bound runtime.
func (r *Runner) Execute(ctx context.Context, req runner.Request) (*runner.Result, error) {
	start := time.Now()
	key, _, err := r.Store.Get("anthropic")
	if err != nil || key == "" {
		return &runner.Result{Err: fmt.Errorf("no API key")}, fmt.Errorf("no API key")
	}
	client := anthropic.NewClient(option.WithAPIKey(key))

	scratch, err := os.MkdirTemp(filepath.Dir(req.OutputDir), "scratch-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(scratch)
	if err := stageInputs(req, scratch); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(req.OutputDir, 0o755); err != nil {
		return nil, err
	}

	sb, err := toolbox.NewSandbox(scratch)
	if err != nil {
		return nil, err
	}

	// Build tool definitions.
	tools := toolDefs()
	system := buildSystem(req)
	model := req.Model
	if model == "" {
		model = DefaultModel
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 4096,
		System:    system,
		Tools:     tools,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.Prompt)),
		},
	}

	res := &runner.Result{ToolCalls: map[string]int{}}
	var transcript strings.Builder

	const maxRounds = 16
	for round := 0; round < maxRounds; round++ {
		resp, err := client.Messages.New(ctx, params)
		if err != nil {
			res.Err = fmt.Errorf("messages.new (round %d): %w", round, err)
			break
		}
		res.PromptTokens += int(resp.Usage.InputTokens)
		res.CompletionTokens += int(resp.Usage.OutputTokens)
		res.TotalTokens = res.PromptTokens + res.CompletionTokens

		toolUses := []anthropic.ContentBlockUnion{}
		for _, block := range resp.Content {
			switch block.Type {
			case "text":
				transcript.WriteString("assistant: ")
				transcript.WriteString(block.Text)
				transcript.WriteString("\n")
			case "tool_use":
				toolUses = append(toolUses, block)
				fmt.Fprintf(&transcript, "tool_use: %s(%s)\n", block.Name, string(block.Input))
				res.ToolCalls[block.Name]++
			}
		}
		// Append assistant message so the next round's context includes
		// what Claude just said.
		asBlocks := make([]anthropic.ContentBlockParamUnion, 0, len(resp.Content))
		for _, b := range resp.Content {
			// Only pass-through types we know about.
			switch b.Type {
			case "text":
				asBlocks = append(asBlocks, anthropic.NewTextBlock(b.Text))
			case "tool_use":
				var input any
				_ = json.Unmarshal(b.Input, &input)
				asBlocks = append(asBlocks, anthropic.NewToolUseBlock(b.ID, input, b.Name))
			}
		}
		params.Messages = append(params.Messages, anthropic.NewAssistantMessage(asBlocks...))

		if len(toolUses) == 0 {
			break // end_turn
		}
		// Execute each tool call and append tool_result blocks.
		var results []anthropic.ContentBlockParamUnion
		for _, tu := range toolUses {
			var args map[string]any
			_ = json.Unmarshal(tu.Input, &args)
			out, err := sb.Call(ctx, tu.Name, args)
			isErr := err != nil
			if isErr {
				out = "error: " + err.Error()
			}
			results = append(results, anthropic.NewToolResultBlock(tu.ID, out, isErr))
			fmt.Fprintf(&transcript, "tool_result: %s -> %s\n", tu.Name, oneLine(out, 80))
		}
		params.Messages = append(params.Messages, anthropic.NewUserMessage(results...))
	}

	// Copy any files the agent wrote under scratch into OutputDir so the
	// grader can inspect them.
	outFiles, _ := collectIntoOutput(scratch, req.OutputDir)
	res.OutputFiles = outFiles
	res.Transcript = []byte(transcript.String())
	res.DurationMs = time.Since(start).Milliseconds()
	res.CostUSD = runner.EstimateCost(defaultPricing, res.PromptTokens, res.CompletionTokens)
	return res, nil
}

// --- helpers ----------------------------------------------------------------

func toolDefs() []anthropic.ToolUnionParam {
	defs := toolbox.DefaultTools()
	out := make([]anthropic.ToolUnionParam, 0, len(defs))
	for _, t := range defs {
		props, _ := t.Schema["properties"].(map[string]any)
		reqArr, _ := t.Schema["required"].([]string)
		schema := anthropic.ToolInputSchemaParam{
			Properties: props,
			Required:   reqArr,
		}
		tp := anthropic.ToolParam{
			Name:        t.Name,
			InputSchema: schema,
		}
		if t.Description != "" {
			tp.Description = anthropic.String(t.Description)
		}
		out = append(out, anthropic.ToolUnionParam{OfTool: &tp})
	}
	return out
}

func buildSystem(req runner.Request) []anthropic.TextBlockParam {
	var blocks []anthropic.TextBlockParam
	if req.SystemPrompt != "" {
		blocks = append(blocks, anthropic.TextBlockParam{Text: req.SystemPrompt})
	}
	if req.SkillDir != "" {
		if body, err := os.ReadFile(filepath.Join(req.SkillDir, "SKILL.md")); err == nil {
			blocks = append(blocks, anthropic.TextBlockParam{
				Text: "You have access to the following skill. Use it as your primary guidance:\n\n" + string(body),
			})
		}
	}
	if len(blocks) == 0 {
		return nil
	}
	return blocks
}

func stageInputs(req runner.Request, scratch string) error {
	inputs := filepath.Join(scratch, "inputs")
	if err := os.MkdirAll(inputs, 0o755); err != nil {
		return err
	}
	for _, f := range req.InputFiles {
		src, err := os.Open(f)
		if err != nil {
			return err
		}
		dst, err := os.Create(filepath.Join(inputs, filepath.Base(f)))
		if err != nil {
			src.Close()
			return err
		}
		if _, err := dst.ReadFrom(src); err != nil {
			src.Close()
			dst.Close()
			return err
		}
		src.Close()
		dst.Close()
	}
	return nil
}

func collectIntoOutput(scratch, outDir string) ([]string, error) {
	var out []string
	err := filepath.Walk(scratch, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		// Skip the inputs/ subtree - that's what we staged in.
		rel, _ := filepath.Rel(scratch, p)
		if strings.HasPrefix(rel, "inputs/") {
			return nil
		}
		dst := filepath.Join(outDir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return err
		}
		out = append(out, rel)
		return nil
	})
	return out, err
}

func oneLine(s string, max int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
