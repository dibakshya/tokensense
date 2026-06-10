package advisor

import (
	"fmt"
	"strings"

	"github.com/dibakshya/tokensense/internal/classifier"
)

// Result holds the full advisor output.
type Result struct {
	Classification *classifier.ClassificationResult
	Recommendations []ModelRecommendation
}

// Advise classifies the input and returns ranked model recommendations.
func Advise(input string, ruleClassifier *classifier.RuleBasedClassifier, fallback *classifier.GeminiFallback, matrix *classifier.ModelMatrix, noCloud bool, confidenceThreshold float64) (*Result, error) {
	if len(strings.Fields(input)) < 3 {
		return nil, fmt.Errorf("please provide a task description of at least 3 words")
	}

	// Step 1: Rule-based classification
	result, err := ruleClassifier.Classify(input)
	if err != nil {
		return nil, fmt.Errorf("classification failed: %w", err)
	}

	// Step 2: API fallback if confidence is low
	if result.Confidence < confidenceThreshold && !noCloud && fallback != nil {
		apiResult, err := fallback.Classify(input)
		if err != nil {
			// Timeout or error — use rule-based result with warning
			result.Source = classifier.SourceAPIFallbackTimeout
		} else {
			result = apiResult
		}
	}

	// Step 3: Rank models
	recs := RankModels(matrix, result.TaskType, result.Complexity)

	return &Result{
		Classification:  result,
		Recommendations: recs,
	}, nil
}

// RenderAdvice renders the advisor result as a terminal-friendly string.
func RenderAdvice(input string, result *Result) string {
	var sb strings.Builder

	cls := result.Classification

	sb.WriteString(fmt.Sprintf("\nClassifying... %s  (confidence: %.2f, %s)\n",
		cls.TaskType, cls.Confidence, cls.Source))
	sb.WriteString(fmt.Sprintf("Complexity: %s\n\n", cls.Complexity))

	sb.WriteString("Model Recommendations\n")
	sb.WriteString("──────────────────────────────────────────────────────────\n")

	for i, rec := range result.Recommendations {
		if i >= 5 {
			break // Show top 5
		}

		prefix := "     "
		if i == 0 {
			prefix = "  ★  "
		}

		sb.WriteString(fmt.Sprintf("%s%-20s ~$%.4f–$%.4f   ~%.0fs   quality: %d/100\n",
			prefix, rec.DisplayName, rec.CostMin, rec.CostMax, rec.SpeedSec, rec.Quality))

		if rec.Warning != "" {
			sb.WriteString(fmt.Sprintf("     ⚠ %s\n", rec.Warning))
		} else {
			sb.WriteString(fmt.Sprintf("     %s\n", rec.Reason))
		}
		sb.WriteString("\n")
	}

	if len(result.Recommendations) > 0 {
		top := result.Recommendations[0]
		sb.WriteString(fmt.Sprintf("→ For %s %s tasks, %s offers the best value.\n",
			cls.Complexity, cls.TaskType, top.DisplayName))
	}
	sb.WriteString("──────────────────────────────────────────────────────────\n")

	return sb.String()
}
