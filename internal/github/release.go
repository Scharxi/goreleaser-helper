package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"

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

	color.Blue("🚀 Creating release v%s for %s/%s...", opts.Version, owner, repoName)

	// Create release
	releaseID, err := createRelease(owner, repoName, opts)
	if err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	color.Green("✅ Release created successfully!")

	// Upload assets
	color.Blue("📦 Uploading assets...")
	if err := uploadAssets(owner, repoName, releaseID, opts); err != nil {
		return fmt.Errorf("failed to upload assets: %w", err)
	}

	color.Green("✅ All assets uploaded successfully!")
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

	// Parse the JSON response to get the release ID
	var respData struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(body, &respData); err != nil {
		return "", fmt.Errorf("failed to parse release response: %w", err)
	}

	return fmt.Sprintf("%d", respData.ID), nil
}

func uploadAssets(owner, repo, releaseID string, opts ReleaseOptions) error {
	// Create progress bar
	bar := progressbar.NewOptions(len(opts.Binaries),
		progressbar.OptionSetDescription("Uploading assets..."),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// Create error channel and wait group
	errChan := make(chan error, len(opts.Binaries))
	var wg sync.WaitGroup

	// Upload assets concurrently
	for _, binary := range opts.Binaries {
		wg.Add(1)
		go func(b build.BuildResult) {
			defer wg.Done()
			if err := uploadSingleAsset(owner, repo, releaseID, opts.Token, b); err != nil {
				errChan <- fmt.Errorf("failed to upload %s: %w", filepath.Base(b.Path), err)
				return
			}
			bar.Add(1)
		}(binary)
	}

	// Wait for all uploads to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func uploadSingleAsset(owner, repo, releaseID, token string, binary build.BuildResult) error {
	file, err := os.Open(binary.Path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", binary.Path, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Ensure file pointer is at the start
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	assetName := filepath.Base(binary.Path)
	uploadURL := fmt.Sprintf(
		"https://uploads.github.com/repos/%s/%s/releases/%s/assets?name=%s",
		owner, repo, releaseID, assetName,
	)

	req, err := http.NewRequest("POST", uploadURL, file)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.ContentLength = fileInfo.Size()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to upload asset: %s", string(body))
	}

	return nil
}
