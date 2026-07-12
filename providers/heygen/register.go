package heygen

import (
	"github.com/plexusone/omniavatar"
	"github.com/plexusone/omniavatar-core/avatar"
	"github.com/plexusone/omniavatar-core/registry"
)

func init() {
	omniavatar.RegisterAvatarProvider("heygen", NewProviderFromConfig, omniavatar.PriorityThick)
}

// NewProviderFromConfig creates a HeyGen provider from registry config.
func NewProviderFromConfig(cfg registry.ProviderConfig) (avatar.Provider, error) {
	return NewProvider(Config{
		APIKey:       cfg.APIKey,
		BaseURL:      cfg.BaseURL,
		AvatarID:     cfg.GetString("avatar_id", ""),
		Sandbox:      cfg.GetBool("sandbox", false),
		VideoQuality: cfg.GetString("video_quality", "high"),
	})
}
