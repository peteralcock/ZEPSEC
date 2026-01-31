package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Scheduler runs autonomous periodic scans defined in the agent config.
// These run independently of server-dispatched jobs, enabling the agent
// to continuously monitor networks post-exploit without server coordination.
type Scheduler struct {
	scans    []ScheduledScan
	scanner  *Scanner
	reporter *Reporter
	network  *Network
	logger   *log.Logger

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewScheduler creates a Scheduler for the given scan definitions.
func NewScheduler(scans []ScheduledScan, scanner *Scanner, reporter *Reporter, network *Network, logger *log.Logger) *Scheduler {
	return &Scheduler{
		scans:    scans,
		scanner:  scanner,
		reporter: reporter,
		network:  network,
		logger:   logger,
		stopCh:   make(chan struct{}),
	}
}

// Start launches a goroutine for each scheduled scan.
func (s *Scheduler) Start() {
	if len(s.scans) == 0 {
		s.logger.Printf("[SCHED] No scheduled scans configured")
		return
	}

	s.logger.Printf("[SCHED] Starting %d scheduled scan(s)", len(s.scans))

	for i := range s.scans {
		scan := s.scans[i]
		interval, err := time.ParseDuration(scan.Interval)
		if err != nil {
			s.logger.Printf("[SCHED] Invalid interval %q for scan %q: %v", scan.Interval, scan.Name, err)
			continue
		}

		s.wg.Add(1)
		go s.runLoop(scan, interval)
	}
}

// Stop signals all scan loops to terminate and waits for them to finish.
func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	s.logger.Printf("[SCHED] All scheduled scans stopped")
}

func (s *Scheduler) runLoop(scan ScheduledScan, interval time.Duration) {
	defer s.wg.Done()

	s.logger.Printf("[SCHED] Scan %q: targets=%q, options=%q, every %s",
		scan.Name, scan.Targets, scan.Options, interval)

	// Run immediately on startup, then on interval
	s.runScan(scan)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.runScan(scan)
		}
	}
}

func (s *Scheduler) runScan(scan ScheduledScan) {
	// Generate a synthetic JID for autonomous scans
	jid := fmt.Sprintf("auto_%s_%d", scan.Name, time.Now().Unix())

	// Build the options string: targets + nmap flags (matching Rails nmap_options_string format)
	options := scan.Targets
	if scan.Options != "" {
		options = fmt.Sprintf("%s %s", scan.Targets, scan.Options)
	}

	s.logger.Printf("[SCHED] Running autonomous scan %q (jid=%s)", scan.Name, jid)

	result, err := s.scanner.Run(jid, options)
	if err != nil {
		s.logger.Printf("[SCHED] Scan %q failed: %v", scan.Name, err)
		return
	}

	externalIP := s.network.ExternalIP()
	payload := s.scanner.ToPayload(jid, externalIP, result)

	if err := s.reporter.SendResults(payload); err != nil {
		s.logger.Printf("[SCHED] Failed to report results for %q: %v", scan.Name, err)
		return
	}

	s.logger.Printf("[SCHED] Scan %q completed: %d hosts reported", scan.Name, len(payload.Hosts))
}
