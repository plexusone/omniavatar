package bithuman

import (
	bithumanrender "github.com/plexusone/bithuman-go/omniavatar"

	"github.com/plexusone/omniavatar"
	"github.com/plexusone/omniavatar-core/live"
	"github.com/plexusone/omniavatar-core/registry"
)

func init() {
	omniavatar.RegisterLiveProvider("bithuman", NewProviderFromConfig, omniavatar.PriorityThick)
	// The render provider lives in the bithuman-go SDK (core-only, no LiveKit).
	omniavatar.RegisterRenderProvider("bithuman", bithumanrender.NewRenderProviderFromConfig, omniavatar.PriorityThick)
}

// NewProviderFromConfig creates a bitHuman live provider from registry config.
func NewProviderFromConfig(cfg registry.ProviderConfig) (live.Provider, error) {
	return NewProvider(Config{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		AgentID: cfg.GetString("agent_id", ""),
	})
}
