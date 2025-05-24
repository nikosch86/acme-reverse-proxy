package main

// This script is created with the help of Anthropic Claude 3.5 Sonnet
// It might not be pretty, but it works and didn't take a lot of time to write

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

const (
	challengeBasePath = "/usr/share/nginx/challenge/.well-known/acme-challenge"
	defaultCADirURL   = "https://acme-staging-v02.api.letsencrypt.org/directory"
	defaultEmail      = "notmy@mail.com"
	defaultCertPath   = "/etc/ssl/private/fullchain.pem"
	defaultKeyPath    = "/etc/ssl/private/key.pem"
	filePerm          = 0644
	dirPerm           = 0755
	defaultExpiryDaysThreshold = 30
	defaultSAN        = ""
)

type Config struct {
	Email    string
	Domain   string
	SAN      []string
	CertPath string
	KeyPath  string
	CADirURL string
	ExpiryDaysThreshold int
}

// MyUser implements the acme.User interface
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

type customHTTP01Provider struct{}

func (d *customHTTP01Provider) Present(domain, token, keyAuth string) error {
	challengePath := filepath.Join(challengeBasePath, token)
	err := os.MkdirAll(challengeBasePath, dirPerm)
	if err != nil {
		return err
	}
	return os.WriteFile(challengePath, []byte(keyAuth), filePerm)
}

func (d *customHTTP01Provider) CleanUp(domain, token, keyAuth string) error {
	challengePath := filepath.Join(challengeBasePath, token)
	return os.Remove(challengePath)
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := loadConfig()

	if err := run(ctx, config); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func loadConfig() *Config {
	expiryDaysThreshold, _ := strconv.Atoi(getEnvWithDefault("EXPIRY_DAYS_THRESHOLD", strconv.Itoa(defaultExpiryDaysThreshold)))
	san := getEnvWithDefault("SAN", defaultSAN)
	var sanSlice []string
	if san != "" {
        sanSlice = append(sanSlice, strings.Split(san, ",")...)
    }
	return &Config{
		Email:    getEnvWithDefault("EMAIL", defaultEmail),
		Domain:   os.Getenv("DOMAIN"),
		SAN:      sanSlice,
		CertPath: getEnvWithDefault("CERT_PATH", defaultCertPath),
		KeyPath:  getEnvWithDefault("KEY_PATH", defaultKeyPath),
		CADirURL: getEnvWithDefault("CA_DIR_URL", defaultCADirURL),
		ExpiryDaysThreshold: expiryDaysThreshold,
	}
}

func run(ctx context.Context, cfg *Config) error {
	if cfg.Domain == "" {
		return fmt.Errorf("DOMAIN environment variable is not set")
	}

	domains := getDomains(cfg)

	needsRenewal, err := checkCertificateExpiration(cfg.CertPath, domains, cfg.ExpiryDaysThreshold)
	if err != nil {
		log.Printf("Error checking certificate expiration: %v", err)
	}

	if !needsRenewal {
		log.Println("Certificate is still valid and not due for renewal")
		return nil
	}

	client, err := setupACMEClient(cfg)
	if err != nil {
		return fmt.Errorf("setting up ACME client: %w", err)
	}

	certificates, err := obtainCertificate(ctx, client, domains)
	if err != nil {
		return fmt.Errorf("obtaining certificate: %w", err)
	}

	if err := saveCertificateAndKey(certificates, cfg.CertPath, cfg.KeyPath); err != nil {
		return fmt.Errorf("saving certificate and key: %w", err)
	}

	if err := reloadNginx(); err != nil {
		return fmt.Errorf("reloading Nginx: %w", err)
	}

	log.Println("Certificate obtained and Nginx reloaded successfully")
	return nil
}

func getDomains(cfg *Config) []string {
    domains := []string{cfg.Domain}
    return append(domains, cfg.SAN...)
}

func checkCertificateExpiration(certPath string, domains []string, expiryDaysThreshold int) (bool, error) {
	certPEMBlock, err := os.ReadFile(certPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil // Certificate doesn't exist, needs to be obtained
		}
		return false, err
	}

	certDERBlock, _ := pem.Decode(certPEMBlock)
	if certDERBlock == nil {
		return false, fmt.Errorf("failed to parse certificate PEM data")
	}

	cert, err := x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Check if the certificate is valid for all required domains
    for _, domain := range domains {
        if err := cert.VerifyHostname(domain); err != nil {
            log.Printf("Certificate is not valid for domain %s: %v", domain, err)
            return true, nil // Certificate needs to be renewed
        }
    }

	daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)
	log.Printf("Certificate expires in %d days", daysUntilExpiry)

	return daysUntilExpiry <= expiryDaysThreshold, nil
}

func setupACMEClient(cfg *Config) (*lego.Client, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating private key: %w", err)
	}

	user := &MyUser{
		Email: cfg.Email,
		key:   privateKey,
	}

	config := lego.NewConfig(user)
	config.CADirURL = cfg.CADirURL

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("creating ACME client: %w", err)
	}

	err = client.Challenge.SetHTTP01Provider(&customHTTP01Provider{})
	if err != nil {
		return nil, fmt.Errorf("setting HTTP-01 provider: %w", err)
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("registering user: %w", err)
	}
	user.Registration = reg

	return client, nil
}

func obtainCertificate(ctx context.Context, client *lego.Client, domains []string) (*certificate.Resource, error) {
    request := certificate.ObtainRequest{
        Domains: domains,
        Bundle:  true,
    }
    return client.Certificate.Obtain(request)
}

func saveCertificateAndKey(cert *certificate.Resource, certPath, keyPath string) error {
	if err := os.WriteFile(certPath, cert.Certificate, filePerm); err != nil {
		return fmt.Errorf("saving certificate: %w", err)
	}
	if err := os.WriteFile(keyPath, cert.PrivateKey, 0600); err != nil {
		return fmt.Errorf("saving private key: %w", err)
	}
	return nil
}

func reloadNginx() error {
	cmd := exec.Command("nginx", "-s", "reload")
	return cmd.Run()
}

func setupGracefulShutdown(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Shutting down gracefully...")
		cancel()
	}()
}
