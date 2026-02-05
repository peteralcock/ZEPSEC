package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Scanner executes nmap and parses results.
type Scanner struct {
	cfg    NmapConfig
	logger *log.Logger
}

// NewScanner creates a Scanner with the given nmap configuration.
func NewScanner(cfg NmapConfig, logger *log.Logger) *Scanner {
	// Ensure temp directory exists
	if err := os.MkdirAll(cfg.TempDir, 0750); err != nil {
		logger.Printf("[WARN] Failed to create temp dir %s: %v", cfg.TempDir, err)
	}
	return &Scanner{cfg: cfg, logger: logger}
}

// Run executes an nmap scan with the given options string (as built by
// ScanJob#nmap_options_string on the Rails side) and returns parsed results.
//
// The options string contains targets and flags, e.g.:
//   "192.168.1.0/24 -sS -Pn -sV -T4 -p 22,80,443"
func (s *Scanner) Run(jid, options string) (*NmapRun, error) {
	xmlPath := s.xmlOutputPath(jid)
	defer os.Remove(xmlPath)

	args := s.buildArgs(options, xmlPath)
	s.logger.Printf("[SCAN] Starting nmap (jid=%s): %s", jid, strings.Join(args, " "))

	start := time.Now()
	if err := s.execute(args); err != nil {
		return nil, fmt.Errorf("nmap execution failed: %w", err)
	}
	s.logger.Printf("[SCAN] Nmap completed (jid=%s) in %s", jid, time.Since(start))

	result, err := s.parseXML(xmlPath)
	if err != nil {
		return nil, fmt.Errorf("parsing nmap XML: %w", err)
	}

	s.logger.Printf("[SCAN] Parsed %d hosts (jid=%s)", len(result.Hosts), jid)
	return result, nil
}

// buildArgs constructs the nmap command-line arguments.
// It injects -oX for XML output and -v for verbosity.
func (s *Scanner) buildArgs(options, xmlPath string) []string {
	var args []string

	if s.cfg.UseSudo {
		args = append(args, "sudo")
	}
	args = append(args, s.cfg.BinaryPath)
	args = append(args, "-oX", xmlPath)

	// Split the options string — it contains both targets and flags
	parts := strings.Fields(options)
	args = append(args, parts...)

	return args
}

// execute runs the nmap command and captures output.
func (s *Scanner) execute(args []string) error {
	var cmd *exec.Cmd
	if len(args) > 1 {
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		cmd = exec.Command(args[0])
	}

	cmd.Stdout = &logWriter{logger: s.logger, prefix: "[NMAP] "}
	cmd.Stderr = &logWriter{logger: s.logger, prefix: "[NMAP-ERR] "}

	return cmd.Run()
}

// parseXML reads and parses the nmap XML output file.
func (s *Scanner) parseXML(xmlPath string) (*NmapRun, error) {
	data, err := os.ReadFile(xmlPath)
	if err != nil {
		return nil, fmt.Errorf("reading XML file %s: %w", xmlPath, err)
	}

	var result NmapRun
	if err := xml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling XML: %w", err)
	}

	return &result, nil
}

// xmlOutputPath generates a unique path for the nmap XML output file.
func (s *Scanner) xmlOutputPath(jid string) string {
	ts := time.Now().Format("2006.01.02-15.04.05.000000")
	filename := fmt.Sprintf("%s_%s_nmap.xml", jid, ts)
	return filepath.Join(s.cfg.TempDir, filename)
}

// ToPayload converts parsed nmap XML results into the ZEPSEC API payload format.
func (s *Scanner) ToPayload(jid, externalIP string, nmapResult *NmapRun) *ScanResultPayload {
	payload := &ScanResultPayload{
		JID:        jid,
		ExternalIP: externalIP,
		Hosts:      make([]HostResult, 0, len(nmapResult.Hosts)),
	}

	for _, host := range nmapResult.Hosts {
		hr := HostResult{
			StartTime: host.StartTime,
			EndTime:   host.EndTime,
			Addresses: make([]Address, 0, len(host.Addresses)),
		}

		for _, addr := range host.Addresses {
			hr.Addresses = append(hr.Addresses, Address{
				AddrType: addr.AddrType,
				Addr:     addr.Addr,
			})
		}

		for _, port := range host.Ports.Ports {
			pr := PortResult{
				ID:       port.PortID,
				Protocol: port.Protocol,
				State:    PortState{State: port.State.State},
			}
			if port.Service != nil {
				pr.Service = &ServiceInfo{
					Name:      port.Service.Name,
					Product:   port.Service.Product,
					Version:   port.Service.Version,
					ExtraInfo: port.Service.ExtraInfo,
				}
			}
			hr.Ports = append(hr.Ports, pr)
		}

		payload.Hosts = append(payload.Hosts, hr)
	}

	return payload
}

// logWriter adapts log.Logger for use as an io.Writer with a prefix.
type logWriter struct {
	logger *log.Logger
	prefix string
}

func (w *logWriter) Write(p []byte) (int, error) {
	lines := strings.Split(strings.TrimRight(string(p), "\n"), "\n")
	for _, line := range lines {
		if line != "" {
			w.logger.Printf("%s%s", w.prefix, line)
		}
	}
	return len(p), nil
}
