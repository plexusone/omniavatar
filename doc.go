// Package omniavatar provides a unified, provider-agnostic interface for AI avatars.
//
// This is the batteries-included package that imports all providers.
// For a minimal dependency footprint, use github.com/plexusone/omniavatar-core instead.
//
// Two surfaces are supported:
//
//   - live: real-time streaming avatar sessions (LiveKit rooms, PCM audio
//     streaming for lip-sync) for conversational agents
//   - render: asynchronous batch avatar video generation (narration audio
//     in, talking-head MP4 out) for offline pipelines such as
//     presentation videos
//
// # Quick Start (live)
//
//	import (
//	    "github.com/plexusone/omniavatar"
//	    _ "github.com/plexusone/omniavatar/providers/all"
//	)
//
//	func main() {
//	    provider, err := omniavatar.GetLiveProvider("heygen",
//	        omniavatar.WithAPIKey(os.Getenv("LIVEAVATAR_API_KEY")),
//	        omniavatar.WithExtension("avatar_id", avatarID),
//	        omniavatar.WithExtension("sandbox", true))
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    session, err := provider.CreateSession(live.SessionConfig{
//	        AudioConfig: live.DefaultAudioConfig(),
//	    })
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Use session with LiveKit or other platform adapters
//	}
//
// # Quick Start (render)
//
//	provider, err := omniavatar.GetRenderProvider("heygen",
//	    omniavatar.WithAPIKey(os.Getenv("HEYGEN_API_KEY")))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	job, err := provider.Generate(ctx, render.GenerateRequest{
//	    AvatarID: avatarID,
//	    AudioURL: narrationURL,
//	})
//	status, err := render.Wait(ctx, provider, job.ID, 5*time.Second)
//	err = provider.Download(ctx, job.ID, outFile)
//
// # Available Providers
//
//   - heygen: HeyGen LiveAvatar (live) and HeyGen Video Generation (render)
//   - tavus: Tavus Conversational Video (live) and Video Generation (render)
//   - bithuman: bitHuman Real-time Avatars (live) and Video Generation
//     (render, with audio upload support)
//
// # Architecture
//
// omniavatar follows the same pattern as omnivoice:
//
//   - omniavatar-core: Core interfaces (live, render) with no provider dependencies
//   - omniavatar: Provider implementations with auto-registration
//   - omniavatar/providers/all: Convenience import for all providers
package omniavatar
