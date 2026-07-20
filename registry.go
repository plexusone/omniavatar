package omniavatar

import (
	"fmt"
	"sync"

	"github.com/plexusone/omniavatar-core/live"
	"github.com/plexusone/omniavatar-core/registry"
	"github.com/plexusone/omniavatar-core/render"
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

	// LiveProviderFactory creates a live.Provider from configuration.
	LiveProviderFactory = registry.LiveProviderFactory

	// RenderProviderFactory creates a render.Provider from configuration.
	RenderProviderFactory = registry.RenderProviderFactory
)

// Re-export option functions from omniavatar-core/registry.
var (
	WithAPIKey    = registry.WithAPIKey
	WithBaseURL   = registry.WithBaseURL
	WithExtension = registry.WithExtension
)

// registeredLiveProvider holds a live factory with its priority.
type registeredLiveProvider struct {
	factory  LiveProviderFactory
	priority int
}

// registeredRenderProvider holds a render factory with its priority.
type registeredRenderProvider struct {
	factory  RenderProviderFactory
	priority int
}

var (
	liveRegistry   = make(map[string]registeredLiveProvider)
	renderRegistry = make(map[string]registeredRenderProvider)
	mu             sync.RWMutex
)

// RegisterLiveProvider registers a live (real-time session) provider factory
// with the given name and priority. Higher priority values override lower
// priority registrations.
//
// Example:
//
//	// In omniavatar/providers/heygen/register.go (thick, priority 10)
//	func init() {
//	    omniavatar.RegisterLiveProvider("heygen", NewProviderFromConfig, omniavatar.PriorityThick)
//	}
func RegisterLiveProvider(name string, factory LiveProviderFactory, priority int) {
	mu.Lock()
	defer mu.Unlock()

	existing, ok := liveRegistry[name]
	if !ok || priority >= existing.priority {
		liveRegistry[name] = registeredLiveProvider{
			factory:  factory,
			priority: priority,
		}
	}
}

// GetLiveProvider creates a live provider instance from the registry.
// Returns an error if the provider is not registered or if creation fails.
//
// Example:
//
//	provider, err := omniavatar.GetLiveProvider("heygen",
//	    omniavatar.WithAPIKey(os.Getenv("LIVEAVATAR_API_KEY")),
//	    omniavatar.WithExtension("avatar_id", avatarID))
func GetLiveProvider(name string, opts ...ProviderOption) (live.Provider, error) {
	mu.RLock()
	rp, ok := liveRegistry[name]
	mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("live provider not registered: %s (available: %v)", name, ListLiveProviders())
	}

	config := registry.ApplyOptions(opts...)
	return rp.factory(config)
}

// ListLiveProviders returns a list of all registered live provider names.
func ListLiveProviders() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(liveRegistry))
	for name := range liveRegistry {
		names = append(names, name)
	}
	return names
}

// HasLiveProvider returns true if a live provider with the given name is registered.
func HasLiveProvider(name string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := liveRegistry[name]
	return ok
}

// GetLiveProviderPriority returns the priority of the registered live provider.
// Returns -1 if the provider is not registered.
func GetLiveProviderPriority(name string) int {
	mu.RLock()
	defer mu.RUnlock()

	if rp, ok := liveRegistry[name]; ok {
		return rp.priority
	}
	return -1
}

// RegisterRenderProvider registers a render (batch generation) provider
// factory with the given name and priority. Higher priority values override
// lower priority registrations.
//
// Example:
//
//	// In omniavatar/providers/heygen/register.go (thick, priority 10)
//	func init() {
//	    omniavatar.RegisterRenderProvider("heygen", NewRenderProviderFromConfig, omniavatar.PriorityThick)
//	}
func RegisterRenderProvider(name string, factory RenderProviderFactory, priority int) {
	mu.Lock()
	defer mu.Unlock()

	existing, ok := renderRegistry[name]
	if !ok || priority >= existing.priority {
		renderRegistry[name] = registeredRenderProvider{
			factory:  factory,
			priority: priority,
		}
	}
}

// GetRenderProvider creates a render provider instance from the registry.
// Returns an error if the provider is not registered or if creation fails.
//
// Example:
//
//	provider, err := omniavatar.GetRenderProvider("heygen",
//	    omniavatar.WithAPIKey(os.Getenv("HEYGEN_API_KEY")))
func GetRenderProvider(name string, opts ...ProviderOption) (render.Provider, error) {
	mu.RLock()
	rp, ok := renderRegistry[name]
	mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("render provider not registered: %s (available: %v)", name, ListRenderProviders())
	}

	config := registry.ApplyOptions(opts...)
	return rp.factory(config)
}

// ListRenderProviders returns a list of all registered render provider names.
func ListRenderProviders() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(renderRegistry))
	for name := range renderRegistry {
		names = append(names, name)
	}
	return names
}

// HasRenderProvider returns true if a render provider with the given name is registered.
func HasRenderProvider(name string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := renderRegistry[name]
	return ok
}

// GetRenderProviderPriority returns the priority of the registered render provider.
// Returns -1 if the provider is not registered.
func GetRenderProviderPriority(name string) int {
	mu.RLock()
	defer mu.RUnlock()

	if rp, ok := renderRegistry[name]; ok {
		return rp.priority
	}
	return -1
}
