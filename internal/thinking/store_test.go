package thinking

import (
	"strings"
	"testing"
	"time"
)

func TestMemoryStoreCreateSessionPopulatesDefaults(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	context := map[string]any{"team": "reasoning"}
	tags := []string{"analysis", "priority"}

	session, err := store.CreateSession("Investigate outage", context, tags)
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	if session.ID == "" {
		t.Fatalf("expected session ID to be generated")
	}
	if session.Status != "active" {
		t.Fatalf("expected status 'active', got %q", session.Status)
	}
	if !strings.Contains(session.InitialAnalysis, session.Problem) {
		t.Fatalf("initial analysis should reference problem, got %q", session.InitialAnalysis)
	}
	if session.CurrentStep != 0 {
		t.Fatalf("expected current step 0, got %d", session.CurrentStep)
	}
	if session.QualityScore != 0.5 {
		t.Fatalf("expected initial quality score 0.5, got %v", session.QualityScore)
	}
	if got := session.Context["team"]; got != "reasoning" {
		t.Fatalf("expected context clone to preserve values, got %v", got)
	}
	if len(session.Tags) != len(tags) {
		t.Fatalf("expected %d tags, got %d", len(tags), len(session.Tags))
	}
}

func TestMemoryStoreAddStepAndConnections(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	session := mustCreateSession(t, store, "Track regression")

	step1 := mustAddStep(t, store, session.ID, "Gather current metrics", StepAnalysis, nil, nil)
	if step1.Number != 1 {
		t.Fatalf("expected first step number 1, got %d", step1.Number)
	}

	parent := step1.Number
	step2 := mustAddStep(t, store, session.ID, "Hypothesis: cache invalidation fault", StepHypothesis, &parent, map[string]any{"confidence": 0.6})
	if step2.Number != 2 {
		t.Fatalf("expected step number 2, got %d", step2.Number)
	}

	cloned := mustGetSession(t, store, session.ID)

	if cloned.CurrentStep != 2 {
		t.Fatalf("expected current step 2, got %d", cloned.CurrentStep)
	}
	expectConnections(t, cloned.Steps[0].Connections, []int{2})
	expectConnections(t, cloned.Steps[1].Connections, []int{1})
	if cloned.Steps[1].Metadata["confidence"] != 0.6 {
		t.Fatalf("expected metadata to persist, got %v", cloned.Steps[1].Metadata)
	}

	if diff := abs(cloned.QualityScore - 0.49); diff > 1e-6 {
		t.Fatalf("expected quality score ≈0.49, got %v", cloned.QualityScore)
	}
}

func TestMemoryStoreAddStepRejectsInvalidType(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	session := mustCreateSession(t, store, "Invalid type guard")

	_, err := store.AddStep(session.ID, "bad type", StepType("unsupported"), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "invalid step type") {
		t.Fatalf("expected invalid step type error, got %v", err)
	}
}

func TestMemoryStoreUpdateStepValidations(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	session := mustCreateSession(t, store, "Improve onboarding")
	mustAddStep(t, store, session.ID, "Collect interview notes", StepAnalysis, nil, nil)

	// Updating a non-existent step should error.
	if _, err := store.UpdateStep(session.ID, 2, nil, nil, nil); err == nil {
		t.Fatalf("expected error updating unknown step")
	}

	content := "Validate assumptions with support tickets"
	stepType := StepVerification
	metadata := map[string]any{"source": "zendesk"}

	_, err := store.UpdateStep(session.ID, 1, &content, &stepType, metadata)
	if err != nil {
		t.Fatalf("UpdateStep returned error: %v", err)
	}

	metadata["source"] = "mutated" // ensure copy-on-write semantics

	cloned := mustGetSession(t, store, session.ID)
	step := cloned.Steps[0]

	if step.Content != content {
		t.Fatalf("expected updated content %q, got %q", content, step.Content)
	}
	if step.Type != stepType {
		t.Fatalf("expected updated type %q, got %q", stepType, step.Type)
	}
	if step.Metadata["source"] != "zendesk" {
		t.Fatalf("expected metadata to be cloned, got %v", step.Metadata)
	}
	if cloned.CurrentStep != 1 {
		t.Fatalf("expected current step 1, got %d", cloned.CurrentStep)
	}
}

func TestMemoryStoreDeleteSession(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	session := mustCreateSession(t, store, "Sunset feature")

	if err := store.DeleteSession(session.ID); err != nil {
		t.Fatalf("DeleteSession returned error: %v", err)
	}

	if err := store.DeleteSession(session.ID); err == nil {
		t.Fatalf("expected error deleting already removed session")
	}
}

func TestMemoryStoreBranchingLifecycle(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	session := mustCreateSession(t, store, "Evaluate strategy")

	mustAddStep(t, store, session.ID, "Baseline performance", StepAnalysis, nil, nil)

	branch, err := store.CreateBranch(session.ID, 1, "Alternative approach")
	if err != nil {
		t.Fatalf("CreateBranch returned error: %v", err)
	}

	mustAddBranchStep(t, store, session.ID, branch.ID, "Explore new market", StepHypothesis, nil, map[string]any{"region": "EU"})

	cloned := mustGetSession(t, store, session.ID)
	storedBranch, ok := cloned.Branches[branch.ID]
	if !ok {
		t.Fatalf("expected branch %s to exist", branch.ID)
	}
	if storedBranch.FromStep != 1 {
		t.Fatalf("expected branch from step 1, got %d", storedBranch.FromStep)
	}
	if len(storedBranch.Steps) != 1 {
		t.Fatalf("expected 1 step in branch, got %d", len(storedBranch.Steps))
	}
	if storedBranch.Steps[0].Metadata["region"] != "EU" {
		t.Fatalf("expected branch metadata to persist, got %v", storedBranch.Steps[0].Metadata)
	}
}

func TestMemoryStoreListSessionsReturnsClones(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	context := map[string]any{"phase": "beta"}
	session, err := store.CreateSession("Rollout plan", context, []string{"ops"})
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	sessions := store.ListSessions()
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	// Mutate the returned clone and ensure store data is unchanged.
	sessions[0].Problem = "tampered"
	sessions[0].Context["phase"] = "tampered"

	cloned := mustGetSession(t, store, session.ID)
	if cloned.Problem != "Rollout plan" {
		t.Fatalf("expected stored problem to remain unchanged, got %q", cloned.Problem)
	}
	if cloned.Context["phase"] != "beta" {
		t.Fatalf("expected stored context to remain unchanged, got %v", cloned.Context["phase"])
	}
}

func TestMemoryStoreGetMetricsRespectsTimeRange(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()

	// Recent session within the week window.
	recent := mustCreateSessionWithTags(t, store, "Stabilize pipeline", []string{"ops", "priority"})
	mustAddStep(t, store, recent.ID, "Analyze logs", StepAnalysis, nil, nil)
	parent := 1
	mustAddStep(t, store, recent.ID, "Hypothesis: flaky worker", StepHypothesis, &parent, nil)
	parent = 2
	mustAddStep(t, store, recent.ID, "Run targeted load test", StepVerification, &parent, nil)
	parent = 3
	mustAddStep(t, store, recent.ID, "Conclusion: upgrade worker pool", StepConclusion, &parent, nil)
	if err := store.UpdateSessionStatus(recent.ID, "completed"); err != nil {
		t.Fatalf("UpdateSessionStatus returned error: %v", err)
	}

	// Older session should be excluded for "week" queries.
	old := mustCreateSessionWithTags(t, store, "Archived incident", []string{"legacy"})
	store.mu.Lock()
	store.sessions[old.ID].Created = time.Now().Add(-10 * 24 * time.Hour)
	store.mu.Unlock()

	metricsWeek := store.GetMetrics("week")
	if metricsWeek.TotalSessions != 1 {
		t.Fatalf("expected 1 session in week range, got %d", metricsWeek.TotalSessions)
	}
	if metricsWeek.CompletedSessions != 1 {
		t.Fatalf("expected completed session count 1, got %d", metricsWeek.CompletedSessions)
	}
	if steps := metricsWeek.StepTypeDistrib[StepConclusion]; steps != 1 {
		t.Fatalf("expected one conclusion step counted, got %d", steps)
	}
	if metricsWeek.AverageSteps != 4 {
		t.Fatalf("expected average steps 4, got %v", metricsWeek.AverageSteps)
	}
	if metricsWeek.TopTags["ops"] != 1 {
		t.Fatalf("expected ops tag count 1, got %d", metricsWeek.TopTags["ops"])
	}
	if metricsWeek.TopTags["priority"] != 1 {
		t.Fatalf("expected priority tag count 1, got %d", metricsWeek.TopTags["priority"])
	}

	metricsAll := store.GetMetrics("all")
	if metricsAll.TotalSessions != 2 {
		t.Fatalf("expected 2 sessions for 'all', got %d", metricsAll.TotalSessions)
	}
	if metricsAll.TopTags["ops"] != 1 {
		t.Fatalf("expected ops tag count 1 in all metrics, got %d", metricsAll.TopTags["ops"])
	}
	if metricsAll.TopTags["priority"] != 1 {
		t.Fatalf("expected priority tag count 1 in all metrics, got %d", metricsAll.TopTags["priority"])
	}
	if metricsAll.TopTags["legacy"] != 1 {
		t.Fatalf("expected legacy tag count 1, got %d", metricsAll.TopTags["legacy"])
	}
	if len(metricsAll.CommonPatterns) == 0 || metricsAll.CommonPatterns[0].Frequency == 0 {
		t.Fatalf("expected hypothesis-verification pattern to be detected, got %v", metricsAll.CommonPatterns)
	}
}

func TestGenerateInitialAnalysisIncludesProblem(t *testing.T) {
	t.Parallel()

	problem := "How do we reduce response time?"
	result := generateInitialAnalysis(problem)
	if !strings.Contains(result, problem) {
		t.Fatalf("expected initial analysis to include the problem statement")
	}
}

func mustCreateSession(t *testing.T, store *MemoryStore, problem string) *ThinkingSession {
	t.Helper()

	return mustCreateSessionWithTags(t, store, problem, nil)
}

func mustCreateSessionWithTags(t *testing.T, store *MemoryStore, problem string, tags []string) *ThinkingSession {
	t.Helper()

	session, err := store.CreateSession(problem, nil, tags)
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}
	return session
}

func mustAddStep(t *testing.T, store *MemoryStore, sessionID, content string, stepType StepType, parent *int, metadata map[string]any) *ThinkingStep {
	t.Helper()

	step, err := store.AddStep(sessionID, content, stepType, parent, metadata)
	if err != nil {
		t.Fatalf("AddStep returned error: %v", err)
	}
	return step
}

func mustAddBranchStep(t *testing.T, store *MemoryStore, sessionID, branchID, content string, stepType StepType, parent *int, metadata map[string]any) *ThinkingStep {
	t.Helper()

	step, err := store.AddStepToBranch(sessionID, branchID, content, stepType, parent, metadata)
	if err != nil {
		t.Fatalf("AddStepToBranch returned error: %v", err)
	}
	return step
}

func mustGetSession(t *testing.T, store *MemoryStore, sessionID string) *ThinkingSession {
	t.Helper()

	session, err := store.GetSession(sessionID)
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}
	return session
}

func expectConnections(t *testing.T, got, want []int) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("expected connections %v, got %v", want, got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("expected connections %v, got %v", want, got)
		}
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
