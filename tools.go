package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool handler: start_thinking
func createStartThinkingHandler(store *MemoryStore) func(context.Context, *mcp.CallToolRequest, StartThinkingInput) (*mcp.CallToolResult, StartThinkingOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input StartThinkingInput) (*mcp.CallToolResult, StartThinkingOutput, error) {
		session, err := store.CreateSession(input.Problem, input.Context, input.Tags)
		if err != nil {
			return nil, StartThinkingOutput{}, err
		}

		suggestedSteps := []string{
			"Break down the problem into smaller components",
			"Identify key assumptions and constraints",
			"Consider multiple perspectives or approaches",
			"Gather relevant information or evidence",
		}

		output := StartThinkingOutput{
			SessionID:       session.ID,
			InitialAnalysis: session.InitialAnalysis,
			SuggestedSteps:  suggestedSteps,
		}

		return nil, output, nil
	}
}

// Tool handler: add_step
func createAddStepHandler(store *MemoryStore) func(context.Context, *mcp.CallToolRequest, AddStepInput) (*mcp.CallToolResult, AddStepOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AddStepInput) (*mcp.CallToolResult, AddStepOutput, error) {
		step, err := store.AddStep(input.SessionID, input.StepContent, input.StepType, input.ParentStep, input.Metadata)
		if err != nil {
			return nil, AddStepOutput{}, err
		}

		session, err := store.GetSession(input.SessionID)
		if err != nil {
			return nil, AddStepOutput{}, err
		}

		progress := generateProgress(session)
		nextSteps := suggestNextSteps(session, step)

		output := AddStepOutput{
			StepNumber:         step.Number,
			CurrentProgress:    progress,
			SuggestedNextSteps: nextSteps,
			QualityScore:       session.QualityScore,
		}

		return nil, output, nil
	}
}

// Tool handler: review_thinking
func createReviewThinkingHandler(store *MemoryStore) func(context.Context, *mcp.CallToolRequest, ReviewThinkingInput) (*mcp.CallToolResult, ReviewThinkingOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ReviewThinkingInput) (*mcp.CallToolResult, ReviewThinkingOutput, error) {
		session, err := store.GetSession(input.SessionID)
		if err != nil {
			return nil, ReviewThinkingOutput{}, err
		}

		format := input.Format
		if format == "" {
			format = "linear"
		}

		connections := buildConnectionsMap(session)
		summary := generateSummary(session, format)
		patterns := detectSessionPatterns(session)

		output := ReviewThinkingOutput{
			Steps:        session.Steps,
			Connections:  connections,
			QualityScore: session.QualityScore,
			Summary:      summary,
			Branches:     session.Branches,
			Patterns:     patterns,
		}

		return nil, output, nil
	}
}

// Tool handler: branch_thinking
func createBranchThinkingHandler(store *MemoryStore) func(context.Context, *mcp.CallToolRequest, BranchThinkingInput) (*mcp.CallToolResult, BranchThinkingOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BranchThinkingInput) (*mcp.CallToolResult, BranchThinkingOutput, error) {
		branch, err := store.CreateBranch(input.SessionID, input.FromStep, input.AlternativeReasoning)
		if err != nil {
			return nil, BranchThinkingOutput{}, err
		}

		// Add the first step to the branch
		_, err = store.AddStepToBranch(input.SessionID, branch.ID, input.AlternativeReasoning, StepHypothesis)
		if err != nil {
			return nil, BranchThinkingOutput{}, err
		}

		summary := fmt.Sprintf("Created alternative reasoning path from step %d: %s", input.FromStep, input.AlternativeReasoning)

		output := BranchThinkingOutput{
			BranchID:        branch.ID,
			DivergencePoint: input.FromStep,
			BranchSummary:   summary,
		}

		return nil, output, nil
	}
}

// Tool handler: merge_insights
func createMergeInsightsHandler(store *MemoryStore) func(context.Context, *mcp.CallToolRequest, MergeInsightsInput) (*mcp.CallToolResult, MergeInsightsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input MergeInsightsInput) (*mcp.CallToolResult, MergeInsightsOutput, error) {
		session, err := store.GetSession(input.SessionID)
		if err != nil {
			return nil, MergeInsightsOutput{}, err
		}

		// Collect insights from main and branches
		insights := []string{}
		conflicts := []string{}

		// Main branch insights
		for _, step := range session.Steps {
			if step.Type == StepConclusion {
				insights = append(insights, step.Content)
			}
		}

		// Branch insights
		branchInsights := make(map[string][]string)
		for _, branchID := range input.BranchIDs {
			branch, exists := session.Branches[branchID]
			if !exists {
				continue
			}
			branchInsights[branchID] = []string{}
			for _, step := range branch.Steps {
				if step.Type == StepConclusion {
					branchInsights[branchID] = append(branchInsights[branchID], step.Content)
				}
			}
		}

		// Detect conflicts (simplified)
		if len(branchInsights) > 1 {
			conflicts = append(conflicts, "Multiple conclusion paths exist - careful synthesis required")
		}

		// Generate synthesis
		synthesis := generateSynthesis(insights, branchInsights)
		confidence := calculateMergeConfidence(insights, conflicts)
		strengths := []string{
			"Multiple perspectives considered",
			"Alternative approaches explored",
			"Conclusions validated across branches",
		}

		output := MergeInsightsOutput{
			Synthesis:  synthesis,
			Conflicts:  conflicts,
			Confidence: confidence,
			Strengths:  strengths,
		}

		return nil, output, nil
	}
}

// Tool handler: validate_logic
func createValidateLogicHandler(store *MemoryStore) func(context.Context, *mcp.CallToolRequest, ValidateLogicInput) (*mcp.CallToolResult, ValidateLogicOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ValidateLogicInput) (*mcp.CallToolResult, ValidateLogicOutput, error) {
		session, err := store.GetSession(input.SessionID)
		if err != nil {
			return nil, ValidateLogicOutput{}, err
		}

		// Determine range
		start := 1
		end := len(session.Steps)
		if input.RangeStart != nil {
			start = *input.RangeStart
		}
		if input.RangeEnd != nil {
			end = *input.RangeEnd
		}

		issues := detectLogicalIssues(session, start, end)
		suggestions := generateValidationSuggestions(issues)
		validityScore := calculateValidityScore(issues, end-start+1)
		strongPoints := identifyStrongPoints(session, start, end)
		assessment := generateAssessment(issues, validityScore, strongPoints)

		output := ValidateLogicOutput{
			Issues:            issues,
			Suggestions:       suggestions,
			ValidityScore:     validityScore,
			StrongPoints:      strongPoints,
			OverallAssessment: assessment,
		}

		return nil, output, nil
	}
}

// Tool handler: export_session
func createExportSessionHandler(store *MemoryStore) func(context.Context, *mcp.CallToolRequest, ExportSessionInput) (*mcp.CallToolResult, ExportSessionOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ExportSessionInput) (*mcp.CallToolResult, ExportSessionOutput, error) {
		session, err := store.GetSession(input.SessionID)
		if err != nil {
			return nil, ExportSessionOutput{}, err
		}

		format := input.Format
		if format == "" {
			format = "markdown"
		}

		var content string
		switch format {
		case "markdown":
			content = exportToMarkdown(session, input.IncludeBranches)
		case "json":
			content = exportToJSON(session, input.IncludeBranches)
		case "text":
			content = exportToText(session, input.IncludeBranches)
		default:
			return nil, ExportSessionOutput{}, fmt.Errorf("unsupported format: %s", format)
		}

		filename := fmt.Sprintf("thinking_%s.%s", session.ID[:8], format)
		if format == "markdown" {
			filename = fmt.Sprintf("thinking_%s.md", session.ID[:8])
		}

		output := ExportSessionOutput{
			Content:  content,
			Format:   format,
			Filename: filename,
		}

		return nil, output, nil
	}
}

// Tool handler: list_sessions
func createListSessionsHandler(store *MemoryStore) func(context.Context, *mcp.CallToolRequest, ListSessionsInput) (*mcp.CallToolResult, ListSessionsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListSessionsInput) (*mcp.CallToolResult, ListSessionsOutput, error) {
		allSessions := store.ListSessions()

		// Filter by status
		var filtered []*ThinkingSession
		for _, s := range allSessions {
			if input.Status != "" && input.Status != "all" && s.Status != input.Status {
				continue
			}

			// Filter by tags
			if len(input.Tags) > 0 {
				hasTag := false
				for _, tag := range input.Tags {
					for _, sTag := range s.Tags {
						if tag == sTag {
							hasTag = true
							break
						}
					}
				}
				if !hasTag {
					continue
				}
			}

			filtered = append(filtered, s)
		}

		// Apply limit
		limit := input.Limit
		if limit == 0 || limit > len(filtered) {
			limit = len(filtered)
		}

		summaries := make([]SessionSummary, 0, limit)
		for i := 0; i < limit; i++ {
			s := filtered[i]
			summaries = append(summaries, SessionSummary{
				ID:           s.ID,
				Problem:      s.Problem,
				StepCount:    len(s.Steps),
				BranchCount:  len(s.Branches),
				Status:       s.Status,
				QualityScore: s.QualityScore,
				Created:      s.Created,
				LastModified: s.LastModified,
				Tags:         s.Tags,
			})
		}

		output := ListSessionsOutput{
			Sessions: summaries,
			Total:    len(filtered),
		}

		return nil, output, nil
	}
}

// Tool handler: delete_session
func createDeleteSessionHandler(store *MemoryStore) func(context.Context, *mcp.CallToolRequest, DeleteSessionInput) (*mcp.CallToolResult, DeleteSessionOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteSessionInput) (*mcp.CallToolResult, DeleteSessionOutput, error) {
		err := store.DeleteSession(input.SessionID)
		if err != nil {
			return nil, DeleteSessionOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to delete session: %v", err),
			}, nil
		}

		output := DeleteSessionOutput{
			Success: true,
			Message: fmt.Sprintf("Session %s deleted successfully", input.SessionID),
		}

		return nil, output, nil
	}
}

// Tool handler: get_metrics
func createGetMetricsHandler(store *MemoryStore) func(context.Context, *mcp.CallToolRequest, GetMetricsInput) (*mcp.CallToolResult, GetMetricsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetMetricsInput) (*mcp.CallToolResult, GetMetricsOutput, error) {
		timeRange := input.TimeRange
		if timeRange == "" {
			timeRange = "all"
		}

		metrics := store.GetMetrics(timeRange)

		output := GetMetricsOutput{
			Metrics: metrics,
		}

		return nil, output, nil
	}
}

// Helper functions

func generateProgress(session *ThinkingSession) string {
	return fmt.Sprintf("Session has %d steps. Current quality score: %.2f. Status: %s",
		len(session.Steps), session.QualityScore, session.Status)
}

func suggestNextSteps(session *ThinkingSession, lastStep *ThinkingStep) []string {
	suggestions := []string{}

	switch lastStep.Type {
	case StepAnalysis:
		suggestions = append(suggestions, "Form a hypothesis based on the analysis")
		suggestions = append(suggestions, "Identify additional factors to analyze")
	case StepHypothesis:
		suggestions = append(suggestions, "Design verification approach")
		suggestions = append(suggestions, "Consider alternative hypotheses")
	case StepVerification:
		suggestions = append(suggestions, "Draw conclusions from verification")
		suggestions = append(suggestions, "Identify limitations of verification")
	case StepConclusion:
		suggestions = append(suggestions, "Summarize key insights")
		suggestions = append(suggestions, "Identify next actions or implications")
	}

	return suggestions
}

func buildConnectionsMap(session *ThinkingSession) map[string][]int {
	connections := make(map[string][]int)
	for _, step := range session.Steps {
		key := fmt.Sprintf("step_%d", step.Number)
		connections[key] = step.Connections
	}
	return connections
}

func generateSummary(session *ThinkingSession, format string) string {
	var sb strings.Builder

	sb.WriteString("Thinking Session Summary\n")
	sb.WriteString(fmt.Sprintf("Problem: %s\n", session.Problem))
	sb.WriteString(fmt.Sprintf("Steps: %d | Branches: %d | Quality: %.2f\n\n",
		len(session.Steps), len(session.Branches), session.QualityScore))

	if format == "tree" {
		sb.WriteString("Reasoning Tree:\n")
		for _, step := range session.Steps {
			indent := ""
			if step.ParentStep != nil {
				indent = "  "
			}
			sb.WriteString(fmt.Sprintf("%s%d. [%s] %s\n", indent, step.Number, step.Type, truncate(step.Content, 60)))
		}
	} else {
		sb.WriteString("Key Steps:\n")
		for _, step := range session.Steps {
			if step.Type == StepConclusion || step.Type == StepHypothesis {
				sb.WriteString(fmt.Sprintf("- Step %d: %s\n", step.Number, truncate(step.Content, 80)))
			}
		}
	}

	return sb.String()
}

func detectSessionPatterns(session *ThinkingSession) []ThinkingPattern {
	patterns := []ThinkingPattern{}

	// Check for hypothesis-verification pattern
	hasHypothesis := false
	hasVerification := false
	for _, step := range session.Steps {
		if step.Type == StepHypothesis {
			hasHypothesis = true
		}
		if step.Type == StepVerification {
			hasVerification = true
		}
	}
	if hasHypothesis && hasVerification {
		patterns = append(patterns, ThinkingPattern{
			Name:        "Hypothesis-Testing",
			Confidence:  0.9,
			Description: "Uses scientific method approach",
		})
	}

	return patterns
}

func generateSynthesis(insights []string, branchInsights map[string][]string) string {
	var sb strings.Builder
	sb.WriteString("Synthesized Insights:\n\n")
	sb.WriteString("Main Branch Conclusions:\n")
	for _, insight := range insights {
		sb.WriteString(fmt.Sprintf("- %s\n", insight))
	}

	if len(branchInsights) > 0 {
		sb.WriteString("\nAlternative Branch Insights:\n")
		for branchID, bis := range branchInsights {
			sb.WriteString(fmt.Sprintf("\nBranch %s:\n", branchID[:8]))
			for _, insight := range bis {
				sb.WriteString(fmt.Sprintf("- %s\n", insight))
			}
		}
	}

	sb.WriteString("\nOverall: Multiple perspectives have been considered and integrated.")
	return sb.String()
}

func calculateMergeConfidence(insights []string, conflicts []string) float64 {
	base := 0.7
	if len(insights) > 2 {
		base += 0.1
	}
	if len(conflicts) > 0 {
		base -= 0.2
	}
	if base < 0 {
		base = 0
	}
	if base > 1 {
		base = 1
	}
	return base
}

func detectLogicalIssues(session *ThinkingSession, start, end int) []LogicalIssue {
	issues := []LogicalIssue{}

	// Check for unsupported conclusions
	for i := start - 1; i < end && i < len(session.Steps); i++ {
		step := session.Steps[i]
		if step.Type == StepConclusion && len(step.Connections) == 0 {
			issues = append(issues, LogicalIssue{
				StepNumber:  step.Number,
				IssueType:   "Unsupported Conclusion",
				Description: "Conclusion drawn without clear supporting evidence",
				Severity:    "medium",
				Suggestion:  "Link this conclusion to previous analysis or verification steps",
			})
		}
	}

	// Check for missing verification
	hasHypothesis := false
	hasVerification := false
	for i := start - 1; i < end && i < len(session.Steps); i++ {
		step := session.Steps[i]
		if step.Type == StepHypothesis {
			hasHypothesis = true
		}
		if step.Type == StepVerification {
			hasVerification = true
		}
	}
	if hasHypothesis && !hasVerification {
		issues = append(issues, LogicalIssue{
			StepNumber:  0,
			IssueType:   "Unverified Hypothesis",
			Description: "Hypotheses formed but not verified",
			Severity:    "high",
			Suggestion:  "Add verification steps to test hypotheses",
		})
	}

	return issues
}

func generateValidationSuggestions(issues []LogicalIssue) []string {
	suggestions := []string{}
	for _, issue := range issues {
		suggestions = append(suggestions, issue.Suggestion)
	}
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Reasoning appears sound - continue developing the argument")
	}
	return suggestions
}

func calculateValidityScore(issues []LogicalIssue, stepCount int) float64 {
	if stepCount == 0 {
		return 0.5
	}

	penaltyPerIssue := 0.15
	score := 1.0 - (float64(len(issues)) * penaltyPerIssue)
	if score < 0 {
		score = 0
	}
	return score
}

func identifyStrongPoints(session *ThinkingSession, start, end int) []string {
	strong := []string{}

	// Check for well-connected reasoning
	wellConnected := 0
	for i := start - 1; i < end && i < len(session.Steps); i++ {
		step := session.Steps[i]
		if len(step.Connections) > 0 {
			wellConnected++
		}
	}
	if wellConnected > (end-start+1)/2 {
		strong = append(strong, "Steps are well-connected and build on each other")
	}

	// Check for diverse step types
	typeMap := make(map[StepType]bool)
	for i := start - 1; i < end && i < len(session.Steps); i++ {
		typeMap[session.Steps[i].Type] = true
	}
	if len(typeMap) >= 3 {
		strong = append(strong, "Uses diverse reasoning approaches")
	}

	return strong
}

func generateAssessment(issues []LogicalIssue, score float64, strengths []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Validity Score: %.2f/1.00\n\n", score))

	if score >= 0.8 {
		sb.WriteString("Overall: Strong logical reasoning with minor or no issues.\n")
	} else if score >= 0.6 {
		sb.WriteString("Overall: Good reasoning with some areas for improvement.\n")
	} else {
		sb.WriteString("Overall: Reasoning needs strengthening in several areas.\n")
	}

	if len(issues) > 0 {
		sb.WriteString(fmt.Sprintf("\n%d logical issues identified.\n", len(issues)))
	}

	if len(strengths) > 0 {
		sb.WriteString("\nStrengths:\n")
		for _, s := range strengths {
			sb.WriteString(fmt.Sprintf("- %s\n", s))
		}
	}

	return sb.String()
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}
