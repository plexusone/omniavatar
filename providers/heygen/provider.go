package heygen

import (
	"fmt"

	"github.com/plexusone/heygen-go/liveavatar"

	"github.com/plexusone/omniavatar-core/live"
)

// Config configures the HeyGen LiveAvatar provider.
type Config struct {
	// APIKey is the LiveAvatar API key.
	// Required.
	APIKey string

	// BaseURL is the LiveAvatar API base URL.
	// Default: https://api.liveavatar.com
	BaseURL string

	// AvatarID is the UUID of the avatar to use.
	// Use liveavatar.SandboxAvatarID for testing.
	// Required.
	AvatarID string

	// Sandbox enables sandbox mode (60s limit, no credits).
	// Recommended for development and testing.
	Sandbox bool

	// VideoQuality sets the avatar video quality.
	// Options: "very_high", "high", "medium", "low"
	// Default: "high"
	VideoQuality string
}

// Provider implements live.Provider for HeyGen LiveAvatar.
type Provider struct {
	client       *liveavatar.Client
	avatarID     string
	sandbox      bool
	videoQuality liveavatar.VideoQuality
}

// NewProvider creates a new HeyGen LiveAvatar provider.
func NewProvider(cfg Config) (*Provider, error) {
	if cfg.APIKey == "" {
		return nil, live.ErrInvalidConfig
	}
	if cfg.AvatarID == "" {
		return nil, live.ErrInvalidConfig
	}

	clientCfg := &liveavatar.Config{
		APIKey: cfg.APIKey,
	}
	if cfg.BaseURL != "" {
		clientCfg.BaseURL = cfg.BaseURL
	}

	client, err := liveavatar.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create LiveAvatar client: %w", err)
	}

	videoQuality := liveavatar.VideoQuality(cfg.VideoQuality)
	if videoQuality == "" {
		videoQuality = liveavatar.QualityHigh
	}

	return &Provider{
		client:       client,
		avatarID:     cfg.AvatarID,
		sandbox:      cfg.Sandbox,
		videoQuality: videoQuality,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "heygen"
}

// CreateSession creates a new HeyGen avatar session.
func (p *Provider) CreateSession(cfg live.SessionConfig) (live.Session, error) {
	audioConfig := cfg.AudioConfig
	if audioConfig.SampleRate == 0 {
		audioConfig = live.DefaultAudioConfig()
	}

	return NewSession(SessionConfig{
		Client:       p.client,
		AvatarID:     p.avatarID,
		Sandbox:      p.sandbox,
		VideoQuality: p.videoQuality,
		AudioConfig:  audioConfig,
	})
}

// Verify interface compliance at compile time.
var _ live.Provider = (*Provider)(nil)
