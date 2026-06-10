package classifier

import (
	"fmt"
	"os"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// ModelMatrix is the root of the model matrix YAML.
type ModelMatrix struct {
	Version     string          `yaml:"version"`
	LastUpdated string          `yaml:"last_updated"`
	TaskTypes   []TaskTypeDef   `yaml:"task_types"`
	Models      []ModelDef      `yaml:"models"`
	LoadedAt    time.Time       `yaml:"-"`
}

// TaskTypeDef describes a task type in the matrix.
type TaskTypeDef struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
}

// ModelDef describes a model and its per-task recommendations.
type ModelDef struct {
	ID                  string                       `yaml:"id"`
	Provider            string                       `yaml:"provider"`
	DisplayName         string                       `yaml:"display_name"`
	Tier                string                       `yaml:"tier"`
	ContextWindow       int                          `yaml:"context_window"`
	Pricing             Pricing                      `yaml:"pricing"`
	TaskRecommendations map[string]TaskRecommendation `yaml:"task_recommendations"`
	LastVerified        string                       `yaml:"last_verified"`
}

// Pricing holds per-million-token costs.
type Pricing struct {
	InputPer1MUSD  float64 `yaml:"input_per_1m_usd"`
	OutputPer1MUSD float64 `yaml:"output_per_1m_usd"`
}

// TaskRecommendation holds quality and complexity recommendations.
type TaskRecommendation struct {
	Quality                  int      `yaml:"quality"`
	RecommendedForComplexity []string `yaml:"recommended_for_complexity"`
}

// ModelRecommendation is a ranked model for a specific task.
type ModelRecommendation struct {
	Model           ModelDef
	Quality         int
	CostPerRequest  float64
	IsRecommended   bool
	Reason          string
}

// LoadMatrix loads a model matrix from a YAML file.
func LoadMatrix(path string) (*ModelMatrix, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read matrix file: %w", err)
	}
	return ParseMatrix(data)
}

// ParseMatrix parses a model matrix from YAML bytes.
func ParseMatrix(data []byte) (*ModelMatrix, error) {
	var matrix ModelMatrix
	if err := yaml.Unmarshal(data, &matrix); err != nil {
		return nil, fmt.Errorf("cannot parse matrix YAML: %w", err)
	}
	matrix.LoadedAt = time.Now()
	return &matrix, nil
}

// CostForRequest calculates the USD cost given model pricing and token counts.
func CostForRequest(pricing Pricing, tokensIn, tokensOut int) float64 {
	inputCost := (float64(tokensIn) / 1_000_000) * pricing.InputPer1MUSD
	outputCost := (float64(tokensOut) / 1_000_000) * pricing.OutputPer1MUSD
	return inputCost + outputCost
}

// FindModel finds a model by ID in the matrix.
func (m *ModelMatrix) FindModel(modelID string) *ModelDef {
	for i := range m.Models {
		if m.Models[i].ID == modelID {
			return &m.Models[i]
		}
	}
	return nil
}

// RankModels returns models ranked by cost-effectiveness for the given task and complexity.
func (m *ModelMatrix) RankModels(taskType, complexity string) []ModelRecommendation {
	var recs []ModelRecommendation

	// Estimate tokens for cost comparison (1000 in, 500 out as baseline)
	estTokensIn := 1000
	estTokensOut := 500

	for _, model := range m.Models {
		rec, ok := model.TaskRecommendations[taskType]
		if !ok {
			continue
		}

		isRecommended := false
		for _, c := range rec.RecommendedForComplexity {
			if c == complexity {
				isRecommended = true
				break
			}
		}

		cost := CostForRequest(model.Pricing, estTokensIn, estTokensOut)

		reason := ""
		if isRecommended {
			reason = fmt.Sprintf("Recommended for %s %s tasks (quality: %d/100)", complexity, taskType, rec.Quality)
		} else {
			reason = fmt.Sprintf("Quality: %d/100 — not optimized for %s complexity", rec.Quality, complexity)
		}

		recs = append(recs, ModelRecommendation{
			Model:          model,
			Quality:        rec.Quality,
			CostPerRequest: cost,
			IsRecommended:  isRecommended,
			Reason:         reason,
		})
	}

	// Sort: recommended first, then by quality-adjusted cost
	sort.Slice(recs, func(i, j int) bool {
		if recs[i].IsRecommended != recs[j].IsRecommended {
			return recs[i].IsRecommended
		}
		// Among recommended, sort by cost (cheapest first)
		if recs[i].IsRecommended && recs[j].IsRecommended {
			return recs[i].CostPerRequest < recs[j].CostPerRequest
		}
		// Among non-recommended, sort by quality (highest first)
		return recs[i].Quality > recs[j].Quality
	})

	return recs
}

// IsStaleDays returns the number of days since the matrix was last updated.
func (m *ModelMatrix) IsStaleDays() int {
	t, err := time.Parse("2006-01-02", m.LastUpdated)
	if err != nil {
		return 999
	}
	return int(time.Since(t).Hours() / 24)
}

// IsStale returns true if the matrix is older than the given threshold in days.
func (m *ModelMatrix) IsStale(thresholdDays int) bool {
	return m.IsStaleDays() > thresholdDays
}
