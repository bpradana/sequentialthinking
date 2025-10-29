package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Prompt handler: problem_breakdown
func createProblemBreakdownPromptHandler() mcp.PromptHandler {
	return func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		problemStatement := req.Params.Arguments["problem_statement"]
		domain := req.Params.Arguments["domain"]

		domainContext := ""
		if domain != "" {
			domainContext = fmt.Sprintf(" in the %s domain", domain)
		}

		prompt := fmt.Sprintf(`You are helping break down a complex problem%s into manageable components for systematic analysis.

Problem Statement:
%s

Please help structure the thinking process by:

1. **Understanding the Core Question**
   - What is the fundamental question being asked?
   - What would constitute a successful answer?
   - What are we trying to achieve or understand?

2. **Identifying Key Components**
   - What are the main parts or aspects of this problem?
   - Which components are interdependent?
   - Which can be analyzed separately?

3. **Recognizing Constraints**
   - What limitations or boundaries exist?
   - What resources are available or unavailable?
   - What are the time, scope, or quality constraints?

4. **Uncovering Assumptions**
   - What assumptions are we making?
   - Which assumptions are critical vs. peripheral?
   - Which assumptions should be tested or validated?

5. **Determining Information Needs**
   - What information do we have?
   - What information do we need?
   - Where can we obtain missing information?

6. **Suggesting Initial Approaches**
   - What methods or frameworks might apply?
   - What similar problems have been solved before?
   - What would be a logical starting point?

Please provide a structured breakdown following this framework.`, domainContext, problemStatement)

		return &mcp.GetPromptResult{
			Description: "Guide for breaking down complex problems into analyzable components",
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: prompt},
				},
			},
		}, nil
	}
}

// Prompt handler: critical_analysis
func createCriticalAnalysisPromptHandler() mcp.PromptHandler {
	return func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		claim := req.Params.Arguments["claim"]
		evidence := req.Params.Arguments["evidence"]

		evidenceSection := ""
		if evidence != "" {
			evidenceSection = fmt.Sprintf(`

Supporting Evidence Provided:
%s`, evidence)
		}

		prompt := fmt.Sprintf(`You are conducting a critical analysis of an argument or claim. Apply rigorous logical thinking to evaluate its validity.

Claim to Analyze:
%s%s

Please conduct a systematic critical analysis:

1. **Clarity and Precision**
   - Is the claim clearly stated and unambiguous?
   - Are key terms well-defined?
   - Could the claim be interpreted in multiple ways?

2. **Logical Structure**
   - What type of argument is being made? (deductive, inductive, analogical, causal)
   - Are the premises clearly stated?
   - Does the conclusion follow logically from the premises?
   - Are there any logical fallacies present?

3. **Evidence Evaluation**
   - What evidence supports the claim?
   - Is the evidence relevant, sufficient, and reliable?
   - What is the quality and source of the evidence?
   - What evidence would be needed to strengthen the claim?
   - What evidence might contradict the claim?

4. **Alternative Perspectives**
   - What are possible counter-arguments?
   - What assumptions underlie the claim?
   - Are there alternative explanations?
   - What would someone who disagrees argue?

5. **Implications and Consequences**
   - If the claim is true, what follows?
   - Are there unintended consequences?
   - What practical implications does this have?

6. **Bias and Context**
   - Are there potential sources of bias?
   - What context is necessary to evaluate this claim?
   - What might be missing from the analysis?

7. **Overall Assessment**
   - What is the strength of the argument?
   - What are the strongest points?
   - What are the weakest points?
   - What modifications would improve the argument?

Please provide a thorough critical analysis following this framework.`, claim, evidenceSection)

		return &mcp.GetPromptResult{
			Description: "Guide for critical evaluation of arguments and claims",
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: prompt},
				},
			},
		}, nil
	}
}

// Prompt handler: synthesis_prompt
func createSynthesisPromptHandler() mcp.PromptHandler {
	return func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		insightsJSON := req.Params.Arguments["insights"]
		goal := req.Params.Arguments["goal"]

		// Try to parse insights as JSON array
		var insights []string
		if err := json.Unmarshal([]byte(insightsJSON), &insights); err != nil {
			// If not valid JSON, treat as single insight
			insights = []string{insightsJSON}
		}

		insightsList := ""
		for i, insight := range insights {
			insightsList += fmt.Sprintf("\n%d. %s", i+1, insight)
		}

		prompt := fmt.Sprintf(`You are synthesizing multiple insights or perspectives into a coherent understanding.

Goal of Synthesis:
%s

Insights to Synthesize:%s

Please create a comprehensive synthesis:

1. **Identify Common Themes**
   - What patterns emerge across the insights?
   - What core ideas are repeated or emphasized?
   - What fundamental principles connect these insights?

2. **Recognize Differences**
   - Where do the insights diverge or conflict?
   - Are these genuine contradictions or different perspectives?
   - Can differences be reconciled or are they fundamental?

3. **Evaluate Relative Strength**
   - Which insights are best supported?
   - Which are more speculative?
   - What is the confidence level of each?

4. **Build Connections**
   - How do these insights relate to each other?
   - Can they be organized in a hierarchy or framework?
   - What causal or logical relationships exist?

5. **Integration**
   - What unified understanding emerges?
   - How can these insights be combined coherently?
   - What is the "big picture" that incorporates all insights?

6. **Identify Gaps**
   - What questions remain unanswered?
   - What additional information would strengthen the synthesis?
   - What are the limitations of the current synthesis?

7. **Actionable Conclusions**
   - What can we conclude with confidence?
   - What are the practical implications?
   - What should be done next?

8. **Quality Check**
   - Is the synthesis internally consistent?
   - Does it respect the nuances of each insight?
   - Is it more valuable than the sum of its parts?

Please provide a structured synthesis that addresses these dimensions.`, goal, insightsList)

		return &mcp.GetPromptResult{
			Description: "Guide for combining multiple insights into coherent understanding",
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: prompt},
				},
			},
		}, nil
	}
}
