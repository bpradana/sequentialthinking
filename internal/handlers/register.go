package handlers

import (
	"context"
	"github.com/bpradana/sequentialthinking/internal/thinking"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterTools wires all tool handlers with the MCP server.
func RegisterTools(server *mcp.Server, store *thinking.MemoryStore) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "start_thinking",
		Description: "Initiate a new thinking session to break down a complex problem",
	}, createStartThinkingHandler(store))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "start_from_template",
		Description: "Initiate a new thinking session using a predefined thinking template",
	}, createStartFromTemplateHandler(store))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_step",
		Description: "Add a reasoning step to an existing thinking session",
	}, createAddStepHandler(store))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_step",
		Description: "Update the content, type, or metadata of an existing reasoning step",
	}, createUpdateStepHandler(store))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "review_thinking",
		Description: "Get complete thinking chain for a session",
	}, createReviewThinkingHandler(store))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "branch_thinking",
		Description: "Create alternative reasoning paths from a specific step",
	}, createBranchThinkingHandler(store))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "merge_insights",
		Description: "Combine insights from multiple reasoning branches",
	}, createMergeInsightsHandler(store))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "validate_logic",
		Description: "Check reasoning for logical fallacies and weaknesses",
	}, createValidateLogicHandler(store))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "export_session",
		Description: "Export a thinking session to the requested format",
	}, createExportSessionHandler(store))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_sessions",
		Description: "List all active thinking sessions",
	}, createListSessionsHandler(store))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_session",
		Description: "Delete a thinking session",
	}, createDeleteSessionHandler(store))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_metrics",
		Description: "Get analytics and metrics for thinking sessions",
	}, createGetMetricsHandler(store))
}

// RegisterResources wires resource handlers with the MCP server.
func RegisterResources(server *mcp.Server, store *thinking.MemoryStore) {
	server.AddResourceTemplate(
		&mcp.ResourceTemplate{
			URITemplate: "thinking://session/{session_id}",
			Name:        "Thinking Session",
			Description: "A complete thinking session with all steps and branches",
		},
		createSessionResourceHandler(store),
	)

	server.AddResource(
		&mcp.Resource{
			URI:         "thinking://sessions/list",
			Name:        "Sessions List",
			Description: "List of all active thinking sessions",
		},
		createSessionListResourceHandler(store),
	)

	server.AddResourceTemplate(
		&mcp.ResourceTemplate{
			URITemplate: "thinking://template/{template_type}",
			Name:        "Thinking Template",
			Description: "Pre-built thinking frameworks",
		},
		createTemplateResourceHandler(store),
	)
}

// RegisterPrompts wires prompt handlers with the MCP server.
func RegisterPrompts(server *mcp.Server) {
	server.AddPrompt(
		&mcp.Prompt{
			Name:        "problem_breakdown",
			Description: "Guide for breaking down complex problems",
			Arguments: []*mcp.PromptArgument{
				{
					Name:        "problem_statement",
					Title:       "Problem Statement",
					Description: "The problem to break down",
					Required:    true,
				},
				{
					Name:        "domain",
					Title:       "Problem Domain",
					Description: "The domain or field of the problem",
					Required:    false,
				},
			},
		},
		createProblemBreakdownPromptHandler(),
	)

	server.AddPrompt(
		&mcp.Prompt{
			Name:        "critical_analysis",
			Description: "Guide for critical evaluation of arguments",
			Arguments: []*mcp.PromptArgument{
				{
					Name:        "claim",
					Title:       "Claim to Analyze",
					Description: "The claim to analyze",
					Required:    true,
				},
				{
					Name:        "evidence",
					Title:       "Supporting Evidence",
					Description: "Evidence supporting the claim",
					Required:    false,
				},
			},
		},
		createCriticalAnalysisPromptHandler(),
	)

	server.AddPrompt(
		&mcp.Prompt{
			Name:        "synthesis_prompt",
			Description: "Guide for combining multiple insights",
			Arguments: []*mcp.PromptArgument{
				{
					Name:        "insights",
					Title:       "Insights to Combine",
					Description: "JSON array of insights to combine",
					Required:    true,
				},
				{
					Name:        "goal",
					Title:       "Synthesis Goal",
					Description: "The goal of synthesis",
					Required:    true,
				},
			},
		},
		createSynthesisPromptHandler(),
	)
}

// CompletionHandler returns the MCP completion handler that powers autocomplete.
func CompletionHandler(store *thinking.MemoryStore) func(context.Context, *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	return func(ctx context.Context, req *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
		ref := req.Params.Ref

		switch ref.Type {
		case "ref/resource":
			if len(ref.URI) > 0 {
				if strings.HasPrefix(ref.URI, "thinking://session") {
					sessions := store.ListSessions()
					var suggestions []string
					for _, s := range sessions {
						suggestions = append(suggestions, s.ID)
					}
					return &mcp.CompleteResult{
						Completion: mcp.CompletionResultDetails{
							Values:  suggestions,
							Total:   len(suggestions),
							HasMore: false,
						},
					}, nil
				}
				if strings.HasPrefix(ref.URI, "thinking://template") {
					templates := getTemplateList()
					var suggestions []string
					for _, t := range templates {
						suggestions = append(suggestions, t)
					}
					return &mcp.CompleteResult{
						Completion: mcp.CompletionResultDetails{
							Values:  suggestions,
							Total:   len(suggestions),
							HasMore: false,
						},
					}, nil
				}
			}

		case "ref/prompt":
			return &mcp.CompleteResult{
				Completion: mcp.CompletionResultDetails{
					Values: []string{
						"problem_breakdown",
						"critical_analysis",
						"synthesis_prompt",
					},
					Total:   3,
					HasMore: false,
				},
			}, nil
		}

		return &mcp.CompleteResult{
			Completion: mcp.CompletionResultDetails{
				Values:  []string{},
				Total:   0,
				HasMore: false,
			},
		}, nil
	}
}
