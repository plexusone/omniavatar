// Package heygen provides a HeyGen LiveAvatar provider for omniavatar.
//
// HeyGen LiveAvatar enables real-time AI avatar video generation using the
// LITE mode, which integrates with LiveKit for audio streaming and video output.
//
// # Quick Start
//
// Import this package to auto-register the heygen provider:
//
//	import (
//	    "github.com/plexusone/omniavatar"
//	    _ "github.com/plexusone/omniavatar/providers/heygen"
//	)
//
//	func main() {
//	    provider, err := omniavatar.GetLiveProvider("heygen",
//	        omniavatar.WithAPIKey(os.Getenv("HEYGEN_API_KEY")),
//	        omniavatar.WithExtension("avatar_id", avatarID),
//	        omniavatar.WithExtension("sandbox", true))
//	}
//
// # Configuration
//
// The heygen provider accepts the following extensions:
//
//   - avatar_id (string, required): UUID of the HeyGen avatar to use
//   - sandbox (bool, optional): Enable sandbox mode (60s limit, no credits)
//   - video_quality (string, optional): "very_high", "high", "medium", "low" (default: "high")
//
// # API Key
//
// Get your LiveAvatar API key from https://app.liveavatar.com/developers
// Note: This is different from the HeyGen video generation API key.
package heygen
