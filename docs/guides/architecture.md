# Architecture

OmniAvatar is split across two Go modules, following the same pattern as
[OmniVoice](https://github.com/plexusone/omnivoice-core).

## Core vs. Batteries-Included

| Module | Import when |
|--------|-------------|
| [`omniavatar-core`](https://github.com/plexusone/omniavatar-core) | You accept avatars as an interface (library code, consumers like videoascode) — zero provider dependencies |
| [`omniavatar`](https://github.com/plexusone/omniavatar) | You construct providers — pulls in provider SDKs and LiveKit |

```
omniavatar-core/              # Core interfaces (no provider deps)
├── live/                     # Real-time sessions
│   ├── provider.go           # live.Provider
│   ├── session.go            # live.Session + callbacks
│   ├── audio.go              # live.AudioDestination
│   └── errors.go
├── render/                   # Batch video generation
│   ├── provider.go           # render.Provider, render.AudioUploader
│   ├── request.go            # render.GenerateRequest
│   ├── job.go                # render.Job, JobStatus, render.Wait
│   └── errors.go
└── registry/                 # Shared config + factory types

omniavatar/                   # Provider implementations
├── registry.go               # Global live + render registries
├── token.go                  # LiveKit token generation
└── providers/
    ├── heygen/               # live + render
    ├── tavus/                # live + render
    ├── bithuman/             # live + render
    └── all/                  # convenience import
```

## The Two Surfaces

The module names its packages after the *mode*, not the vendor concept:

- **`live`** — session-oriented: `Start`, `WaitForJoin`, stream PCM
  frames, `Close`. The avatar joins a room and speaks in real time.
- **`render`** — job-oriented: `Generate`, `Status`/`Wait`, `Download`.
  Audio in, MP4 out, minutes later.

`avatar` is deliberately *not* a package name: with two surfaces,
`live.Session` and `render.Job` are self-describing at every call site,
while `avatar.X` would not say which mode is involved.

## Provider Registry

Providers self-register both surfaces via `init()` when imported:

```go
// Import one provider
import _ "github.com/plexusone/omniavatar/providers/heygen"

// Or all providers
import _ "github.com/plexusone/omniavatar/providers/all"
```

Registration carries a priority so alternative implementations can
coexist:

| Priority | Constant | Meaning |
|----------|----------|---------|
| 0 | `PriorityThin` | Minimal (stdlib-only) implementations |
| 10 | `PriorityThick` | Full SDK implementations |

Higher priority wins for the same name. The registry API is symmetric
across surfaces:

```go
omniavatar.GetLiveProvider("heygen", opts...)     // live.Provider
omniavatar.GetRenderProvider("heygen", opts...)   // render.Provider
omniavatar.ListLiveProviders()                    // []string
omniavatar.ListRenderProviders()                  // []string
```

## Capability Interfaces

Optional provider abilities are modeled as feature-detected interfaces
rather than core methods, so the core interface stays at the
lowest common denominator:

```go
// Audio hosting (HeyGen, bitHuman — not Tavus)
if up, ok := provider.(render.AudioUploader); ok {
    url, err = up.UploadAudio(ctx, filename, r)
}
```

Providers that lack a capability simply don't implement the interface;
`render.ErrAudioUploadUnsupported` is available for callers that need a
typed error.

## Design Documents

The full requirements and technical design for the render surface live
in [Specs](../specs/render/PRD.md).
