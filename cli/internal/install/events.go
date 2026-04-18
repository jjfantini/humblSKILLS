package install

// Phase names one moment in an install run that the UI layer might want to
// reflect (progress bar ticks, spinner labels, log lines).
type Phase string

const (
	// PhaseRunStart fires once, before any step, with Total set to the number
	// of (skill, platform, scope) triples that will be visited.
	PhaseRunStart Phase = "run_start"
	// PhaseStepStart fires once per planned skill, before its targets are
	// processed. IsDep distinguishes transitive deps from the root skill.
	PhaseStepStart Phase = "step_start"
	// PhaseTargetStart fires when work begins on one (skill, platform, scope)
	// triple. The skill may still be shared across targets via staging.
	PhaseTargetStart Phase = "target_start"
	// PhaseTargetDone fires when a single target finishes. Outcome names the
	// terminal state (installed / replaced / skipped / forced).
	PhaseTargetDone Phase = "target_done"
	// PhaseRunDone fires once, after every step has been processed.
	PhaseRunDone Phase = "run_done"
	// PhaseError fires when a target fails; the caller can show Err and abort.
	PhaseError Phase = "error"
)

// Event is a single progress notification emitted by the engine. Not every
// field is populated for every Phase — Total is only meaningful on RunStart,
// Outcome only on TargetDone, Err only on Error.
type Event struct {
	Phase    Phase
	Skill    string
	Platform string
	Scope    string
	IsDep    bool
	Total    int
	Outcome  Outcome
	Err      error
}

// EventSink receives engine progress events. Callers that don't care about
// progress (tests, scripts, --json consumers) can pass nil.
type EventSink func(Event)

// emit is a nil-safe helper so call-sites don't have to guard.
func (s EventSink) emit(ev Event) {
	if s != nil {
		s(ev)
	}
}
