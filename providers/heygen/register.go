package heygen

import (
	heygenrender "github.com/plexusone/heygen-go/omniavatar"

	"github.com/plexusone/omniavatar"
	"github.com/plexusone/omniavatar-core/live"
	"github.com/plexusone/omniavatar-core/registry"
)

func init() {
	omniavatar.RegisterLiveProvider("heygen", NewProviderFromConfig, omniavatar.PriorityThick)
	// The render provider lives in the heygen-go SDK (core-only, no LiveKit).
	omniavatar.RegisterRenderProvider("heygen", heygenrender.NewRenderProviderFromConfig, omniavatar.PriorityThick)
}

// NewProviderFromConfig creates a HeyGen live provider from registry config.
// The live provider uses the LiveAvatar API key (LIVEAVATAR_API_KEY).
func NewProviderFromConfig(cfg registry.ProviderConfig) (live.Provider, error) {
	return NewProvider(Config{
		APIKey:       cfg.APIKey,
		BaseURL:      cfg.BaseURL,
		AvatarID:     cfg.GetString("avatar_id", ""),
		Sandbox:      cfg.GetBool("sandbox", false),
		VideoQuality: cfg.GetString("video_quality", "high"),
	})
}
