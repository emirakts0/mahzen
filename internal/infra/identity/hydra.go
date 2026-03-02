package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/emirakts0/mahzen/internal/config"
)

// HydraClient communicates with the Ory Hydra OAuth2 server.
type HydraClient struct {
	publicURL string
	adminURL  string
	http      *http.Client
}

// TokenIntrospection represents the response from Hydra's token introspection.
type TokenIntrospection struct {
	Active   bool   `json:"active"`
	Sub      string `json:"sub"`
	ClientID string `json:"client_id"`
	Scope    string `json:"scope"`
	Exp      int64  `json:"exp"`
}

// NewHydraClient creates a new Hydra API client.
func NewHydraClient(cfg config.HydraConfig) *HydraClient {
	return &HydraClient{
		publicURL: cfg.PublicURL,
		adminURL:  cfg.AdminURL,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IntrospectToken validates an OAuth2 access token with Hydra.
func (c *HydraClient) IntrospectToken(ctx context.Context, token string) (*TokenIntrospection, error) {
	url := c.adminURL + "/admin/oauth2/introspect"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating introspect request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	q := req.URL.Query()
	q.Set("token", token)
	req.URL.RawQuery = q.Encode()

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling introspect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	var result TokenIntrospection
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding introspection: %w", err)
	}

	return &result, nil
}
