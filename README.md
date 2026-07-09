# Model Registry CLI (ml-reg)

A single-binary CLI that provides Git-style version control for large AI model files, storing blobs in S3-compatible object storage and version metadata in a local SQLite database.

[![Go Report Card](https://goreportcard.com/badge/github.com/shahriar-ahmed-seam/Model-Registry-CLI)](https://goreportcard.com/report/github.com/shahriar-ahmed-seam/Model-Registry-CLI)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/shahriar-ahmed-seam/Model-Registry-CLI)](https://github.com/shahriar-ahmed-seam/Model-Registry-CLI)

## Features

- **Git-like workflow**: Push, pull, and list model versions with familiar commands
- **Content-based deduplication**: Skips uploading identical files to save bandwidth and storage
- **S3-compatible storage**: Works with AWS S3, MinIO, and other S3-compatible object stores
- **Local metadata**: Version information stored in SQLite for fast queries
- **Atomic operations**: All-or-nothing pushes with proper error handling
- **Cross-platform**: Single binary for Linux, macOS, and Windows

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/shahriar-ahmed-seam/Model-Registry-CLI.git
cd ml-reg

# Build the binary
go build -o ml-reg .

# Install globally (optional)
sudo mv ml-reg /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/shahriar-ahmed-seam/Model-Registry-CLI@latest
```

### Download Binary

Check the [Releases](https://github.com/shahriar-ahmed-seam/Model-Registry-CLI/releases) page for pre-built binaries.

## Quick Start

### 1. Initialize a Registry

```bash
# Initialize with MinIO (local S3-compatible storage)
ml-reg init \
  --endpoint http://localhost:9000 \
  --bucket models \
  --access-key minioadmin \
  --secret-key minioadmin

# Initialize with AWS S3
ml-reg init \
  --endpoint https://s3.amazonaws.com \
  --bucket my-model-registry \
  --region us-east-1 \
  --creds-ref aws-profile-name
```

### 2. Push Model Versions

```bash
# Push a model with a version tag
ml-reg push model.pkl --tag v1.0

# Push with accuracy metric
ml-reg push model.pkl --tag v1.1 --accuracy 0.95

# Push and skip if already exists (deduplication)
ml-reg push model.pkl --tag v1.2
```

### 3. List Available Models

```bash
# List all versions
ml-reg list

# Output:
# TAG   HASH                                     SIZE      ACCURACY   CREATED
# v1.0  a1b2c3...                                150MB     0.92       2024-01-15 10:30:00
# v1.1  d4e5f6...                                152MB     0.95       2024-01-15 11:15:00
```

### 4. Pull Models

```bash
# Pull by tag
ml-reg pull v1.0 model-v1.0.pkl

# Pull to current directory with default name
ml-reg pull v1.1
```

## Command Reference

### `ml-reg init`

Initialize a new model registry.

```bash
ml-reg init --endpoint URL --bucket NAME [options]
```

**Required Flags:**
- `--endpoint`: S3-compatible endpoint URL
- `--bucket`: Storage bucket name

**Optional Flags:**
- `--access-key`, `--secret-key`: S3 credentials (can use AWS profile instead)
- `--creds-ref`: AWS credentials profile name
- `--region`: AWS region (defaults to endpoint region)
- `--metadata-path`: SQLite database path (default: `.ml-reg/registry.db`)
- `--force`: Overwrite existing configuration

### `ml-reg push`

Push a model file to the registry.

```bash
ml-reg push MODEL_FILE --tag TAG [options]
```

**Required Flags:**
- `--tag`: Unique version tag (e.g., `v1.0`, `production-2024-01-15`)

**Optional Flags:**
- `--accuracy`: Model accuracy metric (float value)
- `--description`: Human-readable description (not implemented in MVP)

### `ml-reg pull`

Pull a model version from the registry.

```bash
ml-reg pull TAG [DESTINATION_FILE]
```

**Arguments:**
- `TAG`: Version tag to pull
- `DESTINATION_FILE`: Output file path (optional, defaults to tag name)

### `ml-reg list`

List all model versions in the registry.

```bash
ml-reg list
```

## Configuration

### Credentials

The CLI supports multiple authentication methods:

1. **AWS Profile** (recommended):
   ```bash
   ml-reg init --endpoint ... --bucket ... --creds-ref my-profile
   ```

2. **Environment Variables**:
   ```bash
   export AWS_ACCESS_KEY_ID="your-access-key"
   export AWS_SECRET_ACCESS_KEY="your-secret-key"
   ml-reg init --endpoint ... --bucket ...
   ```

3. **Direct Credentials**:
   ```bash
   ml-reg init --endpoint ... --bucket ... \
     --access-key "your-access-key" \
     --secret-key "your-secret-key"
   ```

### Configuration Files

After initialization, configuration is stored in `.ml-reg/config.json`:
```json
{
  "endpoint": "http://localhost:9000",
  "bucket": "models",
  "creds_ref": "minio",
  "metadata_path": ".ml-reg/registry.db"
}
```

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│    CLI Layer    │    │   Registry      │    │   Metadata      │
│  (Cobra/cmd)    │◄──►│  (Orchestrator) │◄──►│   (SQLite)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Error Mapping  │    │   Blob Store    │    │   Hashing       │
│  (Exit Codes)   │    │   (S3/MinIO)    │    │   (SHA256)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Key Components

- **CLI Layer**: Command parsing and user interaction using Cobra
- **Registry**: Core orchestrator coordinating all operations
- **Blob Store**: S3-compatible storage for model files
- **Metadata Store**: SQLite database for version records
- **Hashing**: SHA256 content-based deduplication

## Development

### Prerequisites

- Go 1.20 or later
- SQLite3 development headers
- AWS SDK for Go dependencies

### Building

```bash
# Build for current platform
go build -o ml-reg .

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o ml-reg-linux-amd64 .

# Run tests
go test ./...

# Run integration tests
go test ./cmd_test.go
```

### Project Structure

```
.
├── cmd/              # CLI command implementations
├── internal/
│   ├── blob/        # Blob storage interface and implementations
│   ├── config/      # Configuration management
│   ├── errors/      # Sentinel error definitions
│   ├── hashing/     # Content hashing utilities
│   ├── metadata/    # Metadata store interface and SQLite implementation
│   └── registry/    # Core registry logic
├── main.go          # Entry point
└── cmd_test.go      # Integration tests
```

## Testing

### Unit Tests

```bash
# Run all unit tests
go test ./internal/...

# Run specific package tests
go test ./internal/registry
go test ./internal/metadata
go test ./internal/blob
```

### Integration Tests

```bash
# Run CLI integration tests
go test ./cmd_test.go
```

### Property-Based Tests

The project includes property-based tests to verify core invariants:

```bash
# Run property tests
go test ./internal/registry -run TestPushProperties
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

### Development Guidelines

- Follow Go conventions and style guide
- Write tests for new functionality
- Update documentation for API changes
- Use meaningful commit messages
- Keep the single-binary philosophy

## Roadmap

- [ ] Support for model metadata beyond accuracy
- [ ] Tag aliases and branching
- [ ] Batch operations
- [ ] Webhook notifications
- [ ] REST API server
- [ ] GUI dashboard
- [ ] Plugin system for custom storage backends

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/shahriar-ahmed-seam/Model-Registry-CLI/issues)
- **Discussions**: [GitHub Discussions](https://github.com/shahriar-ahmed-seam/Model-Registry-CLI/discussions)
- **Email**: shahriar-ahmed-seam (via GitHub)

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- Uses [AWS SDK for Go](https://github.com/aws/aws-sdk-go-v2) for S3 operations
- Inspired by Git workflow and content-addressable storage

---

<p align="center">
Made with ❤️ for the ML community
</p>
