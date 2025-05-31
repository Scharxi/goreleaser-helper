package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"goreleaser-helper/internal/build"
	"goreleaser-helper/internal/github"

	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Create a new release",
	Long:  `Create a new release with the specified version tag and repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, _ := cmd.Flags().GetString("repo")
		tag, _ := cmd.Flags().GetString("tag")

		if repo == "" || tag == "" {
			return fmt.Errorf("both --repo and --tag flags are required")
		}

		if os.Getenv("GITHUB_TOKEN") == "" {
			return fmt.Errorf("the GITHUB_TOKEN environment variable must be set")
		}

		// Create release directory structure
		if err := createReleaseStructure(tag); err != nil {
			return fmt.Errorf("failed to create release structure: %w", err)
		}

		// Generate changelog
		if err := generateChangelog(tag); err != nil {
			return fmt.Errorf("failed to generate changelog: %w", err)
		}

		// Build binaries
		if err := buildBinaries(tag); err != nil {
			return fmt.Errorf("failed to build binaries: %w", err)
		}

		// Create GitHub release
		if err := createGitHubRelease(repo, tag); err != nil {
			return fmt.Errorf("failed to create GitHub release: %w", err)
		}

		fmt.Printf("Successfully created release %s\n", tag)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)
	releaseCmd.Flags().StringP("repo", "r", "", "GitHub repository URL (e.g., github.com/user/repo)")
	releaseCmd.Flags().StringP("tag", "t", "", "Release tag (e.g., v1.0.0)")
	releaseCmd.MarkFlagRequired("repo")
	releaseCmd.MarkFlagRequired("tag")
}

func createReleaseStructure(tag string) error {
	distDir := filepath.Join("dist", tag)
	return os.MkdirAll(distDir, 0755)
}

func generateChangelog(tag string) error {
	// TODO: Implement changelog generation
	return nil
}

func buildBinaries(tag string) error {
	distDir := filepath.Join("dist", tag)

	// Prepare build options
	opts := build.BuildOptions{
		OutputDir: distDir,
		Version:   tag,
		MainFile:  "main.go",
		LdFlags:   fmt.Sprintf("-X main.version=%s", tag),
	}

	// Build binaries for all platforms
	return build.Build(opts)
}

func createGitHubRelease(repo, tag string) error {
	// Parse repository URL
	owner, repoName, err := github.ParseRepoURL(repo)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %w", err)
	}

	// Prepare release options
	opts := github.ReleaseOptions{
		Owner:       owner,
		Repo:        repoName,
		Tag:         tag,
		Title:       fmt.Sprintf("Release %s", tag),
		Description: fmt.Sprintf("Release version %s", tag), // TODO: Add changelog content
		AssetsDir:   filepath.Join("dist", tag),
		Draft:       false,
		Prerelease:  false,
	}

	// Create the release
	return github.CreateRelease(opts)
}
