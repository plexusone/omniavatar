package omniavatar

import (
	lksdk "github.com/livekit/server-sdk-go/v2"

	"github.com/plexusone/omniavatar-core/live"
)

// LiveKitStartOptions contains LiveKit-specific start options for avatar sessions.
//
// This is passed to Session.Start() when integrating with LiveKit.
type LiveKitStartOptions struct {
	// Room is the LiveKit room the agent has joined.
	// Required.
	Room *lksdk.Room

	// AgentIdentity is the identity of the agent participant.
	// The avatar will publish tracks on behalf of this identity
	// using the lk.publish_on_behalf attribute.
	// Required.
	AgentIdentity string

	// LiveKitURL is the LiveKit server URL for the avatar to connect to.
	// This should match the URL the agent connected to.
	// Required.
	LiveKitURL string

	// LiveKitAPIKey is used to generate tokens for the avatar.
	// Required.
	LiveKitAPIKey string

	// LiveKitAPISecret is used to generate tokens for the avatar.
	// Required.
	LiveKitAPISecret string

	// Callbacks configures optional event callbacks.
	Callbacks *live.SessionCallbacks

	// AudioDestination is the audio output for streaming TTS audio.
	// If provided, the session will use this instead of creating its own.
	// Optional.
	AudioDestination live.AudioDestination
}

// Validate checks that all required fields are set.
func (o *LiveKitStartOptions) Validate() error {
	if o.Room == nil {
		return live.ErrInvalidConfig
	}
	if o.AgentIdentity == "" {
		return live.ErrInvalidConfig
	}
	if o.LiveKitURL == "" {
		return live.ErrInvalidConfig
	}
	if o.LiveKitAPIKey == "" {
		return live.ErrInvalidConfig
	}
	if o.LiveKitAPISecret == "" {
		return live.ErrInvalidConfig
	}
	return nil
}
