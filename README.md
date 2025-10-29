# Sequential Thinking MCP Server

A Model Context Protocol (MCP) server that helps break down complex problems into step-by-step reasoning processes, making AI reasoning more transparent and structured.

## Features

### 🛠️ Tools (10 total)

1. **start_thinking** - Initiate a new thinking session
2. **add_step** - Add reasoning steps with different types (analysis, hypothesis, verification, conclusion)
3. **review_thinking** - Get complete thinking chain with quality assessment
4. **branch_thinking** - Create alternative reasoning paths
5. **merge_insights** - Combine insights from multiple branches
6. **validate_logic** - Check reasoning for logical fallacies
7. **export_session** - Export sessions to markdown, JSON, or text
8. **list_sessions** - List all thinking sessions with filtering
9. **delete_session** - Remove sessions
10. **get_metrics** - Analytics on thinking patterns and quality

### 📚 Resources

- `thinking://session/{session_id}` - Access individual sessions
- `thinking://sessions/list` - Browse all sessions
- `thinking://template/{template_type}` - Pre-built frameworks:
    - scientific-method
    - five-whys
    - decision-matrix
    - swot-analysis
    - pros-cons
    - first-principles
    - fishbone
    - pareto-analysis

### 💬 Prompts

1. **problem_breakdown** - Guide for decomposing complex problems
2. **critical_analysis** - Framework for evaluating arguments
3. **synthesis_prompt** - Template for combining multiple insights

### ✨ Advanced Features

- **Quality Scoring**: Automatic assessment of reasoning quality
- **Pattern Detection**: Identifies common thinking patterns
- **Logical Validation**: Detects fallacies and weak reasoning
- **Branching**: Explore multiple solution approaches
- **Metrics & Analytics**: Track thinking effectiveness over time
- **Multiple Export Formats**: Markdown, JSON, and plain text
- **Completion Support**: Auto-complete session IDs and prompts
- **Real-time Logging**: Monitor thinking progress

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd sequential-thinking-mcp

# Install dependencies
go mod download

# Build
go build -o sequential-thinking-server

# Run
./sequential-thinking-server
```

## Usage

### With MCP Inspector

```bash
# Install MCP Inspector
npm install -g @modelcontextprotocol/inspector

# Run server with inspector
mcp-inspector sequential-thinking-server
```

### With Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "sequential-thinking": {
      "command": "/path/to/sequential-thinking-server"
    }
  }
}
```

### Example Workflow

```javascript
// 1. Start a thinking session
{
  "tool": "start_thinking",
  "arguments": {
    "problem": "How can we reduce API latency?",
    "tags": ["performance", "architecture"]
  }
}

// 2. Add analysis steps
{
  "tool": "add_step",
  "arguments": {
    "session_id": "<session_id>",
    "step_content": "Current average latency is 250ms, with p95 at 800ms",
    "step_type": "analysis"
  }
}

// 3. Form hypothesis
{
  "tool": "add_step",
  "arguments": {
    "session_id": "<session_id>",
    "step_content": "Database queries are the primary bottleneck",
    "step_type": "hypothesis"
  }
}

// 4. Create alternative branch
{
  "tool": "branch_thinking",
  "arguments": {
    "session_id": "<session_id>",
    "from_step": 2,
    "alternative_reasoning": "Network latency might be the main issue"
  }
}

// 5. Validate logic
{
  "tool": "validate_logic",
  "arguments": {
    "session_id": "<session_id>"
  }
}

// 6. Review complete thinking
{
  "tool": "review_thinking",
  "arguments": {
    "session_id": "<session_id>",
    "format": "tree"
  }
}

// 7. Export session
{
  "tool": "export_session",
  "arguments": {
    "session_id": "<session_id>",
    "format": "markdown",
    "include_branches": true
  }
}
```

## Step Types

- **analysis**: Breaking down and examining the problem
- **hypothesis**: Forming testable explanations
- **verification**: Testing and validating hypotheses
- **conclusion**: Drawing final insights and decisions

## Templates

Access thinking frameworks via resources:

```javascript
{
  "resource": "thinking://template/scientific-method"
}
```

Available templates guide you through proven problem-solving approaches:
- **Scientific Method**: Systematic hypothesis testing
- **Five Whys**: Root cause analysis
- **Decision Matrix**: Weighted option comparison
- **SWOT**: Strengths, Weaknesses, Opportunities, Threats
- **First Principles**: Reasoning from fundamental truths
- **Fishbone**: Cause-effect analysis
- **Pareto**: 80/20 prioritization

## Metrics

Track your thinking effectiveness:

```javascript
{
  "tool": "get_metrics",
  "arguments": {
    "time_range": "week"  // day, week, month, or all
  }
}
```

Returns:
- Total sessions and completion rate
- Average steps per session
- Quality score trends
- Common patterns detected
- Logical issue frequency
- Step type distribution

## Architecture

```
sequential-thinking-server/
├── main.go              # Entry point, server setup
├── types.go             # Data structures
├── store.go             # In-memory storage
├── handlers.go          # Tool implementations
├── export.go            # Export functionality
├── resources.go         # Resource handlers
├── prompts.go           # Prompt handlers
├── go.mod               # Dependencies
└── README.md            # Documentation
```

## Storage

Uses in-memory storage by default. Sessions are lost on restart. All operations are thread-safe using `sync.RWMutex`.

## Error Handling

All tools return structured errors. Sessions that don't exist return appropriate "not found" errors. Invalid inputs are validated before processing.

## Quality Scoring

Automatic quality assessment based on:
- Variety of step types (30%)
- Step connectivity (30%)
- Reasoning depth (40%)

Score ranges:
- 0.8-1.0: Excellent reasoning
- 0.6-0.8: Good reasoning
- 0.4-0.6: Adequate reasoning
- 0.0-0.4: Needs improvement

## Logical Validation

Detects common issues:
- Unsupported conclusions
- Unverified hypotheses
- Weak connections
- Missing evidence
- Logical fallacies

## License

MIT

## Contributing

Contributions welcome! Please open issues or pull requests.

## Development

```bash
# Run tests
go test ./...

# Format code
go fmt ./...

# Lint
golangci-lint run

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o sequential-thinking-linux
GOOS=darwin GOARCH=arm64 go build -o sequential-thinking-mac
GOOS=windows GOARCH=amd64 go build -o sequential-thinking.exe
```

## Future Enhancements

- Persistent storage (SQLite, PostgreSQL)
- Multi-user support with authentication
- Real-time collaboration
- AI-powered suggestion engine
- Visual reasoning graphs
- Import from other formats
- Plugin system for custom validators
- Web dashboard
- API rate limiting
- Session versioning
- Automated quality suggestions
