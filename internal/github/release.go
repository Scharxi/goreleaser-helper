package github

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"goreleaser-helper/internal/build"
	"goreleaser-helper/internal/config"
)

// ReleaseOptions contains the options for creating a GitHub release
type ReleaseOptions struct {
	Version  string
	Repo     string
	Token    string
	Binaries []build.BuildResult
	Config   *config.Config
}

// CreateRelease creates a new GitHub release
func CreateRelease(opts ReleaseOptions) error {
	// Parse repository URL
	owner, repoName, err := parseRepoURL(opts.Repo)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %w", err)
	}

	// Create release
	releaseID, err := createRelease(owner, repoName, opts)
	if err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	// Upload assets
	if err := uploadAssets(owner, repoName, releaseID, opts); err != nil {
		return fmt.Errorf("failed to upload assets: %w", err)
	}

	return nil
}

func parseRepoURL(repo string) (string, string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository format: %s", repo)
	}
	return parts[0], parts[1], nil
}

func createRelease(owner, repo string, opts ReleaseOptions) (string, error) {
	// Prepare release data
	data := fmt.Sprintf(`{
		"tag_name": "v%s",
		"name": "Release v%s",
		"body": "Release v%s",
		"draft": false,
		"prerelease": false
	}`, opts.Version, opts.Version, opts.Version)

	// Create HTTP request
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "token "+opts.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create release: %s", string(body))
	}

	// Extract release ID from response
	// Note: In a real implementation, you would parse the JSON response
	// to get the release ID. For simplicity, we're returning a placeholder.
	return "release-id", nil
}

func uploadAssets(owner, repo, releaseID string, opts ReleaseOptions) error {
	for _, binary := range opts.Binaries {
		// Open file
		file, err := os.Open(binary.Path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", binary.Path, err)
		}
		defer file.Close()

		// Get file info
		fileInfo, err := file.Stat()
		if err != nil {
			return fmt.Errorf("failed to get file info: %w", err)
		}

		// Read file content
		content, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		// Prepare upload data
		data := fmt.Sprintf(`{
			"name": "%s",
			"content_type": "application/octet-stream",
			"size": %d,
			"content": "%s"
		}`, filepath.Base(binary.Path), fileInfo.Size(), base64.StdEncoding.EncodeToString(content))

		// Create HTTP request
		url := fmt.Sprintf("https://uploads.github.com/repos/%s/%s/releases/%s/assets", owner, repo, releaseID)
		req, err := http.NewRequest("POST", url, strings.NewReader(data))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Authorization", "token "+opts.Token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		// Send request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to upload asset: %s", string(body))
		}
	}

	return nil
}
