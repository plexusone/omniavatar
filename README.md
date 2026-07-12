# OmniAvatar

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/omniavatar/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/omniavatar/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/omniavatar/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/omniavatar/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/omniavatar/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/omniavatar/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/omniavatar
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/omniavatar
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/omniavatar/blob/main/LICENSE

Batteries-included package for real-time AI avatars. Provides provider implementations for HeyGen, Tavus, and bitHuman.

For core interfaces only (no provider dependencies), see [omniavatar-core](https://github.com/plexusone/omniavatar-core).

## Quick Start

```go
import (
    "github.com/plexusone/omniavatar"
    _ "github.com/plexusone/omniavatar/providers/all"
)

func main() {
    provider, err := omniavatar.GetAvatarProvider("heygen",
        omniavatar.WithAPIKey(os.Getenv("HEYGEN_API_KEY")),
        omniavatar.WithExtension("avatar_id", avatarID),
        omniavatar.WithExtension("sandbox", true))
    if err != nil {
        log.Fatal(err)
    }

    session, err := provider.CreateSession(avatar.SessionConfig{
        AudioConfig: avatar.DefaultAudioConfig(),
    })
    if err != nil {
        log.Fatal(err)
    }

    // Start with LiveKit
    err = session.Start(ctx, &omniavatar.LiveKitStartOptions{
        Room:             room,
        AgentIdentity:    "agent-123",
        LiveKitURL:       os.Getenv("LIVEKIT_URL"),
        LiveKitAPIKey:    os.Getenv("LIVEKIT_API_KEY"),
        LiveKitAPISecret: os.Getenv("LIVEKIT_API_SECRET"),
    })
}
```

## Architecture

```
omniavatar-core/              # Core interfaces (no provider deps)
├── avatar/
│   ├── provider.go           # Provider interface
│   ├── session.go            # Session interface
│   └── audio.go              # AudioDestination interface
└── registry/
    └── registry.go           # Factory types

omniavatar/                   # Provider implementations (this package)
├── registry.go               # Global registry
├── token.go                  # LiveKit token generation
├── start_options.go          # LiveKitStartOptions
└── providers/
    ├── heygen/               # HeyGen LiveAvatar
    ├── tavus/                # Tavus Conversational Video
    ├── bithuman/             # bitHuman Real-time Avatars
    └── all/                  # Convenience import
```

## Provider Registry

### Priority System

Providers register with a priority level:

| Priority | Constant | Description |
|----------|----------|-------------|
| 0 | `PriorityThin` | Minimal implementations |
| 10 | `PriorityThick` | Full SDK implementations |

Higher priority providers override lower priority registrations for the same name.

### Auto-Registration

Providers auto-register via `init()` when imported:

```go
// Import specific provider
import _ "github.com/plexusone/omniavatar/providers/heygen"

// Or import all providers
import _ "github.com/plexusone/omniavatar/providers/all"
```

### Registry Functions

```go
// Get a provider by name
provider, err := omniavatar.GetAvatarProvider("heygen", opts...)

// List all registered providers
names := omniavatar.ListAvatarProviders()

// Check if provider is registered
if omniavatar.HasAvatarProvider("heygen") {
    // ...
}
```

## Supported Providers

### HeyGen LiveAvatar

Real-time avatar with lip-sync using HeyGen's LITE mode.

```go
provider, err := omniavatar.GetAvatarProvider("heygen",
    omniavatar.WithAPIKey(os.Getenv("HEYGEN_API_KEY")),
    omniavatar.WithExtension("avatar_id", "josh_lite3_20230714"),
    omniavatar.WithExtension("sandbox", true),           // 60s limit, no credits
    omniavatar.WithExtension("video_quality", "high"),   // very_high, high, medium, low
)
```

| Option | Description |
|--------|-------------|
| `avatar_id` | Avatar UUID (required) |
| `sandbox` | Enable sandbox mode (recommended for dev) |
| `video_quality` | Video quality preset |

### Tavus Conversational Video

Real-time avatar using Tavus PAL (Personalized AI Likeness).

```go
provider, err := omniavatar.GetAvatarProvider("tavus",
    omniavatar.WithAPIKey(os.Getenv("TAVUS_API_KEY")),
    omniavatar.WithExtension("pal_id", "pal_xxx"),  // Optional
    omniavatar.WithExtension("face_id", "face_xxx"), // Optional
)
```

| Option | Description |
|--------|-------------|
| `pal_id` | PAL ID (optional, uses stock avatar if not set) |
| `face_id` | Face override (optional) |

### bitHuman Real-time Avatars

Ultra-low latency avatars using bitHuman.

```go
provider, err := omniavatar.GetAvatarProvider("bithuman",
    omniavatar.WithAPIKey(os.Getenv("BITHUMAN_API_KEY")),
    omniavatar.WithExtension("agent_id", "agent_xxx"),
)
```

| Option | Description |
|--------|-------------|
| `agent_id` | bitHuman agent ID (required) |

## Session Lifecycle

```
1. Get Provider    → omniavatar.GetAvatarProvider("heygen", opts...)
2. Create Session  → provider.CreateSession(cfg)
3. Start           → session.Start(ctx, &LiveKitStartOptions{...})
4. Wait for Join   → session.WaitForJoin(ctx, 30*time.Second)
5. Stream Audio    → session.AudioOutput().CaptureFrame(ctx, pcm)
6. Close           → session.Close(ctx)
```

## LiveKit Integration

### Token Generation

Generate tokens for avatar participants to join LiveKit rooms:

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

### Start Options

```go
type LiveKitStartOptions struct {
    Room             *lksdk.Room  // LiveKit room reference
    AgentIdentity    string       // Agent's participant identity
    LiveKitURL       string       // LiveKit server URL
    LiveKitAPIKey    string       // API key for token generation
    LiveKitAPISecret string       // API secret for token generation
}
```

## Audio Format

Default audio configuration:

| Parameter | Value |
|-----------|-------|
| Sample Rate | 24000 Hz |
| Channels | 1 (mono) |
| Encoding | PCM16 (linear16) |

## Provider Comparison

| Provider | Latency | Video Quality | Voice Cloning |
|----------|---------|---------------|---------------|
| HeyGen | ~500ms | Excellent | Yes |
| Tavus | ~300ms | Excellent | Yes (via PAL) |
| bitHuman | ~200ms | Good | No |

## Resources

- [omniavatar-core](https://github.com/plexusone/omniavatar-core) - Core interfaces
- [HeyGen LiveAvatar](https://liveavatar.com/)
- [Tavus](https://www.tavus.io/)
- [bitHuman](https://www.bithuman.io/)
