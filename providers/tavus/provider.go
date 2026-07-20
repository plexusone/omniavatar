package tavus

import (
	"github.com/plexusone/omniavatar-core/live"
)

// Config configures the Tavus avatar provider.
type Config struct {
	// APIKey is the Tavus API key.
	// Required.
	APIKey string

	// BaseURL is the Tavus API base URL.
	// Default: https://tavusapi.com
	BaseURL string

	// PalID is the PAL (Personalized AI Likeness) to use.
	// Default: stock avatar (DefaultPalID)
	PalID string

	// FaceID is an optional face override.
	FaceID string
}

// Provider implements live.Provider for Tavus.
type Provider struct {
	client *Client
	palID  string
	faceID string
}

// NewProvider creates a new Tavus avatar provider.
func NewProvider(cfg Config) (*Provider, error) {
	if cfg.APIKey == "" {
		return nil, live.ErrInvalidConfig
	}

	client, err := NewClient(ClientConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
	})
	if err != nil {
		return nil, err
	}

	palID := cfg.PalID
	if palID == "" {
		palID = DefaultPalID
	}

	return &Provider{
		client: client,
		palID:  palID,
		faceID: cfg.FaceID,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "tavus"
}

// CreateSession creates a new Tavus avatar session.
func (p *Provider) CreateSession(cfg live.SessionConfig) (live.Session, error) {
	audioConfig := cfg.AudioConfig
	if audioConfig.SampleRate == 0 {
		audioConfig = live.DefaultAudioConfig()
	}

	return NewSession(SessionConfig{
		Client:      p.client,
		PalID:       p.palID,
		FaceID:      p.faceID,
		AudioConfig: audioConfig,
	})
}

// Verify interface compliance at compile time.
var _ live.Provider = (*Provider)(nil)
