package main

import (
	"fmt"
	"os"
)

func main() {
	// Load services configuration
	config := loadServicesConfig()

	// Validate configuration for subdomain mode
	if err := validateSubdomainConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Generate nginx configuration
	nginxConfig := generateNginxConfig(config)

	if nginxConfig == "" {
		// No services configured, don't write anything
		fmt.Println("No services configured")
		os.Exit(0)
	}

	// Write to nginx config file
	configPath := "/etc/nginx/conf.d/reverse-proxy.conf"
	if envPath := os.Getenv("NGINX_CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	err := os.WriteFile(configPath, []byte(nginxConfig), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing nginx config: %v\n", err)
		os.Exit(1)
	}

	// Print configuration summary
	routingInfo := "path-based"
	if config.RoutingMode == RoutingModeSubdomain {
		routingInfo = "subdomain-based"
	}
	fmt.Printf("Generated %s nginx config for %d service(s)\n", routingInfo, len(config.Services))
	
	// For subdomain mode, list the domains that need certificates
	if config.RoutingMode == RoutingModeSubdomain && config.Domain != "" {
		domains := getCertificateDomains(config)
		fmt.Println("Certificate domains required:")
		for _, domain := range domains {
			fmt.Printf("  - %s\n", domain)
		}
	}
}