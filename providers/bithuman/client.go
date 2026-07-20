package bithuman

import (
	"context"
	"net/http"
	"time"

	bithumansdk "github.com/plexusone/bithuman-go"
	"github.com/plexusone/bithuman-go/api"

	"github.com/plexusone/omniavatar-core/live"
)

// Client wraps the bithuman-go SDK for avatar session management.
type Client struct {
	sdk *bithumansdk.Client
}

// ClientConfig configures the bitHuman API client.
type ClientConfig struct {
	// APIKey is the bitHuman API key.
	// Required.
	APIKey string

	// BaseURL is the bitHuman API base URL.
	// Default: https://api.bithuman.ai
	BaseURL string

	// HTTPClient is an optional custom HTTP client.
	// Default: http.Client with 30s timeout
	HTTPClient *http.Client
}

// NewClient creates a new bitHuman API client using the bithuman-go SDK.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, live.ErrInvalidConfig
	}

	var opts []bithumansdk.Option
	opts = append(opts, bithumansdk.WithAPIKey(cfg.APIKey))

	if cfg.BaseURL != "" {
		opts = append(opts, bithumansdk.WithBaseURL(cfg.BaseURL))
	}

	if cfg.HTTPClient != nil {
		opts = append(opts, bithumansdk.WithHTTPClient(cfg.HTTPClient))
	} else {
		opts = append(opts, bithumansdk.WithTimeout(30*time.Second))
	}

	sdk, err := bithumansdk.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return &Client{sdk: sdk}, nil
}

// CreateSessionRequest is the request to create a real-time session.
type CreateSessionRequest struct {
	// AgentID is the bitHuman agent to use for this session.
	// Required.
	AgentID string

	// LiveKitURL is the LiveKit WebSocket URL for the avatar to connect to.
	// Optional - if provided, avatar will join the specified LiveKit room.
	LiveKitURL string

	// LiveKitToken is the JWT token for the avatar to join the room.
	// Optional - if provided with LiveKitURL, avatar uses external LiveKit.
	LiveKitToken string
}

// CreateSessionResponse is the response from creating a session.
type CreateSessionResponse struct {
	// SessionID is the unique identifier for the session.
	SessionID string

	// LiveKitURL is the LiveKit WebSocket URL (if using bitHuman's LiveKit).
	LiveKitURL string

	// LiveKitToken is the JWT token for the avatar (if using bitHuman's LiveKit).
	LiveKitToken string
}

// CreateSession creates a new real-time session with the bitHuman API.
func (c *Client) CreateSession(ctx context.Context, req CreateSessionRequest) (*CreateSessionResponse, error) {
	if req.AgentID == "" {
		return nil, live.ErrInvalidConfig
	}

	apiReq := &api.CreateSessionRequest{
		AgentID: req.AgentID,
	}

	if req.LiveKitURL != "" {
		apiReq.LivekitURL = api.NewOptString(req.LiveKitURL)
	}
	if req.LiveKitToken != "" {
		apiReq.LivekitToken = api.NewOptString(req.LiveKitToken)
	}

	resp, err := c.sdk.Sessions().Create(ctx, apiReq)
	if err != nil {
		return nil, live.NewProviderError("bithuman", "create_session", err)
	}

	result := &CreateSessionResponse{
		SessionID: resp.ID,
	}

	if resp.LivekitURL.Set {
		result.LiveKitURL = resp.LivekitURL.Value
	}
	if resp.LivekitToken.Set {
		result.LiveKitToken = resp.LivekitToken.Value
	}

	return result, nil
}

// EndSession ends an active session.
func (c *Client) EndSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return live.ErrInvalidConfig
	}

	if err := c.sdk.Sessions().End(ctx, sessionID); err != nil {
		return live.NewProviderError("bithuman", "end_session", err)
	}

	return nil
}

// SDK returns the underlying bithuman-go SDK client for advanced usage.
func (c *Client) SDK() *bithumansdk.Client {
	return c.sdk
}
