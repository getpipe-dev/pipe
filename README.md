<img src="./pipe.png" height='150' alt="Pipe">

# Pipe

Inspired by Taskfile, this solution provides an intuitive replacement for chained alias commands,
transforming them into a CI-friendly standard that can be easily shared.

## Install

### GitHub Releases

Download the latest binary for your platform from the
[Releases](https://github.com/destis/pipe/releases) page.

### Homebrew

Coming soon.

```
brew install destis/tap/pipe
```

## Dependencies

Pipe is a pure-Go binary with a single external dependency:

| Module | Purpose |
|--------|---------|
| [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml) | YAML pipeline parsing |

Everything else — process execution, retry logic, state persistence, logging — uses the Go standard library.

## Development

### Prerequisites

- Go 1.25+ (version pinned in `go.mod`)

### Running tests

```
go test -v -race ./...
```

### Linting

```
go vet ./...
```

CI runs both automatically on every push and pull request via GitHub Actions.
