package main

import (
	"fmt"
	"os"
	"strconv"
)

type Service struct {
	Name string
	Port string
}

type RoutingMode string

const (
	RoutingModePath      RoutingMode = "path"
	RoutingModeSubdomain RoutingMode = "subdomain"
)

type ServicesConfig struct {
	Services     []Service
	RoutingMode  RoutingMode
	Domain       string // Base domain for subdomain routing
}

func loadServicesConfig() ServicesConfig {
	config := ServicesConfig{
		Services:    []Service{},
		RoutingMode: RoutingModePath, // Default to path routing
		Domain:      os.Getenv("DOMAIN"),
	}

	// Check routing mode
	routingMode := os.Getenv("ROUTING_MODE")
	if routingMode == "subdomain" {
		config.RoutingMode = RoutingModeSubdomain
	} else if routingMode != "" && routingMode != "path" {
		// Invalid routing mode, default to path
		fmt.Fprintf(os.Stderr, "Warning: Invalid ROUTING_MODE '%s', defaulting to 'path'\n", routingMode)
	}

	// Check if NO_HTTP_SERVICE is set
	if os.Getenv("NO_HTTP_SERVICE") == "true" {
		return config
	}

	// First, check for numbered services (SERVICE_1, SERVICE_2, etc.)
	hasNumberedServices := false
	maxServiceNum := 0

	// Find the maximum service number
	for i := 1; i <= 100; i++ { // reasonable upper limit
		if os.Getenv(fmt.Sprintf("SERVICE_%d", i)) != "" {
			hasNumberedServices = true
			maxServiceNum = i
		}
	}

	if hasNumberedServices {
		// Load numbered services, stop at first gap
		for i := 1; i <= maxServiceNum; i++ {
			serviceName := os.Getenv(fmt.Sprintf("SERVICE_%d", i))
			if serviceName == "" {
				// Stop at first gap in numbering
				break
			}

			port := os.Getenv(fmt.Sprintf("PORT_%d", i))
			if port == "" {
				port = "80"
			}

			config.Services = append(config.Services, Service{
				Name: serviceName,
				Port: port,
			})
		}
	} else {
		// Fall back to single service configuration (backward compatibility)
		service := os.Getenv("SERVICE")
		if service == "" {
			service = "service"
		}

		port := os.Getenv("PORT")
		if port == "" {
			port = "80"
		}

		config.Services = append(config.Services, Service{
			Name: service,
			Port: port,
		})
	}

	return config
}

// Helper function to validate port
func isValidPort(port string) bool {
	p, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return p > 0 && p <= 65535
}

// getCertificateDomains returns the list of domains that need certificates
func getCertificateDomains(config ServicesConfig) []string {
	domains := []string{}

	if config.Domain != "" {
		domains = append(domains, config.Domain)
	}

	if config.RoutingMode == RoutingModeSubdomain && config.Domain != "" {
		// Add subdomain for each service
		for _, service := range config.Services {
			subdomain := fmt.Sprintf("%s.%s", service.Name, config.Domain)
			domains = append(domains, subdomain)
		}
	}

	return domains
}

// validateSubdomainConfig validates configuration for subdomain routing
func validateSubdomainConfig(config ServicesConfig) error {
	if config.RoutingMode == RoutingModeSubdomain {
		if config.Domain == "" {
			return fmt.Errorf("DOMAIN environment variable is required for subdomain routing mode")
		}

		// Validate service names are valid for subdomains
		for _, service := range config.Services {
			if !isValidSubdomain(service.Name) {
				return fmt.Errorf("service name '%s' contains invalid characters for subdomain (only lowercase letters, numbers, and hyphens allowed)", service.Name)
			}
		}
	}
	return nil
}

// isValidSubdomain checks if a string is valid as a subdomain
func isValidSubdomain(name string) bool {
	// Subdomain rules: lowercase letters, numbers, hyphens
	// Cannot start or end with hyphen
	if len(name) == 0 {
		return false
	}
	if name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-') {
			return false
		}
	}
	return true
}

// generateNginxConfig creates nginx configuration from services config
func generateNginxConfig(config ServicesConfig) string {
	if len(config.Services) == 0 {
		return ""
	}

	if config.RoutingMode == RoutingModeSubdomain {
		return generateSubdomainConfig(config)
	}

	return generatePathConfig(config)
}

// generatePathConfig generates path-based routing configuration
func generatePathConfig(config ServicesConfig) string {
	if len(config.Services) == 0 {
		return ""
	}

	result := ""

	// Generate upstream blocks for each service
	for _, service := range config.Services {
		result += fmt.Sprintf(`upstream %s_upstream {
    server %s:%s max_fails=3 fail_timeout=30s;
}

`, service.Name, service.Name, service.Port)
	}

	// If only one service, generate simple proxy config
	if len(config.Services) == 1 {
		service := config.Services[0]
		result += fmt.Sprintf(`server {
    listen 443 ssl;
    http2 on;
    include /etc/nginx/conf.d/ssl.conf;
    resolver 127.0.0.11 valid=30s;

    server_name _;

    location / {
      proxy_pass http://%s_upstream;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    }

    server_tokens off;
}
`, service.Name)
		return result
	}

	// Multiple services: use path-based routing
	result += `server {
    listen 443 ssl;
    http2 on;
    include /etc/nginx/conf.d/ssl.conf;
    resolver 127.0.0.11 valid=30s;

    server_name _;

`

	// Add location blocks for each service
	for _, service := range config.Services {
		result += fmt.Sprintf(`    location /%s/ {
      proxy_pass http://%s_upstream/;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    }

`, service.Name, service.Name)
	}

	// Add default location (routes to first service)
	firstService := config.Services[0]
	result += fmt.Sprintf(`    location / {
      proxy_pass http://%s_upstream;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    }

    server_tokens off;
}
`, firstService.Name)

	return result
}

// generateSubdomainConfig generates subdomain-based routing configuration
func generateSubdomainConfig(config ServicesConfig) string {
	if config.Domain == "" || len(config.Services) == 0 {
		return ""
	}

	result := ""

	// Generate upstream blocks for each service
	for _, service := range config.Services {
		result += fmt.Sprintf(`upstream %s_upstream {
    server %s:%s max_fails=3 fail_timeout=30s;
}

`, service.Name, service.Name, service.Port)
	}

	// Create server block for each service subdomain
	for _, service := range config.Services {
		subdomain := fmt.Sprintf("%s.%s", service.Name, config.Domain)
		result += fmt.Sprintf(`server {
    listen 443 ssl;
    http2 on;
    include /etc/nginx/conf.d/ssl.conf;
    resolver 127.0.0.11 valid=30s;

    server_name %s;

    location / {
      proxy_pass http://%s_upstream;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    }

    server_tokens off;
}

`, subdomain, service.Name)
	}

	// Add main domain server block (routes to first service)
	firstService := config.Services[0]
	result += fmt.Sprintf(`server {
    listen 443 ssl;
    http2 on;
    include /etc/nginx/conf.d/ssl.conf;
    resolver 127.0.0.11 valid=30s;

    server_name %s;

    location / {
      proxy_pass http://%s_upstream;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    }

    server_tokens off;
}
`, config.Domain, firstService.Name)

	return result
}