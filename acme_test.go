package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"
	"strings"
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
	os.Setenv("ACME_CA_CERT_PATH", "/test/ca-cert.pem")

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
	if cfg.CACertPath != "/test/ca-cert.pem" {
		t.Errorf("Expected CACertPath to be '/test/ca-cert.pem', got '%s'", cfg.CACertPath)
	}

	// Test with empty ACME_CA_CERT_PATH
	os.Unsetenv("ACME_CA_CERT_PATH")
	cfg2 := loadConfig()
	if cfg2.CACertPath != "" {
		t.Errorf("Expected CACertPath to be empty when env var not set, got '%s'", cfg2.CACertPath)
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

func TestSetupACMEClientWithCACert(t *testing.T) {
	// Create a temporary CA certificate file for testing
	caCertFile, err := os.CreateTemp("", "test-ca-cert-*.pem")
	if err != nil {
		t.Fatalf("Failed to create temporary CA certificate file: %v", err)
	}
	defer os.Remove(caCertFile.Name())

	// Generate a test CA certificate
	caPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate CA private key: %v", err)
	}

	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		IsCA:         true,
		KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	caDerBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		t.Fatalf("Failed to create test CA certificate: %v", err)
	}

	// Write the CA certificate to file
	pem.Encode(caCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: caDerBytes})
	caCertFile.Close()

	// Test with CA certificate path
	cfg := &Config{
		Email:      "test@example.com",
		Domain:     "example.com",
		CADirURL:   "https://test-ca.com/directory",
		CACertPath: caCertFile.Name(),
	}

	// Note: We can't fully test the ACME client creation without a real ACME server,
	// but we can verify that the function handles the CA certificate path correctly
	_, err = setupACMEClient(cfg)
	// The error is expected since we don't have a real ACME server
	if err == nil {
		t.Error("Expected error when connecting to non-existent ACME server")
	}

	// Test without CA certificate path
	cfg2 := &Config{
		Email:    "test@example.com",
		Domain:   "example.com",
		CADirURL: "https://test-ca.com/directory",
	}

	_, err = setupACMEClient(cfg2)
	// The error is expected since we don't have a real ACME server
	if err == nil {
		t.Error("Expected error when connecting to non-existent ACME server")
	}

	// Test with non-existent CA certificate file
	cfg3 := &Config{
		Email:      "test@example.com",
		Domain:     "example.com",
		CADirURL:   "https://test-ca.com/directory",
		CACertPath: "/non/existent/ca-cert.pem",
	}

	_, err = setupACMEClient(cfg3)
	if err == nil {
		t.Error("Expected error when CA certificate file doesn't exist")
	}
	if err != nil && !os.IsNotExist(err) {
		// Check that the error is related to reading the CA certificate
		expectedErrMsg := "reading CA certificate from /non/existent/ca-cert.pem"
		if !strings.Contains(err.Error(), expectedErrMsg) {
			t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrMsg, err.Error())
		}
	}
}

// Add more tests for other functions as needed
