package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const version = "1.0.0"

func main() {
	var (
		configPath  string
		showVersion bool
	)
	flag.StringVar(&configPath, "config", "config.yml", "path to config file")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("zepsec-nmap-agent v%s\n", version)
		os.Exit(0)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmsgprefix)
	logger.Printf("[AGENT] ZEPSEC Nmap Agent v%s starting", version)

	// Load configuration
	cfg, err := LoadConfig(configPath)
	if err != nil {
		logger.Fatalf("[AGENT] Configuration error: %v", err)
	}

	// Initialize components
	network := NewNetwork(logger)
	scanner := NewScanner(cfg.Nmap, logger)
	reporter := NewReporter(cfg.Server, logger)
	srv := NewServer(cfg.Agent.Secret, scanner, reporter, network, logger)

	// Log agent info
	logger.Printf("[AGENT] Listening on %s", cfg.Agent.ListenAddr)
	logger.Printf("[AGENT] ZEPSEC server: %s", cfg.Server.URL)
	logger.Printf("[AGENT] Nmap binary: %s (sudo=%v)", cfg.Nmap.BinaryPath, cfg.Nmap.UseSudo)

	localIPs := network.LocalInterfaces()
	if len(localIPs) > 0 {
		logger.Printf("[AGENT] Local interfaces: %v", localIPs)
	}

	// Start the scheduled scans (autonomous mode)
	scheduler := NewScheduler(cfg.Scans, scanner, reporter, network, logger)
	scheduler.Start()

	// Configure the HTTP server
	httpServer := &http.Server{
		Addr:         cfg.Agent.ListenAddr,
		Handler:      srv.Handler(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start listening
	go func() {
		var err error
		if cfg.Agent.TLSCert != "" && cfg.Agent.TLSKey != "" {
			logger.Printf("[AGENT] TLS enabled")
			httpServer.TLSConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
			err = httpServer.ListenAndServeTLS(cfg.Agent.TLSCert, cfg.Agent.TLSKey)
		} else {
			logger.Printf("[AGENT] TLS disabled (plain HTTP)")
			err = httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			logger.Fatalf("[AGENT] Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logger.Printf("[AGENT] Received signal %s, shutting down...", sig)

	// Graceful shutdown
	scheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Printf("[AGENT] HTTP server shutdown error: %v", err)
	}

	logger.Printf("[AGENT] Shutdown complete")
}
