package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Network provides network utility functions: external IP detection,
// local interface enumeration, and connectivity checks.
type Network struct {
	logger     *log.Logger
	mu         sync.RWMutex
	cachedIP   string
	cacheTime  time.Time
	cacheTTL   time.Duration
}

// NewNetwork creates a Network utility.
func NewNetwork(logger *log.Logger) *Network {
	return &Network{
		logger:   logger,
		cacheTTL: 5 * time.Minute,
	}
}

// ExternalIP returns the agent's external IP address.
// Results are cached for 5 minutes.
func (n *Network) ExternalIP() string {
	n.mu.RLock()
	if n.cachedIP != "" && time.Since(n.cacheTime) < n.cacheTTL {
		ip := n.cachedIP
		n.mu.RUnlock()
		return ip
	}
	n.mu.RUnlock()

	ip := n.fetchExternalIP()

	n.mu.Lock()
	n.cachedIP = ip
	n.cacheTime = time.Now()
	n.mu.Unlock()

	return ip
}

// fetchExternalIP tries multiple services to determine the external IP.
func (n *Network) fetchExternalIP() string {
	services := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
		"https://checkip.amazonaws.com",
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, url := range services {
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		ip := strings.TrimSpace(string(body))
		if net.ParseIP(ip) != nil {
			n.logger.Printf("[NET] External IP: %s (via %s)", ip, url)
			return ip
		}
	}

	// Fallback: try to determine IP from local interfaces
	ip := n.primaryLocalIP()
	if ip != "" {
		n.logger.Printf("[NET] Using local IP as fallback: %s", ip)
		return ip
	}

	n.logger.Printf("[NET] Could not determine external IP")
	return "0.0.0.0"
}

// LocalInterfaces returns all non-loopback IPv4 addresses on the machine.
func (n *Network) LocalInterfaces() []string {
	var addrs []string

	ifaces, err := net.Interfaces()
	if err != nil {
		n.logger.Printf("[NET] Error enumerating interfaces: %v", err)
		return addrs
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		ifAddrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range ifAddrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil {
				addrs = append(addrs, ip.String())
			}
		}
	}

	return addrs
}

// primaryLocalIP returns the first non-loopback IPv4 address.
func (n *Network) primaryLocalIP() string {
	addrs := n.LocalInterfaces()
	if len(addrs) > 0 {
		return addrs[0]
	}
	return ""
}
