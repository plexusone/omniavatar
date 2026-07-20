// Package tavus provides a Tavus Conversational Video provider for omniavatar.
//
// Tavus enables real-time AI avatar video generation that integrates with
// LiveKit for audio streaming and video output.
//
// # Quick Start
//
// Import this package to auto-register the tavus provider:
//
//	import (
//	    "github.com/plexusone/omniavatar"
//	    _ "github.com/plexusone/omniavatar/providers/tavus"
//	)
//
//	func main() {
//	    provider, err := omniavatar.GetLiveProvider("tavus",
//	        omniavatar.WithAPIKey(os.Getenv("TAVUS_API_KEY")),
//	        omniavatar.WithExtension("pal_id", palID))
//	}
//
// # Configuration
//
// The tavus provider accepts the following extensions:
//
//   - pal_id (string, optional): PAL (Personalized AI Likeness) ID. Defaults to stock avatar.
//   - face_id (string, optional): Face override ID.
//
// # API Key
//
// Get your Tavus API key from https://docs.tavus.io
package tavus
