# nginx Reverse Proxy with Automatic ACME Certificate Management

[![CI/CD](https://github.com/nikosch86/acme-reverse-proxy/actions/workflows/ci.yml/badge.svg)](https://github.com/nikosch86/acme-reverse-proxy/actions/workflows/ci.yml)
[![Security Scan](https://github.com/nikosch86/acme-reverse-proxy/actions/workflows/security.yml/badge.svg)](https://github.com/nikosch86/acme-reverse-proxy/actions/workflows/security.yml)

This Docker image provides an nginx reverse proxy with automatic SSL certificate management using the ACME protocol (Let's Encrypt or any ACME-compliant CA). It combines a Go-based ACME client with nginx to automatically obtain, validate, and renew SSL certificates.

## Installation

### Pre-built Binaries

Download the latest release for your platform from [GitHub Releases](https://github.com/nikosch86/acme-reverse-proxy/releases):

**Linux:**
```bash
# AMD64
wget https://github.com/nikosch86/acme-reverse-proxy/releases/latest/download/acme-linux-amd64
chmod +x acme-linux-amd64
sudo mv acme-linux-amd64 /usr/local/bin/acme

# ARM64
wget https://github.com/nikosch86/acme-reverse-proxy/releases/latest/download/acme-linux-arm64
chmod +x acme-linux-arm64
sudo mv acme-linux-arm64 /usr/local/bin/acme
```

**macOS:**
```bash
# Intel
wget https://github.com/nikosch86/acme-reverse-proxy/releases/latest/download/acme-darwin-amd64
chmod +x acme-darwin-amd64
sudo mv acme-darwin-amd64 /usr/local/bin/acme

# Apple Silicon
wget https://github.com/nikosch86/acme-reverse-proxy/releases/latest/download/acme-darwin-arm64
chmod +x acme-darwin-arm64
sudo mv acme-darwin-arm64 /usr/local/bin/acme
```

**Windows:**
```powershell
# Download acme-windows-amd64.exe from releases page
# Add to PATH or place in desired location
```

### Container Images

**GitHub Container Registry:**
```bash
docker pull ghcr.io/nikosch86/acme-reverse-proxy:latest
```

**Docker Hub:**
```bash
docker pull nikosch86/acme-reverse-proxy:latest
```

**Available Tags:**
- `latest` - Latest stable release from main branch
- `main` - Latest main branch build
- `develop` - Latest develop branch build  
- `v1.2.3` - Specific version releases

### Build from Source

```bash
git clone https://github.com/nikosch86/acme-reverse-proxy.git
cd acme-reverse-proxy
go build -o acme .
```

## Quick Start

### Using Standalone Binary

Once you've installed the `acme` binary, you can run it directly:

```bash
# Set required environment variables
export DOMAIN=example.com
export EMAIL=admin@example.com
export CA_DIR_URL=https://acme-v02.api.letsencrypt.org/directory

# Run the ACME client
acme
```

**Note:** The standalone binary is designed to work with nginx. Ensure nginx is installed and properly configured, or use the Docker image for a complete solution.

### Using Pre-built Images

```yaml
services:
  reverse-proxy:
    image: ghcr.io/nikosch86/acme-reverse-proxy:latest
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/sites:/etc/nginx/sites:ro
    environment:
      DOMAIN: example.com
      EMAIL: admin@example.com
      SERVICE: backend-service
      PORT: 3000
      CA_DIR_URL: https://acme-v02.api.letsencrypt.org/directory
```

### Building Locally

```yaml
services:
  reverse-proxy:
    build: .
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/sites:/etc/nginx/sites:ro
    environment:
      DOMAIN: example.com
      EMAIL: admin@example.com
      SERVICE: backend-service
      PORT: 3000
      CA_DIR_URL: https://acme-v02.api.letsencrypt.org/directory
```

## Configuration Options

### Environment Variables

**Required:**
- `DOMAIN` - Primary domain name for certificate generation (mandatory)

**Optional:**
- `SERVICE` - Backend service name for reverse proxy, defaults to `service`
- `PORT` - Backend service port, defaults to `80`
- `EMAIL` - Email for ACME account registration, defaults to `admin@dev.lan`
- `SAN` - Subject Alternative Names (comma-separated), defaults to empty
- `EXPIRY_DAYS_THRESHOLD` - Certificate renewal threshold in days, defaults to `30`
- `RENEWAL_SECONDS` - Certificate check interval in seconds, defaults to `86400` (24 hours)
- `CERT_PATH` - Certificate file path, defaults to `/etc/ssl/private/fullchain.pem`
- `KEY_PATH` - Private key file path, defaults to `/etc/ssl/private/key.pem`
- `CA_DIR_URL` - ACME CA directory URL, defaults to Let's Encrypt staging:
  - Staging: `https://acme-staging-v02.api.letsencrypt.org/directory`
  - Production: `https://acme-v02.api.letsencrypt.org/directory`
- `NO_HTTP_SERVICE` - Set to any value to disable the default reverse proxy configuration

## Configuration Methods

### Method 1: Simple Environment Variables (Default)

For basic use cases, simply set `DOMAIN`, `SERVICE`, and `PORT`. The system will:
1. Generate certificates for the specified domain
2. Configure nginx to proxy requests to `http://SERVICE:PORT`
3. Use the default reverse proxy template with SSL enabled

### Method 2: Custom Site Configurations

For advanced configurations, mount custom nginx site configs to `/etc/nginx/sites/`:

```yaml
volumes:
  - ./nginx/sites:/etc/nginx/sites:ro
environment:
  NO_HTTP_SERVICE: true  # Disable default config
```

Custom site configurations should:
- Listen on port 443 with SSL enabled
- Include the SSL configuration: `include /etc/nginx/conf.d/ssl.conf;`
- Configure appropriate proxy headers

Example custom site config:
```nginx
server {
    listen 443 ssl;
    http2 on;
    include /etc/nginx/conf.d/ssl.conf;
    
    server_name your-domain.com;
    
    location / {
        proxy_pass http://backend:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    server_tokens off;
}
```

### Method 3: Complete Custom Configuration

Mount a custom configuration to `/etc/nginx/conf.d/reverse-proxy.conf` and ensure it includes:
```nginx
include /etc/nginx/conf.d/ssl.conf;
```

## How It Works

### Certificate Management
1. **Startup**: The ACME client checks for existing certificates
2. **Validation**: Verifies certificates cover all required domains (primary + SAN)
3. **Renewal**: Automatically renews certificates approaching expiry threshold
4. **Challenge**: Uses HTTP-01 challenge method via `/.well-known/acme-challenge/`
5. **Reload**: Automatically reloads nginx after certificate updates

### Architecture
- **Go ACME Client**: Handles certificate lifecycle using the lego library
- **nginx**: Serves as reverse proxy with modular configuration
- **Docker**: Multi-stage build combining both components
- **Automation**: Background processes for certificate monitoring and renewal

### Security Features
- **TLS 1.3 Only**: Modern encryption standards
- **HSTS Headers**: HTTP Strict Transport Security enabled
- **Secure Ciphers**: Strong cipher suite configuration
- **HTTP/2**: Enabled by default for better performance

## File Structure

```
/etc/nginx/
├── nginx.conf          # Main nginx configuration
├── http.conf           # HTTP-specific settings and includes
├── conf.d/
│   ├── ssl.conf        # SSL/TLS security configuration
│   ├── challenge.conf  # ACME challenge handling
│   ├── default.conf    # Default server configuration
│   └── reverse-proxy.conf # Default reverse proxy (template-based)
└── sites/              # Custom site configurations
    └── *.conf          # Individual site configs

/usr/share/nginx/challenge/  # ACME challenge directory
/etc/ssl/private/           # Certificate storage
```

## Development

### Prerequisites
- Go 1.24+
- Docker & Docker Compose

### Build and Test

**Manual commands:**
```bash
go test -v ./...                     # Run all tests
go test -short -v ./...              # Run tests (skip slow ones)
go build -o acme .                   # Build binary
docker build -t acme-reverse-proxy . # Build image
docker-compose up                    # Run complete stack
```

### Testing

The project includes test coverage:

- **Unit Tests**: Component testing for core functionality
- **Security Tests**: Automated vulnerability scanning via GitHub Actions

Test the application:
```bash
go test -v ./...             # Run all tests
go test -short -v ./...      # Run tests (skip slow ones)
```

## CI/CD & Automation

### GitHub Actions Workflows

- **CI/CD Pipeline**: Automated testing, building, and publishing
  - Tests run on every branch push
  - Docker images published to GitHub Container Registry and Docker Hub
  - Multi-architecture builds (linux/amd64, linux/arm64)
  - Comprehensive caching for faster builds

- **Security Scanning**: Automated vulnerability detection
  - Go dependency scanning with `govulncheck`
  - Filesystem security scanning with Trivy
  - Results uploaded to GitHub Security tab

- **Dependabot**: Automated dependency updates
  - Weekly updates for Go modules, Docker images, and GitHub Actions
  - Automatic PR creation with changelogs

### Automated Publishing

**Docker Images** are automatically published on:
- **Pushes to main/develop**: Branch-tagged images
- **Version tags**: Semantic versioned releases (v1.2.3 → 1.2.3, 1.2, 1)

**GitHub Releases** are automatically created on:
- **Version tags**: Creates releases with:
  - Pre-built Go binaries for Linux, macOS, Windows (multiple architectures)
  - SHA256 checksums for verification
  - Release notes with Docker image information
  - Links to documentation

### Development Workflow

1. Create feature branch from `develop`
2. Make changes and add tests
3. Run `go test -v ./...` to ensure tests pass
4. Create PR to `develop` branch
5. CI automatically tests and validates
6. Merge to `develop` for testing builds
7. Merge to `main` for stable releases
8. Create version tags for releases

## Certificate Monitoring

- Certificates are checked every `RENEWAL_SECONDS` (default: 24 hours)
- Renewal occurs when expiry is within `EXPIRY_DAYS_THRESHOLD` days (default: 30)
- Comprehensive logging provides detailed certificate status information
- Automatic nginx reload after certificate updates

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and add tests
4. Run the test suite: `go test -v ./...`
5. Ensure tests pass and code builds
6. Commit with clear messages
7. Push and create a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

