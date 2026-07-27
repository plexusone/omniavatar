// Package liveportraitjoyvasa registers the local LivePortrait + JoyVASA
// render provider with the omniavatar registry.
//
// This provider generates audio-driven talking-head videos on Apple Silicon
// without requiring a cloud API. It requires a running Python gRPC server.
//
// Usage:
//
//	import (
//	    "github.com/plexusone/omniavatar"
//	    _ "github.com/plexusone/omniavatar/providers/liveportrait-joyvasa"
//	)
//
//	provider, err := omniavatar.GetRenderProvider("liveportrait-joyvasa")
//
// The provider connects to a Unix socket at /tmp/omniavatar-liveportrait-joyvasa.sock
// by default. Override with the "endpoint" extension:
//
//	provider, err := omniavatar.GetRenderProvider("liveportrait-joyvasa",
//	    omniavatar.WithExtension("endpoint", "unix:///custom/path.sock"),
//	)
package liveportraitjoyvasa

import (
	"github.com/plexusone/omniavatar"
	lp "github.com/plexusone/omniavatar-core/providers/liveportrait-joyvasa"
)

func init() {
	// Local providers don't need an API key, so they register at PriorityThick
	// to be available alongside cloud providers.
	omniavatar.RegisterRenderProvider("liveportrait-joyvasa", lp.NewFromConfig, omniavatar.PriorityThick)
}
