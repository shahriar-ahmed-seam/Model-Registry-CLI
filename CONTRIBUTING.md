# Contributing to Model Registry CLI

Thank you for your interest in contributing to the Model Registry CLI! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please be respectful and considerate of others when contributing to this project.

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [GitHub Issues](https://github.com/shahriar-ahmed-seam/Model-Registry-CLI/issues)
2. If not, create a new issue with:
   - A clear, descriptive title
   - Steps to reproduce the bug
   - Expected behavior
   - Actual behavior
   - Environment details (OS, Go version, etc.)
   - Any relevant logs or error messages

### Suggesting Features

1. Check if the feature has already been suggested
2. Create a new issue with:
   - A clear description of the feature
   - Use cases and benefits
   - Any implementation ideas or references

### Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Add or update tests as needed
5. Ensure all tests pass: `go test ./...`
6. Update documentation if needed
7. Commit with a descriptive message
8. Push to your fork: `git push origin feature/amazing-feature`
9. Open a Pull Request

## Development Setup

### Prerequisites

- Go 1.20 or later
- Git

### Getting Started

1. Clone your fork:
   ```bash
   git clone https://github.com/shahriar-ahmed-seam/Model-Registry-CLI.git
   cd ml-reg
   ```

2. Build the project:
   ```bash
   go build -o ml-reg .
   ```

3. Run tests:
   ```bash
   go test ./...
   ```

### Code Style

- Follow standard Go conventions
- Run `gofmt` on your code before committing
- Use meaningful variable and function names
- Add comments for public functions and complex logic
- Keep functions small and focused

### Testing

- Write tests for new functionality
- Update existing tests when changing behavior
- Run property-based tests to verify invariants
- Ensure integration tests pass

### Documentation

- Update README.md for user-facing changes
- Update code comments for API changes
- Add examples for new features

## Project Structure

```
.
├── cmd/              # CLI command implementations
├── internal/         # Internal packages (not for external use)
│   ├── blob/        # Blob storage interface
│   ├── config/      # Configuration management
│   ├── errors/      # Error definitions
│   ├── hashing/     # Content hashing
│   ├── metadata/    # Metadata storage
│   └── registry/    # Core registry logic
├── main.go          # Entry point
└── README.md        # Project documentation
```

## Release Process

1. Version bumps follow [Semantic Versioning](https://semver.org/)
2. Create a release tag: `git tag v1.0.0`
3. Push the tag: `git push origin v1.0.0`
4. GitHub Actions will automatically build and release binaries

## Getting Help

- Check the [README.md](README.md) for documentation
- Search existing [Issues](https://github.com/shahriar-ahmed-seam/Model-Registry-CLI/issues)
- Start a [Discussion](https://github.com/shahriar-ahmed-seam/Model-Registry-CLI/discussions)

## License

By contributing, you agree that your contributions will be licensed under the project's [MIT License](LICENSE).
