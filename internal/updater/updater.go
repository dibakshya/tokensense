package updater

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.fkinternal.com/dibakshya-c/tokensense/internal/classifier"
	"github.fkinternal.com/dibakshya-c/tokensense/internal/config"
)

const (
	matrixURL       = "https://github.fkinternal.com/dibakshya-c/tokensense/raw/main/data/model-matrix.yaml"
	stalenessWarnDays = 7
	stalenessMaxDays  = 60
)

// Updater handles background model matrix updates.
type Updater struct {
	version string
	installID string
	logger    *log.Logger
}

// New creates a new Updater.
func New(version, installID string, logger *log.Logger) *Updater {
	if logger == nil {
		logger = log.Default()
	}
	return &Updater{
		version:   version,
		installID: installID,
		logger:    logger,
	}
}

// FetchMatrix downloads the latest model matrix from GitHub.
func (u *Updater) FetchMatrix() (*classifier.ModelMatrix, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", matrixURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Set("User-Agent", fmt.Sprintf("tokensense/%s (install_id/%s)", u.version, u.installID))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("matrix fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("matrix fetch returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read matrix response: %w", err)
	}

	matrix, err := classifier.ParseMatrix(data)
	if err != nil {
		return nil, fmt.Errorf("cannot parse fetched matrix: %w", err)
	}

	// Write to cache atomically
	matrixPath, err := config.MatrixPath()
	if err != nil {
		return matrix, err
	}

	tmpPath := matrixPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return matrix, fmt.Errorf("cannot write matrix cache: %w", err)
	}
	if err := os.Rename(tmpPath, matrixPath); err != nil {
		return matrix, fmt.Errorf("cannot rename matrix cache: %w", err)
	}

	u.logger.Printf("Model matrix updated: version %s, %d models", matrix.Version, len(matrix.Models))
	return matrix, nil
}

// LoadCachedMatrix loads the cached matrix from disk, or the bundled one.
func LoadCachedMatrix(bundledData []byte) (*classifier.ModelMatrix, error) {
	matrixPath, err := config.MatrixPath()
	if err != nil {
		return classifier.ParseMatrix(bundledData)
	}

	// Prefer cached version
	if data, err := os.ReadFile(matrixPath); err == nil {
		matrix, err := classifier.ParseMatrix(data)
		if err == nil {
			return matrix, nil
		}
	}

	// Fall back to bundled
	return classifier.ParseMatrix(bundledData)
}

// CheckStaleness returns a warning message if the matrix is stale.
func CheckStaleness(matrix *classifier.ModelMatrix) string {
	days := matrix.IsStaleDays()
	if days > stalenessMaxDays {
		return fmt.Sprintf("⚠ Model matrix is %d days old (last updated: %s). Run: tokensense config set matrix_auto_update true",
			days, matrix.LastUpdated)
	}
	if days > stalenessWarnDays {
		return fmt.Sprintf("Model matrix is %d days old (last updated: %s)", days, matrix.LastUpdated)
	}
	return ""
}
