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

// KratosClient communicates with the Ory Kratos identity server.
type KratosClient struct {
	publicURL string
	adminURL  string
	http      *http.Client
}

// KratosSession represents a validated session from Kratos.
type KratosSession struct {
	ID       string         `json:"id"`
	Active   bool           `json:"active"`
	Identity KratosIdentity `json:"identity"`
}

// KratosIdentity represents the identity within a session.
type KratosIdentity struct {
	ID     string                 `json:"id"`
	Traits map[string]interface{} `json:"traits"`
}

// NewKratosClient creates a new Kratos API client.
func NewKratosClient(cfg config.KratosConfig) *KratosClient {
	return &KratosClient{
		publicURL: cfg.PublicURL,
		adminURL:  cfg.AdminURL,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ValidateSession checks a session token or cookie against Kratos and returns
// the identity information if the session is valid.
func (c *KratosClient) ValidateSession(ctx context.Context, sessionToken string) (*KratosSession, error) {
	url := c.publicURL + "/sessions/whoami"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating whoami request: %w", err)
	}

	req.Header.Set("X-Session-Token", sessionToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling whoami: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("session is not valid")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	var session KratosSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("decoding session: %w", err)
	}

	if !session.Active {
		return nil, fmt.Errorf("session is inactive")
	}

	return &session, nil
}

// GetEmail extracts the email from a Kratos identity's traits.
func (s *KratosSession) GetEmail() string {
	if email, ok := s.Identity.Traits["email"].(string); ok {
		return email
	}
	return ""
}

// GetDisplayName extracts the display name from a Kratos identity's traits.
func (s *KratosSession) GetDisplayName() string {
	if name, ok := s.Identity.Traits["name"].(string); ok {
		return name
	}
	return ""
}
