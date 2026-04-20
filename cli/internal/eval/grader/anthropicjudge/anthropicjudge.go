// Package anthropicjudge is the LLMJudge implementation that sends "llm"
// assertions to the Anthropic Messages API for grading.
//
// Contract follows Anthropic's own agents/grader.md: given the eval prompt,
// the executor transcript, the tree of output files, and a list of
// assertions, return per-assertion {text, passed, evidence}. Everything
// batched into ONE API call so a session with N llm assertions costs one
// round-trip.
package anthropicjudge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/grader"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/scenarios"
)

// DefaultModel is the grader model. Opus is the right call here - graders
// should be smarter than the executor so false positives/negatives are
// minimized.
const DefaultModel = "claude-opus-4-5"

// Judge is an LLM-backed grader.
type Judge struct {
	client anthropic.Client
	model  string
}

// New returns a Judge using the given API key. An empty model falls back
// to DefaultModel.
func New(apiKey, model string) *Judge {
	if model == "" {
		model = DefaultModel
	}
	return &Judge{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
		model:  model,
	}
}

// Grade batches every llm assertion into one API call. Returns results in
// the SAME order as the input assertions.
func (j *Judge) Grade(ctx context.Context, evalPrompt string, transcript []byte, outputs string,
	assertions []scenarios.Assertion) ([]grader.ExpectationResult, error) {

	if len(assertions) == 0 {
		return nil, nil
	}

	// Trim giant transcripts so the judge doesn't OOM and costs stay
	// bounded. 40 KiB each on transcript + outputs is plenty for
	// session-level judgement.
	const cap = 40 * 1024
	transcriptStr := string(transcript)
	if len(transcriptStr) > cap {
		transcriptStr = transcriptStr[:cap] + "\n... (transcript truncated)"
	}
	if len(outputs) > cap {
		outputs = outputs[:cap] + "\n... (outputs truncated)"
	}

	// The judge needs to echo the assertion texts back so we can match
	// results to inputs. Enumerate them.
	var asBullets strings.Builder
	for i, a := range assertions {
		fmt.Fprintf(&asBullets, "  %d. %s\n", i+1, a.Text)
	}

	userPrompt := fmt.Sprintf(`You are grading an agent's output against a set of expectations.

## Eval prompt given to the agent

%s

## Agent output files

%s

## Agent transcript (truncated)

%s

## Expectations to grade

%s

## Instructions

For EACH numbered expectation above, decide if it passes or fails based on
the transcript + output files. Require concrete evidence to PASS -
quoting or referencing the specific text that supports the verdict.
If an expectation asks for a quality score (e.g. "1-10 where 1 is human"),
pass only when the threshold is met and put the numeric score in evidence.

Respond with a JSON object of this exact shape and NOTHING else:

{
  "expectations": [
    { "text": "<exact expectation text from above>", "passed": true|false, "evidence": "<one to two sentences with a quoted phrase>" },
    ...
  ]
}

The array MUST contain one entry per expectation in the order given.`,
		evalPrompt, outputs, transcriptStr, asBullets.String())

	resp, err := j.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(j.model),
		MaxTokens: 4096,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("grader.messages.new: %w", err)
	}

	var reply string
	for _, b := range resp.Content {
		if b.Type == "text" {
			reply += b.Text
		}
	}
	reply = extractJSONObject(reply)
	var parsed struct {
		Expectations []grader.ExpectationResult `json:"expectations"`
	}
	if err := json.Unmarshal([]byte(reply), &parsed); err != nil {
		return nil, fmt.Errorf("parse judge reply: %w (reply=%q)", err, truncate(reply, 200))
	}
	if len(parsed.Expectations) != len(assertions) {
		return nil, fmt.Errorf("judge returned %d expectations, want %d",
			len(parsed.Expectations), len(assertions))
	}
	// Force the text back to the exact assertion text so downstream
	// aggregators can key on it cleanly even if the model paraphrased.
	for i := range parsed.Expectations {
		parsed.Expectations[i].Text = assertions[i].Text
	}
	return parsed.Expectations, nil
}

// extractJSONObject pulls the first {...} block out of a string, since
// models occasionally wrap their JSON in prose or code fences.
func extractJSONObject(s string) string {
	s = strings.TrimSpace(s)
	// Strip code fences.
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start < 0 || end <= start {
		return s
	}
	return s[start : end+1]
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
