package main

import (
	"os"
	"reflect"
	"testing"
)

// Test service configuration loading (single and multi-service)
func TestLoadServicesSingleService(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected ServicesConfig
	}{
		{
			name: "default service",
			envVars: map[string]string{},
			expected: ServicesConfig{
				Services: []Service{
					{Name: "service", Port: "80"},
				},
				RoutingMode: RoutingModePath,
				Domain:      "",
			},
		},
		{
			name: "custom service",
			envVars: map[string]string{
				"SERVICE": "backend",
				"PORT":    "8080",
			},
			expected: ServicesConfig{
				Services: []Service{
					{Name: "backend", Port: "8080"},
				},
				RoutingMode: RoutingModePath,
				Domain:      "",
			},
		},
		{
			name: "service with default port",
			envVars: map[string]string{
				"SERVICE": "api",
			},
			expected: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "80"},
				},
				RoutingMode: RoutingModePath,
				Domain:      "",
			},
		},
		{
			name: "no service when NO_HTTP_SERVICE is set",
			envVars: map[string]string{
				"NO_HTTP_SERVICE": "true",
			},
			expected: ServicesConfig{
				Services:    []Service{},
				RoutingMode: RoutingModePath,
				Domain:      "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Load config
			config := loadServicesConfig()

			// Compare
			if !reflect.DeepEqual(config, tt.expected) {
				t.Errorf("loadServicesConfig() = %v, want %v", config, tt.expected)
			}
		})
	}
}

func TestLoadServicesMultipleServices(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected ServicesConfig
	}{
		{
			name: "two services",
			envVars: map[string]string{
				"SERVICE_1": "api",
				"PORT_1":    "8080",
				"SERVICE_2": "web",
				"PORT_2":    "3000",
			},
			expected: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
					{Name: "web", Port: "3000"},
				},
				RoutingMode: RoutingModePath,
				Domain:      "",
			},
		},
		{
			name: "three services with mixed ports",
			envVars: map[string]string{
				"SERVICE_1": "frontend",
				"PORT_1":    "3000",
				"SERVICE_2": "backend",
				"PORT_2":    "8080",
				"SERVICE_3": "database",
				"PORT_3":    "5432",
			},
			expected: ServicesConfig{
				Services: []Service{
					{Name: "frontend", Port: "3000"},
					{Name: "backend", Port: "8080"},
					{Name: "database", Port: "5432"},
				},
				RoutingMode: RoutingModePath,
				Domain:      "",
			},
		},
		{
			name: "services with gaps in numbering",
			envVars: map[string]string{
				"SERVICE_1": "api",
				"PORT_1":    "8080",
				"SERVICE_3": "web",
				"PORT_3":    "3000",
			},
			expected: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
				},
				RoutingMode: RoutingModePath,
				Domain:      "",
			},
		},
		{
			name: "fallback to single service when no numbered services",
			envVars: map[string]string{
				"SERVICE": "legacy",
				"PORT":    "9000",
			},
			expected: ServicesConfig{
				Services: []Service{
					{Name: "legacy", Port: "9000"},
				},
				RoutingMode: RoutingModePath,
				Domain:      "",
			},
		},
		{
			name: "numbered services override single service",
			envVars: map[string]string{
				"SERVICE":   "legacy",
				"PORT":      "9000",
				"SERVICE_1": "new-api",
				"PORT_1":    "8080",
			},
			expected: ServicesConfig{
				Services: []Service{
					{Name: "new-api", Port: "8080"},
				},
				RoutingMode: RoutingModePath,
				Domain:      "",
			},
		},
		{
			name: "service with default port",
			envVars: map[string]string{
				"SERVICE_1": "api",
				"SERVICE_2": "web",
				"PORT_2":    "3000",
			},
			expected: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "80"},
					{Name: "web", Port: "3000"},
				},
				RoutingMode: RoutingModePath,
				Domain:      "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Load config
			config := loadServicesConfig()

			// Compare
			if !reflect.DeepEqual(config, tt.expected) {
				t.Errorf("loadServicesConfig() = %v, want %v", config, tt.expected)
			}
		})
	}
}

// Test routing mode detection
func TestRoutingModeDetection(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected ServicesConfig
	}{
		{
			name: "default routing mode is path",
			envVars: map[string]string{
				"SERVICE": "api",
				"PORT":    "8080",
			},
			expected: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
				},
				RoutingMode: RoutingModePath,
				Domain:      "",
			},
		},
		{
			name: "explicit path routing mode",
			envVars: map[string]string{
				"SERVICE":      "api",
				"PORT":         "8080",
				"ROUTING_MODE": "path",
			},
			expected: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
				},
				RoutingMode: RoutingModePath,
				Domain:      "",
			},
		},
		{
			name: "subdomain routing mode with domain",
			envVars: map[string]string{
				"SERVICE_1":    "api",
				"PORT_1":       "8080",
				"SERVICE_2":    "web",
				"PORT_2":       "3000",
				"ROUTING_MODE": "subdomain",
				"DOMAIN":       "example.com",
			},
			expected: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
					{Name: "web", Port: "3000"},
				},
				RoutingMode: RoutingModeSubdomain,
				Domain:      "example.com",
			},
		},
		{
			name: "invalid routing mode defaults to path",
			envVars: map[string]string{
				"SERVICE":      "api",
				"PORT":         "8080",
				"ROUTING_MODE": "invalid",
			},
			expected: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
				},
				RoutingMode: RoutingModePath,
				Domain:      "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Load config
			config := loadServicesConfig()

			// Compare
			if !reflect.DeepEqual(config, tt.expected) {
				t.Errorf("loadServicesConfig() = %v, want %v", config, tt.expected)
			}
		})
	}
}

// Test certificate domain generation
func TestGetCertificateDomains(t *testing.T) {
	tests := []struct {
		name     string
		config   ServicesConfig
		expected []string
	}{
		{
			name: "path routing single service",
			config: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
				},
				RoutingMode: RoutingModePath,
				Domain:      "example.com",
			},
			expected: []string{"example.com"},
		},
		{
			name: "path routing multiple services",
			config: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
					{Name: "web", Port: "3000"},
				},
				RoutingMode: RoutingModePath,
				Domain:      "example.com",
			},
			expected: []string{"example.com"},
		},
		{
			name: "subdomain routing single service",
			config: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
				},
				RoutingMode: RoutingModeSubdomain,
				Domain:      "example.com",
			},
			expected: []string{"example.com", "api.example.com"},
		},
		{
			name: "subdomain routing multiple services",
			config: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
					{Name: "web", Port: "3000"},
					{Name: "admin", Port: "9000"},
				},
				RoutingMode: RoutingModeSubdomain,
				Domain:      "example.com",
			},
			expected: []string{"example.com", "api.example.com", "web.example.com", "admin.example.com"},
		},
		{
			name: "subdomain routing no services",
			config: ServicesConfig{
				Services:    []Service{},
				RoutingMode: RoutingModeSubdomain,
				Domain:      "example.com",
			},
			expected: []string{"example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCertificateDomains(tt.config)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("getCertificateDomains() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test subdomain validation
func TestValidateSubdomainConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      ServicesConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid subdomain config",
			config: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
				},
				RoutingMode: RoutingModeSubdomain,
				Domain:      "example.com",
			},
			expectError: false,
		},
		{
			name: "subdomain mode without domain",
			config: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
				},
				RoutingMode: RoutingModeSubdomain,
				Domain:      "",
			},
			expectError: true,
			errorMsg:    "DOMAIN environment variable is required for subdomain routing mode",
		},
		{
			name: "path mode without domain is ok",
			config: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
				},
				RoutingMode: RoutingModePath,
				Domain:      "",
			},
			expectError: false,
		},
		{
			name: "service name with invalid characters for subdomain",
			config: ServicesConfig{
				Services: []Service{
					{Name: "api_service", Port: "8080"},
				},
				RoutingMode: RoutingModeSubdomain,
				Domain:      "example.com",
			},
			expectError: true,
			errorMsg:    "service name 'api_service' contains invalid characters for subdomain (only lowercase letters, numbers, and hyphens allowed)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSubdomainConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("validateSubdomainConfig() expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("validateSubdomainConfig() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateSubdomainConfig() unexpected error: %v", err)
				}
			}
		})
	}
}

// Test port validation
func TestIsValidPort(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected bool
	}{
		{"valid port 80", "80", true},
		{"valid port 8080", "8080", true},
		{"valid port 65535", "65535", true},
		{"invalid port 0", "0", false},
		{"invalid port 65536", "65536", false},
		{"invalid port negative", "-1", false},
		{"invalid port non-numeric", "abc", false},
		{"invalid port empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPort(tt.port)
			if result != tt.expected {
				t.Errorf("isValidPort(%s) = %v, want %v", tt.port, result, tt.expected)
			}
		})
	}
}

// Test subdomain name validation
func TestIsValidSubdomain(t *testing.T) {
	tests := []struct {
		name      string
		subdomain string
		expected  bool
	}{
		{"empty string", "", false},
		{"starts with hyphen", "-api", false},
		{"ends with hyphen", "api-", false},
		{"contains uppercase", "API", false},
		{"contains underscore", "api_service", false},
		{"contains space", "api service", false},
		{"contains dot", "api.service", false},
		{"valid with hyphens", "api-service", true},
		{"valid single char", "a", true},
		{"valid numbers", "api1", true},
		{"valid all numbers", "123", true},
		{"valid mixed", "api-service-1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidSubdomain(tt.subdomain)
			if result != tt.expected {
				t.Errorf("isValidSubdomain(%s) = %v, want %v", tt.subdomain, result, tt.expected)
			}
		})
	}
}