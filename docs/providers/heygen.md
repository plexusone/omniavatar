# HeyGen

| Surface | HeyGen product | API host |
|---------|----------------|----------|
| Live | LiveAvatar (LITE mode) | `api.liveavatar.com` |
| Render | Video Generation v2 | `api.heygen.com` |

!!! warning "Two different API keys"
    HeyGen live and render use **different credentials from different
    dashboards**: the live surface uses `LIVEAVATAR_API_KEY`
    ([app.liveavatar.com/developers](https://app.liveavatar.com/developers)),
    the render surface uses `HEYGEN_API_KEY`
    ([app.heygen.com/settings?nav=API](https://app.heygen.com/settings?nav=API)).

## Live

Real-time avatar with lip-sync using HeyGen's LiveAvatar LITE mode
(~500ms latency, excellent video quality, voice cloning available).

```go
provider, err := omniavatar.GetLiveProvider("heygen",
    omniavatar.WithAPIKey(os.Getenv("LIVEAVATAR_API_KEY")),
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

## Render

Asynchronous talking-head generation via the HeyGen Video Generation API,
wrapped by [heygen-go](https://github.com/plexusone/heygen-go).

```go
provider, err := omniavatar.GetRenderProvider("heygen",
    omniavatar.WithAPIKey(os.Getenv("HEYGEN_API_KEY")),
    omniavatar.WithExtension("avatar_id", avatarID),
)
```

| Option | Level | Description |
|--------|-------|-------------|
| `avatar_id` | provider | Default avatar ID |
| `upload_base_url` | provider | Custom asset upload URL (default: upload.heygen.com) |
| `talking_photo_id` | request | Use a talking photo instead of an avatar |
| `avatar_style` | request | `normal`, `circle`, `closeUp` |
| `voice_id` | request | TTS voice for `Script` input |
| `test` | request | Watermarked test video, no credits |

### Audio Upload

The HeyGen render provider implements `render.AudioUploader` via the
HeyGen asset upload API (`upload.heygen.com`, a separate host from the
main API). MP3 (`audio/mpeg`) is the documented audio asset type — other
formats may be rejected; transcode to MP3 first.

### Downloads

HeyGen video URLs are time-limited signed URLs; `Download` re-fetches
the status immediately before downloading so the URL is always fresh.
`JobStatus.ThumbnailURL` is populated when HeyGen reports one.
