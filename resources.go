package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Resource handler: individual session
func createSessionResourceHandler(store *MemoryStore) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// Extract session ID from URI: thinking://session/{session_id}
		uri := req.Params.URI
		parts := strings.Split(uri, "/")
		if len(parts) < 4 {
			return nil, mcp.ResourceNotFoundError(uri)
		}
		sessionID := parts[3]

		session, err := store.GetSession(sessionID)
		if err != nil {
			return nil, mcp.ResourceNotFoundError(uri)
		}

		// Convert session to JSON
		jsonBytes, err := json.MarshalIndent(session, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal session: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      uri,
					MIMEType: "application/json",
					Text:     string(jsonBytes),
				},
			},
		}, nil
	}
}

// Resource handler: session list
func createSessionListResourceHandler(store *MemoryStore) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		sessions := store.ListSessions()

		summaries := make([]map[string]any, 0, len(sessions))
		for _, s := range sessions {
			summaries = append(summaries, map[string]any{
				"id":            s.ID,
				"problem":       s.Problem,
				"step_count":    len(s.Steps),
				"branch_count":  len(s.Branches),
				"status":        s.Status,
				"quality_score": s.QualityScore,
				"created":       s.Created,
				"last_modified": s.LastModified,
				"tags":          s.Tags,
			})
		}

		jsonBytes, err := json.MarshalIndent(map[string]any{
			"total":    len(sessions),
			"sessions": summaries,
		}, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal session list: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(jsonBytes),
				},
			},
		}, nil
	}
}

// Resource handler: templates
func createTemplateResourceHandler(store *MemoryStore) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// Extract template type from URI: thinking://template/{template_type}
		uri := req.Params.URI
		parts := strings.Split(uri, "/")
		if len(parts) < 4 {
			return nil, mcp.ResourceNotFoundError(uri)
		}
		templateType := parts[3]

		template := getTemplate(templateType)
		if template == nil {
			return nil, mcp.ResourceNotFoundError(uri)
		}

		jsonBytes, err := json.MarshalIndent(template, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal template: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      uri,
					MIMEType: "application/json",
					Text:     string(jsonBytes),
				},
			},
		}, nil
	}
}

// getTemplate returns a thinking template by type
func getTemplate(templateType string) *Template {
	return templates[templateType]
}

// getTemplateList returns a list of available template types
func getTemplateList() []string {
	var templateList []string
	for key := range templates {
		templateList = append(templateList, key)
	}
	return templateList
}
