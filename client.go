package supabasego

import (
	"net/http"
	"time"
)

// Client is the core Supabase API client.
type Client struct {
	BaseURL    string // e.g. https://<project>.supabase.co
	APIKey     string // Supabase anon or service key
	HTTPClient *http.Client
}

// Config holds configuration for the Supabase client.
type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration // Optional: HTTP timeout
}

// NewClient creates a new Supabase API client.
func NewClient(cfg Config) *Client {
	client := &http.Client{}
	if cfg.Timeout > 0 {
		client.Timeout = cfg.Timeout
	}
	return &Client{
		BaseURL:    cfg.BaseURL,
		APIKey:     cfg.APIKey,
		HTTPClient: client,
	}
}

// newRequest creates a new HTTP request with Supabase headers.
func (c *Client) newRequest(method, path string, body interface{}, jwtToken string) (*http.Request, error) {
	// Implementation will handle marshalling body, setting headers, etc.
	// To be filled in as CRUD and auth are implemented.
	return nil, nil
}

// Do sends an HTTP request and returns the response.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.HTTPClient.Do(req)
}
