package bithuman

import (
	"github.com/plexusone/omniavatar"
	"github.com/plexusone/omniavatar-core/avatar"
	"github.com/plexusone/omniavatar-core/registry"
)

func init() {
	omniavatar.RegisterAvatarProvider("bithuman", NewProviderFromConfig, omniavatar.PriorityThick)
}

// NewProviderFromConfig creates a bitHuman provider from registry config.
func NewProviderFromConfig(cfg registry.ProviderConfig) (avatar.Provider, error) {
	return NewProvider(Config{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		AgentID: cfg.GetString("agent_id", ""),
	})
}
