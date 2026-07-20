# OmniAvatar

**Provider-agnostic AI avatars for Go** — one interface, two surfaces, three providers.

OmniAvatar gives PlexusOne applications a unified way to work with AI avatar vendors:

- **`live`** — real-time streaming avatar sessions (LiveKit rooms, PCM audio streaming for lip-sync) for conversational agents such as OmniMeet
- **`render`** — asynchronous batch avatar video generation (narration audio in, talking-head MP4 out) for offline pipelines such as [videoascode](https://github.com/grokify/videoascode) presentation videos

## Supported Providers

| Provider | Live | Render | Render Audio Upload |
|----------|------|--------|---------------------|
| [HeyGen](providers/heygen.md) | LiveAvatar LITE mode | Video Generation v2 | Yes (MP3) |
| [Tavus](providers/tavus.md) | Conversational Video (CVI) | Video Generation (replicas) | No — URL only |
| [bitHuman](providers/bithuman.md) | Real-time Avatars | Video Generation | Yes |

## Installation

```bash
go get github.com/plexusone/omniavatar
```

For interfaces only (no provider dependencies):

```bash
go get github.com/plexusone/omniavatar-core
```

## Quick Example

```go
import (
    "github.com/plexusone/omniavatar"
    "github.com/plexusone/omniavatar-core/render"
    _ "github.com/plexusone/omniavatar/providers/all"
)

provider, err := omniavatar.GetRenderProvider("heygen",
    omniavatar.WithAPIKey(os.Getenv("HEYGEN_API_KEY")))

job, err := provider.Generate(ctx, render.GenerateRequest{
    AvatarID: avatarID,
    AudioURL: narrationURL, // lip-sync driven by your own audio
})
status, err := render.Wait(ctx, provider, job.ID, 5*time.Second)
err = provider.Download(ctx, job.ID, outFile)
```

## Module Layout

OmniAvatar follows the same pattern as [OmniVoice](https://github.com/plexusone/omnivoice-core):

| Module | Contents |
|--------|----------|
| [`omniavatar-core`](https://github.com/plexusone/omniavatar-core) | Interfaces only (`live`, `render`, `registry`) — no provider dependencies |
| [`omniavatar`](https://github.com/plexusone/omniavatar) | Provider implementations with auto-registration |

See the [Architecture guide](guides/architecture.md) for how the split works and when to import which.

## Next Steps

- [Live Quick Start](getting-started/live.md) — real-time avatar in a LiveKit room
- [Render Quick Start](getting-started/render.md) — narration audio to talking-head MP4
- [Render PRD](specs/render/PRD.md) / [TRD](specs/render/TRD.md) — design background
