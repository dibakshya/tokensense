package proxy

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/dibakshya/tokensense/internal/classifier"
	"github.com/dibakshya/tokensense/internal/storage"
)

// Server is the local HTTPS proxy server.
type Server struct {
	addr        string
	listener    net.Listener
	httpServer  *http.Server
	caCert      *x509.Certificate
	caKey       *ecdsa.PrivateKey
	db          *storage.DB
	classifier  classifier.Classifier
	matrix      *classifier.ModelMatrix
	contentMode bool
	logger      *log.Logger
}

// Config holds proxy server configuration.
type Config struct {
	Addr        string
	CACert      *x509.Certificate
	CAKey       *ecdsa.PrivateKey
	DB          *storage.DB
	Classifier  classifier.Classifier
	Matrix      *classifier.ModelMatrix
	ContentMode bool
	Logger      *log.Logger
}

// New creates a new proxy server.
func New(cfg Config) *Server {
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	s := &Server{
		addr:        cfg.Addr,
		caCert:      cfg.CACert,
		caKey:       cfg.CAKey,
		db:          cfg.DB,
		classifier:  cfg.Classifier,
		matrix:      cfg.Matrix,
		contentMode: cfg.ContentMode,
		logger:      cfg.Logger,
	}
	return s
}

// ListenAndServe starts the proxy server. It binds to 127.0.0.1 only.
func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("cannot listen on %s: %w", s.addr, err)
	}
	s.listener = ln
	s.logger.Printf("Tokensense proxy listening on %s", s.addr)

	handler := &proxyHandler{
		server: s,
	}

	s.httpServer = &http.Server{
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return s.httpServer.Serve(ln)
}

// Shutdown gracefully shuts down the proxy server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// Addr returns the address the proxy is listening on.
func (s *Server) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}
