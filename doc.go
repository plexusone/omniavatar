// Package omniavatar provides a unified, provider-agnostic interface for real-time AI avatars.
//
// This is the batteries-included package that imports all providers.
// For a minimal dependency footprint, use github.com/plexusone/omniavatar-core instead.
//
// # Quick Start
//
//	import (
//	    "github.com/plexusone/omniavatar"
//	    _ "github.com/plexusone/omniavatar/providers/all"
//	)
//
//	func main() {
//	    provider, err := omniavatar.GetAvatarProvider("heygen",
//	        omniavatar.WithAPIKey(os.Getenv("HEYGEN_API_KEY")),
//	        omniavatar.WithExtension("avatar_id", avatarID),
//	        omniavatar.WithExtension("sandbox", true))
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    session, err := provider.CreateSession(avatar.SessionConfig{
//	        AudioConfig: avatar.DefaultAudioConfig(),
//	    })
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Use session with LiveKit or other platform adapters
//	}
//
// # Available Providers
//
//   - heygen: HeyGen LiveAvatar (LITE mode)
//   - tavus: Tavus Conversational Video
//   - bithuman: bitHuman Real-time Avatars
//
// # Architecture
//
// omniavatar follows the same pattern as omnivoice:
//
//   - omniavatar-core: Core interfaces with no provider dependencies
//   - omniavatar: Provider implementations with auto-registration
//   - omniavatar/providers/all: Convenience import for all providers
package omniavatar
