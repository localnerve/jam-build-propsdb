package helpers

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// AssertStatus verifies the HTTP status code
func AssertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		t.Errorf("Expected status %d, got %d", expected, resp.StatusCode)
	}
}

// ParseJSON decodes the response body into the target
func ParseJSON(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	defer resp.Body.Close()

	if err := json.Unmarshal(body, target); err != nil {
		t.Fatalf("Failed to decode JSON: %v. Body: %s", err, string(body))
	}
}

// AssertNoContent verifies that the response body is empty (for 204s)
func AssertNoContent(t *testing.T, resp *http.Response) {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	defer resp.Body.Close()

	if len(body) > 0 {
		t.Errorf("Expected empty body for 204 No Content, got: %s", string(body))
	}
}
