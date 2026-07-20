package tavus

import (
	"context"
	"net/http"
	"time"

	tavussdk "github.com/plexusone/tavus-go"
	"github.com/plexusone/tavus-go/api"

	"github.com/plexusone/omniavatar-core/live"
)

// DefaultPalID is the stock Tavus PAL for testing.
const DefaultPalID = "pb87e71797da"

// Client wraps the tavus-go SDK for avatar session management.
type Client struct {
	sdk *tavussdk.Client
}

// ClientConfig configures the Tavus API client.
type ClientConfig struct {
	// APIKey is the Tavus API key.
	// Required.
	APIKey string

	// BaseURL is the Tavus API base URL.
	// Default: https://tavusapi.com
	BaseURL string

	// HTTPClient is an optional custom HTTP client.
	// Default: http.Client with 30s timeout
	HTTPClient *http.Client
}

// NewClient creates a new Tavus API client using the tavus-go SDK.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, live.ErrInvalidConfig
	}

	var opts []tavussdk.Option
	opts = append(opts, tavussdk.WithAPIKey(cfg.APIKey))

	if cfg.BaseURL != "" {
		opts = append(opts, tavussdk.WithBaseURL(cfg.BaseURL))
	}

	if cfg.HTTPClient != nil {
		opts = append(opts, tavussdk.WithHTTPClient(cfg.HTTPClient))
	} else {
		opts = append(opts, tavussdk.WithTimeout(30*time.Second))
	}

	sdk, err := tavussdk.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return &Client{sdk: sdk}, nil
}

// CreateConversationRequest is the request to create a conversation.
type CreateConversationRequest struct {
	// PalID is the PAL (Personalized AI Likeness) to use.
	// Required.
	PalID string

	// FaceID is an optional face override.
	FaceID string

	// LiveKitURL is the LiveKit WebSocket URL for the avatar to connect to.
	LiveKitURL string

	// LiveKitToken is the JWT token for the avatar to join the room.
	LiveKitToken string

	// ConversationName is an optional name for the conversation.
	// Auto-generated if not provided.
	ConversationName string
}

// CreateConversationResponse is the response from creating a conversation.
type CreateConversationResponse struct {
	// ConversationID is the unique identifier for the conversation.
	ConversationID string

	// ConversationURL is the URL for the conversation (if available).
	ConversationURL string
}

// CreateConversation creates a new conversation with the Tavus API.
func (c *Client) CreateConversation(ctx context.Context, req CreateConversationRequest) (*CreateConversationResponse, error) {
	palID := req.PalID
	if palID == "" {
		palID = DefaultPalID
	}

	// Build properties for LiveKit integration
	var properties api.OptConversationProperties
	properties.SetTo(api.ConversationProperties{
		LivekitWsURL:     api.NewOptString(req.LiveKitURL),
		LivekitRoomToken: api.NewOptString(req.LiveKitToken),
	})

	apiReq := &api.CreateConversationRequest{
		PalID:      palID,
		FaceID:     req.FaceID,
		Properties: properties,
	}

	if req.ConversationName != "" {
		apiReq.ConversationName = api.NewOptString(req.ConversationName)
	}

	resp, err := c.sdk.Conversations().Create(ctx, apiReq)
	if err != nil {
		return nil, live.NewProviderError("tavus", "create_conversation", err)
	}

	result := &CreateConversationResponse{
		ConversationID: resp.ConversationID.Value,
	}

	if resp.ConversationURL.Set {
		result.ConversationURL = resp.ConversationURL.Value.String()
	}

	return result, nil
}

// EndConversation ends an active conversation.
func (c *Client) EndConversation(ctx context.Context, conversationID string) error {
	if conversationID == "" {
		return live.ErrInvalidConfig
	}

	if err := c.sdk.Conversations().End(ctx, conversationID); err != nil {
		return live.NewProviderError("tavus", "end_conversation", err)
	}

	return nil
}

// SDK returns the underlying tavus-go SDK client for advanced usage.
func (c *Client) SDK() *tavussdk.Client {
	return c.sdk
}
