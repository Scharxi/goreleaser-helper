package github

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// ReleaseOptions contains the configuration for creating a GitHub release
type ReleaseOptions struct {
	Owner       string
	Repo        string
	Tag         string
	Title       string
	Description string
	AssetsDir   string
	Draft       bool
	Prerelease  bool
}

// CreateRelease creates a new GitHub release with the specified options
func CreateRelease(opts ReleaseOptions) error {
	// Create GitHub client
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Create release
	release := &github.RepositoryRelease{
		TagName:         &opts.Tag,
		Name:            &opts.Title,
		Body:            &opts.Description,
		Draft:           &opts.Draft,
		Prerelease:      &opts.Prerelease,
		TargetCommitish: github.String("main"), // or master, depending on your default branch
	}

	// Create the release
	createdRelease, _, err := client.Repositories.CreateRelease(ctx, opts.Owner, opts.Repo, release)
	if err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	// Upload assets
	if err := uploadAssets(ctx, client, opts.Owner, opts.Repo, createdRelease.GetID(), opts.AssetsDir); err != nil {
		return fmt.Errorf("failed to upload assets: %w", err)
	}

	return nil
}

func uploadAssets(ctx context.Context, client *github.Client, owner, repo string, releaseID int64, assetsDir string) error {
	// Read all files in the assets directory
	files, err := os.ReadDir(assetsDir)
	if err != nil {
		return fmt.Errorf("failed to read assets directory: %w", err)
	}

	// Upload each file
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(assetsDir, file.Name())
		fileHandle, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", filePath, err)
		}
		defer fileHandle.Close()

		// Create upload options
		uploadOpts := &github.UploadOptions{
			Name: file.Name(),
		}

		// Upload the file
		_, _, err = client.Repositories.UploadReleaseAsset(
			ctx,
			owner,
			repo,
			releaseID,
			uploadOpts,
			fileHandle,
		)
		if err != nil {
			return fmt.Errorf("failed to upload asset %s: %w", file.Name(), err)
		}

		fmt.Printf("Uploaded asset: %s\n", file.Name())
	}

	return nil
}

// ParseRepoURL extracts owner and repo name from a GitHub repository URL
func ParseRepoURL(repoURL string) (owner, repo string, err error) {
	// Remove protocol and domain if present
	parts := strings.Split(repoURL, "github.com/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid GitHub repository URL: %s", repoURL)
	}

	// Split owner and repo
	ownerRepo := strings.Split(parts[1], "/")
	if len(ownerRepo) != 2 {
		return "", "", fmt.Errorf("invalid GitHub repository URL format: %s", repoURL)
	}

	return ownerRepo[0], ownerRepo[1], nil
}
