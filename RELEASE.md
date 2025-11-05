# JARVIS Release Process

This document describes the automated release process for JARVIS using GoReleaser and GitHub Actions.

## Overview

JARVIS uses GoReleaser to automate the entire release process, including:
- Building binaries for multiple platforms (Linux, macOS, Windows)
- Creating distribution archives
- Generating checksums and SBOMs
- Building and publishing multi-arch Docker images
- Creating GitHub releases with auto-generated changelogs

## Release Workflow

### Triggering a Release

Releases are automatically triggered when you push a new tag:

```bash
# Create and push a new tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### What Happens During a Release

1. **Build Binaries** - Compiles JARVIS for multiple platforms:
   - Linux (amd64, arm64, arm/v7)
   - macOS (amd64, arm64, Universal Binary)
   - Windows (amd64, arm64)

2. **Create Archives** - Packages binaries with documentation:
   - tar.gz for Linux/macOS
   - zip for Windows
   - Includes README, LICENSE, docs, and config.yaml

3. **Generate Security Artifacts**:
   - SHA256 checksums for all artifacts
   - Software Bill of Materials (SBOM) for compliance

4. **Build Docker Images**:
   - Multi-arch images (amd64, arm64)
   - Published to GitHub Container Registry (ghcr.io)
   - Tagged with version and `latest`

5. **Create GitHub Release**:
   - Auto-generated changelog
   - All artifacts attached
   - Release notes with installation instructions

## Docker Images

### Available Images

After each release, Docker images are available at:

```bash
# Version-specific (multi-arch)
ghcr.io/dipjyotimetia/jarvis:v1.0.0

# Latest (multi-arch)
ghcr.io/dipjyotimetia/jarvis:latest

# Architecture-specific
ghcr.io/dipjyotimetia/jarvis:v1.0.0-amd64
ghcr.io/dipjyotimetia/jarvis:v1.0.0-arm64
```

### Using Docker Images

```bash
# Pull the latest image
docker pull ghcr.io/dipjyotimetia/jarvis:latest

# Run JARVIS in a container
docker run --rm ghcr.io/dipjyotimetia/jarvis:latest version

# Run with volume mounts for config and specs
docker run --rm \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/specs:/app/specs \
  -p 8080:8080 \
  -p 9090:9090 \
  ghcr.io/dipjyotimetia/jarvis:latest proxy --record

# Interactive mode
docker run --rm -it \
  ghcr.io/dipjyotimetia/jarvis:latest \
  /bin/sh
```

### Docker Image Features

- **Minimal Size**: Based on Alpine Linux
- **Security**: Runs as non-root user
- **Multi-arch**: Supports AMD64 and ARM64
- **Health Check**: Built-in health check endpoint
- **Labels**: OCI-compliant labels for metadata

## Build Configuration

### Optimizations Applied

The GoReleaser configuration includes several optimizations:

1. **Build Flags**:
   - `-s -w`: Strip debug symbols (reduces binary size by 60-70%)
   - `-trimpath`: Remove file system paths from binary
   - `-mod=readonly`: Ensure reproducible builds

2. **LDFlags**:
   ```go
   -X github.com/dipjyotimetia/jarvis/cmd.Version={{ .Version }}
   -X github.com/dipjyotimetia/jarvis/cmd.Commit={{ .Commit }}
   -X github.com/dipjyotimetia/jarvis/cmd.BuildDate={{ .Date }}
   -X github.com/dipjyotimetia/jarvis/cmd.BuiltBy=goreleaser
   ```

3. **Multi-Platform Support**:
   - Linux: amd64, arm64, arm/v7
   - macOS: amd64, arm64, Universal Binary
   - Windows: amd64, arm64

4. **Docker Optimizations**:
   - Multi-stage builds
   - Layer caching (GitHub Actions cache)
   - Multi-arch manifest

## Changelog

The changelog is automatically generated from commit messages using conventional commits:

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New features
- `fix`: Bug fixes
- `perf`: Performance improvements
- `refactor`: Code refactoring
- `docs`: Documentation changes
- `test`: Test updates
- `chore`: Build process or auxiliary tool changes
- `ci`: CI configuration changes

### Example Commits

```bash
# Feature
git commit -m "feat(api): add GraphQL support for API queries"

# Bug fix
git commit -m "fix(proxy): resolve connection timeout issue"

# Performance
git commit -m "perf(db): optimize database connection pooling"

# Breaking change
git commit -m "feat(cli)!: change command structure

BREAKING CHANGE: renamed 'generate' to 'gen' command"
```

## Release Checklist

Before creating a release:

- [ ] Update VERSION in Makefile
- [ ] Run tests: `make test`
- [ ] Run linter: `make lint`
- [ ] Update CHANGELOG.md (if needed)
- [ ] Update documentation
- [ ] Test build locally: `goreleaser build --snapshot --clean`
- [ ] Create and push tag

## Local Testing

Test the release process locally without publishing:

```bash
# Install GoReleaser
brew install goreleaser

# Test build (no publish)
goreleaser build --snapshot --clean

# Test full release (no publish)
goreleaser release --snapshot --clean

# Check generated artifacts
ls -la dist/
```

## Artifacts Generated

After a successful release, the following artifacts are created:

```
dist/
├── jarvis_1.0.0_linux_x86_64.tar.gz
├── jarvis_1.0.0_linux_arm64.tar.gz
├── jarvis_1.0.0_darwin_x86_64.tar.gz
├── jarvis_1.0.0_darwin_arm64.tar.gz
├── jarvis_1.0.0_darwin_all.tar.gz        # Universal Binary
├── jarvis_1.0.0_windows_x86_64.zip
├── jarvis_1.0.0_windows_arm64.zip
├── checksums.txt                         # SHA256 checksums
├── jarvis_1.0.0_linux_x86_64.spdx.json  # SBOMs
├── jarvis_1.0.0_linux_arm64.spdx.json
└── ...
```

## Docker Build Process

### Build Stages

1. **QEMU Setup**: Enables multi-architecture emulation
2. **Buildx Setup**: Configures Docker Buildx for multi-platform builds
3. **Login**: Authenticates with GitHub Container Registry
4. **Build**: Builds images for each architecture
5. **Manifest**: Creates multi-arch manifest
6. **Push**: Pushes images to registry

### Build Flags

```dockerfile
--pull                                    # Always pull base images
--platform=linux/amd64                    # Target platform
--cache-from=type=gha                     # Use GitHub Actions cache
--cache-to=type=gha,mode=max              # Save to cache
--label=org.opencontainers.image.*       # OCI labels
```

## Security

### SBOM (Software Bill of Materials)

SBOMs are automatically generated for:
- All binary artifacts
- All archive artifacts

SBOMs are in SPDX format and can be used for:
- Security audits
- Compliance verification
- Dependency tracking
- Vulnerability scanning

### Checksums

SHA256 checksums are generated for all artifacts:

```bash
# Verify a downloaded binary
sha256sum -c checksums.txt
```

## Troubleshooting

### Release Fails

If a release fails:

1. Check GitHub Actions logs
2. Verify all tests pass: `make test`
3. Test locally: `goreleaser release --snapshot --clean`
4. Check for breaking changes in GoReleaser config

### Docker Build Fails

If Docker build fails:

1. Verify Dockerfile.goreleaser exists
2. Check binary is being copied correctly
3. Verify config.yaml exists
4. Test local Docker build:
   ```bash
   docker buildx build \
     --platform linux/amd64,linux/arm64 \
     -f Dockerfile.goreleaser \
     -t test:latest \
     .
   ```

### Tag Already Exists

If you need to recreate a tag:

```bash
# Delete local tag
git tag -d v1.0.0

# Delete remote tag
git push origin :refs/tags/v1.0.0

# Recreate and push
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

## CI/CD Pipeline

### GitHub Actions Workflow

```yaml
Trigger: Push tag (v*)
├── Checkout code
├── Setup Go
├── Setup QEMU (multi-arch emulation)
├── Setup Docker Buildx
├── Login to GitHub Container Registry
└── Run GoReleaser
    ├── Build binaries (all platforms)
    ├── Create archives
    ├── Generate checksums
    ├── Generate SBOMs
    ├── Build Docker images (amd64, arm64)
    ├── Push Docker images
    ├── Create GitHub release
    └── Upload artifacts
```

### Required Secrets

- `GITHUB_TOKEN`: Automatically provided by GitHub Actions

No additional secrets required! 🎉

## Version Management

### Semantic Versioning

JARVIS follows [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH

1.0.0 → Initial release
1.1.0 → New features (backward compatible)
1.1.1 → Bug fixes (backward compatible)
2.0.0 → Breaking changes
```

### Pre-releases

Pre-release versions are automatically detected:

```bash
v1.0.0-alpha.1   # Alpha release
v1.0.0-beta.1    # Beta release
v1.0.0-rc.1      # Release candidate
```

## Best Practices

1. **Test Before Release**:
   ```bash
   make ci  # Run all checks
   ```

2. **Use Conventional Commits**:
   - Enables automatic changelog generation
   - Groups changes by type

3. **Tag Format**:
   - Always prefix with `v`: `v1.0.0`
   - Use semantic versioning

4. **Release Notes**:
   - Review auto-generated notes
   - Add manual notes if needed

5. **Docker Images**:
   - Test images before tagging
   - Use specific versions in production

## Monitoring Releases

### GitHub Release Page

View all releases: https://github.com/dipjyotimetia/jarvis/releases

### Container Registry

View Docker images: https://github.com/dipjyotimetia/jarvis/pkgs/container/jarvis

### Download Statistics

GitHub provides download statistics for each release.

## Advanced Configuration

### Customizing the Release

Edit `.goreleaser.yaml` to:
- Add new platforms
- Change archive formats
- Modify changelog groups
- Add homebrew tap
- Enable signing (GPG)
- Add announcements (Discord, Slack)

### Example Customizations

```yaml
# Add signing
signs:
  - artifacts: checksum
    args: ["--batch", "--local-user", "{{ .Env.GPG_FINGERPRINT }}", ...]

# Add Homebrew tap
brews:
  - name: jarvis
    repository:
      owner: dipjyotimetia
      name: homebrew-tap

# Add Discord announcement
announce:
  discord:
    enabled: true
    message_template: 'JARVIS {{ .Tag }} is out!'
```

## Resources

- [GoReleaser Documentation](https://goreleaser.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [OCI Image Spec](https://github.com/opencontainers/image-spec)
- [SPDX](https://spdx.dev/)

## Support

For issues with the release process:
1. Check [GitHub Actions logs](https://github.com/dipjyotimetia/jarvis/actions)
2. Review [GoReleaser docs](https://goreleaser.com/)
3. Open an issue on GitHub

---

**Happy Releasing!** 🚀
