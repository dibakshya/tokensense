package proxy

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"
)

// certCache caches per-host TLS certificates to avoid re-signing on every request.
type certCache struct {
	mu    sync.RWMutex
	certs map[string]*cachedCert
}

type cachedCert struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
}

var hostCertCache = &certCache{
	certs: make(map[string]*cachedCert),
}

func signForHostWithCA(host string, caCert *x509.Certificate, caKey *ecdsa.PrivateKey) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	// Check cache first
	hostCertCache.mu.RLock()
	if cached, ok := hostCertCache.certs[host]; ok {
		hostCertCache.mu.RUnlock()
		return cached.cert, cached.key, nil
	}
	hostCertCache.mu.RUnlock()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate host key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate serial: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   host,
			Organization: []string{"Tokensense"},
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
	}

	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = []net.IP{ip}
	} else {
		template.DNSNames = []string{host}
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create host certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse host certificate: %w", err)
	}

	// Cache it
	hostCertCache.mu.Lock()
	hostCertCache.certs[host] = &cachedCert{cert: cert, key: key}
	hostCertCache.mu.Unlock()

	return cert, key, nil
}
