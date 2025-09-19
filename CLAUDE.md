# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

cat-server is a Go-based project that implements a cat command alternative. The project is in early development stages with a specification-driven development workflow using the Specify framework.

## Development Commands

### Quality Gates (Required for All Changes)
The project constitution mandates these commands must pass before any code is considered complete:

```bash
go vet      # Static analysis
go fmt      # Code formatting
go test     # Run all tests
go build    # Compilation
```

### Development Workflow
Run these commands in order during development:
```bash
go test -v ./...              # Run all tests with verbose output
go test -run TestFunctionName # Run specific test
go test -cover               # Run tests with coverage report
```

### REST API Development Commands
For the current health endpoint implementation:
```bash
go run src/main.go            # Start the HTTP server on :8080
curl http://localhost:8080/health  # Test health endpoint (JSON)
curl -H "Accept: text/html" http://localhost:8080/health  # Test HTML response
curl -H "Accept: text/plain" http://localhost:8080/health # Test text response
go test ./tests/unit/... -v   # Run unit tests
go test ./tests/integration/... -v  # Run integration tests
go test ./tests/contract/... -v     # Run OpenAPI contract tests
go test ./tests/performance/... -v  # Run load tests
go build -o bin/cat-server src/main.go  # Build production binary
```

### Project Structure (Current Implementation)
```
src/
├── server/          # HTTP server implementation
├── handlers/        # HTTP request handlers
└── main.go         # Application entry point

tests/
├── unit/           # Unit tests for individual components
├── integration/    # Integration tests for full workflows
├── contract/       # OpenAPI specification compliance tests
└── performance/    # Load and performance tests

bin/                # Compiled binaries
```

## Project Architecture

### Directory Structure
- `/specs/` - Feature specifications created via Specify framework
- `/.specify/` - Specify framework configuration and templates
- `/.claude/` - Claude Code specific commands and settings
- `/.serena/` - Serena MCP server configuration

### Development Framework
The project uses the Specify framework for specification-driven development:

1. **Feature Creation**: Use `/specify` command to create new features with branch and spec file
2. **Planning Phase**: Features go through planning with `/plan` command
3. **Implementation**: Code implementation follows the generated specifications
4. **Task Management**: Use `/tasks` for breaking down implementation work

### Language and Testing Requirements

From the project constitution (`.specify/memory/constitution.md`):

- **Language**: Go programming language exclusively
- **Testing**: Every function/method MUST have unit tests using Go's built-in testing framework
- **Code Style**: Follow Go community best practices and idiomatic patterns
- **Documentation Language Policy**:
  - **Internal development documentation** (specifications, plans, tasks, design docs): Write in **Japanese** to support the Japanese-speaking development team
  - **External documentation** (README, API docs, user guides, public documentation): Write in **English** to maintain international accessibility
  - **Code comments and commit messages**: Write in **English** for consistency with Go community standards

### Quality Standards

The project enforces strict quality gates:
- All quality gate commands (`go vet`, `go fmt`, `go test`, `go build`) must pass
- Test-Driven Development (TDD) approach required
- Main branch must always be deployable
- Code reviews must verify constitutional compliance

## Specify Framework Integration

When working with features:
- Feature specs are in `/specs/{feature-number}-{feature-name}/spec.md`
- Use Specify commands to maintain proper workflow
- Follow the specification-driven development process
- Ensure all constitutional principles are upheld during implementation

## Serena MCP Configuration

The project is configured as a bash project in Serena (`.serena/project.yml`) with symbolic code analysis tools available for navigation and refactoring.

## Git Management Best Practices

### Files to Include in Git
**Always commit these implementation artifacts**:
- `src/` - Source code
- `tests/` - Test code (unit, integration, contract, performance)
- `go.mod` - Go module definition
- `CLAUDE.md` - Project documentation updates
- `specs/{feature}/` - Feature specifications and design documents

### Files to Exclude from Git
**Never commit these files**:
- `bin/` - Binary files (generated artifacts, can be rebuilt)
- `.claude/` - Claude Code personal settings
- `.serena/` - Serena MCP personal settings
- `.specify/` - Specify framework personal configuration

### Security Checks Before Commit
**Always run these checks before git add**:
```bash
# Check for sensitive information patterns
grep -r -i "password\|secret\|key\|token\|api_key\|private" src/ tests/ go.mod

# Verify localhost usage (acceptable in tests only)
grep -r "localhost" src/ tests/
```

### Pre-Commit Checklist
- [ ] Source code included (`src/`)
- [ ] Tests included (`tests/`)
- [ ] Documentation updated (`CLAUDE.md`, `specs/`)
- [ ] No binary files (`bin/` excluded)
- [ ] No personal settings (`.claude/`, `.serena/`, `.specify/` excluded)
- [ ] Security scan completed (no passwords/keys found)
- [ ] Localhost usage verified (tests only)

### Example Git Commands
```bash
# Proper file selection
git add go.mod src/ tests/ CLAUDE.md specs/{feature-name}/

# Security verification
grep -r -i "password\|secret\|key\|token" src/ tests/ || echo "✅ No sensitive data"

# Standard commit with attribution
git commit -m "Feature description

Implementation details here.

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```