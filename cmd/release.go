package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"goreleaser-helper/internal/build"
	"goreleaser-helper/internal/changelog"
	"goreleaser-helper/internal/config"
	"goreleaser-helper/internal/github"
)

var (
	version     string
	repo        string
	configPath  string
	generateChg bool
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Create a new release",
	Long:  `Create a new release with binaries and changelog`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Use default repository from config if not specified
		if repo == "" {
			repo = cfg.GitHub.DefaultRepo
		}

		// Check required flags
		if version == "" {
			return fmt.Errorf("version is required")
		}
		if repo == "" {
			return fmt.Errorf("repository is required")
		}

		// Check for GitHub token
		token := os.Getenv(cfg.GitHub.TokenEnv)
		if token == "" {
			return fmt.Errorf("GitHub token not found in environment variable %s", cfg.GitHub.TokenEnv)
		}

		// Generate changelog if enabled
		if cfg.Release.Changelog.Enabled || generateChg {
			gen := changelog.NewGenerator(cfg, repo)
			if err := gen.Generate(version); err != nil {
				return fmt.Errorf("failed to generate changelog: %w", err)
			}
		}

		// Build binaries
		buildOpts := build.BuildOptions{
			Version:  version,
			Config:   cfg,
			MainFile: cfg.Build.MainFile,
			LdFlags:  cfg.Build.LdFlags,
		}

		binaries, err := build.BuildBinaries(buildOpts)
		if err != nil {
			return fmt.Errorf("failed to build binaries: %w", err)
		}

		// Create GitHub release
		releaseOpts := github.ReleaseOptions{
			Version:  version,
			Repo:     repo,
			Token:    token,
			Binaries: binaries,
			Config:   cfg,
		}

		if err := github.CreateRelease(releaseOpts); err != nil {
			return fmt.Errorf("failed to create release: %w", err)
		}

		fmt.Printf("Successfully created release %s\n", version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)

	// Add flags
	releaseCmd.Flags().StringVarP(&version, "version", "v", "", "Version to release")
	releaseCmd.Flags().StringVarP(&repo, "repo", "r", "", "GitHub repository (owner/repo)")
	releaseCmd.Flags().StringVarP(&configPath, "config", "c", "goreleaser.yaml", "Path to configuration file")
	releaseCmd.Flags().BoolVarP(&generateChg, "changelog", "g", false, "Generate changelog")
}
