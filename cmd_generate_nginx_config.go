//go:build config_gen
// +build config_gen

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

type Service struct {
	Name string
	Port string
}

type NginxConfig struct {
	Domain          string
	Services        []Service
	EnableWebsocket bool
	RoutingMode     string
}

func generateNginxConfig() error {
	config := NginxConfig{
		Domain:          os.Getenv("DOMAIN"),
		EnableWebsocket: os.Getenv("ENABLE_WEBSOCKET") == "true",
		RoutingMode:     getEnvWithDefault("ROUTING_MODE", "path"),
		Services:        []Service{},
	}

	if config.Domain == "" {
		return fmt.Errorf("DOMAIN environment variable is not set")
	}

	// Check for multi-service configuration (SERVICE_1, SERVICE_2, etc.)
	serviceFound := false
	for i := 1; i <= 100; i++ { // Support up to 100 services
		serviceName := os.Getenv(fmt.Sprintf("SERVICE_%d", i))
		servicePort := os.Getenv(fmt.Sprintf("PORT_%d", i))

		if serviceName == "" {
			if i == 1 {
				// No SERVICE_1, check for single service mode
				break
			}
			// End of sequential services
			continue
		}

		if servicePort == "" {
			servicePort = "80" // Default port
		}

		config.Services = append(config.Services, Service{
			Name: serviceName,
			Port: servicePort,
		})
		serviceFound = true
	}

	// If no multi-service config found, check for single service mode
	if !serviceFound {
		serviceName := getEnvWithDefault("SERVICE", "")
		servicePort := getEnvWithDefault("PORT", "80")

		if serviceName != "" {
			config.Services = append(config.Services, Service{
				Name: serviceName,
				Port: servicePort,
			})
		}
	}

	// Validate subdomain mode
	if config.RoutingMode == "subdomain" && len(config.Services) > 0 {
		for _, service := range config.Services {
			if strings.Contains(service.Name, ".") || strings.Contains(service.Name, "_") {
				return fmt.Errorf("service name '%s' contains invalid characters for subdomain routing (no dots or underscores allowed)", service.Name)
			}
		}

		// Log the required certificate domains
		log.Printf("Subdomain routing enabled. Certificate must include:")
		log.Printf("  - Primary: %s", config.Domain)
		for _, service := range config.Services {
			log.Printf("  - Subdomain: %s.%s", service.Name, config.Domain)
		}
	}

	// Parse and execute template
	tmplContent, err := os.ReadFile("/etc/nginx/templates/reverse-proxy.conf.tmpl")
	if err != nil {
		// Try local path for development
		tmplContent, err = os.ReadFile("nginx/templates/reverse-proxy.conf.tmpl")
		if err != nil {
			return fmt.Errorf("failed to read template file: %w", err)
		}
	}

	tmpl, err := template.New("nginx").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Output to stdout by default, or to file if OUTPUT_FILE is set
	outputFile := os.Getenv("OUTPUT_FILE")
	var output *os.File

	if outputFile != "" {
		output, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer output.Close()
	} else {
		output = os.Stdout
	}

	if err := tmpl.Execute(output, config); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Log configuration summary
	if len(config.Services) == 0 {
		log.Println("No services configured")
	} else if len(config.Services) == 1 {
		log.Printf("Single service mode: %s:%s", config.Services[0].Name, config.Services[0].Port)
	} else {
		log.Printf("Multi-service mode (%s routing): %d services configured", config.RoutingMode, len(config.Services))
		for _, service := range config.Services {
			log.Printf("  - %s:%s", service.Name, service.Port)
		}
	}

	return nil
}

// Build this as a standalone binary with: go build -o generate_nginx_config config_generator.go
func main() {
	if err := generateNginxConfig(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
