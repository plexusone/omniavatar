# OmniAvatar

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/omniavatar/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/omniavatar/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/omniavatar/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/omniavatar/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/omniavatar/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/omniavatar/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/omniavatar
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/omniavatar
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.dev/omniavatar
 [viz-svg]: https://img.shields.io/badge/Go-visualizaton-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fomniavatar
 [loc-svg]: https://tokei.rs/b1/github/plexusone/omniavatar
 [repo-url]: https://github.com/plexusone/omniavatar
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/omniavatar/blob/main/LICENSE

Batteries-included package for AI avatars with two surfaces:

- **live** — real-time streaming avatars (LiveKit sessions, PCM audio streaming for lip-sync) for conversational agents
- **render** — asynchronous batch avatar video generation (narration audio in, talking-head MP4 out) for offline pipelines such as presentation videos

Provides provider implementations for HeyGen, Tavus, and bitHuman.

For core interfaces only (no provider dependencies), see [omniavatar-core](https://github.com/plexusone/omniavatar-core).

## Quick Start (live)

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
        omniavatar.WithExtension("sandbox", true))
    if err != nil {
        log.Fatal(err)
    }

    session, err := provider.CreateSession(live.SessionConfig{
        AudioConfig: live.DefaultAudioConfig(),
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

## Quick Start (render)

```go
import (
    "github.com/plexusone/omniavatar"
    "github.com/plexusone/omniavatar-core/render"
    _ "github.com/plexusone/omniavatar/providers/all"
)

func main() {
    provider, err := omniavatar.GetRenderProvider("bithuman",
        omniavatar.WithAPIKey(os.Getenv("BITHUMAN_API_KEY")),
        omniavatar.WithExtension("agent_id", agentID))
    if err != nil {
        log.Fatal(err)
    }

    // Providers with hosting support can upload local narration audio.
    audioURL := ""
    if up, ok := provider.(render.AudioUploader); ok {
        audioURL, err = up.UploadAudio(ctx, "narration.mp3", audioFile)
        if err != nil {
            log.Fatal(err)
        }
    }

    job, err := provider.Generate(ctx, render.GenerateRequest{
        AvatarID: agentID,
        AudioURL: audioURL,
    })
    if err != nil {
        log.Fatal(err)
    }

    status, err := render.Wait(ctx, provider, job.ID, 5*time.Second)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("video ready: %s (%.1fs)", status.VideoURL, status.Duration)

    out, err := os.Create("presenter.mp4")
    if err != nil {
        log.Fatal(err)
    }
    defer out.Close()

    if err := provider.Download(ctx, job.ID, out); err != nil {
        log.Fatal(err)
    }
}
```

## Architecture

Adapters follow the PlexusOne convention: **render** adapters live in each
provider SDK repo (`heygen-go/omniavatar`, …), depending only on
`omniavatar-core`, so provider-specific knowledge stays with the SDK. The
**live** adapters live here in the batteries-included package, because their
LiveKit integration (`LiveKitStartOptions`, token generation) lives here.
This package re-exports both, registered by name via `providers/all`.

```
omniavatar-core/              # Core interfaces + shared helpers (no provider deps)
├── live/                     # Real-time session interfaces
├── render/                   # Batch generation: Provider, AudioUploader,
│                             #   AvatarLister, GenerateRequest, Wait,
│                             #   AudioContentType/DownloadURL helpers
└── registry/                 # Factory types

heygen-go/omniavatar/         # HeyGen RENDER adapter (core-only) — in the SDK repo
tavus-go/omniavatar/          # Tavus RENDER adapter
bithuman-go/omniavatar/       # bitHuman RENDER adapter

omniavatar/                   # Batteries-included (this package)
├── registry.go               # Global live + render registries
├── token.go / start_options.go  # LiveKit token + start options
└── providers/
    ├── heygen/               # HeyGen LIVE adapter (LiveAvatar); registers the SDK render adapter
    ├── tavus/                # Tavus LIVE adapter (CVI)
    ├── bithuman/             # bitHuman LIVE adapter
    └── all/                  # Convenience import (registers every provider)
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

Providers auto-register both surfaces via `init()` when imported:

```go
// Import specific provider
import _ "github.com/plexusone/omniavatar/providers/heygen"

// Or import all providers
import _ "github.com/plexusone/omniavatar/providers/all"
```

### Registry Functions

```go
// Live (real-time sessions)
provider, err := omniavatar.GetLiveProvider("heygen", opts...)
names := omniavatar.ListLiveProviders()
ok := omniavatar.HasLiveProvider("heygen")

// Render (batch video generation)
provider, err := omniavatar.GetRenderProvider("heygen", opts...)
names := omniavatar.ListRenderProviders()
ok := omniavatar.HasRenderProvider("heygen")
```

## Supported Providers

### HeyGen

Live: real-time avatar with lip-sync using HeyGen LiveAvatar LITE mode.
Render: HeyGen Video Generation API (v2).

Note: the live surface uses the **LiveAvatar API key** (`LIVEAVATAR_API_KEY`);
the render surface uses the **HeyGen API key** (`HEYGEN_API_KEY`). They are
different credentials.

```go
// Live
provider, err := omniavatar.GetLiveProvider("heygen",
    omniavatar.WithAPIKey(os.Getenv("LIVEAVATAR_API_KEY")),
    omniavatar.WithExtension("avatar_id", "josh_lite3_20230714"),
    omniavatar.WithExtension("sandbox", true),           // 60s limit, no credits
    omniavatar.WithExtension("video_quality", "high"),   // very_high, high, medium, low
)

// Render
provider, err := omniavatar.GetRenderProvider("heygen",
    omniavatar.WithAPIKey(os.Getenv("HEYGEN_API_KEY")),
    omniavatar.WithExtension("avatar_id", avatarID),
)
```

| Surface | Option | Description |
|---------|--------|-------------|
| live | `avatar_id` | Avatar UUID (required) |
| live | `sandbox` | Enable sandbox mode (recommended for dev) |
| live | `video_quality` | Video quality preset |
| render | `avatar_id` | Default avatar ID |
| render | `upload_base_url` | Custom asset upload service URL (default: upload.heygen.com) |
| render request | `talking_photo_id` | Use a talking photo instead of an avatar |
| render request | `avatar_style` | normal, circle, closeUp |
| render request | `voice_id` | TTS voice for Script input |
| render request | `test` | Watermarked test video, no credits |

The HeyGen render provider implements `render.AudioUploader` via the HeyGen
asset upload API (MP3/`audio/mpeg` is the documented audio asset type).

### Tavus

Live: real-time avatar using Tavus PAL (Personalized AI Likeness).
Render: Tavus Video Generation using replicas.

```go
// Live
provider, err := omniavatar.GetLiveProvider("tavus",
    omniavatar.WithAPIKey(os.Getenv("TAVUS_API_KEY")),
    omniavatar.WithExtension("pal_id", "pal_xxx"),   // Optional
    omniavatar.WithExtension("face_id", "face_xxx"), // Optional
)

// Render
provider, err := omniavatar.GetRenderProvider("tavus",
    omniavatar.WithAPIKey(os.Getenv("TAVUS_API_KEY")),
    omniavatar.WithExtension("replica_id", "rep_xxx"),
)
```

| Surface | Option | Description |
|---------|--------|-------------|
| live | `pal_id` | PAL ID (optional, uses stock avatar if not set) |
| live | `face_id` | Face override (optional) |
| render | `replica_id` | Default replica ID |
| render request | `fast` | Faster generation (disables some features) |
| render request | `callback_url` | Completion webhook URL |

Tavus has no audio upload API; supply a publicly fetchable
`GenerateRequest.AudioURL` (.wav or .mp3).

### bitHuman

Live: ultra-low latency real-time avatars.
Render: bitHuman video generation, including audio upload support
(`render.AudioUploader`).

```go
// Live
provider, err := omniavatar.GetLiveProvider("bithuman",
    omniavatar.WithAPIKey(os.Getenv("BITHUMAN_API_KEY")),
    omniavatar.WithExtension("agent_id", "agent_xxx"),
)

// Render
provider, err := omniavatar.GetRenderProvider("bithuman",
    omniavatar.WithAPIKey(os.Getenv("BITHUMAN_API_KEY")),
    omniavatar.WithExtension("agent_id", "agent_xxx"),
)
```

| Surface | Option | Description |
|---------|--------|-------------|
| live | `agent_id` | bitHuman agent ID (required) |
| render | `agent_id` | Default agent ID |
| render request | `voice_id` | TTS voice for Script input |

## Session Lifecycle (live)

```
1. Get Provider    → omniavatar.GetLiveProvider("heygen", opts...)
2. Create Session  → provider.CreateSession(cfg)
3. Start           → session.Start(ctx, &LiveKitStartOptions{...})
4. Wait for Join   → session.WaitForJoin(ctx, 30*time.Second)
5. Stream Audio    → session.AudioOutput().CaptureFrame(ctx, pcm)
6. Close           → session.Close(ctx)
```

## Job Lifecycle (render)

```
1. Get Provider    → omniavatar.GetRenderProvider("heygen", opts...)
2. Upload Audio    → provider.(render.AudioUploader).UploadAudio(...)  [optional]
3. Generate        → provider.Generate(ctx, render.GenerateRequest{...})
4. Wait            → render.Wait(ctx, provider, job.ID, interval)
5. Download        → provider.Download(ctx, job.ID, dst)
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

## Audio Format (live)

Default audio configuration:

| Parameter | Value |
|-----------|-------|
| Sample Rate | 24000 Hz |
| Channels | 1 (mono) |
| Encoding | PCM16 (linear16) |

## Provider Comparison

| Provider | Live Latency | Video Quality | Voice Cloning | Render Audio Upload |
|----------|--------------|---------------|---------------|---------------------|
| HeyGen | ~500ms | Excellent | Yes | Yes (asset API, MP3) |
| Tavus | ~300ms | Excellent | Yes (via PAL) | No (URL only) |
| bitHuman | ~200ms | Good | No | Yes |

## Specs

- [Render PRD](docs/specs/render/PRD.md)
- [Render TRD](docs/specs/render/TRD.md)
- [Render Plan](docs/specs/render/PLAN.md)
- [Render Roadmap](docs/specs/render/ROADMAP.md)

## Resources

- [omniavatar-core](https://github.com/plexusone/omniavatar-core) - Core interfaces
- [HeyGen LiveAvatar](https://liveavatar.com/)
- [HeyGen API](https://docs.heygen.com/)
- [Tavus](https://www.tavus.io/)
- [bitHuman](https://www.bithuman.io/)
