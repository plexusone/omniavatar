# bitHuman

| Surface | bitHuman product |
|---------|------------------|
| Live | Real-time Avatars (ultra-low latency, ~200ms) |
| Render | Video Generation |

Both surfaces use the same `BITHUMAN_API_KEY`. SDK:
[bithuman-go](https://github.com/plexusone/bithuman-go).

## Live

Ultra-low latency real-time avatars — the lowest-latency option of the
three providers (video quality is good rather than excellent, and voice
cloning is not available).

```go
provider, err := omniavatar.GetLiveProvider("bithuman",
    omniavatar.WithAPIKey(os.Getenv("BITHUMAN_API_KEY")),
    omniavatar.WithExtension("agent_id", "agent_xxx"),
)
```

| Option | Description |
|--------|-------------|
| `agent_id` | bitHuman agent ID (required) |

## Render

Asynchronous talking-head generation using bitHuman agents.

```go
provider, err := omniavatar.GetRenderProvider("bithuman",
    omniavatar.WithAPIKey(os.Getenv("BITHUMAN_API_KEY")),
    omniavatar.WithExtension("agent_id", "agent_xxx"),
)
```

| Option | Level | Description |
|--------|-------|-------------|
| `agent_id` | provider | Default agent ID |
| `voice_id` | request | TTS voice for `Script` input |

### Audio Upload

The bitHuman render provider implements `render.AudioUploader` via the
bitHuman file API (base64 upload returning a hosted URL). Common audio
formats are accepted; the MIME type is derived from the filename
extension.

### Limitations

Width, height, and background options are not supported by the bitHuman
video API and are ignored by this provider.
