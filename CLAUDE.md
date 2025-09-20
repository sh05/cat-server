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
For the current health endpoint and file list endpoint implementation:
```bash
# Server startup commands
go run src/main.go                      # Start server with default directory (./files/)
go run src/main.go -dir ./custom-dir    # Start server with custom directory

# Health endpoint testing
curl http://localhost:8080/health       # Test health endpoint (JSON)
curl -H "Accept: text/html" http://localhost:8080/health  # Test HTML response
curl -H "Accept: text/plain" http://localhost:8080/health # Test text response

# File list endpoint testing (/ls)
curl http://localhost:8080/ls         # Get file list from configured directory
curl -s http://localhost:8080/ls | jq .  # Pretty print JSON response
curl -w "%{time_total}\n" -s http://localhost:8080/ls >/dev/null  # Measure response time

# File content endpoint testing (/cat/{filename})
curl http://localhost:8080/cat/example.txt  # Get specific file content
curl -s http://localhost:8080/cat/config.json | jq .  # Pretty print file content JSON
curl -s http://localhost:8080/cat/.env | jq .content  # Extract only content field
curl -w "%{time_total}\n" -s http://localhost:8080/cat/README.md >/dev/null  # Measure response time

# Testing with different directories
mkdir -p ./test-files && echo "test content" > ./test-files/sample.txt
go run src/main.go -dir ./test-files     # Test with custom directory

# Error case testing (/ls)
curl -X POST http://localhost:8080/ls # Test method not allowed (should return 405)

# Error case testing (/cat/{filename})
curl http://localhost:8080/cat/nonexistent.txt  # Test file not found (should return 404)
curl http://localhost:8080/cat/../etc/passwd    # Test path traversal (should return 400)
curl -X POST http://localhost:8080/cat/test.txt # Test method not allowed (should return 405)

# Development testing commands
go test ./tests/unit/... -v             # Run unit tests
go test ./tests/integration/... -v      # Run integration tests
go test ./tests/contract/... -v         # Run OpenAPI contract tests
go test ./tests/performance/... -v      # Run load tests
go test ./specs/004-list-get-request/contracts/ -v  # Run ls endpoint contract tests
go test ./specs/005-cat-filename-ls/contracts/ -v  # Run cat endpoint contract tests

# Build commands
go build -o bin/cat-server src/main.go  # Build production binary
./bin/cat-server -dir ./files           # Run production binary
```

### Docker Development Commands
For containerized deployment and development:
```bash
# Docker image build
docker build -t cat-server:latest .     # Build Docker image with Alpine base
docker build --no-cache -t cat-server:latest .  # Build without cache

# Container management
docker run -d --name cat-server -p 8080:8080 cat-server:latest  # Start container
docker run -d --name cat-server -p 8080:8080 -v $(pwd)/files:/app/files cat-server:latest  # With volume mount
docker run -it --rm cat-server:latest /bin/sh  # Interactive debugging

# Container inspection and debugging
docker ps                              # List running containers
docker logs cat-server                 # View container logs
docker exec cat-server ps aux          # Check processes inside container
docker inspect cat-server --format='{{.State.Health.Status}}'  # Health status

# Image analysis
docker images cat-server               # List cat-server images
docker history cat-server:latest       # Show image layers
docker inspect cat-server:latest       # Detailed image information

# Container testing
curl http://localhost:8080/health       # Test containerized health endpoint
curl http://localhost:8080/ls          # Test containerized file list
curl http://localhost:8080/cat/go.mod   # Test containerized file content

# Cleanup
docker stop cat-server                 # Stop container
docker rm cat-server                   # Remove container
docker rmi cat-server:latest           # Remove image
docker system prune -f                 # Clean up unused resources

# Performance monitoring
docker stats cat-server --no-stream    # Check resource usage
time docker run --rm cat-server:latest --help  # Measure startup time

# Security checks
docker run --rm cat-server:latest whoami  # Verify non-root user
docker scout cves cat-server:latest    # Vulnerability scan (if available)
```

### Project Structure (Current Implementation)
```
src/
├── server/          # HTTP server implementation
├── handlers/        # HTTP request handlers (health.go, list.go, cat.go)
├── services/        # Business logic services (directory.go)
└── main.go         # Application entry point

tests/
├── unit/           # Unit tests for individual components
├── integration/    # Integration tests for full workflows
├── contract/       # OpenAPI specification compliance tests
└── performance/    # Load and performance tests

specs/              # Feature specifications (Specify framework)
├── 004-list-get-request/
│   ├── spec.md         # Feature specification
│   ├── plan.md         # Implementation plan
│   ├── research.md     # Technical research
│   ├── data-model.md   # Data models and entities
│   ├── quickstart.md   # Demo and testing guide
│   └── contracts/      # OpenAPI specs and contract tests

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
**These files are automatically excluded by .gitignore**:
- `bin/` - Binary files (generated artifacts, can be rebuilt)
- `.claude/` - Claude Code personal settings
- `.serena/` - Serena MCP personal settings
- `.specify/` - Specify framework personal configuration
- OS generated files (`.DS_Store`, `Thumbs.db`, etc.)
- Log files (`*.log`)

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
- [ ] .gitignore properly excludes build artifacts and personal settings
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