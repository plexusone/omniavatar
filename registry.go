package omniavatar

import (
	"fmt"
	"sync"

	"github.com/plexusone/omniavatar-core/avatar"
	"github.com/plexusone/omniavatar-core/registry"
)

// Priority constants for provider registration.
// Higher priority values override lower priority registrations.
const (
	// PriorityThin is the priority for thin (stdlib-only) provider implementations.
	// These have no external dependencies beyond the standard library.
	PriorityThin = 0

	// PriorityThick is the priority for thick (official SDK) provider implementations.
	// These use official provider SDKs for full feature support.
	PriorityThick = 10
)

// Re-export types from omniavatar-core/registry for convenience.
type (
	// ProviderConfig holds common configuration options for creating providers.
	ProviderConfig = registry.ProviderConfig

	// ProviderOption configures a ProviderConfig.
	ProviderOption = registry.ProviderOption

	// ProviderFactory creates a Provider from configuration.
	ProviderFactory = registry.ProviderFactory
)

// Re-export option functions from omniavatar-core/registry.
var (
	WithAPIKey    = registry.WithAPIKey
	WithBaseURL   = registry.WithBaseURL
	WithExtension = registry.WithExtension
)

// registeredProvider holds a factory with its priority.
type registeredProvider struct {
	factory  ProviderFactory
	priority int
}

var (
	avatarRegistry = make(map[string]registeredProvider)
	mu             sync.RWMutex
)

// RegisterAvatarProvider registers an avatar provider factory with the given name and priority.
// Higher priority values override lower priority registrations.
//
// Example:
//
//	// In omniavatar/providers/heygen/register.go (thick, priority 10)
//	func init() {
//	    omniavatar.RegisterAvatarProvider("heygen", NewProviderFromConfig, omniavatar.PriorityThick)
//	}
func RegisterAvatarProvider(name string, factory ProviderFactory, priority int) {
	mu.Lock()
	defer mu.Unlock()

	existing, ok := avatarRegistry[name]
	if !ok || priority >= existing.priority {
		avatarRegistry[name] = registeredProvider{
			factory:  factory,
			priority: priority,
		}
	}
}

// GetAvatarProvider creates an avatar provider instance from the registry.
// Returns an error if the provider is not registered or if creation fails.
//
// Example:
//
//	provider, err := omniavatar.GetAvatarProvider("heygen",
//	    omniavatar.WithAPIKey(os.Getenv("HEYGEN_API_KEY")),
//	    omniavatar.WithExtension("avatar_id", avatarID))
func GetAvatarProvider(name string, opts ...ProviderOption) (avatar.Provider, error) {
	mu.RLock()
	rp, ok := avatarRegistry[name]
	mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("avatar provider not registered: %s (available: %v)", name, ListAvatarProviders())
	}

	config := registry.ApplyOptions(opts...)
	return rp.factory(config)
}

// ListAvatarProviders returns a list of all registered avatar provider names.
func ListAvatarProviders() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(avatarRegistry))
	for name := range avatarRegistry {
		names = append(names, name)
	}
	return names
}

// HasAvatarProvider returns true if an avatar provider with the given name is registered.
func HasAvatarProvider(name string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := avatarRegistry[name]
	return ok
}

// GetAvatarProviderPriority returns the priority of the registered avatar provider.
// Returns -1 if the provider is not registered.
func GetAvatarProviderPriority(name string) int {
	mu.RLock()
	defer mu.RUnlock()

	if rp, ok := avatarRegistry[name]; ok {
		return rp.priority
	}
	return -1
}
