package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
)

// Server handles incoming scan requests from the ZEPSEC Rails application.
// The Rails server POSTs to /scans with query params: id (jid) and options (nmap flags).
type Server struct {
	secret   string
	scanner  *Scanner
	reporter *Reporter
	network  *Network
	logger   *log.Logger

	mu       sync.Mutex
	running  map[string]*ScanJob
}

// NewServer creates a Server wired to the scanner, reporter, and network utilities.
func NewServer(secret string, scanner *Scanner, reporter *Reporter, network *Network, logger *log.Logger) *Server {
	return &Server{
		secret:   secret,
		scanner:  scanner,
		reporter: reporter,
		network:  network,
		logger:   logger,
		running:  make(map[string]*ScanJob),
	}
}

// Handler returns an http.Handler with the agent's routes.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/scans", s.handleScan)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/status", s.handleStatus)
	return mux
}

// handleScan receives a scan job from the ZEPSEC server.
// Expected: POST /scans?id=<jid>&options=<nmap_options_string>
// Auth: Authorization: Bearer <agent_secret>
func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.authenticate(r) {
		s.jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	jid := r.URL.Query().Get("id")
	options := r.URL.Query().Get("options")

	if jid == "" {
		s.jsonError(w, "missing required parameter: id", http.StatusBadRequest)
		return
	}
	if options == "" {
		s.jsonError(w, "missing required parameter: options", http.StatusBadRequest)
		return
	}

	// Track the job
	s.mu.Lock()
	if _, exists := s.running[jid]; exists {
		s.mu.Unlock()
		s.jsonError(w, "scan job already running", http.StatusConflict)
		return
	}
	job := &ScanJob{JID: jid, Options: options, Status: "running"}
	s.running[jid] = job
	s.mu.Unlock()

	s.logger.Printf("[SERVER] Accepted scan job (jid=%s, options=%q)", jid, options)

	// Respond immediately — the scan runs asynchronously
	s.jsonResponse(w, map[string]string{"message": "accepted"}, http.StatusOK)

	// Execute scan in background
	go s.executeScan(job)
}

// executeScan runs nmap, parses results, and reports back to ZEPSEC.
func (s *Server) executeScan(job *ScanJob) {
	defer func() {
		s.mu.Lock()
		delete(s.running, job.JID)
		s.mu.Unlock()
	}()

	result, err := s.scanner.Run(job.JID, job.Options)
	if err != nil {
		job.Status = "failed"
		s.logger.Printf("[SERVER] Scan failed (jid=%s): %v", job.JID, err)
		return
	}

	externalIP := s.network.ExternalIP()
	payload := s.scanner.ToPayload(job.JID, externalIP, result)

	if err := s.reporter.SendResults(payload); err != nil {
		job.Status = "failed"
		s.logger.Printf("[SERVER] Failed to report results (jid=%s): %v", job.JID, err)
		return
	}

	job.Status = "completed"
	s.logger.Printf("[SERVER] Scan job completed (jid=%s)", job.JID)
}

// handleHealth provides a simple health check endpoint.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, map[string]string{"status": "ok"}, http.StatusOK)
}

// handleStatus returns the count and IDs of currently running scans.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if !s.authenticate(r) {
		s.jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	s.mu.Lock()
	jobs := make([]map[string]string, 0, len(s.running))
	for _, job := range s.running {
		jobs = append(jobs, map[string]string{
			"jid":    job.JID,
			"status": job.Status,
		})
	}
	s.mu.Unlock()

	s.jsonResponse(w, map[string]interface{}{
		"running_scans": len(jobs),
		"scans":         jobs,
	}, http.StatusOK)
}

// authenticate checks the Bearer token against the configured agent secret.
func (s *Server) authenticate(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return false
	}
	token := strings.TrimPrefix(auth, prefix)
	return token == s.secret
}

func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) jsonError(w http.ResponseWriter, message string, status int) {
	s.jsonResponse(w, map[string]interface{}{
		"error":  message,
		"status": status,
	}, status)
}
