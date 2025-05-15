# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build/Run/Test Commands
- Start the demo: `tilt up`
- Run tests: `go test ./...`
- Run single test: `go test -v ./toxic/toxics -run TestDelayHandler`
- Build Go code: `go build ./...`
- Format Go code: `go fmt ./...`
- Vet Go code: `go vet ./...`

## Code Style Guidelines
- **Formatting**: Use standard Go formatting with `go fmt`
- **Imports**: Group standard library imports first, then third-party, then local packages
- **Error Handling**: Check errors and provide context in error messages
- **Naming**: Follow Go conventions (CamelCase for exported, camelCase for internal)
- **Testing**: Write table-driven tests where appropriate
- **Comments**: Use meaningful comments that explain "why" not "what"
- **Dependencies**: Use Go modules for dependency management
- **Logging**: Use github.com/charmbracelet/log for logging