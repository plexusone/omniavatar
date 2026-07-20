package bithuman

import (
	"github.com/plexusone/omniavatar-core/live"
)

// Config configures the bitHuman avatar provider.
type Config struct {
	// APIKey is the bitHuman API key.
	// Required.
	APIKey string

	// BaseURL is the bitHuman API base URL.
	// Default: https://api.bithuman.ai
	BaseURL string

	// AgentID is the bitHuman agent to use for sessions.
	// Required.
	AgentID string
}

// Provider implements live.Provider for bitHuman.
type Provider struct {
	client  *Client
	agentID string
}

// NewProvider creates a new bitHuman avatar provider.
func NewProvider(cfg Config) (*Provider, error) {
	if cfg.APIKey == "" {
		return nil, live.ErrInvalidConfig
	}
	if cfg.AgentID == "" {
		return nil, live.ErrInvalidConfig
	}

	client, err := NewClient(ClientConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
	})
	if err != nil {
		return nil, err
	}

	return &Provider{
		client:  client,
		agentID: cfg.AgentID,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "bithuman"
}

// CreateSession creates a new bitHuman avatar session.
func (p *Provider) CreateSession(cfg live.SessionConfig) (live.Session, error) {
	audioConfig := cfg.AudioConfig
	if audioConfig.SampleRate == 0 {
		audioConfig = live.DefaultAudioConfig()
	}

	return NewSession(SessionConfig{
		Client:      p.client,
		AgentID:     p.agentID,
		AudioConfig: audioConfig,
	})
}

// Verify interface compliance at compile time.
var _ live.Provider = (*Provider)(nil)
