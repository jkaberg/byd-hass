package netutil

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ReliableDNSTransport creates an HTTP transport with system DNS resolution
// TLS certificate verification is disabled by default for Android/Termux compatibility
func NewReliableDNSTransport(logger *logrus.Logger) *http.Transport {
	// Create transport with standard dialer - no custom DNS to avoid system call issues
	transport := &http.Transport{
		DialContext:           createDialContext(logger),
		TLSClientConfig:       getTLSConfig(logger),
		TLSHandshakeTimeout:   10 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
	}

	return transport
}

func createDialContext(logger *logrus.Logger) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		// Log the connection attempt
		if isLocalOrPrivateHost(host) {
			logger.WithField("host", host).Debug("Connecting to local/private host using system DNS")
		} else {
			logger.WithField("host", host).Debug("Connecting to external host using system DNS")
		}

		// Use system DNS for all connections to avoid system call compatibility issues
		dialer := net.Dialer{}
		return dialer.DialContext(ctx, network, addr)
	}
}

// isLocalOrPrivateHost checks if a hostname is localhost or a private network address
func isLocalOrPrivateHost(host string) bool {
	// Check for localhost variations
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}

	// Check for localhost-like names
	if strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".localhost") {
		return true
	}

	// Try to parse as IP address
	ip := net.ParseIP(host)
	if ip == nil {
		// Not an IP address, check if it's a private domain
		// Consider domains like "homeassistant.local", "router.local", etc. as private
		if strings.Contains(host, ".local") || strings.Contains(host, ".lan") {
			return true
		}
		return false // External domain name
	}

	// Check if IP is private
	return isPrivateIP(ip)
}

// isPrivateIP checks if an IP address is in a private network range
func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}

	// Private IPv4 ranges
	private := []string{
		"10.0.0.0/8",     // Class A private
		"172.16.0.0/12",  // Class B private
		"192.168.0.0/16", // Class C private
		"169.254.0.0/16", // Link-local
	}

	for _, cidr := range private {
		_, network, _ := net.ParseCIDR(cidr)
		if network != nil && network.Contains(ip) {
			return true
		}
	}

	// Private IPv6 ranges
	if ip.To4() == nil { // IPv6
		// Check for IPv6 private ranges
		if strings.HasPrefix(ip.String(), "fc") || strings.HasPrefix(ip.String(), "fd") {
			return true // Unique local addresses
		}
		if strings.HasPrefix(ip.String(), "fe80:") {
			return true // Link-local
		}
	}

	return false
}

// getTLSConfig creates a tls.Config with certificate verification disabled
// This is the default behavior for Android/Termux compatibility
func getTLSConfig(logger *logrus.Logger) *tls.Config {
	logger.Debug("TLS certificate verification is disabled for Android/Termux compatibility")

	return &tls.Config{
		InsecureSkipVerify: true,             // Always skip certificate verification
		MinVersion:         tls.VersionTLS12, // Set minimum TLS version for security
	}
}

// NewHTTPClientWithReliableDNS creates an HTTP client with system DNS resolution
func NewHTTPClientWithReliableDNS(timeout time.Duration, logger *logrus.Logger) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: NewReliableDNSTransport(logger),
	}
}
