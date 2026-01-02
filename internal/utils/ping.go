package utils

import (
	"fmt"
	"net"
	"net/url"
	"time"
)

// PingService checks if a service is reachable at the given URL
func PingService(serviceURL string, timeout time.Duration) error {
	parsedURL, err := url.Parse(serviceURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	host := parsedURL.Hostname()
	port := parsedURL.Port()

	// Default ports if not specified
	if port == "" {
		switch parsedURL.Scheme {
		case "https":
			port = "443"
		case "http":
			port = "80"
		default:
			port = "80"
		}
	}

	address := net.JoinHostPort(host, port)

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	defer conn.Close()

	return nil
}

// PingAuthorizer checks if the Authorizer service is reachable
func PingAuthorizer(authzURL string) error {
	return PingService(authzURL, 1500*time.Millisecond)
}
