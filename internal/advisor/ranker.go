package advisor

import (
	"github.com/dibakshya/tokensense/internal/classifier"
)

// ModelRecommendation is a ranked model recommendation for display.
type ModelRecommendation struct {
	ModelID     string
	DisplayName string
	Provider    string
	Tier        string
	Quality     int
	CostMin     float64
	CostMax     float64
	SpeedSec    float64
	Recommended bool
	Reason      string
	Warning     string
}

// RankModels returns ranked model recommendations for a given task type and complexity.
func RankModels(matrix *classifier.ModelMatrix, taskType, complexity string) []ModelRecommendation {
	if matrix == nil {
		return nil
	}

	matrixRecs := matrix.RankModels(taskType, complexity)

	var results []ModelRecommendation
	for _, rec := range matrixRecs {
		// Estimate speed based on tier
		speedSec := estimateSpeed(rec.Model.Tier)

		// Cost range for a typical request (500-2000 input tokens, 200-1000 output tokens)
		costMin := classifier.CostForRequest(rec.Model.Pricing, 500, 200)
		costMax := classifier.CostForRequest(rec.Model.Pricing, 2000, 1000)

		mr := ModelRecommendation{
			ModelID:     rec.Model.ID,
			DisplayName: rec.Model.DisplayName,
			Provider:    rec.Model.Provider,
			Tier:        rec.Model.Tier,
			Quality:     rec.Quality,
			CostMin:     costMin,
			CostMax:     costMax,
			SpeedSec:    speedSec,
			Recommended: rec.IsRecommended,
			Reason:      rec.Reason,
		}

		if !rec.IsRecommended {
			mr.Warning = "Not recommended for this complexity level"
		}

		results = append(results, mr)
	}

	return results
}

func estimateSpeed(tier string) float64 {
	switch tier {
	case "fast":
		return 1.0
	case "balanced":
		return 2.5
	case "premium":
		return 5.0
	default:
		return 3.0
	}
}
