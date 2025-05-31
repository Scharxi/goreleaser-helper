# goreleaser-helper

Ein CLI-Tool, um Go-Projekte einfach für Windows, MacOS und Linux zu bauen und als Release auf GitHub bereitzustellen.

## Voraussetzungen
- [Go](https://golang.org/dl/)
- [GoReleaser](https://goreleaser.com/install/)
- Ein gesetztes `GITHUB_TOKEN` mit Repo-Rechten

## Installation

```sh
go build -o goreleaser-helper main.go
```

## Benutzung

```sh
./goreleaser-helper --repo <repo-url> --tag <tag> [--config <pfad-zur-goreleaser.yaml>]
```

- `--repo`: GitHub Repository (z.B. github.com/user/repo)
- `--tag`: Release Tag (z.B. v1.0.0)
- `--config`: Optionaler Pfad zu einer eigenen GoReleaser-Konfiguration

## Beispiel

```sh
GITHUB_TOKEN=... ./goreleaser-helper --repo github.com/deinuser/deinprojekt --tag v1.0.0
```

## Hinweise
- GoReleaser muss installiert und im PATH verfügbar sein.
- Die GoReleaser-Konfiguration kann individuell angepasst werden.
