package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/dibakshya/tokensense/internal/cert"
	"github.com/dibakshya/tokensense/internal/classifier"
	"github.com/dibakshya/tokensense/internal/config"
	"github.com/dibakshya/tokensense/internal/proxy"
	"github.com/dibakshya/tokensense/internal/storage"
	"github.com/dibakshya/tokensense/internal/updater"
)

// BundledMatrix is the embedded model matrix YAML, set from main.go.
var BundledMatrix []byte

// BundledGuide is the embedded token optimization guide, set from main.go.
var BundledGuide []byte

var foreground bool

func init() {
	startCmd.Flags().BoolVar(&foreground, "foreground", false, "Run in foreground (used by OS service)")
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Tokensense proxy daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.LoadConfig(); err != nil {
			return err
		}

		// Open database
		dbPath, err := config.DBPath()
		if err != nil {
			return err
		}
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("cannot open database: %w", err)
		}
		defer db.Close()

		// Load CA
		if !cert.CAExists() {
			return fmt.Errorf("CA certificate not found. Run: tokensense setup")
		}
		caCert, caKey, err := cert.LoadCA()
		if err != nil {
			return fmt.Errorf("cannot load CA: %w", err)
		}

		// Load model matrix
		matrix, err := updater.LoadCachedMatrix(BundledMatrix)
		if err != nil {
			log.Printf("Warning: cannot load model matrix: %v", err)
		}

		// Create classifier
		ruleClassifier := classifier.NewRuleBasedClassifier()

		// Configure proxy
		host := config.GetString("proxy_host")
		port := config.GetInt("proxy_port")
		addr := fmt.Sprintf("%s:%d", host, port)
		contentMode := config.GetString("privacy_mode") == "content"

		logger := log.New(os.Stderr, "[tokensense] ", log.LstdFlags)

		srv := proxy.New(proxy.Config{
			Addr:        addr,
			CACert:      caCert,
			CAKey:       caKey,
			DB:          db,
			Classifier:  ruleClassifier,
			Matrix:      matrix,
			ContentMode: contentMode,
			Logger:      logger,
		})

		// Graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigCh
			logger.Println("Shutting down proxy...")
			srv.Shutdown(ctx)
			cancel()
		}()

		fmt.Println()
		fmt.Println(bold("  ╔══════════════════════════════════════════════════╗"))
		fmt.Println(bold("  ║") + green("  ▶️   Tokensense proxy is running               ") + bold("║"))
		fmt.Println(bold("  ╚══════════════════════════════════════════════════╝"))
		fmt.Println()
		fmt.Printf("  Listening on:  %s\n", cyan(addr))
		fmt.Printf("  Privacy mode:  %s\n", cyan(contentModeLabel(contentMode)))
		fmt.Println()
		fmt.Println(dim("  Your AI tool calls are now being tracked silently."))
		fmt.Println(dim("  No prompts or responses are stored — only metadata."))
		fmt.Println()
		fmt.Println(bold("  What to do next:"))
		fmt.Printf("  %-35s →  %s\n", "Check today's stats", cyan("tokensense status"))
		fmt.Printf("  %-35s →  %s\n", "View cost breakdown", cyan("tokensense report"))
		fmt.Printf("  %-35s →  %s\n", "Open visual report", cyan("tokensense report --html --open"))
		fmt.Printf("  %-35s →  %s\n", "For developers/agents", cyan("tokensense api"))
		fmt.Println()
		fmt.Println(dim("  Press Ctrl+C to stop."))
		fmt.Println()
		return srv.ListenAndServe()
	},
}

func contentModeLabel(contentMode bool) string {
	if contentMode {
		return "content (full classification)"
	}
	return "metadata only (privacy-safe)"
}
