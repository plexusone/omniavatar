package tavus

import (
	"github.com/plexusone/omniavatar"
	"github.com/plexusone/omniavatar-core/avatar"
	"github.com/plexusone/omniavatar-core/registry"
)

func init() {
	omniavatar.RegisterAvatarProvider("tavus", NewProviderFromConfig, omniavatar.PriorityThick)
}

// NewProviderFromConfig creates a Tavus provider from registry config.
func NewProviderFromConfig(cfg registry.ProviderConfig) (avatar.Provider, error) {
	return NewProvider(Config{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		PalID:   cfg.GetString("pal_id", ""),
		FaceID:  cfg.GetString("face_id", ""),
	})
}
