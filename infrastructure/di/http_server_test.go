package di_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/config"
	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/di"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	cfg := &config.Config{
		HTTPAddr:        ":0",
		MaxLimit:        10000,
		RateLimitPerSec: 1000,
		RateLimitBurst:  1000,
		LogLevel:        "error",
	}

	c := di.NewContainer(context.Background(), cfg)
	ts := httptest.NewServer(c.HTTPHandler())

	t.Cleanup(ts.Close)

	return ts
}

func get(t *testing.T, url string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}

	return resp
}

func TestIntegration_GenerateThenStats(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t)

	resp := get(t, ts.URL+"/fizzbuzz?int1=3&int2=5&limit=5&str1=fizz&str2=buzz")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("generate status = %d, want 200", resp.StatusCode)
	}

	var arr []string
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		t.Fatalf("decode generate body: %v", err)
	}

	if len(arr) != 5 {
		t.Fatalf("got %d items, want 5", len(arr))
	}

	statsResp := get(t, ts.URL+"/fizzbuzz/stats")
	defer func() { _ = statsResp.Body.Close() }()

	if statsResp.StatusCode != http.StatusOK {
		t.Fatalf("stats status = %d, want 200", statsResp.StatusCode)
	}
}

func TestIntegration_StatsEmpty404(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t)

	resp := get(t, ts.URL+"/fizzbuzz/stats")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("stats status = %d, want 404", resp.StatusCode)
	}
}
