# GoReleaser Helper

A command-line tool to simplify the process of creating and managing Go releases, with support for changelog generation, multi-platform builds, and GitHub releases.

## Features

- 🚀 Easy-to-use CLI interface
- 📝 Automatic changelog generation from conventional commits
- 🔨 Multi-platform binary builds (Linux, macOS, Windows)
- 📦 GitHub release creation with asset uploads
- ⚙️ YAML-based configuration
- 🔒 Support for GitHub authentication
- 🎯 Customizable build options and release settings

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/goreleaser-helper.git
cd goreleaser-helper

# Build the tool
go build -o goreleaser-helper

# Move to your PATH (optional)
mv goreleaser-helper /usr/local/bin/
```

## Configuration

Create a `goreleaser.yaml` file in your project root:

```yaml
project:
  name: your-project-name
  description: Your project description
  version: "1.0.0"
  license: MIT
  authors:
    - Your Name <your.email@example.com>

build:
  mainFile: main.go
  outputDir: dist
  platforms:
    - os: darwin
      arch: amd64
    - os: darwin
      arch: arm64
    - os: linux
      arch: amd64
    - os: linux
      arch: arm64
    - os: windows
      arch: amd64
    - os: windows
      arch: arm64
  ldflags: "-X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}"
  env:
    CGO_ENABLED: "0"
  before:
    - go mod tidy
    - go vet ./...
  after:
    - go test ./...

release:
  defaultBranch: main
  changelog:
    enabled: true
    path: CHANGELOG.md
    format: markdown
  assets:
    include:
      - "LICENSE"
      - "README.md"
      - "CHANGELOG.md"
    exclude:
      - "*.tmp"
      - "*.log"

github:
  defaultRepo: "owner/repo"  # Your GitHub repository
  tokenEnv: GITHUB_TOKEN     # Environment variable for GitHub token
  labels:
    - "enhancement"
    - "bug"
    - "documentation"
```

## Usage

### Basic Release

```bash
# Create a release with version 1.0.0
goreleaser-helper release --version 1.0.0 --repo owner/repo
```

### With Changelog Generation

```bash
# Generate changelog and create release
goreleaser-helper release --version 1.0.0 --repo owner/repo --changelog
```

### Using Custom Configuration

```bash
# Use a custom configuration file
goreleaser-helper release --version 1.0.0 --repo owner/repo --config custom-config.yaml
```

### Environment Setup

1. Set your GitHub token:
   ```bash
   export GITHUB_TOKEN=your_github_token
   ```

2. Make sure your repository is properly configured in the `goreleaser.yaml` file.

## Commit Message Format

The changelog generator uses conventional commit messages. Follow this format for your commits:

```
type(scope): description
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or modifying tests
- `build`: Build system changes
- `ci`: CI configuration changes
- `chore`: General maintenance

Example:
```
feat(auth): add OAuth2 support
fix(api): handle rate limiting errors
docs(readme): update installation instructions
```

## Generated Changelog Format

The tool generates changelogs in the following format:

```markdown
# Changelog for v1.0.0

Release date: 2024-03-21

## Features
- (auth) add OAuth2 support
- (api) implement rate limiting

## Bug Fixes
- (api) handle rate limiting errors
- (auth) fix token refresh

## Documentation
- (readme) update installation instructions

## Contributors
- John Doe
- Jane Smith
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
