package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bpradana/sequentialthinking/internal/thinking"
)

// Tool handler: start_thinking
func createStartThinkingHandler(store *thinking.MemoryStore) func(context.Context, *mcp.CallToolRequest, thinking.StartThinkingInput) (*mcp.CallToolResult, thinking.StartThinkingOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input thinking.StartThinkingInput) (*mcp.CallToolResult, thinking.StartThinkingOutput, error) {
		templateID := strings.TrimSpace(input.Template)
		problem := strings.TrimSpace(input.Problem)

		var contextMap map[string]any
		if len(input.Context) > 0 {
			contextMap = make(map[string]any, len(input.Context))
			for k, v := range input.Context {
				contextMap[k] = v
			}
		}

		suggestedSteps := []string{
			"Break down the problem into smaller components",
			"Identify key assumptions and constraints",
			"Consider multiple perspectives or approaches",
			"Gather relevant information or evidence",
		}

		var templateData *thinking.Template
		if templateID != "" {
			templateData = getTemplate(templateID)
			if templateData == nil {
				return nil, thinking.StartThinkingOutput{}, fmt.Errorf("template %q not found", templateID)
			}

			if contextMap == nil {
				contextMap = make(map[string]any, len(input.Context)+2)
			}
			contextMap["template_type"] = templateData.Type
			contextMap["template_name"] = templateData.Name

			suggestedSteps = make([]string, len(templateData.Steps))
			copy(suggestedSteps, templateData.Steps)
		}

		session, err := store.CreateSession(problem, contextMap, input.Tags)
		if err != nil {
			return nil, thinking.StartThinkingOutput{}, err
		}

		output := thinking.StartThinkingOutput{
			SessionID:       session.ID,
			InitialAnalysis: session.InitialAnalysis,
			SuggestedSteps:  suggestedSteps,
		}

		return nil, output, nil
	}
}

// Tool handler: add_step
func createAddStepHandler(store *thinking.MemoryStore) func(context.Context, *mcp.CallToolRequest, thinking.AddStepInput) (*mcp.CallToolResult, thinking.AddStepOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input thinking.AddStepInput) (*mcp.CallToolResult, thinking.AddStepOutput, error) {
		var (
			step     *thinking.ThinkingStep
			err      error
			branchID = strings.TrimSpace(input.BranchID)
		)

		if branchID != "" {
			step, err = store.AddStepToBranch(input.SessionID, branchID, input.StepContent, input.StepType, input.ParentStep, input.Metadata)
		} else {
			step, err = store.AddStep(input.SessionID, input.StepContent, input.StepType, input.ParentStep, input.Metadata)
		}
		if err != nil {
			return nil, thinking.AddStepOutput{}, err
		}

		session, err := store.GetSession(input.SessionID)
		if err != nil {
			return nil, thinking.AddStepOutput{}, err
		}

		var targetStep *thinking.ThinkingStep
		if branchID != "" {
			branch, exists := session.Branches[branchID]
			if !exists || len(branch.Steps) < step.Number {
				return nil, thinking.AddStepOutput{}, fmt.Errorf("branch %s not found in session after update", branchID)
			}
			targetStep = branch.Steps[step.Number-1]
		} else {
			targetStep = step
		}

		progress := generateProgress(session)
		nextSteps := suggestNextSteps(session, targetStep)

		output := thinking.AddStepOutput{
			StepNumber:         step.Number,
			BranchID:           branchID,
			CurrentProgress:    progress,
			SuggestedNextSteps: nextSteps,
			QualityScore:       session.QualityScore,
		}

		return nil, output, nil
	}
}

// Tool handler: update_step
func createUpdateStepHandler(store *thinking.MemoryStore) func(context.Context, *mcp.CallToolRequest, thinking.UpdateStepInput) (*mcp.CallToolResult, thinking.UpdateStepOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input thinking.UpdateStepInput) (*mcp.CallToolResult, thinking.UpdateStepOutput, error) {
		if input.StepNumber <= 0 {
			return nil, thinking.UpdateStepOutput{}, fmt.Errorf("step_number must be greater than zero")
		}
		if input.StepContent == nil && input.StepType == nil && input.Metadata == nil {
			return nil, thinking.UpdateStepOutput{}, fmt.Errorf("at least one of step_content, step_type, or metadata must be provided")
		}

		_, err := store.UpdateStep(input.SessionID, input.StepNumber, input.StepContent, input.StepType, input.Metadata)
		if err != nil {
			return nil, thinking.UpdateStepOutput{}, err
		}

		session, err := store.GetSession(input.SessionID)
		if err != nil {
			return nil, thinking.UpdateStepOutput{}, err
		}
		if input.StepNumber > len(session.Steps) {
			return nil, thinking.UpdateStepOutput{}, fmt.Errorf("step %d not found after update", input.StepNumber)
		}

		updatedStep := session.Steps[input.StepNumber-1]
		progress := generateProgress(session)
		nextSteps := suggestNextSteps(session, updatedStep)

		output := thinking.UpdateStepOutput{
			StepNumber:         updatedStep.Number,
			UpdatedStep:        updatedStep,
			CurrentProgress:    progress,
			SuggestedNextSteps: nextSteps,
			QualityScore:       session.QualityScore,
			LastModified:       session.LastModified,
			MetadataChanged:    input.Metadata != nil,
			TypeChanged:        input.StepType != nil,
			ContentChanged:     input.StepContent != nil,
		}

		return nil, output, nil
	}
}

// Tool handler: review_thinking
func createReviewThinkingHandler(store *thinking.MemoryStore) func(context.Context, *mcp.CallToolRequest, thinking.ReviewThinkingInput) (*mcp.CallToolResult, thinking.ReviewThinkingOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input thinking.ReviewThinkingInput) (*mcp.CallToolResult, thinking.ReviewThinkingOutput, error) {
		session, err := store.GetSession(input.SessionID)
		if err != nil {
			return nil, thinking.ReviewThinkingOutput{}, err
		}

		format := input.Format
		if format == "" {
			format = "linear"
		}

		connections := buildConnectionsMap(session)
		summary := generateSummary(session, format)
		patterns := detectSessionPatterns(session)

		output := thinking.ReviewThinkingOutput{
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
func createBranchThinkingHandler(store *thinking.MemoryStore) func(context.Context, *mcp.CallToolRequest, thinking.BranchThinkingInput) (*mcp.CallToolResult, thinking.BranchThinkingOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input thinking.BranchThinkingInput) (*mcp.CallToolResult, thinking.BranchThinkingOutput, error) {
		branch, err := store.CreateBranch(input.SessionID, input.FromStep, input.AlternativeReasoning)
		if err != nil {
			return nil, thinking.BranchThinkingOutput{}, err
		}

		// Add the first step to the branch
		_, err = store.AddStepToBranch(input.SessionID, branch.ID, input.AlternativeReasoning, thinking.StepHypothesis, nil, nil)
		if err != nil {
			return nil, thinking.BranchThinkingOutput{}, err
		}

		summary := fmt.Sprintf("Created alternative reasoning path from step %d: %s", input.FromStep, input.AlternativeReasoning)

		output := thinking.BranchThinkingOutput{
			BranchID:        branch.ID,
			DivergencePoint: input.FromStep,
			BranchSummary:   summary,
		}

		return nil, output, nil
	}
}

// Tool handler: merge_insights
func createMergeInsightsHandler(store *thinking.MemoryStore) func(context.Context, *mcp.CallToolRequest, thinking.MergeInsightsInput) (*mcp.CallToolResult, thinking.MergeInsightsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input thinking.MergeInsightsInput) (*mcp.CallToolResult, thinking.MergeInsightsOutput, error) {
		session, err := store.GetSession(input.SessionID)
		if err != nil {
			return nil, thinking.MergeInsightsOutput{}, err
		}

		// Collect insights from main and branches
		insights := []string{}
		conflicts := []string{}

		// Main branch insights
		for _, step := range session.Steps {
			if step.Type == thinking.StepConclusion {
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
				if step.Type == thinking.StepConclusion {
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

		output := thinking.MergeInsightsOutput{
			Synthesis:  synthesis,
			Conflicts:  conflicts,
			Confidence: confidence,
			Strengths:  strengths,
		}

		return nil, output, nil
	}
}

// Tool handler: validate_logic
func createValidateLogicHandler(store *thinking.MemoryStore) func(context.Context, *mcp.CallToolRequest, thinking.ValidateLogicInput) (*mcp.CallToolResult, thinking.ValidateLogicOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input thinking.ValidateLogicInput) (*mcp.CallToolResult, thinking.ValidateLogicOutput, error) {
		session, err := store.GetSession(input.SessionID)
		if err != nil {
			return nil, thinking.ValidateLogicOutput{}, err
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

		output := thinking.ValidateLogicOutput{
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
func createExportSessionHandler(store *thinking.MemoryStore) func(context.Context, *mcp.CallToolRequest, thinking.ExportSessionInput) (*mcp.CallToolResult, thinking.ExportSessionOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input thinking.ExportSessionInput) (*mcp.CallToolResult, thinking.ExportSessionOutput, error) {
		session, err := store.GetSession(input.SessionID)
		if err != nil {
			return nil, thinking.ExportSessionOutput{}, err
		}

		format := input.Format
		if format == "" {
			format = "markdown"
		}

		var content string
		switch format {
		case "markdown":
			content = thinking.ExportToMarkdown(session, input.IncludeBranches)
		case "json":
			content = thinking.ExportToJSON(session, input.IncludeBranches)
		case "text":
			content = thinking.ExportToText(session, input.IncludeBranches)
		default:
			return nil, thinking.ExportSessionOutput{}, fmt.Errorf("unsupported format: %s", format)
		}

		filename := fmt.Sprintf("thinking_%s.%s", session.ID[:8], format)
		if format == "markdown" {
			filename = fmt.Sprintf("thinking_%s.md", session.ID[:8])
		}

		output := thinking.ExportSessionOutput{
			Content:  content,
			Format:   format,
			Filename: filename,
		}

		return nil, output, nil
	}
}

// Tool handler: list_sessions
func createListSessionsHandler(store *thinking.MemoryStore) func(context.Context, *mcp.CallToolRequest, thinking.ListSessionsInput) (*mcp.CallToolResult, thinking.ListSessionsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input thinking.ListSessionsInput) (*mcp.CallToolResult, thinking.ListSessionsOutput, error) {
		allSessions := store.ListSessions()

		// Filter by status
		var filtered []*thinking.ThinkingSession
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

		summaries := make([]thinking.SessionSummary, 0, limit)
		for i := 0; i < limit; i++ {
			s := filtered[i]
			summaries = append(summaries, thinking.SessionSummary{
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

		output := thinking.ListSessionsOutput{
			Sessions: summaries,
			Total:    len(filtered),
		}

		return nil, output, nil
	}
}

// Tool handler: delete_session
func createDeleteSessionHandler(store *thinking.MemoryStore) func(context.Context, *mcp.CallToolRequest, thinking.DeleteSessionInput) (*mcp.CallToolResult, thinking.DeleteSessionOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input thinking.DeleteSessionInput) (*mcp.CallToolResult, thinking.DeleteSessionOutput, error) {
		err := store.DeleteSession(input.SessionID)
		if err != nil {
			return nil, thinking.DeleteSessionOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to delete session: %v", err),
			}, nil
		}

		output := thinking.DeleteSessionOutput{
			Success: true,
			Message: fmt.Sprintf("Session %s deleted successfully", input.SessionID),
		}

		return nil, output, nil
	}
}

// Tool handler: get_metrics
func createGetMetricsHandler(store *thinking.MemoryStore) func(context.Context, *mcp.CallToolRequest, thinking.GetMetricsInput) (*mcp.CallToolResult, thinking.GetMetricsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input thinking.GetMetricsInput) (*mcp.CallToolResult, thinking.GetMetricsOutput, error) {
		timeRange := input.TimeRange
		if timeRange == "" {
			timeRange = "all"
		}

		metrics := store.GetMetrics(timeRange)

		output := thinking.GetMetricsOutput{
			Metrics: metrics,
		}

		return nil, output, nil
	}
}

// Helper functions

func generateProgress(session *thinking.ThinkingSession) string {
	return fmt.Sprintf("Session has %d steps. Current quality score: %.2f. Status: %s",
		len(session.Steps), session.QualityScore, session.Status)
}

func suggestNextSteps(session *thinking.ThinkingSession, lastStep *thinking.ThinkingStep) []string {
	suggestions := []string{}

	switch lastStep.Type {
	case thinking.StepAnalysis:
		suggestions = append(suggestions, "Form a hypothesis based on the analysis")
		suggestions = append(suggestions, "Identify additional factors to analyze")
	case thinking.StepHypothesis:
		suggestions = append(suggestions, "Design verification approach")
		suggestions = append(suggestions, "Consider alternative hypotheses")
	case thinking.StepVerification:
		suggestions = append(suggestions, "Draw conclusions from verification")
		suggestions = append(suggestions, "Identify limitations of verification")
	case thinking.StepConclusion:
		suggestions = append(suggestions, "Summarize key insights")
		suggestions = append(suggestions, "Identify next actions or implications")
	}

	return suggestions
}

func buildConnectionsMap(session *thinking.ThinkingSession) map[string][]int {
	connections := make(map[string][]int)
	for _, step := range session.Steps {
		key := fmt.Sprintf("step_%d", step.Number)
		connections[key] = append([]int{}, step.Connections...)
	}
	return connections
}

func generateSummary(session *thinking.ThinkingSession, format string) string {
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
			if step.Type == thinking.StepConclusion || step.Type == thinking.StepHypothesis {
				sb.WriteString(fmt.Sprintf("- Step %d: %s\n", step.Number, truncate(step.Content, 80)))
			}
		}
	}

	return sb.String()
}

func detectSessionPatterns(session *thinking.ThinkingSession) []thinking.ThinkingPattern {
	patterns := []thinking.ThinkingPattern{}

	// Check for hypothesis-verification pattern
	hasHypothesis := false
	hasVerification := false
	for _, step := range session.Steps {
		if step.Type == thinking.StepHypothesis {
			hasHypothesis = true
		}
		if step.Type == thinking.StepVerification {
			hasVerification = true
		}
	}
	if hasHypothesis && hasVerification {
		patterns = append(patterns, thinking.ThinkingPattern{
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

func detectLogicalIssues(session *thinking.ThinkingSession, start, end int) []thinking.LogicalIssue {
	issues := []thinking.LogicalIssue{}

	// Check for unsupported conclusions
	for i := start - 1; i < end && i < len(session.Steps); i++ {
		step := session.Steps[i]
		if step.Type == thinking.StepConclusion && len(step.Connections) == 0 {
			issues = append(issues, thinking.LogicalIssue{
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
		if step.Type == thinking.StepHypothesis {
			hasHypothesis = true
		}
		if step.Type == thinking.StepVerification {
			hasVerification = true
		}
	}
	if hasHypothesis && !hasVerification {
		issues = append(issues, thinking.LogicalIssue{
			StepNumber:  0,
			IssueType:   "Unverified Hypothesis",
			Description: "Hypotheses formed but not verified",
			Severity:    "high",
			Suggestion:  "Add verification steps to test hypotheses",
		})
	}

	return issues
}

func generateValidationSuggestions(issues []thinking.LogicalIssue) []string {
	suggestions := []string{}
	for _, issue := range issues {
		suggestions = append(suggestions, issue.Suggestion)
	}
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Reasoning appears sound - continue developing the argument")
	}
	return suggestions
}

func calculateValidityScore(issues []thinking.LogicalIssue, stepCount int) float64 {
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

func identifyStrongPoints(session *thinking.ThinkingSession, start, end int) []string {
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
	typeMap := make(map[thinking.StepType]bool)
	for i := start - 1; i < end && i < len(session.Steps); i++ {
		typeMap[session.Steps[i].Type] = true
	}
	if len(typeMap) >= 3 {
		strong = append(strong, "Uses diverse reasoning approaches")
	}

	return strong
}

func generateAssessment(issues []thinking.LogicalIssue, score float64, strengths []string) string {
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
