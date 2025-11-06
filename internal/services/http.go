package services

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rgnote/TheneKunda/internal/config"
	"github.com/rgnote/TheneKunda/internal/logger"
)

// HTTPService implements an HTTP honeypot
type HTTPService struct {
	config config.ServiceConfig
	logger *logger.Logger
	server *http.Server
}

// NewHTTPService creates a new HTTP honeypot service
func NewHTTPService(cfg config.ServiceConfig, log *logger.Logger) *HTTPService {
	return &HTTPService{
		config: cfg,
		logger: log,
	}
}

// Start starts the HTTP honeypot service
func (h *HTTPService) Start() error {
	addr := fmt.Sprintf("0.0.0.0:%d", h.config.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.handleRequest)

	h.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	h.logger.Info("HTTP honeypot listening on %s", addr)

	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.Error("HTTP server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the HTTP honeypot service
func (h *HTTPService) Stop() error {
	if h.server != nil {
		return h.server.Close()
	}
	return nil
}

func (h *HTTPService) handleRequest(w http.ResponseWriter, r *http.Request) {
	ip := h.getClientIP(r)

	// Log the request
	extraData := map[string]interface{}{
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
		"headers":    h.sanitizeHeaders(r.Header),
	}

	// Check for authentication attempts
	username, password, hasAuth := r.BasicAuth()
	if hasAuth {
		h.logger.LogConnection("HTTP", ip, 0, username, password)
		extraData["auth_type"] = "basic"
	}

	// Log request
	message := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	h.logger.LogEvent("HTTP", "http_request", ip, message, extraData)

	// Parse POST data for credentials
	if r.Method == "POST" {
		r.ParseForm()
		if user := r.FormValue("username"); user != "" {
			pass := r.FormValue("password")
			h.logger.LogConnection("HTTP", ip, 0, user, pass)
			extraData["form_auth"] = true
		}
	}

	// Set server banner
	if h.config.ServerBanner != "" {
		w.Header().Set("Server", h.config.ServerBanner)
	}

	// Send fake response based on path
	h.sendFakeResponse(w, r)
}

func (h *HTTPService) sendFakeResponse(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Common paths that attackers probe
	switch {
	case path == "/" || path == "/index.html":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Welcome</title></head>
<body>
<h1>It works!</h1>
<p>This is the default web page for this server.</p>
</body>
</html>`))

	case strings.HasPrefix(path, "/admin") || strings.HasPrefix(path, "/login"):
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Login</title></head>
<body>
<h2>Login</h2>
<form method="post">
<input type="text" name="username" placeholder="Username"><br>
<input type="password" name="password" placeholder="Password"><br>
<button type="submit">Login</button>
</form>
</body>
</html>`))

	case strings.Contains(path, "phpmyadmin") || strings.Contains(path, "wp-admin"):
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))

	case strings.HasSuffix(path, ".php") || strings.HasSuffix(path, ".asp"):
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<!-- Fake page -->"))

	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>404 Not Found</title></head>
<body>
<h1>Not Found</h1>
<p>The requested URL was not found on this server.</p>
</body>
</html>`))
	}
}

func (h *HTTPService) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use RemoteAddr
	ip, _ := getIPPort(r.RemoteAddr)
	return ip
}

func (h *HTTPService) sanitizeHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}
