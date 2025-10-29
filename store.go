package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// MemoryStore provides in-memory storage for thinking sessions
type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*ThinkingSession
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]*ThinkingSession),
	}
}

// generateID generates a random session ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// CreateSession creates a new thinking session
func (s *MemoryStore) CreateSession(problem string, context map[string]any, tags []string) (*ThinkingSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := &ThinkingSession{
		ID:              generateID(),
		Problem:         problem,
		Context:         context,
		Steps:           make([]*ThinkingStep, 0),
		Branches:        make(map[string]*Branch),
		Created:         time.Now(),
		LastModified:    time.Now(),
		CurrentStep:     0,
		QualityScore:    0.5,
		Status:          "active",
		Tags:            tags,
		InitialAnalysis: generateInitialAnalysis(problem),
	}

	s.sessions[session.ID] = session
	return session, nil
}

// GetSession retrieves a session by ID
func (s *MemoryStore) GetSession(id string) (*ThinkingSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return session, nil
}

// AddStep adds a step to a session
func (s *MemoryStore) AddStep(sessionID string, content string, stepType StepType, parentStep *int, metadata map[string]any) (*ThinkingStep, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	stepNumber := len(session.Steps) + 1
	step := &ThinkingStep{
		Number:      stepNumber,
		Type:        stepType,
		Content:     content,
		Timestamp:   time.Now(),
		ParentStep:  parentStep,
		Metadata:    metadata,
		Connections: make([]int, 0),
	}

	// Link to parent
	if parentStep != nil && *parentStep > 0 && *parentStep <= len(session.Steps) {
		step.Connections = append(step.Connections, *parentStep)
		session.Steps[*parentStep-1].Connections = append(session.Steps[*parentStep-1].Connections, stepNumber)
	}

	session.Steps = append(session.Steps, step)
	session.CurrentStep = stepNumber
	session.LastModified = time.Now()
	session.QualityScore = calculateQualityScore(session)

	return step, nil
}

// CreateBranch creates a new branch from a specific step
func (s *MemoryStore) CreateBranch(sessionID string, fromStep int, alternativeDesc string) (*Branch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	if fromStep < 1 || fromStep > len(session.Steps) {
		return nil, fmt.Errorf("invalid step number: %d", fromStep)
	}

	branch := &Branch{
		ID:              generateID(),
		FromStep:        fromStep,
		Steps:           make([]*ThinkingStep, 0),
		Created:         time.Now(),
		AlternativeDesc: alternativeDesc,
	}

	session.Branches[branch.ID] = branch
	session.LastModified = time.Now()

	return branch, nil
}

// AddStepToBranch adds a step to a specific branch
func (s *MemoryStore) AddStepToBranch(sessionID, branchID string, content string, stepType StepType) (*ThinkingStep, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	branch, exists := session.Branches[branchID]
	if !exists {
		return nil, fmt.Errorf("branch not found: %s", branchID)
	}

	stepNumber := len(branch.Steps) + 1
	step := &ThinkingStep{
		Number:      stepNumber,
		Type:        stepType,
		Content:     content,
		Timestamp:   time.Now(),
		Connections: make([]int, 0),
	}

	branch.Steps = append(branch.Steps, step)
	session.LastModified = time.Now()

	return step, nil
}

// ListSessions returns all sessions
func (s *MemoryStore) ListSessions() []*ThinkingSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*ThinkingSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// DeleteSession deletes a session
func (s *MemoryStore) DeleteSession(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[id]; !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	delete(s.sessions, id)
	return nil
}

// UpdateSessionStatus updates the status of a session
func (s *MemoryStore) UpdateSessionStatus(id, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	session.Status = status
	session.LastModified = time.Now()
	return nil
}

// Helper functions

func generateInitialAnalysis(problem string) string {
	return fmt.Sprintf("Initial analysis of problem: '%s'\n\nThis problem requires systematic breakdown. Key aspects to consider:\n1. Understanding the core question\n2. Identifying relevant factors\n3. Evaluating potential approaches\n4. Considering constraints and assumptions", problem)
}

func calculateQualityScore(session *ThinkingSession) float64 {
	if len(session.Steps) == 0 {
		return 0.5
	}

	// Simple heuristic: variety of step types, connections, and depth
	typeMap := make(map[StepType]int)
	totalConnections := 0

	for _, step := range session.Steps {
		typeMap[step.Type]++
		totalConnections += len(step.Connections)
	}

	varietyScore := float64(len(typeMap)) / 4.0 // 4 types possible
	connectionScore := float64(totalConnections) / float64(len(session.Steps))
	depthScore := float64(len(session.Steps)) / 20.0 // 20 steps considered thorough

	if depthScore > 1.0 {
		depthScore = 1.0
	}
	if connectionScore > 1.0 {
		connectionScore = 1.0
	}

	score := (varietyScore*0.3 + connectionScore*0.3 + depthScore*0.4)
	if score > 1.0 {
		score = 1.0
	}
	return score
}

// GetMetrics calculates metrics for all sessions
func (s *MemoryStore) GetMetrics(timeRange string) SessionMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := SessionMetrics{
		CommonIssues:    make(map[string]int),
		StepTypeDistrib: make(map[StepType]int),
		SessionsByDay:   make(map[string]int),
		TopTags:         make(map[string]int),
	}

	totalSteps := 0
	totalQuality := 0.0
	totalBranches := 0

	for _, session := range s.sessions {
		// Filter by time range
		if !matchesTimeRange(session.Created, timeRange) {
			continue
		}

		metrics.TotalSessions++

		if session.Status == "active" {
			metrics.ActiveSessions++
		} else if session.Status == "completed" {
			metrics.CompletedSessions++
		}

		totalSteps += len(session.Steps)
		totalQuality += session.QualityScore
		totalBranches += len(session.Branches)

		// Count step types
		for _, step := range session.Steps {
			metrics.StepTypeDistrib[step.Type]++
		}

		// Count sessions by day
		day := session.Created.Format("2006-01-02")
		metrics.SessionsByDay[day]++

		// Count tags
		for _, tag := range session.Tags {
			metrics.TopTags[tag]++
		}
	}

	if metrics.TotalSessions > 0 {
		metrics.AverageSteps = float64(totalSteps) / float64(metrics.TotalSessions)
		metrics.AverageQuality = totalQuality / float64(metrics.TotalSessions)
		metrics.AverageBranches = float64(totalBranches) / float64(metrics.TotalSessions)
	}

	// Detect common patterns
	metrics.CommonPatterns = detectPatterns(s.sessions)

	return metrics
}

func matchesTimeRange(t time.Time, timeRange string) bool {
	now := time.Now()
	switch timeRange {
	case "day":
		return t.After(now.Add(-24 * time.Hour))
	case "week":
		return t.After(now.Add(-7 * 24 * time.Hour))
	case "month":
		return t.After(now.Add(-30 * 24 * time.Hour))
	case "all", "":
		return true
	default:
		return true
	}
}

func detectPatterns(sessions map[string]*ThinkingSession) []ThinkingPattern {
	patterns := []ThinkingPattern{
		{
			Name:        "Hypothesis-Verification",
			Frequency:   0,
			Confidence:  0.8,
			Description: "Pattern of forming hypotheses followed by verification steps",
		},
		{
			Name:        "Progressive-Refinement",
			Frequency:   0,
			Confidence:  0.75,
			Description: "Pattern of iteratively refining conclusions",
		},
		{
			Name:        "Branching-Synthesis",
			Frequency:   0,
			Confidence:  0.85,
			Description: "Pattern of exploring alternatives and merging insights",
		},
	}

	for _, session := range sessions {
		// Detect hypothesis-verification pattern
		hasHypothesis := false
		hasVerification := false
		for _, step := range session.Steps {
			if step.Type == StepHypothesis {
				hasHypothesis = true
			}
			if step.Type == StepVerification && hasHypothesis {
				hasVerification = true
			}
		}
		if hasHypothesis && hasVerification {
			patterns[0].Frequency++
		}

		// Detect progressive refinement
		conclusionCount := 0
		for _, step := range session.Steps {
			if step.Type == StepConclusion {
				conclusionCount++
			}
		}
		if conclusionCount > 1 {
			patterns[1].Frequency++
		}

		// Detect branching-synthesis
		if len(session.Branches) > 0 {
			patterns[2].Frequency++
		}
	}

	return patterns
}
