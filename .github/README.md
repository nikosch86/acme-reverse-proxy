# GitHub Actions Setup

This repository includes three GitHub Actions workflows for automated CI/CD:

## Workflows

### 1. CI/CD Pipeline (`ci.yml`)
**Triggers:** Push to any branch, tags, pull requests to main/develop

**Jobs:**
- **Test**: Runs Go tests with race detection and coverage (all branches)
- **Build**: Builds and tests Docker image functionality (all branches)
- **Publish Branch**: Pushes development images to Docker Hub (main/develop only)
- **Publish Release**: Pushes semantic versioned images to Docker Hub (tags only)

### 2. Security Scan (`security.yml`)
**Triggers:** Push to main/develop, pull requests to main/develop, weekly schedule

**Jobs:**
- **Go Security**: Runs Gosec and govulncheck for Go vulnerabilities
- **Docker Security**: Scans Docker image with Trivy for container vulnerabilities

### 3. Release (`release.yml`)
**Triggers:** Semantic version tags (v1.2.3, v1.2.3-alpha.1, etc.)

**Jobs:**
- **Validate Tag**: Ensures tag follows semantic versioning pattern
- **Create Release**: Builds cross-platform binaries and creates GitHub release

## Required Secrets

Configure these secrets in your GitHub repository settings:

### Docker Hub Publishing
```
DOCKER_USERNAME: your-dockerhub-username
DOCKER_PASSWORD: your-dockerhub-password-or-token
```

## Setup Instructions

1. **Fork/Clone this repository**

2. **Set up Docker Hub secrets:**
   - Go to repository Settings → Secrets and variables → Actions
   - Add `DOCKER_USERNAME` with your Docker Hub username
   - Add `DOCKER_PASSWORD` with your Docker Hub password or access token

3. **Update image name (optional):**
   - Edit `.github/workflows/ci.yml`
   - Change `IMAGE_NAME` environment variable to your preferred name
   - Update the Docker Hub repository path in the metadata extraction step

4. **Test the workflows:**
   - Push a commit to trigger the CI pipeline
   - Create a tag like `v1.0.0` to trigger a release

## Workflow Features

### Multi-Architecture Support
- Builds for `linux/amd64` and `linux/arm64`
- Uses Docker Buildx for cross-platform builds

### Caching
- Go module caching for faster builds
- Docker layer caching using GitHub Actions cache

### Security
- Vulnerability scanning for Go dependencies
- Container image security scanning
- SARIF upload for GitHub Security tab integration

### Release Management
- Automatic binary builds for multiple platforms
- Checksum generation for release artifacts
- GitHub release creation with release notes

## Tag Naming Convention

Use strict semantic versioning for releases:
- `v1.0.0` - Major release
- `v1.0.1` - Patch release  
- `v1.1.0` - Minor release
- `v1.0.0-alpha.1` - Pre-release (marked as prerelease)
- `v1.0.0-beta.2` - Beta release
- `v1.0.0-rc.1` - Release candidate

**Note:** Tags must follow the exact pattern `v[major].[minor].[patch][-prerelease]` or the release workflow will fail validation.

## Docker Hub Integration

The workflow automatically:
- **Branch Images**: Builds `main` and `develop` branch images
- **Release Images**: Creates semantic versioned tags (v1.2.3, v1.2, v1)
- **Multi-Architecture**: Supports linux/amd64 and linux/arm64
- **Latest Tag**: Only applied to main branch releases
- **Description Updates**: Syncs README to Docker Hub (main branch only)