package utils

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

// FetchSubscription fetches the subscription content from the given URL
func FetchSubscription(url string) (string, error) {
	// Create a custom HTTP client that ignores HTTPS certificate errors
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second, // Add timeout for requests
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}