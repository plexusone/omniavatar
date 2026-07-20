# Tavus

| Surface | Tavus product |
|---------|---------------|
| Live | Conversational Video Interface (CVI) with PAL |
| Render | Video Generation (replicas) |

Both surfaces use the same `TAVUS_API_KEY`. SDK:
[tavus-go](https://github.com/plexusone/tavus-go).

## Live

Real-time avatar using Tavus PAL (Personalized AI Likeness), ~300ms
latency with voice cloning via PAL.

```go
provider, err := omniavatar.GetLiveProvider("tavus",
    omniavatar.WithAPIKey(os.Getenv("TAVUS_API_KEY")),
    omniavatar.WithExtension("pal_id", "pal_xxx"),   // optional
    omniavatar.WithExtension("face_id", "face_xxx"), // optional
)
```

| Option | Description |
|--------|-------------|
| `pal_id` | PAL ID (optional, uses stock avatar if not set) |
| `face_id` | Face override (optional) |

## Render

Asynchronous talking-head generation using Tavus replicas.

```go
provider, err := omniavatar.GetRenderProvider("tavus",
    omniavatar.WithAPIKey(os.Getenv("TAVUS_API_KEY")),
    omniavatar.WithExtension("replica_id", "rep_xxx"),
)
```

| Option | Level | Description |
|--------|-------|-------------|
| `replica_id` | provider | Default replica ID |
| `fast` | request | Faster generation (disables some features) |
| `callback_url` | request | Completion webhook URL |

### No Audio Upload

!!! note
    Tavus has no audio upload API, so the Tavus render provider does
    **not** implement `render.AudioUploader`. Supply a publicly fetchable
    `GenerateRequest.AudioURL` (`.wav` or `.mp3`) — for example from your
    own object storage.

### Backgrounds and Downloads

`Background{Type: "video"}` maps to `background_source_url`; color and
image backgrounds are not supported by the Tavus API. Downloads prefer
the `download_url` and fall back to the `stream_url`.
