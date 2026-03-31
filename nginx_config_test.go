package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

// Service represents a service configuration for nginx
type ServiceTest struct {
	Name string
	Port string
}

// NginxConfig represents the nginx configuration
type NginxConfigTest struct {
	Domain          string
	Services        []ServiceTest
	EnableWebsocket bool
	RoutingMode     string
}

// This duplicates the getEnvWithDefault function for testing purposes
func getEnvWithDefaultTest(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// This creates a testable version of generateNginxConfig
func generateNginxConfigTest(templateContent string, outputFile string) error {
	config := NginxConfigTest{
		Domain:          os.Getenv("DOMAIN"),
		EnableWebsocket: os.Getenv("ENABLE_WEBSOCKET") == "true",
		RoutingMode:     getEnvWithDefaultTest("ROUTING_MODE", "path"),
		Services:        []ServiceTest{},
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

		config.Services = append(config.Services, ServiceTest{
			Name: serviceName,
			Port: servicePort,
		})
		serviceFound = true
	}

	// If no multi-service config found, check for single service mode
	if !serviceFound {
		serviceName := getEnvWithDefaultTest("SERVICE", "")
		servicePort := getEnvWithDefaultTest("PORT", "80")

		if serviceName != "" {
			config.Services = append(config.Services, ServiceTest{
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
	}

	// Parse and execute template
	tmpl, err := template.New("nginx").Parse(templateContent)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Output to file
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

	return nil
}

func TestGetEnvWithDefaultNginxConfig(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{"env var set", "TEST_KEY", "default", "custom", "custom"},
		{"env var empty", "TEST_KEY_EMPTY", "default", "", "default"},
		{"env var unset", "TEST_KEY_UNSET", "default", "", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up env before test
			os.Unsetenv(tt.key)
			
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvWithDefaultTest(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvWithDefault(%q, %q) = %q; want %q", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGenerateNginxConfigFunction(t *testing.T) {
	// Simple test template
	templateContent := `{{- if .Services -}}
# Services configured: {{len .Services}}
{{- range .Services }}
# Service: {{.Name}}:{{.Port}}
{{- end }}
{{- if eq .RoutingMode "subdomain" }}
# Subdomain routing enabled
{{- else }}
# Path routing enabled  
{{- end }}
{{- if .EnableWebsocket }}
# WebSocket support enabled
{{- end }}
{{- else -}}
# No services configured
{{- end -}}
`

	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		contains    []string
		notContains []string
	}{
		{
			name: "no domain set",
			envVars: map[string]string{
				"DOMAIN": "",
			},
			expectError: true,
		},
		{
			name: "single service mode",
			envVars: map[string]string{
				"DOMAIN":  "example.com",
				"SERVICE": "app",
				"PORT":    "3000",
			},
			expectError: false,
			contains:    []string{"# Services configured: 1", "# Service: app:3000", "# Path routing enabled"},
			notContains: []string{"# Subdomain routing enabled", "# WebSocket support enabled", "# No services configured"},
		},
		{
			name: "multi-service path routing",
			envVars: map[string]string{
				"DOMAIN":       "example.com",
				"SERVICE_1":    "frontend",
				"PORT_1":       "80",
				"SERVICE_2":    "backend",
				"PORT_2":       "8080",
				"ROUTING_MODE": "path",
			},
			expectError: false,
			contains:    []string{"# Services configured: 2", "# Service: frontend:80", "# Service: backend:8080", "# Path routing enabled"},
			notContains: []string{"# Subdomain routing enabled", "# WebSocket support enabled", "# No services configured"},
		},
		{
			name: "multi-service subdomain routing",
			envVars: map[string]string{
				"DOMAIN":       "example.com",
				"SERVICE_1":    "api",
				"PORT_1":       "8080",
				"SERVICE_2":    "web",
				"PORT_2":       "80",
				"ROUTING_MODE": "subdomain",
			},
			expectError: false,
			contains:    []string{"# Services configured: 2", "# Service: api:8080", "# Service: web:80", "# Subdomain routing enabled"},
			notContains: []string{"# Path routing enabled", "# WebSocket support enabled", "# No services configured"},
		},
		{
			name: "subdomain routing with websocket",
			envVars: map[string]string{
				"DOMAIN":          "example.com",
				"SERVICE_1":       "chat",
				"PORT_1":          "3000",
				"ROUTING_MODE":    "subdomain",
				"ENABLE_WEBSOCKET": "true",
			},
			expectError: false,
			contains:    []string{"# Services configured: 1", "# Service: chat:3000", "# Subdomain routing enabled", "# WebSocket support enabled"},
			notContains: []string{"# Path routing enabled", "# No services configured"},
		},
		{
			name: "no services configured",
			envVars: map[string]string{
				"DOMAIN": "example.com",
			},
			expectError: false,
			contains:    []string{"# No services configured"},
			notContains: []string{"# Services configured:", "# Service:", "# Subdomain routing enabled", "# Path routing enabled"},
		},
		{
			name: "service with default port",
			envVars: map[string]string{
				"DOMAIN":    "example.com",
				"SERVICE_1": "webapp",
				// PORT_1 not set, should default to "80"
			},
			expectError: false,
			contains:    []string{"# Services configured: 1", "# Service: webapp:80"},
		},
		{
			name: "invalid service name for subdomain routing",
			envVars: map[string]string{
				"DOMAIN":       "example.com",
				"SERVICE_1":    "invalid.service",
				"ROUTING_MODE": "subdomain",
			},
			expectError: true,
		},
		{
			name: "invalid service name with underscore for subdomain routing",
			envVars: map[string]string{
				"DOMAIN":       "example.com",
				"SERVICE_1":    "invalid_service",
				"ROUTING_MODE": "subdomain",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			for _, key := range []string{"DOMAIN", "SERVICE", "PORT", "SERVICE_1", "PORT_1", "SERVICE_2", "PORT_2", "ROUTING_MODE", "ENABLE_WEBSOCKET"} {
				os.Unsetenv(key)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Create a temporary output file
			tempDir, err := os.MkdirTemp("", "nginx-config-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			outputFile := filepath.Join(tempDir, "output.conf")

			err = generateNginxConfigTest(templateContent, outputFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Read the output file
			output, err := os.ReadFile(outputFile)
			if err != nil {
				t.Errorf("Failed to read output file: %v", err)
				return
			}

			outputStr := string(output)

			// Check expected content
			for _, expected := range tt.contains {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain %q, but it didn't. Output:\n%s", expected, outputStr)
				}
			}

			// Check content that should not be present
			for _, notExpected := range tt.notContains {
				if strings.Contains(outputStr, notExpected) {
					t.Errorf("Expected output to NOT contain %q, but it did. Output:\n%s", notExpected, outputStr)
				}
			}

			// Clean up environment after test
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestNginxConfigServiceParsing(t *testing.T) {
	// Test the service parsing logic more thoroughly
	templateContent := `Services: {{len .Services}}`
	
	tests := []struct {
		name         string
		envVars      map[string]string
		wantServices int
	}{
		{
			name: "services 1-3 with gap",
			envVars: map[string]string{
				"DOMAIN":    "example.com",
				"SERVICE_1": "service1",
				"PORT_1":    "8001",
				"SERVICE_3": "service3",
				"PORT_3":    "8003",
			},
			wantServices: 2, // Should get both service1 and service3 (continues through gap)
		},
		{
			name: "services 1-5 sequential",
			envVars: map[string]string{
				"DOMAIN":    "example.com",
				"SERVICE_1": "service1",
				"PORT_1":    "8001",
				"SERVICE_2": "service2",
				"PORT_2":    "8002",
				"SERVICE_3": "service3",
				"PORT_3":    "8003",
				"SERVICE_4": "service4",
				"PORT_4":    "8004",
				"SERVICE_5": "service5",
				"PORT_5":    "8005",
			},
			wantServices: 5,
		},
		{
			name: "fallback to single service mode",
			envVars: map[string]string{
				"DOMAIN":  "example.com",
				"SERVICE": "singleservice",
				"PORT":    "9000",
			},
			wantServices: 1,
		},
		{
			name: "no services",
			envVars: map[string]string{
				"DOMAIN": "example.com",
			},
			wantServices: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			for i := 1; i <= 10; i++ {
				os.Unsetenv(fmt.Sprintf("SERVICE_%d", i))
				os.Unsetenv(fmt.Sprintf("PORT_%d", i))
			}
			os.Unsetenv("SERVICE")
			os.Unsetenv("PORT")
			os.Unsetenv("DOMAIN")

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			tempDir, err := os.MkdirTemp("", "nginx-service-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			outputFile := filepath.Join(tempDir, "output.conf")

			err = generateNginxConfigTest(templateContent, outputFile)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Read the output
			output, err := os.ReadFile(outputFile)
			if err != nil {
				t.Errorf("Failed to read output file: %v", err)
				return
			}

			expectedOutput := fmt.Sprintf("Services: %d", tt.wantServices)
			if strings.TrimSpace(string(output)) != expectedOutput {
				t.Errorf("Expected %q, got %q", expectedOutput, strings.TrimSpace(string(output)))
			}

			// Clean up
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestNginxConfigTemplateExecution(t *testing.T) {
	// Test template execution with actual nginx template patterns
	realTemplateContent := `{{- if .Services -}}
{{- if eq .RoutingMode "subdomain" -}}
{{- range .Services }}
server {
    listen 443 ssl;
    server_name {{.Name}}.{{$.Domain}};
    location / {
        proxy_pass http://{{.Name}}:{{.Port}};
        {{- if $.EnableWebsocket }}
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        {{- end }}
    }
}
{{end -}}
{{- else -}}
server {
    listen 443 ssl;
    server_name {{.Domain}};
    {{- range .Services }}
    location /{{.Name}}/ {
        proxy_pass http://{{.Name}}:{{.Port}}/;
    }
    {{- end }}
}
{{- end -}}
{{- else -}}
server {
    listen 443 ssl;
    server_name _;
    location / {
        return 503 "No backend service configured";
    }
}
{{- end -}}`

	tests := []struct {
		name     string
		envVars  map[string]string
		contains []string
	}{
		{
			name: "subdomain routing",
			envVars: map[string]string{
				"DOMAIN":       "example.com",
				"SERVICE_1":    "api",
				"PORT_1":       "8080",
				"ROUTING_MODE": "subdomain",
			},
			contains: []string{
				"server_name api.example.com",
				"proxy_pass http://api:8080",
			},
		},
		{
			name: "path routing",
			envVars: map[string]string{
				"DOMAIN":    "example.com",
				"SERVICE_1": "api",
				"PORT_1":    "8080",
			},
			contains: []string{
				"server_name example.com",
				"location /api/",
				"proxy_pass http://api:8080/",
			},
		},
		{
			name: "websocket enabled",
			envVars: map[string]string{
				"DOMAIN":          "example.com",
				"SERVICE_1":       "chat",
				"PORT_1":          "3000",
				"ROUTING_MODE":    "subdomain",
				"ENABLE_WEBSOCKET": "true",
			},
			contains: []string{
				"proxy_http_version 1.1",
				"proxy_set_header Upgrade $http_upgrade",
			},
		},
		{
			name: "no services",
			envVars: map[string]string{
				"DOMAIN": "example.com",
			},
			contains: []string{
				"return 503 \"No backend service configured\"",
				"server_name _",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			for _, key := range []string{"DOMAIN", "SERVICE_1", "PORT_1", "ROUTING_MODE", "ENABLE_WEBSOCKET"} {
				os.Unsetenv(key)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			tempDir, err := os.MkdirTemp("", "nginx-template-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			outputFile := filepath.Join(tempDir, "output.conf")

			err = generateNginxConfigTest(realTemplateContent, outputFile)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			output, err := os.ReadFile(outputFile)
			if err != nil {
				t.Errorf("Failed to read output file: %v", err)
				return
			}

			outputStr := string(output)

			for _, expected := range tt.contains {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain %q, but it didn't. Output:\n%s", expected, outputStr)
				}
			}

			// Clean up
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}