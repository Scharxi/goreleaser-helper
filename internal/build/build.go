package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

	// Create output directory
	outputDir := filepath.Join(opts.Config.Build.OutputDir, opts.Version)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build for each platform
	for _, platform := range opts.Config.Build.Platforms {
		result, err := buildForPlatform(opts, platform.OS, platform.Arch, outputDir)
		if err != nil {
			return nil, fmt.Errorf("failed to build for %s/%s: %w", platform.OS, platform.Arch, err)
		}
		results = append(results, result)
	}

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
	args := []string{"build"}
	if opts.MainFile != "" {
		args = append(args, opts.MainFile)
	}
	if opts.LdFlags != "" {
		args = append(args, "-ldflags", opts.LdFlags)
	}
	args = append(args, "-o", outputPath)

	// Execute build command
	cmd := exec.Command("go", args...)
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		return BuildResult{}, fmt.Errorf("build command failed: %w", err)
	}

	return BuildResult{
		Path:     outputPath,
		Platform: goos,
		Arch:     arch,
	}, nil
}
