package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// BuildOptions contains the configuration for building binaries
type BuildOptions struct {
	OutputDir string
	Version   string
	MainFile  string
	LdFlags   string
}

// Build creates binaries for all specified platforms
func Build(opts BuildOptions) error {
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Build for each platform
	for _, platform := range platforms {
		if err := buildForPlatform(platform, opts, wd); err != nil {
			return fmt.Errorf("failed to build for %s/%s: %w", platform.OS, platform.Arch, err)
		}
	}

	return nil
}

func buildForPlatform(platform Platform, opts BuildOptions, wd string) error {
	// Set the output filename
	outputName := fmt.Sprintf("%s_%s_%s", filepath.Base(wd), platform.OS, platform.Arch)
	if platform.OS == "windows" {
		outputName += ".exe"
	}
	outputPath := filepath.Join(opts.OutputDir, outputName)

	// Set build environment variables
	env := os.Environ()
	env = append(env, fmt.Sprintf("GOOS=%s", platform.OS))
	env = append(env, fmt.Sprintf("GOARCH=%s", platform.Arch))

	// Prepare build command
	args := []string{"build"}
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Building for %s/%s...\n", platform.OS, platform.Arch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	return nil
}
