package classifier

const (
	TaskCodeGeneration = "code_generation"
	TaskCodeReview     = "code_review"
	TaskDebugging      = "debugging"
	TaskTestGeneration = "test_generation"
	TaskDocumentation  = "documentation"
	TaskReasoning      = "reasoning"
	TaskDataExtraction = "data_extraction"
	TaskSummarization  = "summarization"
	TaskChat           = "chat"
)

const (
	ComplexityLow    = "low"
	ComplexityMedium = "medium"
	ComplexityHigh   = "high"
)

const (
	SourceRuleBased          = "rule_based"
	SourceAPIFallback        = "api_fallback"
	SourceAPIFallbackTimeout = "api_fallback_timeout"
	SourceMetadataOnly       = "metadata_only"
)

// ConfidenceThreshold is the default threshold below which the API fallback is triggered.
const ConfidenceThreshold = 0.6

// AllTaskTypes returns all valid task types in priority order.
func AllTaskTypes() []string {
	return []string{
		TaskDebugging,
		TaskCodeReview,
		TaskReasoning,
		TaskCodeGeneration,
		TaskTestGeneration,
		TaskDocumentation,
		TaskDataExtraction,
		TaskSummarization,
		TaskChat,
	}
}

// ClassificationResult holds the result of classifying a task.
type ClassificationResult struct {
	TaskType   string  `json:"task_type"`
	Complexity string  `json:"complexity"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source,omitempty"`
}

// Classifier is the interface for task classification.
type Classifier interface {
	Classify(input string) (*ClassificationResult, error)
}

// Provider constants for identifying AI providers.
const (
	ProviderAnthropic = "anthropic"
	ProviderGoogle    = "google"
	ProviderOpenAI    = "openai"
	ProviderMistral   = "mistral"
	ProviderCohere    = "cohere"
	ProviderGroq      = "groq"
	ProviderXAI       = "xai"
	ProviderUnknown   = "unknown"
)

// ProviderFromHost maps API hostnames to provider identifiers.
func ProviderFromHost(host string) string {
	switch {
	case contains(host, "anthropic"):
		return ProviderAnthropic
	case contains(host, "googleapis") || contains(host, "generativelanguage"):
		return ProviderGoogle
	case contains(host, "openai"):
		return ProviderOpenAI
	case contains(host, "mistral"):
		return ProviderMistral
	case contains(host, "cohere"):
		return ProviderCohere
	case contains(host, "groq"):
		return ProviderGroq
	case contains(host, "x.ai"):
		return ProviderXAI
	default:
		return ProviderUnknown
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
