package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/certificate"
)

// Test ACME configuration loading
func TestLoadConfig(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("EMAIL", "test@example.com")
	os.Setenv("DOMAIN", "example.com")
	os.Setenv("SAN", "www.example.com,api.example.com")
	os.Setenv("CERT_PATH", "/test/cert.pem")
	os.Setenv("KEY_PATH", "/test/key.pem")
	os.Setenv("CA_DIR_URL", "https://test-ca.com/directory")
	os.Setenv("EXPIRY_DAYS_THRESHOLD", "15")

	cfg := loadConfig()

	if cfg.Email != "test@example.com" {
		t.Errorf("Expected Email to be 'test@example.com', got '%s'", cfg.Email)
	}
	if cfg.Domain != "example.com" {
		t.Errorf("Expected Domain to be 'example.com', got '%s'", cfg.Domain)
	}
	if len(cfg.SAN) != 2 || cfg.SAN[0] != "www.example.com" || cfg.SAN[1] != "api.example.com" {
		t.Errorf("Unexpected SAN value: %v", cfg.SAN)
	}
	if cfg.CertPath != "/test/cert.pem" {
		t.Errorf("Expected CertPath to be '/test/cert.pem', got '%s'", cfg.CertPath)
	}
	if cfg.KeyPath != "/test/key.pem" {
		t.Errorf("Expected KeyPath to be '/test/key.pem', got '%s'", cfg.KeyPath)
	}
	if cfg.CADirURL != "https://test-ca.com/directory" {
		t.Errorf("Expected CADirURL to be 'https://test-ca.com/directory', got '%s'", cfg.CADirURL)
	}
	if cfg.ExpiryDaysThreshold != 15 {
		t.Errorf("Expected ExpiryDaysThreshold to be 15, got %d", cfg.ExpiryDaysThreshold)
	}
}

// Test basic domain handling (backward compatibility)
func TestGetDomainsBasic(t *testing.T) {
	cfg := &Config{
		Domain: "example.com",
		SAN:    []string{"www.example.com", "api.example.com"},
	}

	// Clear environment to test basic functionality
	os.Clearenv()

	domains := getDomains(cfg)

	expected := []string{"example.com", "www.example.com", "api.example.com"}
	if !reflect.DeepEqual(domains, expected) {
		t.Errorf("getDomains() = %v, want %v", domains, expected)
	}
}

// Test domain integration with services (subdomain mode)
func TestGetDomainsWithSubdomainIntegration(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		envVars  map[string]string
		expected []string
	}{
		{
			name: "basic domain only",
			config: &Config{
				Domain: "example.com",
				SAN:    []string{},
			},
			envVars: map[string]string{
				"ROUTING_MODE": "path",
			},
			expected: []string{"example.com"},
		},
		{
			name: "domain with SAN",
			config: &Config{
				Domain: "example.com",
				SAN:    []string{"www.example.com"},
			},
			envVars: map[string]string{
				"ROUTING_MODE": "path",
			},
			expected: []string{"example.com", "www.example.com"},
		},
		{
			name: "subdomain mode adds service subdomains",
			config: &Config{
				Domain: "example.com",
				SAN:    []string{"www.example.com"},
			},
			envVars: map[string]string{
				"DOMAIN":       "example.com",
				"ROUTING_MODE": "subdomain",
				"SERVICE_1":    "api",
				"PORT_1":       "8080",
				"SERVICE_2":    "web",
				"PORT_2":       "3000",
			},
			expected: []string{"example.com", "www.example.com", "api.example.com", "web.example.com"},
		},
		{
			name: "subdomain mode with duplicate domains",
			config: &Config{
				Domain: "example.com",
				SAN:    []string{"api.example.com"}, // Already in SAN
			},
			envVars: map[string]string{
				"DOMAIN":       "example.com",
				"ROUTING_MODE": "subdomain",
				"SERVICE_1":    "api",
				"PORT_1":       "8080",
			},
			expected: []string{"example.com", "api.example.com"}, // No duplicate
		},
		{
			name: "subdomain mode with different domain",
			config: &Config{
				Domain: "other.com",
				SAN:    []string{},
			},
			envVars: map[string]string{
				"DOMAIN":       "example.com", // Different domain
				"ROUTING_MODE": "subdomain",
				"SERVICE_1":    "api",
				"PORT_1":       "8080",
			},
			expected: []string{"other.com"}, // No subdomains added for different domain
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

			// Get domains
			result := getDomains(tt.config)

			// Compare
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("getDomains() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test environment variable handling
func TestGetEnvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "env var set",
			envKey:       "TEST_VAR",
			envValue:     "custom_value",
			defaultValue: "default_value",
			expected:     "custom_value",
		},
		{
			name:         "env var not set",
			envKey:       "UNSET_VAR",
			envValue:     "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set environment variable if provided
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
			}

			// Test function
			result := getEnvWithDefault(tt.envKey, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("getEnvWithDefault(%s, %s) = %s, want %s", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

// Test certificate expiration checking
func TestCheckCertificateExpiration(t *testing.T) {
	// Create a temporary certificate file
	certFile, err := os.CreateTemp("", "test-cert-*.pem")
	if err != nil {
		t.Fatalf("Failed to create temporary certificate file: %v", err)
	}
	defer os.Remove(certFile.Name())

	// Generate a private key for the certificate
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Generate a test certificate
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(40 * 24 * time.Hour), // 40 days from now
		DNSNames:     []string{"example.com", "www.example.com"},
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	// Encode and write the certificate to the temporary file
	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certFile.Close()

	// Test cases
	testCases := []struct {
		name                 string
		domains              []string
		expiryDaysThreshold  int
		expectedNeedsRenewal bool
	}{
		{"Valid certificate", []string{"example.com", "www.example.com"}, 30, false},
		{"Expiring soon", []string{"example.com", "www.example.com"}, 50, true},
		{"Missing domain", []string{"example.com", "missing.example.com"}, 30, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			needsRenewal, err := checkCertificateExpiration(certFile.Name(), tc.domains, tc.expiryDaysThreshold)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if needsRenewal != tc.expectedNeedsRenewal {
				t.Errorf("Expected needsRenewal to be %v, got %v", tc.expectedNeedsRenewal, needsRenewal)
			}
		})
	}
}

// Test certificate and key saving
func TestSaveCertificateAndKey(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cert-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	certPath := tempDir + "/cert.pem"
	keyPath := tempDir + "/key.pem"

	cert := &certificate.Resource{
		Certificate: []byte("TEST CERTIFICATE"),
		PrivateKey:  []byte("TEST PRIVATE KEY"),
	}

	err = saveCertificateAndKey(cert, certPath, keyPath)
	if err != nil {
		t.Fatalf("Failed to save certificate and key: %v", err)
	}

	// Check if files were created with correct content
	certContent, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("Failed to read certificate file: %v", err)
	}
	if string(certContent) != "TEST CERTIFICATE" {
		t.Errorf("Certificate content mismatch")
	}

	keyContent, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("Failed to read key file: %v", err)
	}
	if string(keyContent) != "TEST PRIVATE KEY" {
		t.Errorf("Private key content mismatch")
	}
}