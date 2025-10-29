package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bpradana/sequentialthinking/internal/server"
	"github.com/bpradana/sequentialthinking/internal/thinking"
)

type (
	StartThinkingInput   = thinking.StartThinkingInput
	StartThinkingOutput  = thinking.StartThinkingOutput
	AddStepInput         = thinking.AddStepInput
	AddStepOutput        = thinking.AddStepOutput
	ReviewThinkingInput  = thinking.ReviewThinkingInput
	ReviewThinkingOutput = thinking.ReviewThinkingOutput
	ValidateLogicInput   = thinking.ValidateLogicInput
	ValidateLogicOutput  = thinking.ValidateLogicOutput
	ExportSessionInput   = thinking.ExportSessionInput
	ExportSessionOutput  = thinking.ExportSessionOutput
)

const (
	StepAnalysis     = thinking.StepAnalysis
	StepHypothesis   = thinking.StepHypothesis
	StepVerification = thinking.StepVerification
	StepConclusion   = thinking.StepConclusion
)

func TestMCPIntegrationFlow(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store := thinking.NewMemoryStore()
	srv := server.New(store)

	client := mcp.NewClient(
		&mcp.Implementation{Name: "integration-client", Version: "test"},
		nil,
	)

	t1, t2 := mcp.NewInMemoryTransports()
	serverSession, err := srv.Connect(ctx, t1, nil)
	if err != nil {
		t.Fatalf("server connect failed: %v", err)
	}
	defer func() {
		if err := serverSession.Wait(); err != nil {
			t.Fatalf("server session wait failed: %v", err)
		}
	}()

	clientSession, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect failed: %v", err)
	}
	defer func() {
		if err := clientSession.Close(); err != nil {
			t.Fatalf("client close failed: %v", err)
		}
	}()

	startRes, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "start_thinking",
		Arguments: StartThinkingInput{
			Problem: "Integration flow validation",
			Tags:    []string{"integration"},
		},
	})
	if err != nil {
		t.Fatalf("start_thinking call failed: %v", err)
	}
	startOut := decodeStructured[StartThinkingOutput](t, startRes.StructuredContent)
	if startOut.SessionID == "" {
		t.Fatalf("expected session id, got %+v", startOut)
	}
	if len(startOut.SuggestedSteps) == 0 {
		t.Fatalf("expected suggested steps in start output")
	}

	addRes, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "add_step",
		Arguments: AddStepInput{
			SessionID:   startOut.SessionID,
			StepContent: "Initial analysis step from integration test",
			StepType:    StepAnalysis,
		},
	})
	if err != nil {
		t.Fatalf("add_step call failed: %v", err)
	}
	addOut := decodeStructured[AddStepOutput](t, addRes.StructuredContent)
	if addOut.StepNumber != 1 {
		t.Fatalf("expected step number 1, got %d", addOut.StepNumber)
	}

	hypRes, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "add_step",
		Arguments: AddStepInput{
			SessionID:   startOut.SessionID,
			StepContent: "Form hypothesis about bottleneck",
			StepType:    StepHypothesis,
			ParentStep:  intPtr(addOut.StepNumber),
		},
	})
	if err != nil {
		t.Fatalf("add_step hypothesis failed: %v", err)
	}
	hypOut := decodeStructured[AddStepOutput](t, hypRes.StructuredContent)

	verifyRes, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "add_step",
		Arguments: AddStepInput{
			SessionID:   startOut.SessionID,
			StepContent: "Verify hypothesis with experiment",
			StepType:    StepVerification,
			ParentStep:  intPtr(hypOut.StepNumber),
		},
	})
	if err != nil {
		t.Fatalf("add_step verification failed: %v", err)
	}
	verifyOut := decodeStructured[AddStepOutput](t, verifyRes.StructuredContent)

	_, err = clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "add_step",
		Arguments: AddStepInput{
			SessionID:   startOut.SessionID,
			StepContent: "Draw conclusion from results",
			StepType:    StepConclusion,
			ParentStep:  intPtr(verifyOut.StepNumber),
		},
	})
	if err != nil {
		t.Fatalf("add_step conclusion failed: %v", err)
	}

	reviewRes, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "review_thinking",
		Arguments: ReviewThinkingInput{
			SessionID: startOut.SessionID,
			Format:    "summary",
		},
	})
	if err != nil {
		t.Fatalf("review_thinking call failed: %v", err)
	}
	reviewOut := decodeStructured[ReviewThinkingOutput](t, reviewRes.StructuredContent)
	if len(reviewOut.Steps) != 4 {
		t.Fatalf("expected 4 steps in review, got %d", len(reviewOut.Steps))
	}

	validateRes, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "validate_logic",
		Arguments: ValidateLogicInput{
			SessionID: startOut.SessionID,
		},
	})
	if err != nil {
		t.Fatalf("validate_logic call failed: %v", err)
	}
	validateOut := decodeStructured[ValidateLogicOutput](t, validateRes.StructuredContent)
	if len(validateOut.StrongPoints) == 0 {
		t.Fatalf("expected strong points from validation output")
	}

	exportRes, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "export_session",
		Arguments: ExportSessionInput{
			SessionID:       startOut.SessionID,
			Format:          "markdown",
			IncludeBranches: true,
		},
	})
	if err != nil {
		t.Fatalf("export_session call failed: %v", err)
	}
	exportOut := decodeStructured[ExportSessionOutput](t, exportRes.StructuredContent)
	if len(exportOut.Content) == 0 || exportOut.Format != "markdown" {
		t.Fatalf("unexpected export output: %+v", exportOut)
	}

	resourceRes, err := clientSession.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: fmt.Sprintf("thinking://session/%s", startOut.SessionID),
	})
	if err != nil {
		t.Fatalf("read resource failed: %v", err)
	}
	if len(resourceRes.Contents) != 1 {
		t.Fatalf("expected single resource content, got %d", len(resourceRes.Contents))
	}

	promptRes, err := clientSession.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "problem_breakdown",
		Arguments: map[string]string{
			"problem_statement": "Validate integration",
		},
	})
	if err != nil {
		t.Fatalf("get_prompt failed: %v", err)
	}
	if len(promptRes.Messages) == 0 {
		t.Fatalf("expected prompt messages, got 0")
	}

	completeRes, err := clientSession.Complete(ctx, &mcp.CompleteParams{
		Ref: &mcp.CompleteReference{
			Type: "ref/resource",
			URI:  "thinking://session/",
		},
	})
	if err != nil {
		t.Fatalf("complete call failed: %v", err)
	}
	if len(completeRes.Completion.Values) == 0 {
		t.Fatalf("expected completion suggestions, got 0")
	}
}

func decodeStructured[T any](t *testing.T, structured any) T {
	t.Helper()

	var zero T
	if structured == nil {
		t.Fatalf("structured content is nil")
		return zero
	}
	data, err := json.Marshal(structured)
	if err != nil {
		t.Fatalf("failed to marshal structured content: %v", err)
		return zero
	}
	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("failed to unmarshal structured content: %v", err)
		return zero
	}
	return out
}

func intPtr(v int) *int {
	return &v
}
