# Live Quick Start

Real-time streaming avatars: create a session, connect it to a LiveKit
room, and stream TTS audio for lip-sync playback.

## Prerequisites

- A provider API key (see [Providers](../providers/heygen.md) — note that
  HeyGen live uses the **LiveAvatar** key, not the HeyGen key)
- A LiveKit server and API credentials

## Create and Start a Session

```go
import (
    "github.com/plexusone/omniavatar"
    "github.com/plexusone/omniavatar-core/live"
    _ "github.com/plexusone/omniavatar/providers/all"
)

func main() {
    provider, err := omniavatar.GetLiveProvider("heygen",
        omniavatar.WithAPIKey(os.Getenv("LIVEAVATAR_API_KEY")),
        omniavatar.WithExtension("avatar_id", avatarID),
        omniavatar.WithExtension("sandbox", true)) // 60s limit, no credits
    if err != nil {
        log.Fatal(err)
    }

    session, err := provider.CreateSession(live.SessionConfig{
        AudioConfig: live.DefaultAudioConfig(),
    })
    if err != nil {
        log.Fatal(err)
    }

    err = session.Start(ctx, &omniavatar.LiveKitStartOptions{
        Room:             room,
        AgentIdentity:    "agent-123",
        LiveKitURL:       os.Getenv("LIVEKIT_URL"),
        LiveKitAPIKey:    os.Getenv("LIVEKIT_API_KEY"),
        LiveKitAPISecret: os.Getenv("LIVEKIT_API_SECRET"),
    })
    if err != nil {
        log.Fatal(err)
    }

    if err := session.WaitForJoin(ctx, 30*time.Second); err != nil {
        log.Fatal(err)
    }

    // Stream PCM16 audio frames from your TTS pipeline
    audioOut := session.AudioOutput()
    _ = audioOut.CaptureFrame(ctx, pcmFrame)
    _ = audioOut.Flush(ctx)

    defer session.Close(ctx)
}
```

## Session Lifecycle

```
1. Get Provider    → omniavatar.GetLiveProvider("heygen", opts...)
2. Create Session  → provider.CreateSession(cfg)
3. Start           → session.Start(ctx, &LiveKitStartOptions{...})
4. Wait for Join   → session.WaitForJoin(ctx, 30*time.Second)
5. Stream Audio    → session.AudioOutput().CaptureFrame(ctx, pcm)
6. Close           → session.Close(ctx)
```

## Audio Format

| Parameter | Default |
|-----------|---------|
| Sample Rate | 24000 Hz |
| Channels | 1 (mono) |
| Encoding | PCM16 (linear16) |

## Session Callbacks

```go
session.SetCallbacks(&live.SessionCallbacks{
    OnAvatarJoined: func(identity string) {
        log.Printf("Avatar joined: %s", identity)
    },
    OnPlaybackStarted: func() {
        log.Print("Avatar started speaking")
    },
    OnPlaybackFinished: func(position float64, interrupted bool) {
        log.Printf("Finished at %.2fs (interrupted: %v)", position, interrupted)
    },
    OnError: func(err error) {
        log.Printf("Avatar error: %v", err)
    },
})
```

## Next Steps

- [LiveKit Integration](../guides/livekit.md) — token generation and start options
- [Provider pages](../providers/heygen.md) — per-provider options and latency
