package thinking

import (
	"time"
)

// StepType represents the type of reasoning step
type StepType string

const (
	StepAnalysis     StepType = "analysis"
	StepHypothesis   StepType = "hypothesis"
	StepVerification StepType = "verification"
	StepConclusion   StepType = "conclusion"
)

var (
	stepTypeValues = []StepType{
		StepAnalysis,
		StepHypothesis,
		StepVerification,
		StepConclusion,
	}
	stepTypeSet = map[StepType]struct{}{
		StepAnalysis:     {},
		StepHypothesis:   {},
		StepVerification: {},
		StepConclusion:   {},
	}
)

// IsValid reports whether the step type is one of the supported values.
func (t StepType) IsValid() bool {
	_, ok := stepTypeSet[t]
	return ok
}

// AllowedStepTypeStrings returns the list of supported step types as strings.
func AllowedStepTypeStrings() []string {
	values := make([]string, len(stepTypeValues))
	for i, v := range stepTypeValues {
		values[i] = string(v)
	}
	return values
}

// ThinkingStep represents a single step in the reasoning process
type ThinkingStep struct {
	Number      int            `json:"number"`
	Type        StepType       `json:"type"`
	Content     string         `json:"content"`
	Timestamp   time.Time      `json:"timestamp"`
	ParentStep  *int           `json:"parent_step,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Connections []int          `json:"connections,omitempty"`
}

// Branch represents an alternative reasoning path
type Branch struct {
	ID              string          `json:"id"`
	FromStep        int             `json:"from_step"`
	Steps           []*ThinkingStep `json:"steps"`
	Created         time.Time       `json:"created"`
	AlternativeDesc string          `json:"alternative_desc"`
}

// ThinkingSession represents a complete thinking session
type ThinkingSession struct {
	ID              string             `json:"id"`
	Problem         string             `json:"problem"`
	Context         map[string]any     `json:"context,omitempty"`
	Steps           []*ThinkingStep    `json:"steps"`
	Branches        map[string]*Branch `json:"branches"`
	Created         time.Time          `json:"created"`
	LastModified    time.Time          `json:"last_modified"`
	CurrentStep     int                `json:"current_step"`
	QualityScore    float64            `json:"quality_score"`
	Status          string             `json:"status"`
	Tags            []string           `json:"tags,omitempty"`
	InitialAnalysis string             `json:"initial_analysis"`
}

// LogicalIssue represents a potential problem in reasoning
type LogicalIssue struct {
	StepNumber  int      `json:"step_number"`
	IssueType   string   `json:"issue_type"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Suggestion  string   `json:"suggestion"`
	Examples    []string `json:"examples,omitempty"`
}

// ThinkingPattern represents a detected reasoning pattern
type ThinkingPattern struct {
	Name        string  `json:"name"`
	Frequency   int     `json:"frequency"`
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

// SessionMetrics represents analytics for thinking sessions
type SessionMetrics struct {
	TotalSessions     int               `json:"total_sessions"`
	ActiveSessions    int               `json:"active_sessions"`
	CompletedSessions int               `json:"completed_sessions"`
	AverageSteps      float64           `json:"average_steps"`
	AverageQuality    float64           `json:"average_quality"`
	CommonPatterns    []ThinkingPattern `json:"common_patterns"`
	CommonIssues      map[string]int    `json:"common_issues"`
	StepTypeDistrib   map[StepType]int  `json:"step_type_distribution"`
	AverageBranches   float64           `json:"average_branches"`
	SessionsByDay     map[string]int    `json:"sessions_by_day"`
	TopTags           map[string]int    `json:"top_tags"`
}

// Template represents a thinking framework template
type Template struct {
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
	WhenToUse   string   `json:"when_to_use"`
	Example     string   `json:"example,omitempty"`
}

// Input/Output types for tools

type StartThinkingInput struct {
	Template string         `json:"template,omitempty" jsonschema:"The template identifier to initialize from,enum=scientific-method,enum=five-whys,enum=root-cause-analysis,enum=decision-matrix,enum=swot-analysis,enum=pros-cons,enum=five-principles,enum=fishbone,enum=pareto-analysis"`
	Problem  string         `json:"problem" jsonschema:"The problem to analyze"`
	Context  map[string]any `json:"context,omitempty" jsonschema:"Optional background information"`
	Tags     []string       `json:"tags,omitempty" jsonschema:"Tags to categorize the session"`
}

type StartThinkingOutput struct {
	SessionID       string   `json:"session_id" jsonschema:"Unique identifier for the session"`
	InitialAnalysis string   `json:"initial_analysis" jsonschema:"First thoughts on the problem"`
	SuggestedSteps  []string `json:"suggested_steps" jsonschema:"Recommended next steps"`
}

type AddStepInput struct {
	SessionID   string         `json:"session_id" jsonschema:"The session to add to"`
	BranchID    string         `json:"branch_id,omitempty" jsonschema:"Branch identifier if adding to a branch"`
	StepContent string         `json:"step_content" jsonschema:"The reasoning for this step"`
	StepType    StepType       `json:"step_type" jsonschema:"Type of reasoning step,enum=analysis,enum=hypothesis,enum=verification,enum=conclusion"`
	ParentStep  *int           `json:"parent_step,omitempty" jsonschema:"Parent step number if building on previous step"`
	Metadata    map[string]any `json:"metadata,omitempty" jsonschema:"Additional metadata for the step"`
}

type AddStepOutput struct {
	StepNumber         int      `json:"step_number" jsonschema:"The number of the added step"`
	BranchID           string   `json:"branch_id,omitempty" jsonschema:"Branch identifier when the step belongs to a branch"`
	CurrentProgress    string   `json:"current_progress" jsonschema:"Summary of thinking so far"`
	SuggestedNextSteps []string `json:"suggested_next_steps" jsonschema:"Potential next steps"`
	QualityScore       float64  `json:"quality_score" jsonschema:"Current reasoning quality score"`
}

type UpdateStepInput struct {
	SessionID   string         `json:"session_id" jsonschema:"The session containing the step"`
	StepNumber  int            `json:"step_number" jsonschema:"The step number to update"`
	StepContent *string        `json:"step_content,omitempty" jsonschema:"Updated reasoning content"`
	StepType    *StepType      `json:"step_type,omitempty" jsonschema:"Updated step type"`
	Metadata    map[string]any `json:"metadata,omitempty" jsonschema:"Replacement metadata for the step"`
}

type UpdateStepOutput struct {
	StepNumber         int           `json:"step_number" jsonschema:"The number of the updated step"`
	UpdatedStep        *ThinkingStep `json:"updated_step" jsonschema:"The updated step details"`
	CurrentProgress    string        `json:"current_progress" jsonschema:"Summary of thinking so far"`
	SuggestedNextSteps []string      `json:"suggested_next_steps" jsonschema:"Potential next steps"`
	QualityScore       float64       `json:"quality_score" jsonschema:"Current reasoning quality score"`
	LastModified       time.Time     `json:"last_modified" jsonschema:"Timestamp when the session was last modified"`
	MetadataChanged    bool          `json:"metadata_changed" jsonschema:"Whether metadata was replaced"`
	TypeChanged        bool          `json:"type_changed" jsonschema:"Whether the step type was updated"`
	ContentChanged     bool          `json:"content_changed" jsonschema:"Whether the step content was updated"`
}

type ReviewThinkingInput struct {
	SessionID string `json:"session_id" jsonschema:"The session to review"`
	Format    string `json:"format,omitempty" jsonschema:"Output format: linear, tree, or summary"`
}

type ReviewThinkingOutput struct {
	Steps        []*ThinkingStep    `json:"steps" jsonschema:"All thinking steps"`
	Connections  map[string][]int   `json:"connections" jsonschema:"Relationships between steps"`
	QualityScore float64            `json:"quality_score" jsonschema:"Assessment of reasoning quality"`
	Summary      string             `json:"summary" jsonschema:"Overall summary of the thinking process"`
	Branches     map[string]*Branch `json:"branches,omitempty" jsonschema:"Alternative reasoning paths"`
	Patterns     []ThinkingPattern  `json:"patterns,omitempty" jsonschema:"Detected reasoning patterns"`
}

type BranchThinkingInput struct {
	SessionID            string `json:"session_id" jsonschema:"The session to branch"`
	FromStep             int    `json:"from_step" jsonschema:"Step number to branch from"`
	AlternativeReasoning string `json:"alternative_reasoning" jsonschema:"The alternative reasoning path"`
}

type BranchThinkingOutput struct {
	BranchID        string `json:"branch_id" jsonschema:"Identifier for the new branch"`
	DivergencePoint int    `json:"divergence_point" jsonschema:"Where the paths split"`
	BranchSummary   string `json:"branch_summary" jsonschema:"Summary of the new branch"`
}

type MergeInsightsInput struct {
	SessionID string   `json:"session_id" jsonschema:"The session containing branches"`
	BranchIDs []string `json:"branch_ids" jsonschema:"Branch IDs to merge"`
}

type MergeInsightsOutput struct {
	Synthesis  string   `json:"synthesis" jsonschema:"Combined insights from all branches"`
	Conflicts  []string `json:"conflicts" jsonschema:"Contradictions found between branches"`
	Confidence float64  `json:"confidence" jsonschema:"Overall confidence in the merged conclusion"`
	Strengths  []string `json:"strengths" jsonschema:"Strengths of the combined reasoning"`
}

type ValidateLogicInput struct {
	SessionID  string `json:"session_id" jsonschema:"The session to validate"`
	RangeStart *int   `json:"range_start,omitempty" jsonschema:"Range start of step"`
	RangeEnd   *int   `json:"range_end,omitempty" jsonschema:"Range end of step"`
}

type ValidateLogicOutput struct {
	Issues            []LogicalIssue `json:"issues" jsonschema:"Logical problems found"`
	Suggestions       []string       `json:"suggestions" jsonschema:"Recommendations for improvement"`
	ValidityScore     float64        `json:"validity_score" jsonschema:"Overall validity rating (0-1)"`
	StrongPoints      []string       `json:"strong_points" jsonschema:"Well-reasoned aspects"`
	OverallAssessment string         `json:"overall_assessment" jsonschema:"Summary of logical quality"`
}

type ExportSessionInput struct {
	SessionID       string `json:"session_id" jsonschema:"The session to export"`
	IncludeBranches bool   `json:"include_branches,omitempty" jsonschema:"Include alternative branches"`
	Format          string `json:"format,omitempty" jsonschema:"Export format: markdown, json, or text"`
}

type ExportSessionOutput struct {
	Content  string `json:"content" jsonschema:"The exported session content"`
	Format   string `json:"format" jsonschema:"The format of the export"`
	Filename string `json:"filename" jsonschema:"Suggested filename"`
}

type ListSessionsInput struct {
	Status string   `json:"status,omitempty" jsonschema:"Filter by status: active, completed, all"`
	Tags   []string `json:"tags,omitempty" jsonschema:"Filter by tags"`
	Limit  int      `json:"limit,omitempty" jsonschema:"Maximum number of sessions to return"`
}

type SessionSummary struct {
	ID           string    `json:"id"`
	Problem      string    `json:"problem"`
	StepCount    int       `json:"step_count"`
	BranchCount  int       `json:"branch_count"`
	Status       string    `json:"status"`
	QualityScore float64   `json:"quality_score"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"last_modified"`
	Tags         []string  `json:"tags,omitempty"`
}

type ListSessionsOutput struct {
	Sessions []SessionSummary `json:"sessions" jsonschema:"List of session summaries"`
	Total    int              `json:"total" jsonschema:"Total number of matching sessions"`
}

type DeleteSessionInput struct {
	SessionID string `json:"session_id" jsonschema:"The session to delete"`
}

type DeleteSessionOutput struct {
	Success bool   `json:"success" jsonschema:"Whether deletion was successful"`
	Message string `json:"message" jsonschema:"Confirmation or error message"`
}

type GetMetricsInput struct {
	TimeRange string `json:"time_range,omitempty" jsonschema:"Time range: day, week, month, all"`
}

type GetMetricsOutput struct {
	Metrics SessionMetrics `json:"metrics" jsonschema:"Analytics and metrics for thinking sessions"`
}
