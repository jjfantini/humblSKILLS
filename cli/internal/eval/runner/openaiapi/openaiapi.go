// Package openaiapi is the pure-Go runner that speaks directly to the
// OpenAI Chat Completions API. Mirrors the anthropic-api package: same
// tool surface (Read/Write/Bash/Glob/Grep), same sandboxing, same Result
// shape, so scenarios are portable across runners.
package openaiapi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner/toolbox"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
)

// DefaultModel is a sensible default for coding-agent evals.
const DefaultModel = "gpt-5"

var defaultPricing = &runner.Pricing{
	PromptUSDPerMtok:     2.5,
	CompletionUSDPerMtok: 10,
}

// Runner is the openai-api backend.
type Runner struct {
	Store secrets.Store
}

// New returns a Runner.
func New(store secrets.Store) *Runner { return &Runner{Store: store} }

func (r *Runner) Name() string { return "openai-api" }

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
			Available:   false,
			Reason:      "secrets store not configured",
			Fix:         "run `humblskills eval set-key openai`",
			RequiresKey: "openai",
		}
	}
	_, src, err := r.Store.Get("openai")
	if err != nil || src == secrets.SourceAbsent {
		return runner.DoctorCheck{
			Available:   false,
			Reason:      "OPENAI_API_KEY not set",
			Fix:         "export OPENAI_API_KEY=... or run `humblskills eval set-key openai`",
			RequiresKey: "openai",
		}
	}
	return runner.DoctorCheck{
		Available:   true,
		Version:     "openai-go v1.12.0",
		RequiresKey: "openai",
	}
}

func (r *Runner) Execute(ctx context.Context, req runner.Request) (*runner.Result, error) {
	start := time.Now()
	key, _, err := r.Store.Get("openai")
	if err != nil || key == "" {
		return &runner.Result{Err: fmt.Errorf("no API key")}, fmt.Errorf("no API key")
	}
	client := openai.NewClient(option.WithAPIKey(key))

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

	model := req.Model
	if model == "" {
		model = DefaultModel
	}

	msgs := []openai.ChatCompletionMessageParamUnion{}
	if system := buildSystem(req); system != "" {
		msgs = append(msgs, openai.SystemMessage(system))
	}
	msgs = append(msgs, openai.UserMessage(req.Prompt))

	tools := toolDefs()

	res := &runner.Result{ToolCalls: map[string]int{}}
	var transcript strings.Builder

	const maxRounds = 16
	for round := 0; round < maxRounds; round++ {
		resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model:               shared.ChatModel(model),
			Messages:            msgs,
			Tools:               tools,
			MaxCompletionTokens: openai.Int(4096),
		})
		if err != nil {
			res.Err = fmt.Errorf("chat.completions.new (round %d): %w", round, err)
			break
		}
		if len(resp.Choices) == 0 {
			break
		}
		choice := resp.Choices[0]
		msg := choice.Message
		res.PromptTokens += int(resp.Usage.PromptTokens)
		res.CompletionTokens += int(resp.Usage.CompletionTokens)
		res.TotalTokens = res.PromptTokens + res.CompletionTokens

		if msg.Content != "" {
			transcript.WriteString("assistant: ")
			transcript.WriteString(msg.Content)
			transcript.WriteString("\n")
		}
		// Append the assistant message (with tool_calls) so the next
		// round's context carries the pending calls.
		msgs = append(msgs, msg.ToParam())

		if len(msg.ToolCalls) == 0 {
			break
		}
		for _, tc := range msg.ToolCalls {
			res.ToolCalls[tc.Function.Name]++
			var args map[string]any
			_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
			out, err := sb.Call(ctx, tc.Function.Name, args)
			if err != nil {
				out = "error: " + err.Error()
			}
			fmt.Fprintf(&transcript, "tool_result: %s -> %s\n", tc.Function.Name, oneLine(out, 80))
			msgs = append(msgs, openai.ToolMessage(out, tc.ID))
		}
	}

	outFiles, _ := collectIntoOutput(scratch, req.OutputDir)
	res.OutputFiles = outFiles
	res.Transcript = []byte(transcript.String())
	res.DurationMs = time.Since(start).Milliseconds()
	res.CostUSD = runner.EstimateCost(defaultPricing, res.PromptTokens, res.CompletionTokens)
	return res, nil
}

// --- helpers ----------------------------------------------------------------

func toolDefs() []openai.ChatCompletionToolParam {
	defs := toolbox.DefaultTools()
	out := make([]openai.ChatCompletionToolParam, 0, len(defs))
	for _, t := range defs {
		params := shared.FunctionParameters(t.Schema)
		fd := shared.FunctionDefinitionParam{
			Name:       t.Name,
			Parameters: params,
		}
		if t.Description != "" {
			fd.Description = openai.String(t.Description)
		}
		out = append(out, openai.ChatCompletionToolParam{Function: fd})
	}
	return out
}

func buildSystem(req runner.Request) string {
	var sb strings.Builder
	if req.SystemPrompt != "" {
		sb.WriteString(req.SystemPrompt)
		sb.WriteString("\n\n")
	}
	if req.SkillDir != "" {
		if body, err := os.ReadFile(filepath.Join(req.SkillDir, "SKILL.md")); err == nil {
			sb.WriteString("You have access to the following skill. Use it as your primary guidance:\n\n")
			sb.WriteString(string(body))
		}
	}
	return strings.TrimSpace(sb.String())
}

func stageInputs(req runner.Request, scratch string) error {
	inputs := filepath.Join(scratch, "inputs")
	if err := os.MkdirAll(inputs, 0o755); err != nil {
		return err
	}
	for _, f := range req.InputFiles {
		data, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(inputs, filepath.Base(f)), data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func collectIntoOutput(scratch, outDir string) ([]string, error) {
	var out []string
	err := filepath.Walk(scratch, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
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
