package handlers

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestProblemBreakdownPromptHandler(t *testing.T) {
	t.Parallel()

	handler := createProblemBreakdownPromptHandler()
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name: "problem_breakdown",
			Arguments: map[string]string{
				"problem_statement": "How do we reduce latency?",
				"domain":            "infrastructure",
			},
		},
	}

	res, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if len(res.Messages) != 1 {
		t.Fatalf("expected single message, got %d", len(res.Messages))
	}
	content := res.Messages[0].Content.(*mcp.TextContent).Text
	if !strings.Contains(content, "reduce latency") || !strings.Contains(content, "infrastructure") {
		t.Fatalf("unexpected prompt content:\n%s", content)
	}
}

func TestCriticalAnalysisPromptHandler(t *testing.T) {
	t.Parallel()

	handler := createCriticalAnalysisPromptHandler()
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name: "critical_analysis",
			Arguments: map[string]string{
				"claim":    "Caching improves performance",
				"evidence": "Benchmarks show 40% latency reduction",
			},
		},
	}

	res, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	content := res.Messages[0].Content.(*mcp.TextContent).Text
	if !strings.Contains(content, "Caching improves performance") || !strings.Contains(content, "Benchmarks") {
		t.Fatalf("unexpected prompt content:\n%s", content)
	}
}

func TestSynthesisPromptHandler(t *testing.T) {
	t.Parallel()

	handler := createSynthesisPromptHandler()
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name: "synthesis_prompt",
			Arguments: map[string]string{
				"goal":     "Create launch plan",
				"insights": `["Marketing ready","Engineering confident"]`,
			},
		},
	}

	res, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	content := res.Messages[0].Content.(*mcp.TextContent).Text
	if !strings.Contains(content, "Marketing ready") || !strings.Contains(content, "Engineering confident") {
		t.Fatalf("unexpected prompt content:\n%s", content)
	}

	// Non-JSON path should still render.
	req.Params.Arguments["insights"] = "Single insight"
	res, err = handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error on fallback path: %v", err)
	}
	content = res.Messages[0].Content.(*mcp.TextContent).Text
	if !strings.Contains(content, "Single insight") {
		t.Fatalf("expected fallback insight to be included, got:\n%s", content)
	}
}
