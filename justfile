# aerospace-utils justfile
# Run `just` to see available recipes

# Default recipe: list available commands
default:
    @just --list

# Build the binary (development)
build:
    nix develop -c go build

# Build the binary (release, stripped)
build-release:
    nix develop -c go build -ldflags="-s -w"

# Run the binary with arguments
run *ARGS:
    nix develop -c go run . -- {{ARGS}}

# Run all tests
test:
    nix develop -c go test ./...

# Run tests with verbose output
test-verbose:
    nix develop -c go test -v ./...

# Run a specific test by name
test-run NAME:
    nix develop -c go test -run {{NAME}} ./...

# Run tests for a specific package
test-pkg PKG:
    nix develop -c go test ./{{PKG}}

# Format code
fmt:
    nix develop -c go fmt ./...

# Run static analysis
vet:
    nix develop -c go vet ./...

# Run comprehensive linting
lint:
    nix develop -c golangci-lint run

# Run all checks (format, vet, lint, test)
check: fmt vet lint test

# Tidy go modules
tidy:
    nix develop -c go mod tidy

# Update gomod2nix.toml after dependency changes
gomod2nix:
    nix develop -c gomod2nix

# Build with nix
nix-build:
    nix build

# Run with nix
nix-run *ARGS:
    nix run -- {{ARGS}}
