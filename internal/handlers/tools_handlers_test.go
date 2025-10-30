package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bpradana/sequentialthinking/internal/thinking"
)

func TestCreateStartThinkingHandler(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	handler := createStartThinkingHandler(store)

	_, out, err := handler(context.Background(), &mcp.CallToolRequest{}, thinking.StartThinkingInput{
		Problem: "Reduce latency",
		Tags:    []string{"performance"},
	})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if out.SessionID == "" {
		t.Fatalf("expected session ID to be generated")
	}
	if !strings.Contains(out.InitialAnalysis, "Reduce latency") {
		t.Fatalf("initial analysis should mention the problem:\n%s", out.InitialAnalysis)
	}
	if len(out.SuggestedSteps) == 0 {
		t.Fatalf("expected suggested steps")
	}

	sessions := store.ListSessions()
	if len(sessions) != 1 || sessions[0].ID != out.SessionID {
		t.Fatalf("expected store to contain created session, got %v", sessions)
	}
}

func TestCreateStartFromTemplateHandler(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	handler := createStartFromTemplateHandler(store)

	_, out, err := handler(context.Background(), &mcp.CallToolRequest{}, thinking.StartFromTemplateInput{
		Template: "five-whys",
		Problem:  "Investigate outage recurrence",
		Tags:     []string{"incident"},
	})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if out.SessionID == "" {
		t.Fatalf("expected session ID to be generated")
	}
	if out.Template == nil || out.Template.Type != "five-whys" {
		t.Fatalf("expected template metadata for five-whys, got %+v", out.Template)
	}
	if len(out.SuggestedSteps) != len(out.Template.Steps) {
		t.Fatalf("expected suggested steps to mirror template steps")
	}
	if !strings.Contains(out.InitialAnalysis, "Investigate outage recurrence") {
		t.Fatalf("expected initial analysis to reference problem; got %q", out.InitialAnalysis)
	}

	sessions := store.ListSessions()
	if len(sessions) != 1 || sessions[0].ID != out.SessionID {
		t.Fatalf("expected store to contain created session, got %v", sessions)
	}
	if sessions[0].Context["template_type"] != "five-whys" {
		t.Fatalf("expected session context to include template metadata, got %+v", sessions[0].Context)
	}

	_, _, err = handler(context.Background(), &mcp.CallToolRequest{}, thinking.StartFromTemplateInput{
		Template: "unknown-template",
	})
	if err == nil {
		t.Fatalf("expected error for unknown template")
	}
}

func TestCreateAddStepHandlerMainAndBranch(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	session := mustCreateSession(t, store, "Ship new feature")

	addHandler := createAddStepHandler(store)
	parent := 0

	_, out1, err := addHandler(context.Background(), &mcp.CallToolRequest{}, thinking.AddStepInput{
		SessionID:   session.ID,
		StepContent: "Analyze current architecture",
		StepType:    thinking.StepAnalysis,
		ParentStep:  nil,
	})
	if err != nil {
		t.Fatalf("add handler returned error: %v", err)
	}
	if out1.StepNumber != 1 {
		t.Fatalf("expected first step number 1, got %d", out1.StepNumber)
	}
	if !strings.Contains(out1.CurrentProgress, "Session has 1 steps") {
		t.Fatalf("unexpected progress message: %q", out1.CurrentProgress)
	}

	parent = out1.StepNumber
	branch, err := store.CreateBranch(session.ID, parent, "Explore alternative architecture")
	if err != nil {
		t.Fatalf("CreateBranch returned error: %v", err)
	}

	_, out2, err := addHandler(context.Background(), &mcp.CallToolRequest{}, thinking.AddStepInput{
		SessionID:   session.ID,
		BranchID:    branch.ID,
		StepContent: "Propose service-oriented option",
		StepType:    thinking.StepHypothesis,
	})
	if err != nil {
		t.Fatalf("branch add handler returned error: %v", err)
	}
	if out2.BranchID != branch.ID {
		t.Fatalf("expected branch ID %s, got %s", branch.ID, out2.BranchID)
	}
	if len(out2.SuggestedNextSteps) == 0 {
		t.Fatalf("expected suggested next steps")
	}
}

func TestCreateUpdateStepHandler(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	session := mustCreateSession(t, store, "Improve onboarding")
	mustAddStep(t, store, session.ID, "Gather feedback", thinking.StepAnalysis, nil, nil)

	handler := createUpdateStepHandler(store)
	newContent := "Validate assumptions via survey"
	newType := thinking.StepVerification
	metadata := map[string]any{"owner": "research"}

	_, out, err := handler(context.Background(), &mcp.CallToolRequest{}, thinking.UpdateStepInput{
		SessionID:   session.ID,
		StepNumber:  1,
		StepContent: &newContent,
		StepType:    &newType,
		Metadata:    metadata,
	})
	if err != nil {
		t.Fatalf("update handler returned error: %v", err)
	}
	if out.UpdatedStep.Content != newContent {
		t.Fatalf("expected updated content %q, got %q", newContent, out.UpdatedStep.Content)
	}
	if !out.MetadataChanged || !out.TypeChanged || !out.ContentChanged {
		t.Fatalf("expected all change indicators to be true: %+v", out)
	}
	if out.QualityScore <= 0 {
		t.Fatalf("expected quality score to be positive, got %v", out.QualityScore)
	}

	// ensure metadata cloning occurred
	metadata["owner"] = "mutated"
	cloned := mustGetSession(t, store, session.ID)
	if cloned.Steps[0].Metadata["owner"] != "research" {
		t.Fatalf("expected metadata to be cloned, got %v", cloned.Steps[0].Metadata)
	}
}

func TestCreateReviewThinkingHandler(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	session := mustCreateSession(t, store, "Stabilize pipeline")
	step1 := mustAddStep(t, store, session.ID, "Collect metrics", thinking.StepAnalysis, nil, nil)
	step2 := mustAddStep(t, store, session.ID, "Hypothesis: worker saturation", thinking.StepHypothesis, &step1.Number, nil)
	mustAddStep(t, store, session.ID, "Verify with load test", thinking.StepVerification, &step2.Number, nil)

	handler := createReviewThinkingHandler(store)
	_, out, err := handler(context.Background(), &mcp.CallToolRequest{}, thinking.ReviewThinkingInput{
		SessionID: session.ID,
		Format:    "tree",
	})
	if err != nil {
		t.Fatalf("review handler returned error: %v", err)
	}
	if len(out.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(out.Steps))
	}
	if _, ok := out.Connections["step_1"]; !ok {
		t.Fatalf("expected connections to include step_1")
	}
	if !strings.Contains(out.Summary, "Reasoning Tree") {
		t.Fatalf("expected tree summary, got:\n%s", out.Summary)
	}
	if len(out.Patterns) == 0 {
		t.Fatalf("expected patterns to be detected")
	}
}

func TestCreateBranchThinkingHandler(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	session := mustCreateSession(t, store, "Evaluate expansion")
	step := mustAddStep(t, store, session.ID, "Assess current markets", thinking.StepAnalysis, nil, nil)

	handler := createBranchThinkingHandler(store)
	_, out, err := handler(context.Background(), &mcp.CallToolRequest{}, thinking.BranchThinkingInput{
		SessionID:            session.ID,
		FromStep:             step.Number,
		AlternativeReasoning: "Consider emerging regions",
	})
	if err != nil {
		t.Fatalf("branch handler returned error: %v", err)
	}
	if out.BranchID == "" {
		t.Fatalf("expected branch id to be set")
	}
	if !strings.Contains(out.BranchSummary, "alternative reasoning path") {
		t.Fatalf("unexpected summary: %s", out.BranchSummary)
	}

	cloned := mustGetSession(t, store, session.ID)
	branch := cloned.Branches[out.BranchID]
	if len(branch.Steps) != 1 || branch.Steps[0].Type != thinking.StepHypothesis {
		t.Fatalf("expected initial hypothesis step in branch, got %v", branch.Steps)
	}
}

func TestCreateMergeInsightsHandler(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	session := mustCreateSession(t, store, "Decide on architecture")
	mustAddStep(t, store, session.ID, "Document current bottlenecks", thinking.StepAnalysis, nil, nil)
	conclusion := mustAddStep(t, store, session.ID, "Adopt event-driven design", thinking.StepConclusion, nil, nil)

	branch, err := store.CreateBranch(session.ID, conclusion.Number, "Keep monolith")
	if err != nil {
		t.Fatalf("CreateBranch returned error: %v", err)
	}

	mustAddBranchStep(t, store, session.ID, branch.ID, "Monolith scales with caching", thinking.StepConclusion, nil, nil)

	handler := createMergeInsightsHandler(store)
	_, out, err := handler(context.Background(), &mcp.CallToolRequest{}, thinking.MergeInsightsInput{
		SessionID: session.ID,
		BranchIDs: []string{branch.ID},
	})
	if err != nil {
		t.Fatalf("merge handler returned error: %v", err)
	}
	if !strings.Contains(out.Synthesis, "Synthesized Insights") {
		t.Fatalf("expected synthesis text, got %s", out.Synthesis)
	}
	if out.Confidence <= 0 {
		t.Fatalf("expected confidence > 0, got %v", out.Confidence)
	}
	if len(out.Strengths) == 0 {
		t.Fatalf("expected strengths")
	}
}

func TestCreateValidateLogicHandler(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	session := mustCreateSession(t, store, "Validate reasoning")
	step1 := mustAddStep(t, store, session.ID, "Initial data review", thinking.StepAnalysis, nil, nil)
	mustAddStep(t, store, session.ID, "Hypothesis: caching issue", thinking.StepHypothesis, &step1.Number, nil)
	// Conclusion without connections should trigger issue
	mustAddStep(t, store, session.ID, "Conclusion: rewrite cache layer", thinking.StepConclusion, nil, nil)

	handler := createValidateLogicHandler(store)
	_, out, err := handler(context.Background(), &mcp.CallToolRequest{}, thinking.ValidateLogicInput{
		SessionID: session.ID,
	})
	if err != nil {
		t.Fatalf("validate handler returned error: %v", err)
	}
	if len(out.Issues) < 1 {
		t.Fatalf("expected at least one logical issue, got %v", out.Issues)
	}
	if out.ValidityScore >= 1 {
		t.Fatalf("expected validity score less than 1, got %v", out.ValidityScore)
	}
	if len(out.Suggestions) == 0 {
		t.Fatalf("expected suggestions for issues")
	}
	if !strings.Contains(out.OverallAssessment, "Validity Score") {
		t.Fatalf("unexpected assessment: %s", out.OverallAssessment)
	}
}

func TestCreateExportSessionHandler(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	session := mustCreateSession(t, store, "Export reasoning")
	mustAddStep(t, store, session.ID, "Draft plan", thinking.StepAnalysis, nil, nil)

	handler := createExportSessionHandler(store)

	formats := []string{"markdown", "json", "text"}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			_, out, err := handler(context.Background(), &mcp.CallToolRequest{}, thinking.ExportSessionInput{
				SessionID:       session.ID,
				Format:          format,
				IncludeBranches: true,
			})
			if err != nil {
				t.Fatalf("export handler returned error: %v", err)
			}
			if out.Format != format {
				t.Fatalf("expected format %s, got %s", format, out.Format)
			}
			if out.Filename == "" || !strings.HasPrefix(out.Filename, "thinking_") {
				t.Fatalf("unexpected filename: %s", out.Filename)
			}
			if !strings.Contains(out.Content, session.Problem) {
				t.Fatalf("expected content to include problem: %s", out.Content)
			}
		})
	}

	_, _, err := handler(context.Background(), &mcp.CallToolRequest{}, thinking.ExportSessionInput{
		SessionID: session.ID,
		Format:    "unsupported",
	})
	if err == nil {
		t.Fatalf("expected error for unsupported format")
	}
}

func TestCreateListSessionsHandler(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	session1 := mustCreateSessionWithTags(t, store, "Session A", []string{"ops"})
	session2 := mustCreateSessionWithTags(t, store, "Session B", []string{"research"})

	if err := store.UpdateSessionStatus(session2.ID, "completed"); err != nil {
		t.Fatalf("UpdateSessionStatus returned error: %v", err)
	}

	handler := createListSessionsHandler(store)
	_, out, err := handler(context.Background(), &mcp.CallToolRequest{}, thinking.ListSessionsInput{
		Status: "completed",
	})
	if err != nil {
		t.Fatalf("list handler returned error: %v", err)
	}
	if out.Total != 1 || out.Sessions[0].ID != session2.ID {
		t.Fatalf("expected only completed session, got %+v", out.Sessions)
	}

	_, limited, err := handler(context.Background(), &mcp.CallToolRequest{}, thinking.ListSessionsInput{
		Status: "all",
		Limit:  1,
	})
	if err != nil {
		t.Fatalf("list handler with limit returned error: %v", err)
	}
	if limited.Total != 2 || len(limited.Sessions) != 1 {
		t.Fatalf("expected limit to truncate sessions, got %+v", limited)
	}

	_, tagged, err := handler(context.Background(), &mcp.CallToolRequest{}, thinking.ListSessionsInput{
		Tags: []string{"ops"},
	})
	if err != nil {
		t.Fatalf("list handler tagged returned error: %v", err)
	}
	if tagged.Total != 1 || tagged.Sessions[0].ID != session1.ID {
		t.Fatalf("expected tag filter to select session1, got %+v", tagged.Sessions)
	}
}

func TestCreateDeleteSessionHandler(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	session := mustCreateSession(t, store, "Delete me")

	handler := createDeleteSessionHandler(store)
	_, out, err := handler(context.Background(), &mcp.CallToolRequest{}, thinking.DeleteSessionInput{
		SessionID: session.ID,
	})
	if err != nil {
		t.Fatalf("delete handler returned error: %v", err)
	}
	if !out.Success {
		t.Fatalf("expected success true, got %+v", out)
	}

	_, out2, err := handler(context.Background(), &mcp.CallToolRequest{}, thinking.DeleteSessionInput{
		SessionID: session.ID,
	})
	if err != nil {
		t.Fatalf("delete handler second call returned error: %v", err)
	}
	if out2.Success {
		t.Fatalf("expected failure on deleting missing session, got %+v", out2)
	}
}

func TestCreateGetMetricsHandler(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	session := mustCreateSessionWithTags(t, store, "Metrics session", []string{"ops"})
	mustAddStep(t, store, session.ID, "Analyze incident", thinking.StepAnalysis, nil, nil)

	handler := createGetMetricsHandler(store)
	_, out, err := handler(context.Background(), &mcp.CallToolRequest{}, thinking.GetMetricsInput{})
	if err != nil {
		t.Fatalf("metrics handler returned error: %v", err)
	}
	if out.Metrics.TotalSessions != 1 {
		t.Fatalf("expected total sessions 1, got %d", out.Metrics.TotalSessions)
	}
}

func TestGenerateProgress(t *testing.T) {
	t.Parallel()

	session := &thinking.ThinkingSession{
		Steps:        []*thinking.ThinkingStep{{}},
		QualityScore: 0.66,
		Status:       "active",
	}
	got := generateProgress(session)
	if !strings.Contains(got, "1 steps") || !strings.Contains(got, "0.66") {
		t.Fatalf("unexpected progress string: %s", got)
	}
}

func TestSuggestNextSteps(t *testing.T) {
	t.Parallel()

	session := &thinking.ThinkingSession{}
	tests := []struct {
		stepType thinking.StepType
		want     string
	}{
		{thinking.StepAnalysis, "Form a hypothesis"},
		{thinking.StepHypothesis, "Design verification"},
		{thinking.StepVerification, "Draw conclusions"},
		{thinking.StepConclusion, "Summarize key insights"},
	}

	for _, tc := range tests {
		last := &thinking.ThinkingStep{Type: tc.stepType}
		got := suggestNextSteps(session, last)
		if len(got) == 0 || !strings.Contains(got[0], tc.want) {
			t.Fatalf("expected suggestion containing %q for type %s, got %v", tc.want, tc.stepType, got)
		}
	}
}

func TestBuildConnectionsMap(t *testing.T) {
	t.Parallel()

	session := &thinking.ThinkingSession{
		Steps: []*thinking.ThinkingStep{
			{Number: 1, Connections: []int{2}},
			{Number: 2, Connections: []int{1, 3}},
		},
	}
	got := buildConnectionsMap(session)
	if len(got["step_1"]) != 1 || got["step_1"][0] != 2 {
		t.Fatalf("unexpected connections map: %v", got)
	}
}

func TestGenerateSummaryFormats(t *testing.T) {
	t.Parallel()

	session := &thinking.ThinkingSession{
		Problem:      "Example problem",
		QualityScore: 0.7,
		Branches:     map[string]*thinking.Branch{},
		Steps: []*thinking.ThinkingStep{
			{Number: 1, Type: thinking.StepAnalysis, Content: "Analyze situation"},
			{Number: 2, Type: thinking.StepHypothesis, Content: "Hypothesis 1"},
			{Number: 3, Type: thinking.StepConclusion, Content: "Conclusion", ParentStep: ptr(2)},
		},
	}

	tree := generateSummary(session, "tree")
	if !strings.Contains(tree, "Reasoning Tree") || !strings.Contains(tree, "2. [hypothesis]") {
		t.Fatalf("unexpected tree summary:\n%s", tree)
	}

	linear := generateSummary(session, "")
	if !strings.Contains(linear, "Key Steps") || !strings.Contains(linear, "- Step 2") {
		t.Fatalf("unexpected linear summary:\n%s", linear)
	}
}

func TestDetectSessionPatterns(t *testing.T) {
	t.Parallel()

	session := &thinking.ThinkingSession{
		Steps: []*thinking.ThinkingStep{
			{Type: thinking.StepAnalysis},
			{Type: thinking.StepHypothesis},
			{Type: thinking.StepVerification},
		},
	}
	got := detectSessionPatterns(session)
	if len(got) != 1 || got[0].Name != "Hypothesis-Testing" {
		t.Fatalf("expected hypothesis pattern, got %v", got)
	}
}

func TestGenerateSynthesisIncludesBranches(t *testing.T) {
	t.Parallel()

	insights := []string{"Main conclusion"}
	branchInsights := map[string][]string{
		"branch-12345678": {"Alternative conclusion"},
	}

	got := generateSynthesis(insights, branchInsights)
	if !strings.Contains(got, "Main Branch Conclusions") || !strings.Contains(got, "Alternative Branch Insights") {
		t.Fatalf("unexpected synthesis:\n%s", got)
	}
}

func TestCalculateMergeConfidence(t *testing.T) {
	t.Parallel()

	if got := calculateMergeConfidence([]string{"a", "b", "c"}, nil); got <= 0.7 {
		t.Fatalf("expected confidence boost for multiple insights, got %v", got)
	}
	if got := calculateMergeConfidence([]string{}, []string{"conflict"}); got >= 0.7 {
		t.Fatalf("expected confidence drop for conflicts, got %v", got)
	}
}

func TestDetectLogicalIssues(t *testing.T) {
	t.Parallel()

	session := &thinking.ThinkingSession{
		Steps: []*thinking.ThinkingStep{
			{Number: 1, Type: thinking.StepHypothesis},
			{Number: 2, Type: thinking.StepConclusion},
		},
	}
	issues := detectLogicalIssues(session, 1, 2)
	if len(issues) != 2 {
		t.Fatalf("expected two issues (unsupported conclusion and unverified hypothesis), got %v", issues)
	}
}

func TestGenerateValidationSuggestionsNoIssues(t *testing.T) {
	t.Parallel()

	got := generateValidationSuggestions(nil)
	if len(got) != 1 || !strings.Contains(got[0], "Reasoning appears sound") {
		t.Fatalf("unexpected suggestions %v", got)
	}
}

func TestCalculateValidityScore(t *testing.T) {
	t.Parallel()

	issues := []thinking.LogicalIssue{{}, {}, {}}
	got := calculateValidityScore(issues, 5)
	if got != 1.0-0.45 {
		t.Fatalf("unexpected validity score: %v", got)
	}
	if zero := calculateValidityScore(nil, 0); zero != 0.5 {
		t.Fatalf("expected default score 0.5 for zero steps, got %v", zero)
	}
}

func TestIdentifyStrongPoints(t *testing.T) {
	t.Parallel()

	session := &thinking.ThinkingSession{
		Steps: []*thinking.ThinkingStep{
			{Connections: []int{2}, Type: thinking.StepAnalysis},
			{Connections: []int{1}, Type: thinking.StepHypothesis},
			{Connections: []int{2}, Type: thinking.StepConclusion},
		},
	}
	got := identifyStrongPoints(session, 1, 3)
	if len(got) < 2 {
		t.Fatalf("expected multiple strengths, got %v", got)
	}
}

func TestGenerateAssessment(t *testing.T) {
	t.Parallel()

	issues := []thinking.LogicalIssue{{IssueType: "Test"}}
	strengths := []string{"Strong linkage"}
	got := generateAssessment(issues, 0.65, strengths)
	if !strings.Contains(got, "Validity Score: 0.65") || !strings.Contains(got, "Strengths") {
		t.Fatalf("unexpected assessment:\n%s", got)
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	if got := truncate("short", 10); got != "short" {
		t.Fatalf("expected string unchanged, got %s", got)
	}
	if got := truncate("this is long", 4); got != "this..." {
		t.Fatalf("expected truncated string, got %s", got)
	}
}

func TestCompletionHandlerSuggestsSessions(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	session := mustCreateSessionWithTags(t, store, "Completion test", nil)
	handler := CompletionHandler(store)

	req := &mcp.CompleteRequest{
		Params: &mcp.CompleteParams{
			Ref: &mcp.CompleteReference{
				Type: "ref/resource",
				URI:  "thinking://session/",
			},
		},
	}
	res, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("completion handler returned error: %v", err)
	}
	if len(res.Completion.Values) == 0 || res.Completion.Values[0] != session.ID {
		t.Fatalf("expected completion values with session id, got %+v", res.Completion.Values)
	}

	req.Params.Ref.URI = "thinking://template/"
	res, err = handler(context.Background(), req)
	if err != nil {
		t.Fatalf("completion handler template returned error: %v", err)
	}
	if len(res.Completion.Values) == 0 {
		t.Fatalf("expected template names, got %v", res.Completion.Values)
	}

	req.Params.Ref.Type = "ref/prompt"
	res, err = handler(context.Background(), req)
	if err != nil {
		t.Fatalf("completion handler prompt returned error: %v", err)
	}
	if !contains(res.Completion.Values, "problem_breakdown") {
		t.Fatalf("expected prompt completions, got %v", res.Completion.Values)
	}
}

func TestRegisterFunctionsDoNotPanic(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	server := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "0.0.1"},
		nil,
	)

	RegisterTools(server, store)
	RegisterResources(server, store)
	RegisterPrompts(server)
}

// Helpers --------------------------------------------------------------------

func ptr[T any](v T) *T {
	return &v
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func TestMarshalExportSessionJSON(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	session := mustCreateSessionWithTags(t, store, "Export JSON", []string{"ops"})
	mustAddStep(t, store, session.ID, "Collect requirements", thinking.StepAnalysis, nil, nil)

	content := thinking.ExportToJSON(session, true)
	if !strings.Contains(content, "\"problem\": \"Export JSON\"") {
		t.Fatalf("expected json export to include problem, got %s", content)
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(content), &decoded); err != nil {
		t.Fatalf("expected json export to be valid JSON: %v", err)
	}
}

func TestExportSessionTextWraps(t *testing.T) {
	t.Parallel()

	session := &thinking.ThinkingSession{
		ID:           "test-id",
		Problem:      "Text export problem",
		Status:       "active",
		QualityScore: 0.75,
		Steps: []*thinking.ThinkingStep{
			{
				Number:    1,
				Type:      thinking.StepAnalysis,
				Content:   strings.Repeat("word ", 20),
				Timestamp: time.Now(),
			},
		},
	}
	content := thinking.ExportToText(session, false)
	if !strings.Contains(content, "Text export problem") {
		t.Fatalf("expected text export to include problem, got %s", content)
	}
	if !strings.Contains(content, "\n") {
		t.Fatalf("expected text export to include wrapped lines")
	}
}

func mustCreateSession(t *testing.T, store *thinking.MemoryStore, problem string) *thinking.ThinkingSession {
	t.Helper()

	return mustCreateSessionWithTags(t, store, problem, nil)
}

func mustCreateSessionWithTags(t *testing.T, store *thinking.MemoryStore, problem string, tags []string) *thinking.ThinkingSession {
	t.Helper()

	session, err := store.CreateSession(problem, nil, tags)
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}
	return session
}

func mustAddStep(t *testing.T, store *thinking.MemoryStore, sessionID, content string, stepType thinking.StepType, parent *int, metadata map[string]any) *thinking.ThinkingStep {
	t.Helper()

	step, err := store.AddStep(sessionID, content, stepType, parent, metadata)
	if err != nil {
		t.Fatalf("AddStep returned error: %v", err)
	}
	return step
}

func mustAddBranchStep(t *testing.T, store *thinking.MemoryStore, sessionID, branchID, content string, stepType thinking.StepType, parent *int, metadata map[string]any) *thinking.ThinkingStep {
	t.Helper()

	step, err := store.AddStepToBranch(sessionID, branchID, content, stepType, parent, metadata)
	if err != nil {
		t.Fatalf("AddStepToBranch returned error: %v", err)
	}
	return step
}

func mustGetSession(t *testing.T, store *thinking.MemoryStore, sessionID string) *thinking.ThinkingSession {
	t.Helper()

	session, err := store.GetSession(sessionID)
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}
	return session
}
