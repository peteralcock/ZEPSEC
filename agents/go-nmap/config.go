package main

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config holds all agent configuration.
type Config struct {
	Agent  AgentConfig  `yaml:"agent"`
	Server ServerConfig `yaml:"server"`
	Nmap   NmapConfig   `yaml:"nmap"`
	Scans  []ScheduledScan `yaml:"scheduled_scans"`
}

// AgentConfig defines the HTTP listener for incoming scan requests.
type AgentConfig struct {
	ListenAddr string `yaml:"listen_addr"` // e.g. "0.0.0.0:8000"
	Secret     string `yaml:"secret"`      // Bearer token the ZEPSEC server uses
	TLSCert    string `yaml:"tls_cert"`    // Path to TLS certificate (optional)
	TLSKey     string `yaml:"tls_key"`     // Path to TLS key (optional)
}

// ServerConfig defines connection to the ZEPSEC Rails server.
type ServerConfig struct {
	URL       string `yaml:"url"`        // e.g. "https://zepsec.example.com"
	APIToken  string `yaml:"api_token"`  // User auth_token for /api/v1/ra_api
	VerifyTLS bool   `yaml:"verify_tls"` // Whether to verify server TLS certificate
}

// NmapConfig defines nmap binary settings.
type NmapConfig struct {
	BinaryPath string `yaml:"binary_path"` // Path to nmap binary, default: "nmap"
	UseSudo    bool   `yaml:"use_sudo"`    // Run nmap with sudo
	TempDir    string `yaml:"temp_dir"`    // Directory for temp XML files, default: "/tmp/zepsec-nmap"
}

// LoadConfig reads a YAML config file and applies environment variable overrides.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		Agent: AgentConfig{
			ListenAddr: "0.0.0.0:8000",
		},
		Server: ServerConfig{
			VerifyTLS: true,
		},
		Nmap: NmapConfig{
			BinaryPath: "nmap",
			UseSudo:    true,
			TempDir:    "/tmp/zepsec-nmap",
		},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading config %s: %w", path, err)
		}
		// File not found — use defaults + env overrides only
	} else {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config %s: %w", path, err)
		}
	}

	applyEnvOverrides(cfg)

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("ZEPSEC_LISTEN_ADDR"); v != "" {
		cfg.Agent.ListenAddr = v
	}
	if v := os.Getenv("ZEPSEC_AGENT_SECRET"); v != "" {
		cfg.Agent.Secret = v
	}
	if v := os.Getenv("ZEPSEC_TLS_CERT"); v != "" {
		cfg.Agent.TLSCert = v
	}
	if v := os.Getenv("ZEPSEC_TLS_KEY"); v != "" {
		cfg.Agent.TLSKey = v
	}
	if v := os.Getenv("ZEPSEC_SERVER_URL"); v != "" {
		cfg.Server.URL = v
	}
	if v := os.Getenv("ZEPSEC_API_TOKEN"); v != "" {
		cfg.Server.APIToken = v
	}
	if v := os.Getenv("ZEPSEC_VERIFY_TLS"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Server.VerifyTLS = b
		}
	}
	if v := os.Getenv("ZEPSEC_NMAP_PATH"); v != "" {
		cfg.Nmap.BinaryPath = v
	}
	if v := os.Getenv("ZEPSEC_NMAP_SUDO"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Nmap.UseSudo = b
		}
	}
	if v := os.Getenv("ZEPSEC_NMAP_TEMP_DIR"); v != "" {
		cfg.Nmap.TempDir = v
	}
}

func validateConfig(cfg *Config) error {
	if cfg.Agent.Secret == "" {
		return fmt.Errorf("agent secret is required (set agent.secret in config or ZEPSEC_AGENT_SECRET env)")
	}
	if cfg.Server.URL == "" {
		return fmt.Errorf("server URL is required (set server.url in config or ZEPSEC_SERVER_URL env)")
	}
	if cfg.Server.APIToken == "" {
		return fmt.Errorf("server API token is required (set server.api_token in config or ZEPSEC_API_TOKEN env)")
	}
	return nil
}
