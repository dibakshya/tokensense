package classifier

import (
	"strings"
)

// taskKeywordEntry holds a keyword and its weight for scoring.
type taskKeywordEntry struct {
	keyword string
	weight  int // multi-word phrases get higher weight
}

// buildKeywords creates weighted keyword entries; multi-word phrases score 3, single words score 1.
func buildKeywords(keywords []string) []taskKeywordEntry {
	var entries []taskKeywordEntry
	for _, kw := range keywords {
		w := 1
		if strings.Contains(kw, " ") {
			w = 3
		}
		entries = append(entries, taskKeywordEntry{keyword: kw, weight: w})
	}
	return entries
}

var taskKeywords = map[string][]taskKeywordEntry{
	TaskCodeGeneration: buildKeywords([]string{
		"implement", "create function", "build a", "generate code",
		"refactor", "add method", "scaffold", "make a function",
		"create a class", "develop", "code a", "program a",
		"write a function", "write code",
	}),
	TaskCodeReview: buildKeywords([]string{
		"review this", "review", "audit this", "audit", "check this code",
		"is this correct", "code quality", "what's wrong",
		"looks good", "lgtm", "race condition", "memory leak",
		"is this implementation", "correct and secure",
	}),
	TaskDebugging: buildKeywords([]string{
		"fix", "bug", "error", "exception", "crash", "not working",
		"why is", "broken", "failing", "doesn't work", "debug",
		"traceback", "stack trace", "exits immediately",
		"null pointer", "returns 500", "segfault", "deadlock",
		"re-rendering", "indexerror",
	}),
	TaskTestGeneration: buildKeywords([]string{
		"unit test", "integration test", "test for", "write test",
		"write tests", "jest test", "pytest", "vitest", "test case",
		"test suite", "coverage test", "spec for",
		"generate test", "create test", "write a test",
		"mock for", "mocks for",
	}),
	TaskDocumentation: buildKeywords([]string{
		"document this", "docstring", "readme", "explain this code",
		"explain this", "explain to", "annotate", "docs for",
		"describe this", "api documentation", "write a readme",
		"write a docstring", "inline comment", "write docs",
		"document this api",
	}),
	TaskReasoning: buildKeywords([]string{
		"analyze", "think through", "tradeoff", "tradeoffs",
		"implication", "implications", "compare", "evaluate",
		"pros and cons", "should i use", "should i",
		"which is better", "which database", "synthesize",
		"architectural", "decision", "event sourcing",
		"recommend", "vs ", " vs ",
	}),
	TaskDataExtraction: buildKeywords([]string{
		"extract", "find all", "list all", "json from",
		"table from", "pull out", "get all", "scrape",
		"structured data", "parse this",
	}),
	TaskSummarization: buildKeywords([]string{
		"summarize", "summary", "tldr", "key points",
		"brief summary", "short version", "condense",
		"main points", "bullet points",
	}),
}

// complexityKeywords provides hints for complexity determination.
var complexityKeywords = map[string][]string{
	ComplexityHigh: {
		"distributed", "architecture", "security", "microservices",
		"concurrent", "scalable", "production", "enterprise",
		"tradeoffs", "implications", "edge cases", "complex",
		"advanced", "optimization", "performance", "system design",
	},
	ComplexityMedium: {
		"module", "service", "integration", "api", "database",
		"authentication", "middleware", "component", "workflow",
		"pipeline", "context", "understanding",
	},
	ComplexityLow: {
		"simple", "basic", "hello world", "example", "quick",
		"template", "boilerplate", "routine", "straightforward",
	},
}

// RuleBasedClassifier classifies tasks using keyword matching.
type RuleBasedClassifier struct{}

// NewRuleBasedClassifier creates a new rule-based classifier.
func NewRuleBasedClassifier() *RuleBasedClassifier {
	return &RuleBasedClassifier{}
}

// Classify classifies the input text using keyword matching.
func (c *RuleBasedClassifier) Classify(input string) (*ClassificationResult, error) {
	lower := strings.ToLower(input)

	scores := make(map[string]int)
	for taskType, entries := range taskKeywords {
		for _, e := range entries {
			if strings.Contains(lower, e.keyword) {
				scores[taskType] += e.weight
			}
		}
	}

	// "write" alone should not strongly signal code_generation if another category has multi-word matches
	// Suppress generic "write" boost if test/doc keywords are present
	if scores[TaskTestGeneration] > 0 && scores[TaskCodeGeneration] > 0 {
		// Only count code_generation if it has multi-word matches beyond generic "write"
		hasSpecific := false
		for _, e := range taskKeywords[TaskCodeGeneration] {
			if e.weight >= 3 && strings.Contains(lower, e.keyword) {
				hasSpecific = true
				break
			}
		}
		if !hasSpecific {
			scores[TaskCodeGeneration] = 0
		}
	}
	if scores[TaskDocumentation] > 0 && scores[TaskCodeGeneration] > 0 {
		hasSpecific := false
		for _, e := range taskKeywords[TaskCodeGeneration] {
			if e.weight >= 3 && strings.Contains(lower, e.keyword) {
				hasSpecific = true
				break
			}
		}
		if !hasSpecific {
			scores[TaskCodeGeneration] = 0
		}
	}

	// Find best match using priority order for ties
	bestType := TaskChat
	bestScore := 0
	priorityOrder := AllTaskTypes()

	for _, taskType := range priorityOrder {
		if score, ok := scores[taskType]; ok && score > bestScore {
			bestScore = score
			bestType = taskType
		}
	}

	// Calculate confidence based on match strength
	confidence := calculateConfidence(bestScore, scores)

	// Determine complexity
	complexity := classifyComplexity(lower)

	return &ClassificationResult{
		TaskType:   bestType,
		Complexity: complexity,
		Confidence: confidence,
		Source:     SourceRuleBased,
	}, nil
}

func calculateConfidence(bestScore int, allScores map[string]int) float64 {
	if bestScore == 0 {
		return 0.2 // Low confidence for no matches (defaults to chat)
	}

	totalMatches := 0
	for _, s := range allScores {
		totalMatches += s
	}

	// Base confidence from match count
	base := float64(bestScore) / float64(totalMatches+1)

	// Scale to useful range (0.4 - 0.95)
	confidence := 0.4 + (base * 0.55)

	// Bonus for multiple keyword matches
	if bestScore >= 3 {
		confidence += 0.15
	} else if bestScore >= 2 {
		confidence += 0.08
	}

	// Cap at 0.95
	if confidence > 0.95 {
		confidence = 0.95
	}

	return confidence
}

func classifyComplexity(input string) string {
	highScore := 0
	for _, kw := range complexityKeywords[ComplexityHigh] {
		if strings.Contains(input, kw) {
			highScore++
		}
	}

	medScore := 0
	for _, kw := range complexityKeywords[ComplexityMedium] {
		if strings.Contains(input, kw) {
			medScore++
		}
	}

	lowScore := 0
	for _, kw := range complexityKeywords[ComplexityLow] {
		if strings.Contains(input, kw) {
			lowScore++
		}
	}

	// Length-based heuristic
	wordCount := len(strings.Fields(input))
	if wordCount > 30 {
		highScore++
	} else if wordCount < 10 {
		lowScore++
	}

	if highScore > medScore && highScore > lowScore {
		return ComplexityHigh
	}
	if medScore > lowScore {
		return ComplexityMedium
	}
	return ComplexityLow
}
