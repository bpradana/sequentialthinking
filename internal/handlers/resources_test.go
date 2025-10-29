package handlers

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bpradana/sequentialthinking/internal/thinking"
)

func TestSessionResourceHandlerSuccess(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	session := mustCreateSession(t, store, "Resource session")
	mustAddStep(t, store, session.ID, "First step", thinking.StepAnalysis, nil, nil)

	handler := createSessionResourceHandler(store)
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "thinking://session/" + session.ID,
		},
	}

	res, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if len(res.Contents) != 1 {
		t.Fatalf("expected single resource content, got %d", len(res.Contents))
	}
	if !strings.Contains(res.Contents[0].Text, session.ID) {
		t.Fatalf("expected JSON payload to include session id, got %s", res.Contents[0].Text)
	}
}

func TestSessionResourceHandlerNotFound(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	handler := createSessionResourceHandler(store)
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "thinking://session/missing",
		},
	}
	_, err := handler(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error for missing session")
	}
}

func TestSessionListResourceHandler(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	mustCreateSessionWithTags(t, store, "List session", []string{"ops"})

	handler := createSessionListResourceHandler(store)
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "thinking://sessions/list",
		},
	}

	res, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if !strings.Contains(res.Contents[0].Text, "\"total\": 1") {
		t.Fatalf("expected listing to include total count, got %s", res.Contents[0].Text)
	}
}

func TestTemplateResourceHandler(t *testing.T) {
	t.Parallel()

	store := thinking.NewMemoryStore()
	handler := createTemplateResourceHandler(store)

	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "thinking://template/scientific-method",
		},
	}

	res, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if !strings.Contains(res.Contents[0].Text, "Scientific Method") {
		t.Fatalf("expected template payload, got %s", res.Contents[0].Text)
	}

	req.Params.URI = "thinking://template/unknown"
	_, err = handler(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error for unknown template")
	}
}
