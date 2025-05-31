package changelog

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"goreleaser-helper/internal/config"
)

// Entry represents a single changelog entry
type Entry struct {
	Type        string
	Scope       string
	Description string
	Hash        string
	Author      string
	Date        time.Time
}

// Generator handles changelog generation
type Generator struct {
	config *config.Config
	repo   string
}

// NewGenerator creates a new changelog generator
func NewGenerator(cfg *config.Config, repo string) *Generator {
	return &Generator{
		config: cfg,
		repo:   repo,
	}
}

// Generate creates a changelog for the given version
func (g *Generator) Generate(version string) error {
	// Get the last tag
	lastTag, err := g.getLastTag()
	if err != nil {
		return fmt.Errorf("failed to get last tag: %w", err)
	}

	// Get commits since last tag
	entries, err := g.getCommits(lastTag)
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	// Generate changelog content
	content, err := g.formatChangelog(version, entries)
	if err != nil {
		return fmt.Errorf("failed to format changelog: %w", err)
	}

	// Write changelog file
	if err := g.writeChangelog(content); err != nil {
		return fmt.Errorf("failed to write changelog: %w", err)
	}

	return nil
}

func (g *Generator) getLastTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		// If no tags exist, return empty string
		if strings.Contains(err.Error(), "No names found") {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (g *Generator) getCommits(since string) ([]Entry, error) {
	args := []string{"log", "--pretty=format:%H|%an|%ad|%s"}
	if since != "" {
		args = append(args, since+"..HEAD")
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var entries []Entry
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) != 4 {
			continue
		}

		hash := parts[0]
		author := parts[1]
		date, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", parts[2])
		if err != nil {
			continue
		}
		message := parts[3]

		// Parse conventional commit message
		entry := parseCommitMessage(message)
		entry.Hash = hash
		entry.Author = author
		entry.Date = date

		entries = append(entries, entry)
	}

	// Sort entries by date
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date.After(entries[j].Date)
	})

	return entries, nil
}

func parseCommitMessage(message string) Entry {
	// Conventional commit format: type(scope): description
	re := regexp.MustCompile(`^(\w+)(?:\(([\w-]+)\))?:\s*(.+)$`)
	matches := re.FindStringSubmatch(message)

	entry := Entry{
		Type:        "other",
		Description: message,
	}

	if len(matches) >= 4 {
		entry.Type = matches[1]
		if matches[2] != "" {
			entry.Scope = matches[2]
		}
		entry.Description = matches[3]
	}

	return entry
}

func (g *Generator) formatChangelog(version string, entries []Entry) (string, error) {
	var content strings.Builder

	// Write header
	content.WriteString(fmt.Sprintf("# Changelog for %s\n\n", version))
	content.WriteString(fmt.Sprintf("Release date: %s\n\n", time.Now().Format("2006-01-02")))

	// Group entries by type
	groups := make(map[string][]Entry)
	for _, entry := range entries {
		groups[entry.Type] = append(groups[entry.Type], entry)
	}

	// Write entries by type
	types := []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "other"}
	for _, t := range types {
		if entries, ok := groups[t]; ok {
			content.WriteString(fmt.Sprintf("## %s\n\n", formatType(t)))
			for _, entry := range entries {
				scope := ""
				if entry.Scope != "" {
					scope = fmt.Sprintf("(%s) ", entry.Scope)
				}
				content.WriteString(fmt.Sprintf("- %s%s\n", scope, entry.Description))
			}
			content.WriteString("\n")
		}
	}

	// Write contributors
	content.WriteString("## Contributors\n\n")
	authors := make(map[string]bool)
	for _, entry := range entries {
		authors[entry.Author] = true
	}
	for author := range authors {
		content.WriteString(fmt.Sprintf("- %s\n", author))
	}

	return content.String(), nil
}

func formatType(t string) string {
	types := map[string]string{
		"feat":     "Features",
		"fix":      "Bug Fixes",
		"docs":     "Documentation",
		"style":    "Styles",
		"refactor": "Code Refactoring",
		"perf":     "Performance Improvements",
		"test":     "Tests",
		"build":    "Builds",
		"ci":       "Continuous Integration",
		"chore":    "Chores",
		"other":    "Other Changes",
	}
	if formatted, ok := types[t]; ok {
		return formatted
	}
	return strings.Title(t)
}

func (g *Generator) writeChangelog(content string) error {
	path := g.config.Release.Changelog.Path
	if path == "" {
		path = "CHANGELOG.md"
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create changelog directory: %w", err)
	}

	// Read existing changelog if it exists
	var existingContent string
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read existing changelog: %w", err)
		}
		existingContent = string(data)
	}

	// Write new content
	if err := os.WriteFile(path, []byte(content+"\n\n"+existingContent), 0644); err != nil {
		return fmt.Errorf("failed to write changelog: %w", err)
	}

	return nil
}
