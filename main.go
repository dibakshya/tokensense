package main

import (
	_ "embed"
	"os"

	"github.fkinternal.com/dibakshya-c/tokensense/cmd"
)

// Build-time variables set by goreleaser ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

//go:embed data/model-matrix.yaml
var bundledMatrix []byte

//go:embed docs/token-optimization-guide.md
var bundledGuide []byte

func main() {
	cmd.SetVersionInfo(version, commit, date)
	cmd.BundledMatrix = bundledMatrix
	cmd.BundledGuide = bundledGuide
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
