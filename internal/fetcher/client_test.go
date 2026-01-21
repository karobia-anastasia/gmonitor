package fetcher

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// override HTTP client for testing
func withCustomClient(rt http.RoundTripper, testFunc func()) {
	defaultTransport := http.DefaultTransport
	http.DefaultTransport = rt
	defer func() { http.DefaultTransport = defaultTransport }()
	testFunc()
}

// mockRoundTripper is used to simulate HTTP responses
type mockRoundTripper struct {
	fn func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.fn(req)
}

func TestMakeGitHubRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test_token" {
			t.Errorf("expected Authorization header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	resp, err := makeGitHubRequest(server.URL, "test_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "success") {
		t.Errorf("expected success in response, got: %s", string(body))
	}
}

func TestMakeGitHubRequest_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	resp, err := makeGitHubRequest(server.URL, "")
	if err == nil {
		t.Fatalf("expected error for non-200 response")
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func TestMakeGitHubRequest_RequestCreationError(t *testing.T) {
	// invalid URL will trigger request creation error
	_, err := makeGitHubRequest("http://[::1]:namedport", "")
	if err == nil || !strings.Contains(err.Error(), "error creating request") {
		t.Errorf("expected request creation error, got: %v", err)
	}
}
