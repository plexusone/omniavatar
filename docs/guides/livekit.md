# LiveKit Integration

The live surface connects avatar participants to LiveKit rooms. The
avatar provider publishes video (and lip-synced audio) into the room as
its own participant.

## Token Generation

Generate tokens for avatar participants to join rooms:

```go
token, err := omniavatar.GenerateAvatarToken(omniavatar.TokenOptions{
    APIKey:        os.Getenv("LIVEKIT_API_KEY"),
    APISecret:     os.Getenv("LIVEKIT_API_SECRET"),
    RoomName:      "my-room",
    Identity:      "avatar-heygen-abc123",
    Provider:      "heygen",
    AgentIdentity: "agent-123",
    TTL:           time.Hour,
})
```

## Start Options

`Session.Start` takes platform-specific options; for LiveKit pass
`*omniavatar.LiveKitStartOptions`:

```go
type LiveKitStartOptions struct {
    Room             *lksdk.Room  // LiveKit room reference
    AgentIdentity    string       // Agent's participant identity
    LiveKitURL       string       // LiveKit server URL
    LiveKitAPIKey    string       // API key for token generation
    LiveKitAPISecret string       // API secret for token generation
}
```

## Streaming Audio

After `WaitForJoin` succeeds, stream PCM16 frames from your TTS pipeline
to the avatar for lip-sync playback:

```go
audioOut := session.AudioOutput()

// 20ms frames at the configured sample rate (default 24kHz mono PCM16)
err := audioOut.CaptureFrame(ctx, frame)

// Mark end of utterance
err = audioOut.Flush(ctx)

// Interrupt playback (e.g., user barge-in)
err = audioOut.ClearBuffer(ctx)
```

## Provider Latency

| Provider | Join-to-speech latency (approx.) |
|----------|----------------------------------|
| bitHuman | ~200ms |
| Tavus | ~300ms |
| HeyGen | ~500ms |
