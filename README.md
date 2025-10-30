# Sequential Thinking MCP Server

Sequential Thinking is a Model Context Protocol (MCP) server that operationalizes deliberate reasoning. It lets MCP clients create structured thinking sessions, capture sequential steps (analysis → hypothesis → verification → conclusion), branch into alternative lines of thought, and export or audit the full chain of reasoning.

## Overview

- Written in Go 1.24 with the MCP Go SDK as its only direct dependency.
- Ships an in-memory session store, branch management, logical validation, and reasoning-quality heuristics.
- Provides eleven MCP tools, resource endpoints for sessions and templates, and three reasoning prompts.
- Supports stdio (default) and streamable HTTP transports, plus a minimal scratch-based Docker image published as `bpradana/sequentialthinking`.
- Backed by unit tests across handlers and an end-to-end MCP integration test.

## Quick Start

### Prerequisites

- Go 1.24 or newer
- Git (or another way to obtain the source)

### Build and Run (stdio transport)

```bash
go mod download
go build -o sequentialthinking ./cmd/sequentialthinking
./sequentialthinking
```

The binary listens on stdio and is ready to be launched by MCP-aware clients.

### Run as HTTP server

```bash
./sequentialthinking -http :8080
```

The `-http` flag enables streamable HTTP transport on the supplied address.

### Docker (prebuilt image)

```bash
docker pull bpradana/sequentialthinking
docker run --rm -i bpradana/sequentialthinking
```

`-i` keeps stdin open so MCP clients can speak stdio with the containerised server.

### Docker (build locally)

```bash
docker build -t sequentialthinking .
docker run --rm sequentialthinking
```

The multi-stage Dockerfile builds a static binary and runs it from a scratch image (UID 65532).

## Integrating with MCP Clients

- **MCP Inspector**
  ```bash
  npm install -g @modelcontextprotocol/inspector
  mcp-inspector ./sequentialthinking
  ```
- **Claude Desktop via Docker (macOS path shown)**
  ```json
  {
    "mcpServers": {
      "sequentialthinking": {
        "command": "docker",
        "args": [
          "run",
          "--rm",
          "-i",
          "bpradana/sequentialthinking"
        ]
      }
    }
  }
  ```
  Pull the image first with `docker pull bpradana/sequentialthinking`. Claude keeps the container alive while the MCP session is active and re-creates it on demand.

### Kick off a Template Session

Once connected (Inspector, Claude, or another MCP client), call the `sequentialthinking.start_from_template` tool to seed a new session:

```json
{
  "name": "sequentialthinking.start_from_template",
  "arguments": {
    "template": "scientific-method",
    "problem": "Why did the nightly build fail?",
    "context": {
      "repo": "widgets-service",
      "last_green_commit": "abc1234"
    }
  }
}
```

The server responds with the new `session_id`, populated steps from the requested template, and suggested next actions. Subsequent calls can extend or branch the reasoning as usual.

## Tooling Surface

| Tool                  | Purpose                                                              | Notable Arguments                                                   |
|-----------------------|----------------------------------------------------------------------|---------------------------------------------------------------------|
| `start_thinking`      | Create a new session with initial analysis and suggested next steps. | `problem`, optional `context`, optional `tags`                      |
| `start_from_template` | Create a session seeded with a predefined thinking template.         | `template`, optional `problem`, optional `context`, optional `tags` |
| `add_step`            | Append a reasoning step to the main flow or a branch.                | `session_id`, optional `branch_id`, `step_type`, `step_content`     |
| `update_step`         | Edit an existing step’s content, type, or metadata.                  | `session_id`, `step_number`, optional fields to update              |
| `review_thinking`     | Retrieve the full chain, connections, patterns, and summary.         | `session_id`, optional `format` (`linear`, `tree`, `summary`)       |
| `branch_thinking`     | Fork the reasoning path from an existing step and seed a branch.     | `session_id`, `from_step`, `alternative_reasoning`                  |
| `merge_insights`      | Synthesize conclusions across multiple branches.                     | `session_id`, `branch_ids`                                          |
| `validate_logic`      | Flag logical issues and describe strengths.                          | `session_id`, optional `range_start`, `range_end`                   |
| `export_session`      | Export a session as markdown, JSON, or plain text.                   | `session_id`, optional `include_branches`, `format`                 |
| `list_sessions`       | Filterable session directory with quality scores.                    | optional `status`, `tags`, `limit`                                  |
| `delete_session`      | Remove a session.                                                    | `session_id`                                                        |
| `get_metrics`         | Aggregate reasoning metrics across sessions.                         | optional `time_range` (`day`, `week`, `month`, `all`)               |

## Resources and Prompts

- **Resources**
  - `thinking://session/{session_id}` – JSON snapshot of a single session.
  - `thinking://sessions/list` – JSON list of all sessions in memory.
  - `thinking://template/{template_type}` – JSON representation of thinking frameworks (`scientific-method`, `five-whys`, `root-cause-analysis`, `decision-matrix`, `swot-analysis`, `pros-cons`, `first-principles`, `fishbone`, `pareto-analysis`).

- **Prompts**
  - `problem_breakdown` – Structures complex problem decomposition.
  - `critical_analysis` – Guides evaluation of claims and evidence.
  - `synthesis_prompt` – Helps combine multiple insights into coherent output.

## Reasoning Model Highlights

- **Step Types**: `analysis`, `hypothesis`, `verification`, `conclusion`. Parent references and connection tracking let clients build reasoning graphs.
- **Branching**: Any step can branch into alternative reasoning paths. Branch steps are numbered locally and remain tied to the originating step.
- **Quality Scoring**: Heuristic combines step-type diversity (30%), connection density (30%), and depth (40%). Scores stay in `[0,1]` and recompute after each update.
- **Logical Validation**: Detects unsupported conclusions and unverified hypotheses. Returns structured issues, suggestions, strength highlights, and an overall validity score.
- **Pattern Detection**: Identifies recurring reasoning patterns (hypothesis-verification, progressive refinement, branching synthesis) for analytics.
- **Exports**: Markdown, JSON, and text exporters include metadata, connections, branches, and suggested filenames.

## Operational Notes

- **Storage**: In-memory (`internal/thinking.MemoryStore`) guarded by RW mutexes. Sessions are ephemeral and cleared on restart.
- **Concurrency**: Reads return deep clones to prevent external mutation; writes update timestamps, current step counters, and quality scores.
- **Logging**: Uses `log/slog` for initialization messages and `mcp.LoggingTransport` for stdio diagnostics.
- **Completion Support**: Autocomplete handler surfaces session IDs and template names for MCP completion requests.

## Development Workflow

```bash
# Run unit and integration tests
go test ./...

# Format and lint
go fmt ./...
golangci-lint run    # optional, if installed
```

### Project Layout

```
cmd/sequentialthinking   CLI entry point and flag parsing
internal/server          MCP server assembly, transports, integration tests
internal/handlers        Tool, resource, prompt handlers and helper logic
internal/thinking        Core domain types, in-memory store, exports, metrics
```

### Release Tips

- Build with `CGO_ENABLED=0` for static binaries (already used in Dockerfile).
- Bump the version in `internal/server.New` when publishing server changes.
- Use the provided Dockerfile as the canonical container build.

## License

MIT License (use repository root `LICENSE` if present).
