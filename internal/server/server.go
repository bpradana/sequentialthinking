package server

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bpradana/sequentialthinking/internal/handlers"
	"github.com/bpradana/sequentialthinking/internal/thinking"
)

// New constructs the MCP server with all registered handlers.
func New(store *thinking.MemoryStore) *mcp.Server {
	srv := mcp.NewServer(
		&mcp.Implementation{
			Name:    "sequential-thinking",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{
			CompletionHandler: handlers.CompletionHandler(store),
			InitializedHandler: func(ctx context.Context, req *mcp.InitializedRequest) {
				slog.Info("Server initialized")
			},
		},
	)

	handlers.RegisterTools(srv, store)
	handlers.RegisterResources(srv, store)
	handlers.RegisterPrompts(srv)

	return srv
}

// Run starts the MCP server using either stdio or HTTP transport.
func Run(ctx context.Context, srv *mcp.Server, httpAddr string) error {
	if httpAddr != "" {
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return srv
		}, nil)
		log.Printf("sequential thinking MCP server listening at %s", httpAddr)
		return http.ListenAndServe(httpAddr, handler)
	}

	t := &mcp.LoggingTransport{Transport: &mcp.StdioTransport{}, Writer: os.Stderr}
	return srv.Run(ctx, t)
}
