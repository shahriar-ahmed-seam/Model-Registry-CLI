#!/bin/bash
# Setup script for initializing git repository with proper configuration

echo "Initializing git repository..."

# Initialize git repo
git init

# Add all files
git add .

# Create initial commit
git commit -m "Initial commit: Model Registry CLI v1.0

Features:
- Git-style version control for AI model files
- S3-compatible blob storage with content-based deduplication
- SQLite metadata storage
- Single binary CLI with init, push, pull, list commands
- Comprehensive error handling with proper exit codes
- Unit and integration tests

This commit includes:
- Core registry implementation
- CLI interface using Cobra
- Configuration management
- Content hashing and deduplication
- Professional documentation
- GitHub Actions workflows
- License and contribution guidelines"

echo ""
echo "Repository initialized successfully!"
echo ""
echo "To connect to GitHub:"
echo "1. Create a new repository on GitHub"
echo "2. Run: git remote add origin https://github.com/YOUR_USERNAME/ml-reg.git"
echo "3. Run: git push -u origin main"
echo ""
echo "Tags for GitHub repository:"
echo "- cli"
echo "- machine-learning"
echo "- model-registry"
echo "- version-control"
echo "- s3"
echo "- golang"
echo "- ai"
echo "- mlops"
echo "- devops"
echo "- storage"