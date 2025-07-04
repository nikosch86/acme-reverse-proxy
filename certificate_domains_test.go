package main

import (
	"reflect"
	"testing"
)

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