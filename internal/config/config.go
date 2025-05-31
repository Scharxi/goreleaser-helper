package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	// Project configuration
	Project struct {
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
		Version     string   `yaml:"version"`
		License     string   `yaml:"license"`
		Authors     []string `yaml:"authors"`
	} `yaml:"project"`

	// Build configuration
	Build struct {
		MainFile  string `yaml:"mainFile"`
		OutputDir string `yaml:"outputDir"`
		Platforms []struct {
			OS   string `yaml:"os"`
			Arch string `yaml:"arch"`
		} `yaml:"platforms"`
		LdFlags string            `yaml:"ldflags"`
		Env     map[string]string `yaml:"env"`
		Before  []string          `yaml:"before"` // Commands to run before build
		After   []string          `yaml:"after"`  // Commands to run after build
	} `yaml:"build"`

	// Release configuration
	Release struct {
		DefaultBranch string `yaml:"defaultBranch"`
		Changelog     struct {
			Enabled bool   `yaml:"enabled"`
			Path    string `yaml:"path"`
			Format  string `yaml:"format"` // markdown, json, etc.
		} `yaml:"changelog"`
		Assets struct {
			Include []string `yaml:"include"` // Glob patterns for files to include
			Exclude []string `yaml:"exclude"` // Glob patterns for files to exclude
		} `yaml:"assets"`
		Sign struct {
			Enabled bool   `yaml:"enabled"`
			Key     string `yaml:"key"`
			Pass    string `yaml:"pass"`
		} `yaml:"sign"`
	} `yaml:"release"`

	// GitHub configuration
	GitHub struct {
		DefaultRepo string   `yaml:"defaultRepo"`
		TokenEnv    string   `yaml:"tokenEnv"`
		Labels      []string `yaml:"labels"`
		Milestones  []string `yaml:"milestones"`
		Teams       []string `yaml:"teams"`
	} `yaml:"github"`

	// Notifications configuration
	Notifications struct {
		Slack struct {
			Enabled bool   `yaml:"enabled"`
			Webhook string `yaml:"webhook"`
		} `yaml:"slack"`
		Discord struct {
			Enabled bool   `yaml:"enabled"`
			Webhook string `yaml:"webhook"`
		} `yaml:"discord"`
		Email struct {
			Enabled bool     `yaml:"enabled"`
			From    string   `yaml:"from"`
			To      []string `yaml:"to"`
			SMTP    struct {
				Host     string `yaml:"host"`
				Port     int    `yaml:"port"`
				Username string `yaml:"username"`
				Password string `yaml:"password"`
			} `yaml:"smtp"`
		} `yaml:"email"`
	} `yaml:"notifications"`
}

// Load reads and parses the configuration file
func Load(configPath string) (*Config, error) {
	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults if not specified
	setDefaults(&config)

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for configuration fields
func setDefaults(config *Config) {
	// Project defaults
	if config.Project.Name == "" {
		config.Project.Name = filepath.Base(getCurrentDir())
	}

	// Build defaults
	if config.Build.MainFile == "" {
		config.Build.MainFile = "main.go"
	}
	if config.Build.OutputDir == "" {
		config.Build.OutputDir = "dist"
	}

	// Set default platforms if none specified
	if len(config.Build.Platforms) == 0 {
		config.Build.Platforms = []struct {
			OS   string `yaml:"os"`
			Arch string `yaml:"arch"`
		}{
			{OS: "darwin", Arch: "amd64"},
			{OS: "darwin", Arch: "arm64"},
			{OS: "linux", Arch: "amd64"},
			{OS: "linux", Arch: "arm64"},
			{OS: "windows", Arch: "amd64"},
			{OS: "windows", Arch: "arm64"},
		}
	}

	// Release defaults
	if config.Release.DefaultBranch == "" {
		config.Release.DefaultBranch = "main"
	}
	if config.Release.Changelog.Format == "" {
		config.Release.Changelog.Format = "markdown"
	}

	// GitHub defaults
	if config.GitHub.TokenEnv == "" {
		config.GitHub.TokenEnv = "GITHUB_TOKEN"
	}
}

// validateConfig validates the configuration values
func validateConfig(config *Config) error {
	// Validate project name
	if !isValidProjectName(config.Project.Name) {
		return fmt.Errorf("invalid project name: %s", config.Project.Name)
	}

	// Validate version format
	if config.Project.Version != "" && !isValidVersion(config.Project.Version) {
		return fmt.Errorf("invalid version format: %s", config.Project.Version)
	}

	// Validate platforms
	for _, platform := range config.Build.Platforms {
		if !isValidPlatform(platform.OS, platform.Arch) {
			return fmt.Errorf("invalid platform: %s/%s", platform.OS, platform.Arch)
		}
	}

	// Validate GitHub configuration
	if config.GitHub.DefaultRepo != "" && !isValidRepoURL(config.GitHub.DefaultRepo) {
		return fmt.Errorf("invalid GitHub repository URL: %s", config.GitHub.DefaultRepo)
	}

	return nil
}

// ProcessTemplate processes template strings in the configuration
func (c *Config) ProcessTemplate(data interface{}) error {
	// Process build ldflags
	if c.Build.LdFlags != "" {
		tmpl, err := template.New("ldflags").Parse(c.Build.LdFlags)
		if err != nil {
			return fmt.Errorf("invalid ldflags template: %w", err)
		}
		var buf strings.Builder
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("failed to process ldflags template: %w", err)
		}
		c.Build.LdFlags = buf.String()
	}

	return nil
}

// Helper functions for validation
func isValidProjectName(name string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*$`).MatchString(name)
}

func isValidVersion(version string) bool {
	return regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$`).MatchString(version)
}

func isValidPlatform(os, arch string) bool {
	validOS := map[string]bool{
		"darwin":  true,
		"linux":   true,
		"windows": true,
	}
	validArch := map[string]bool{
		"amd64": true,
		"arm64": true,
	}
	return validOS[os] && validArch[arch]
}

func isValidRepoURL(url string) bool {
	return regexp.MustCompile(`^github\.com/[a-zA-Z0-9-]+/[a-zA-Z0-9-]+$`).MatchString(url)
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

// Save writes the configuration to a file
func (c *Config) Save(configPath string) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal the config to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write the file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
