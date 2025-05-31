package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"

	"goreleaser-helper/internal/config"
)

// Platform represents a build target platform
type Platform struct {
	OS   string
	Arch string
}

// Common platforms to build for
var platforms = []Platform{
	{OS: "darwin", Arch: "amd64"},
	{OS: "darwin", Arch: "arm64"},
	{OS: "linux", Arch: "amd64"},
	{OS: "linux", Arch: "arm64"},
	{OS: "windows", Arch: "amd64"},
	{OS: "windows", Arch: "arm64"},
}

// BuildOptions contains the options for building binaries
type BuildOptions struct {
	Version  string
	Config   *config.Config
	MainFile string
	LdFlags  string
}

// BuildResult represents the result of a build
type BuildResult struct {
	Path     string
	Platform string
	Arch     string
}

// BuildBinaries builds binaries for all configured platforms
func BuildBinaries(opts BuildOptions) ([]BuildResult, error) {
	var results []BuildResult
	var mu sync.Mutex

	// Create output directory
	outputDir := filepath.Join(opts.Config.Build.OutputDir, opts.Version)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	color.Blue("🔨 Building binaries for %d platforms...", len(opts.Config.Build.Platforms))

	// Create progress bar
	bar := progressbar.NewOptions(len(opts.Config.Build.Platforms),
		progressbar.OptionSetDescription("Building binaries..."),
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
	errChan := make(chan error, len(opts.Config.Build.Platforms))
	var wg sync.WaitGroup

	// Build for each platform concurrently
	for _, platform := range opts.Config.Build.Platforms {
		wg.Add(1)
		go func(p struct {
			OS   string `yaml:"os"`
			Arch string `yaml:"arch"`
		}) {
			defer wg.Done()
			result, err := buildForPlatform(opts, p.OS, p.Arch, outputDir)
			if err != nil {
				errChan <- fmt.Errorf("failed to build for %s/%s: %w", p.OS, p.Arch, err)
				return
			}
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
			bar.Add(1)
		}(platform)
	}

	// Wait for all builds to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	color.Green("✅ All binaries built successfully!")
	return results, nil
}

func buildForPlatform(opts BuildOptions, goos, arch, outputDir string) (BuildResult, error) {
	// Set environment variables
	env := os.Environ()
	env = append(env, fmt.Sprintf("GOOS=%s", goos))
	env = append(env, fmt.Sprintf("GOARCH=%s", arch))
	env = append(env, fmt.Sprintf("CGO_ENABLED=0"))

	// Add custom environment variables from config
	for k, v := range opts.Config.Build.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Prepare output path
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s_%s", opts.Config.Project.Name, goos, arch))
	if goos == "windows" {
		outputPath += ".exe"
	}

	// Prepare build command
	args := []string{"build", "-v"}
	if opts.LdFlags != "" {
		args = append(args, "-ldflags", opts.LdFlags)
	}
	args = append(args, "-o", outputPath)
	if opts.MainFile != "" {
		args = append(args, opts.MainFile)
	}

	// Execute build command
	cmd := exec.Command("go", args...)
	cmd.Env = env

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return BuildResult{}, fmt.Errorf("build command failed: %w\nOutput: %s", err, string(output))
	}

	return BuildResult{
		Path:     outputPath,
		Platform: goos,
		Arch:     arch,
	}, nil
}
