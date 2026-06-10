package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/dibakshya/tokensense/internal/advisor"
	"github.com/dibakshya/tokensense/internal/classifier"
	"github.com/dibakshya/tokensense/internal/config"
	"github.com/dibakshya/tokensense/internal/updater"
)

var (
	askNoCloud bool
)

func init() {
	askCmd.Flags().BoolVar(&askNoCloud, "no-cloud", false, "Force rule-based only, no API fallback")
	rootCmd.AddCommand(askCmd)
}

var askCmd = &cobra.Command{
	Use:   "ask",
	Short: "Get model recommendations for a described task",
	Long:  `Describe a task and get ranked model recommendations based on task type and complexity.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.LoadConfig(); err != nil {
			return err
		}

		input := strings.Join(args, " ")

		// Load matrix
		matrix, err := updater.LoadCachedMatrix(BundledMatrix)
		if err != nil {
			fmt.Println("Warning: cannot load model matrix. Showing classification only.")
		}

		// Create classifiers
		ruleClassifier := classifier.NewRuleBasedClassifier()

		var fallback *classifier.GeminiFallback
		if !askNoCloud && config.GetBool("cloud_fallback") {
			apiKey := config.GetString("gemini_api_key")
			fallback = classifier.NewGeminiFallback(apiKey)
		}

		threshold := config.GetFloat64("confidence_threshold")
		if threshold == 0 {
			threshold = classifier.ConfidenceThreshold
		}

		result, err := advisor.Advise(input, ruleClassifier, fallback, matrix, askNoCloud, threshold)
		if err != nil {
			return fmt.Errorf("Classification failed: %w. Showing generic guidance.", err)
		}

		fmt.Print(advisor.RenderAdvice(input, result))
		return nil
	},
}
