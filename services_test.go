package main

import (
	"os"
	"reflect"
	"testing"
)

func TestLoadServicesSingleService(t *testing.T) {
	// Test current single-service behavior
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
	// Test multiple services configuration
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