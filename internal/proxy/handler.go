package proxy

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/dibakshya/tokensense/internal/classifier"
	"github.com/dibakshya/tokensense/internal/storage"
)

// proxyHandler handles CONNECT and direct HTTP requests.
type proxyHandler struct {
	server *Server
}

func (h *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		h.handleConnect(w, r)
	} else {
		http.Error(w, "Only CONNECT method is supported", http.StatusMethodNotAllowed)
	}
}

func (h *proxyHandler) handleConnect(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if !strings.Contains(host, ":") {
		host += ":443"
	}

	hostname := strings.Split(host, ":")[0]
	provider := classifier.ProviderFromHost(hostname)

	// If not an AI API or in metadata-only mode for cert-pinned tools, tunnel transparently
	if provider == classifier.ProviderUnknown || !h.server.contentMode {
		h.tunnelConnect(w, r, hostname, provider)
		return
	}

	// TLS termination mode for known AI APIs
	h.interceptConnect(w, r, hostname, provider)
}

func (h *proxyHandler) interceptConnect(w http.ResponseWriter, r *http.Request, hostname, provider string) {
	// Hijack the client connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		h.server.logger.Printf("Hijack failed: %v", err)
		return
	}
	defer clientConn.Close()

	// Send 200 Connection Established
	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// Sign a cert for this host
	hostCert, hostKey, err := h.signForHost(hostname)
	if err != nil {
		h.server.logger.Printf("Cannot sign cert for %s: %v", hostname, err)
		return
	}

	tlsCert := tls.Certificate{
		Certificate: [][]byte{hostCert.Raw},
		PrivateKey:  hostKey,
	}

	// TLS handshake with client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}
	tlsConn := tls.Server(clientConn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		h.server.logger.Printf("TLS handshake with client failed for %s: %v", hostname, err)
		return
	}
	defer tlsConn.Close()

	// Read the actual HTTP request from the TLS connection.
	// We cap at 256KB to bound memory per request; large prompt bodies beyond
	// this limit are forwarded correctly but classification uses the truncated prefix.
	const maxRequestRead = 256 * 1024
	reqBuf := make([]byte, maxRequestRead)
	n, err := tlsConn.Read(reqBuf)
	if err != nil {
		h.server.logger.Printf("Cannot read from TLS conn: %v", err)
		return
	}
	requestData := reqBuf[:n]

	startTime := time.Now()

	// Forward to upstream — prefer IPv4 to avoid "bad file descriptor" on
	// environments where IPv6 is unavailable (e.g. split-tunnel VPNs).
	rawConn, err := dialUpstream(hostname, "443")
	if err != nil {
		h.server.logger.Printf("Cannot connect to upstream %s: %v", hostname, err)
		return
	}
	upstreamTLS := tls.Client(rawConn, &tls.Config{ServerName: hostname})
	if err := upstreamTLS.Handshake(); err != nil {
		rawConn.Close()
		h.server.logger.Printf("TLS handshake with upstream %s failed: %v", hostname, err)
		return
	}
	defer upstreamTLS.Close()

	_, err = upstreamTLS.Write(requestData)
	if err != nil {
		h.server.logger.Printf("Cannot write to upstream: %v", err)
		return
	}

	// Read and forward upstream response, classifying in process
	// content is in-process only; never persisted or logged
	responseData, err := readAndForward(upstreamTLS, tlsConn)
	latencyMs := time.Since(startTime).Milliseconds()

	if err != nil {
		h.server.logger.Printf("Stream error: %v", err)
	}

	// Extract metadata and classify
	go func() {
		defer func() {
			if r := recover(); r != nil {
				h.server.logger.Printf("Classifier panic recovered: %v", r)
			}
		}()
		h.classifyAndStore(requestData, responseData, hostname, provider, latencyMs)
	}()
}

func (h *proxyHandler) tunnelConnect(w http.ResponseWriter, r *http.Request, hostname, provider string) {
	host := r.Host
	if !strings.Contains(host, ":") {
		host += ":443"
	}

	// Connect to upstream — prefer IPv4 (same reason as interceptConnect).
	hostname2 := strings.Split(host, ":")[0]
	port2 := "443"
	if idx := strings.LastIndex(host, ":"); idx >= 0 {
		port2 = host[idx+1:]
	}
	upstream, err := dialUpstream(hostname2, port2)
	if err != nil {
		http.Error(w, "Cannot connect to upstream", http.StatusBadGateway)
		return
	}
	defer upstream.Close()

	// Hijack client connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		h.server.logger.Printf("Hijack failed: %v", err)
		return
	}
	defer clientConn.Close()

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// Store metadata-only row if it's a known AI provider
	if provider != classifier.ProviderUnknown {
		go func() {
			req := &storage.RequestMetadata{
				ID:          uuid.New().String(),
				Timestamp:   time.Now().UnixMilli(),
				DayDate:     time.Now().Format("2006-01-02"),
				Provider:    provider,
				Model:       "unknown",
				ContentMode: 0,
				Intercepted: 0,
			}
			src := classifier.SourceMetadataOnly
			req.ClassifierSource = &src
			if err := h.server.db.Insert(req); err != nil {
				h.server.logger.Printf("Cannot store metadata row: %v", err)
			}
		}()
	}

	// Bidirectional tunnel
	tunnel(clientConn, upstream)
}

func (h *proxyHandler) classifyAndStore(requestData, responseData []byte, hostname, provider string, latencyMs int64) {
	// content is in-process only; never persisted or logged
	model := extractModel(requestData)
	tokensIn, tokensOut := extractTokenCounts(responseData)
	toolSource := "direct"

	var result *classifier.ClassificationResult
	if h.server.classifier != nil && h.server.contentMode {
		// content is in-process only; never persisted or logged
		promptText := extractPromptText(requestData)
		var err error
		result, err = h.server.classifier.Classify(promptText)
		if err != nil {
			h.server.logger.Printf("Classification failed: %v", err)
		}
		// promptText is now eligible for GC
	}

	var cost *float64
	if h.server.matrix != nil && model != "" {
		if m := h.server.matrix.FindModel(model); m != nil {
			c := classifier.CostForRequest(m.Pricing, tokensIn, tokensOut)
			cost = &c
		}
	}

	req := &storage.RequestMetadata{
		ID:          uuid.New().String(),
		Timestamp:   time.Now().UnixMilli(),
		DayDate:     time.Now().Format("2006-01-02"),
		Provider:    provider,
		Model:       model,
		ContentMode: 1,
		Intercepted: 1,
	}

	if tokensIn > 0 {
		req.TokensIn = &tokensIn
	}
	if tokensOut > 0 {
		req.TokensOut = &tokensOut
	}
	if cost != nil {
		req.CostUSD = cost
	}
	lat := int(latencyMs)
	req.LatencyMs = &lat
	ts := toolSource
	req.ToolSource = &ts

	if result != nil {
		req.TaskType = &result.TaskType
		req.Complexity = &result.Complexity
		req.ClassifierSource = &result.Source
		req.ClassifierConfidence = &result.Confidence
	}

	if err := h.server.db.Insert(req); err != nil {
		h.server.logger.Printf("Cannot store request row: %v", err)
	}
}

func (h *proxyHandler) signForHost(hostname string) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	return signForHostWithCA(hostname, h.server.caCert, h.server.caKey)
}

// dialUpstream connects to host:port, trying IPv4 first then IPv6.
// Go's default resolver picks whichever address comes first in DNS, which is
// often IPv6. On environments with a split-tunnel VPN or broken IPv6 (common
// on corporate networks), IPv6 dials fail with "bad file descriptor".
// Trying tcp4 first, then falling back to tcp, ensures we always get a working
// connection when IPv4 is available.
func dialUpstream(hostname, port string) (net.Conn, error) {
	addr := net.JoinHostPort(hostname, port)
	if conn, err := net.DialTimeout("tcp4", addr, 10*time.Second); err == nil {
		return conn, nil
	}
	return net.DialTimeout("tcp", addr, 10*time.Second)
}

func extractModel(data []byte) string {
	// Try to parse as JSON to extract model field
	var body map[string]interface{}
	// Find JSON body after HTTP headers
	idx := bytes.Index(data, []byte("\r\n\r\n"))
	if idx < 0 {
		return "unknown"
	}
	jsonBody := data[idx+4:]
	if err := json.Unmarshal(jsonBody, &body); err != nil {
		return "unknown"
	}
	if model, ok := body["model"].(string); ok {
		return model
	}
	return "unknown"
}

func extractPromptText(data []byte) string {
	// content is in-process only; never persisted or logged
	idx := bytes.Index(data, []byte("\r\n\r\n"))
	if idx < 0 {
		return ""
	}
	jsonBody := data[idx+4:]
	var body map[string]interface{}
	if err := json.Unmarshal(jsonBody, &body); err != nil {
		return ""
	}

	// Try Anthropic / OpenAI format (both use "messages" array)
	if messages, ok := body["messages"].([]interface{}); ok {
		var parts []string
		for _, msg := range messages {
			if m, ok := msg.(map[string]interface{}); ok {
				switch c := m["content"].(type) {
				case string:
					parts = append(parts, c)
				case []interface{}:
					// Anthropic multimodal format: content is array of {type, text/image} blocks
					for _, block := range c {
						if bm, ok := block.(map[string]interface{}); ok {
							if bm["type"] == "text" {
								if text, ok := bm["text"].(string); ok {
									parts = append(parts, text)
								}
							}
						}
					}
				}
			}
		}
		return strings.Join(parts, " ")
	}

	// Try Google Gemini format ("contents" array with "parts")
	if contents, ok := body["contents"].([]interface{}); ok {
		var parts []string
		for _, c := range contents {
			if cm, ok := c.(map[string]interface{}); ok {
				if ps, ok := cm["parts"].([]interface{}); ok {
					for _, p := range ps {
						if pm, ok := p.(map[string]interface{}); ok {
							if text, ok := pm["text"].(string); ok {
								parts = append(parts, text)
							}
						}
					}
				}
			}
		}
		return strings.Join(parts, " ")
	}

	return ""
}

type usageResponse struct {
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		PromptTokens int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func extractTokenCounts(data []byte) (int, int) {
	// Try to find the usage block in response
	idx := bytes.Index(data, []byte("\r\n\r\n"))
	if idx < 0 {
		idx = 0
	} else {
		idx += 4
	}
	responseBody := data[idx:]

	var resp usageResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		return 0, 0
	}

	tokensIn := resp.Usage.InputTokens
	if tokensIn == 0 {
		tokensIn = resp.Usage.PromptTokens
	}
	tokensOut := resp.Usage.OutputTokens
	if tokensOut == 0 {
		tokensOut = resp.Usage.CompletionTokens
	}

	return tokensIn, tokensOut
}
