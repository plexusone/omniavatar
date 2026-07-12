// Package bithuman provides a bitHuman Real-time Avatars provider for omniavatar.
//
// bitHuman enables real-time AI avatar video generation that integrates with
// LiveKit for audio streaming and video output.
//
// # Quick Start
//
// Import this package to auto-register the bithuman provider:
//
//	import (
//	    "github.com/plexusone/omniavatar"
//	    _ "github.com/plexusone/omniavatar/providers/bithuman"
//	)
//
//	func main() {
//	    provider, err := omniavatar.GetAvatarProvider("bithuman",
//	        omniavatar.WithAPIKey(os.Getenv("BITHUMAN_API_KEY")),
//	        omniavatar.WithExtension("agent_id", agentID))
//	}
//
// # Configuration
//
// The bithuman provider accepts the following extensions:
//
//   - agent_id (string, required): bitHuman agent ID to use for the session.
//
// # API Key
//
// Get your bitHuman API key from https://www.bithuman.ai
package bithuman
