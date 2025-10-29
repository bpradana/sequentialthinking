package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var httpAddr = flag.String("http", "", "if set, use streamable HTTP at this address, instead of stdin/stdout")

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	// Create in-memory store
	store := NewMemoryStore()

	// Create server
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "sequential-thinking",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{
			CompletionHandler: createCompletionHandler(store),
			InitializedHandler: func(ctx context.Context, req *mcp.InitializedRequest) {
				slog.Info("Server initialized")
			},
		},
	)

	// Register all tools
	registerTools(server, store)

	// Register all resources
	registerResources(server, store)

	// Register all prompts
	registerPrompts(server, store)

	// Connect using stdio transport
	if *httpAddr != "" {
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return server
		}, nil)
		log.Printf("sequential thinking MCP server listening at %s", *httpAddr)
		if err := http.ListenAndServe(*httpAddr, handler); err != nil {
			log.Fatal(err)
		}
	} else {
		t := &mcp.LoggingTransport{Transport: &mcp.StdioTransport{}, Writer: os.Stderr}
		if err := server.Run(ctx, t); err != nil {
			log.Printf("server failed: %v", err)
		}
	}

	return nil
}

func registerTools(server *mcp.Server, store *MemoryStore) {
	// Tool: start_thinking
	mcp.AddTool(server, &mcp.Tool{
		Name:        "start_thinking",
		Description: "Initiate a new thinking session to break down a complex problem",
	}, createStartThinkingHandler(store))

	// Tool: add_step
	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_step",
		Description: "Add a reasoning step to an existing thinking session",
	}, createAddStepHandler(store))

	// Tool: review_thinking
	mcp.AddTool(server, &mcp.Tool{
		Name:        "review_thinking",
		Description: "Get complete thinking chain for a session",
	}, createReviewThinkingHandler(store))

	// Tool: branch_thinking
	mcp.AddTool(server, &mcp.Tool{
		Name:        "branch_thinking",
		Description: "Create alternative reasoning paths from a specific step",
	}, createBranchThinkingHandler(store))

	// Tool: merge_insights
	mcp.AddTool(server, &mcp.Tool{
		Name:        "merge_insights",
		Description: "Combine insights from multiple reasoning branches",
	}, createMergeInsightsHandler(store))

	// Tool: validate_logic
	mcp.AddTool(server, &mcp.Tool{
		Name:        "validate_logic",
		Description: "Check reasoning for logical fallacies and weaknesses",
	}, createValidateLogicHandler(store))

	// Tool: export_session
	mcp.AddTool(server, &mcp.Tool{
		Name:        "export_session",
		Description: "Export a thinking session to markdown format",
	}, createExportSessionHandler(store))

	// Tool: list_sessions
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_sessions",
		Description: "List all active thinking sessions",
	}, createListSessionsHandler(store))

	// Tool: delete_session
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_session",
		Description: "Delete a thinking session",
	}, createDeleteSessionHandler(store))

	// Tool: get_metrics
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_metrics",
		Description: "Get analytics and metrics for thinking sessions",
	}, createGetMetricsHandler(store))
}

func registerResources(server *mcp.Server, store *MemoryStore) {
	// Resource: Individual session
	server.AddResourceTemplate(
		&mcp.ResourceTemplate{
			URITemplate: "thinking://session/{session_id}",
			Name:        "Thinking Session",
			Description: "A complete thinking session with all steps and branches",
		},
		createSessionResourceHandler(store),
	)

	// Resource: Session list
	server.AddResource(
		&mcp.Resource{
			URI:         "thinking://sessions/list",
			Name:        "Sessions List",
			Description: "List of all active thinking sessions",
		},
		createSessionListResourceHandler(store),
	)

	// Resource: Templates
	server.AddResourceTemplate(
		&mcp.ResourceTemplate{
			URITemplate: "thinking://template/{template_type}",
			Name:        "Thinking Template",
			Description: "Pre-built thinking frameworks",
		},
		createTemplateResourceHandler(store),
	)
}

func registerPrompts(server *mcp.Server, store *MemoryStore) {
	// Prompt: problem_breakdown
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

	// Prompt: critical_analysis
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

	// Prompt: synthesis_prompt
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

func createCompletionHandler(store *MemoryStore) func(context.Context, *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	return func(ctx context.Context, req *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
		ref := req.Params.Ref

		switch ref.Type {
		case "ref/resource":
			// Complete session IDs
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
					templates_ := getTemplateList()
					var suggestions []string
					for _, t := range templates_ {
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
			// Complete prompt arguments
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
