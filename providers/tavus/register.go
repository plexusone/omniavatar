package tavus

import (
	tavusrender "github.com/plexusone/tavus-go/omniavatar"

	"github.com/plexusone/omniavatar"
	"github.com/plexusone/omniavatar-core/live"
	"github.com/plexusone/omniavatar-core/registry"
)

func init() {
	omniavatar.RegisterLiveProvider("tavus", NewProviderFromConfig, omniavatar.PriorityThick)
	// The render provider lives in the tavus-go SDK (core-only, no LiveKit).
	omniavatar.RegisterRenderProvider("tavus", tavusrender.NewRenderProviderFromConfig, omniavatar.PriorityThick)
}

// NewProviderFromConfig creates a Tavus live provider from registry config.
func NewProviderFromConfig(cfg registry.ProviderConfig) (live.Provider, error) {
	return NewProvider(Config{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		PalID:   cfg.GetString("pal_id", ""),
		FaceID:  cfg.GetString("face_id", ""),
	})
}
