# nginx based reverse proxy with automatic ACME certificate generation using any CA supporting ACME protocol

This Docker image provides an nginx reverse proxy with automatic SSL certificate management using the ACME protocol (Let's Encrypt or any ACME-compliant CA). It combines a Go-based ACME client with nginx to automatically obtain, validate, and renew SSL certificates.

## Quick Start

See the included `docker-compose.yml` for a complete example:

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

**Build and test:**
```bash
go test                              # Run tests
go build -o acme .                   # Build binary
docker build -t reverse-proxy-acme . # Build image
docker-compose up                    # Run complete stack
```

**Certificate monitoring:**
- Certificates are checked every `RENEWAL_SECONDS`
- Renewal occurs when expiry is within `EXPIRY_DAYS_THRESHOLD` days
- Logs provide detailed information about certificate status

