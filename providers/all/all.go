// Package all imports all avatar providers for side-effect registration.
//
// Import this package to register all available avatar providers:
//
//	import (
//	    "github.com/plexusone/omniavatar"
//	    _ "github.com/plexusone/omniavatar/providers/all"
//	)
//
//	func main() {
//	    // All providers are now available
//	    providers := omniavatar.ListAvatarProviders()
//	    // ["heygen", "tavus", "bithuman"]
//	}
package all

import (
	// HeyGen LiveAvatar provider
	_ "github.com/plexusone/omniavatar/providers/bithuman"
	// Tavus Conversational Video provider
	_ "github.com/plexusone/omniavatar/providers/heygen"
	// bitHuman Real-time Avatars provider
	_ "github.com/plexusone/omniavatar/providers/tavus"
)
