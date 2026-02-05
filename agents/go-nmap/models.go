package main

import "encoding/xml"

// --- Inbound: scan request from ZEPSEC server (POST /scans) ---

// ScanRequest is received from the ZEPSEC Rails server when it dispatches
// a scan job to this agent. Fields arrive as query parameters.
type ScanRequest struct {
	ID      string `json:"id"      form:"id"`      // Sidekiq job ID (jid)
	Options string `json:"options" form:"options"` // Nmap command-line string
}

// --- Outbound: scan results posted back to ZEPSEC (POST /api/v1/ra_api) ---

// ScanResultPayload is the top-level JSON body sent to the ZEPSEC ra_api endpoint.
type ScanResultPayload struct {
	JID        string       `json:"jid"`
	ExternalIP string       `json:"externalip"`
	Hosts      []HostResult `json:"hosts"`
}

// HostResult represents a single scanned host in the result payload.
type HostResult struct {
	Addresses []Address    `json:"addresses"`
	StartTime int64        `json:"starttime"`
	EndTime   int64        `json:"endtime"`
	Ports     []PortResult `json:"ports,omitempty"`
}

// Address holds an IP address and its type for the result payload.
type Address struct {
	AddrType string `json:"addrtype"`
	Addr     string `json:"addr"`
}

// PortResult represents a single port finding.
type PortResult struct {
	ID       int          `json:"id"`
	Protocol string       `json:"protocol"`
	State    PortState    `json:"state"`
	Service  *ServiceInfo `json:"service,omitempty"`
}

// PortState holds the nmap state string.
type PortState struct {
	State string `json:"state"`
}

// ServiceInfo holds service detection results for a port.
type ServiceInfo struct {
	Name      string `json:"name"`
	Product   string `json:"product"`
	Version   string `json:"version"`
	ExtraInfo string `json:"extrainfo"`
}

// --- Nmap XML output structures ---

// NmapRun is the root element of nmap XML output.
type NmapRun struct {
	XMLName xml.Name   `xml:"nmaprun"`
	Hosts   []NmapHost `xml:"host"`
}

// NmapHost represents a <host> element in nmap XML.
type NmapHost struct {
	StartTime int64           `xml:"starttime,attr"`
	EndTime   int64           `xml:"endtime,attr"`
	Addresses []NmapAddress   `xml:"address"`
	Ports     NmapPorts       `xml:"ports"`
	Hostnames NmapHostnames   `xml:"hostnames"`
	Status    NmapHostStatus  `xml:"status"`
	OS        *NmapOS         `xml:"os"`
}

// NmapHostStatus represents the <status> element.
type NmapHostStatus struct {
	State  string `xml:"state,attr"`
	Reason string `xml:"reason,attr"`
}

// NmapAddress represents an <address> element.
type NmapAddress struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}

// NmapPorts wraps the <ports> element containing multiple <port> entries.
type NmapPorts struct {
	Ports []NmapPort `xml:"port"`
}

// NmapPort represents a <port> element.
type NmapPort struct {
	Protocol string       `xml:"protocol,attr"`
	PortID   int          `xml:"portid,attr"`
	State    NmapState    `xml:"state"`
	Service  *NmapService `xml:"service"`
}

// NmapState represents the <state> element within a port.
type NmapState struct {
	State  string `xml:"state,attr"`
	Reason string `xml:"reason,attr"`
}

// NmapService represents the <service> element within a port.
type NmapService struct {
	Name      string `xml:"name,attr"`
	Product   string `xml:"product,attr"`
	Version   string `xml:"version,attr"`
	ExtraInfo string `xml:"extrainfo,attr"`
	Method    string `xml:"method,attr"`
	Conf      string `xml:"conf,attr"`
}

// NmapHostnames wraps hostname entries.
type NmapHostnames struct {
	Hostnames []NmapHostname `xml:"hostname"`
}

// NmapHostname represents a <hostname> element.
type NmapHostname struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

// NmapOS wraps OS detection results.
type NmapOS struct {
	OSMatches []NmapOSMatch `xml:"osmatch"`
}

// NmapOSMatch represents an <osmatch> element.
type NmapOSMatch struct {
	Name     string `xml:"name,attr"`
	Accuracy string `xml:"accuracy,attr"`
}

// --- Internal types ---

// ScanJob tracks an in-flight scan.
type ScanJob struct {
	JID     string
	Options string
	Status  string // "running", "completed", "failed"
}

// ScheduledScan defines an autonomous periodic scan.
type ScheduledScan struct {
	Name     string `yaml:"name"`
	Targets  string `yaml:"targets"`
	Options  string `yaml:"options"`
	Interval string `yaml:"interval"`
}
