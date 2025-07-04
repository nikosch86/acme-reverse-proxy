package main

import (
	"os"
	"reflect"
	"testing"
)

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