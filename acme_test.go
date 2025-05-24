package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/certificate"
)

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

func TestGetDomains(t *testing.T) {
	cfg := &Config{
		Domain: "example.com",
		SAN:    []string{"www.example.com", "api.example.com"},
	}

	domains := getDomains(cfg)

	expected := []string{"example.com", "www.example.com", "api.example.com"}
	if len(domains) != len(expected) {
		t.Errorf("Expected %d domains, got %d", len(expected), len(domains))
	}
	for i, domain := range domains {
		if domain != expected[i] {
			t.Errorf("Expected domain %s at index %d, got %s", expected[i], i, domain)
		}
	}
}

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

// Add more tests for other functions as needed
