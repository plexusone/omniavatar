# LivePortrait + JoyVASA (Local)

On-device audio-driven talking-head video generation on Apple Silicon.

## Overview

The LivePortrait + JoyVASA provider generates talking-head videos locally
without requiring a cloud API. It uses:

- **LivePortrait** (MIT) — portrait animation renderer
- **JoyVASA** (MIT) — audio-to-motion diffusion stage
- **mediapipe** (Apache-2.0) — face detection (commercial-clean)

The Go client connects to a Python gRPC server running the inference pipeline.

## Quick Start

```go
import (
    "github.com/plexusone/omniavatar"
    "github.com/plexusone/omniavatar-core/render"
    _ "github.com/plexusone/omniavatar/providers/liveportrait-joyvasa"
)

// No API key needed
provider, _ := omniavatar.GetRenderProvider("liveportrait-joyvasa")

// Upload local audio
audioURL, _ := provider.(render.AudioUploader).UploadAudio(ctx, "narration.wav", f)

// Generate video
job, _ := provider.Generate(ctx, render.GenerateRequest{
    AvatarID: "john",  // avatar bundle name
    AudioURL: audioURL,
})

// Wait and download
status, _ := render.Wait(ctx, provider, job.ID, 3*time.Second)
provider.Download(ctx, job.ID, outFile)
```

## Setup

### 1. Start the Python Server

```bash
cd $GOPATH/src/github.com/plexusone/omniavatar-core/providers/liveportrait-joyvasa/server

# Create venv
arch -arm64 python3 -m venv .venv
source .venv/bin/activate

# Install and run
pip install -r requirements.txt
./generate_proto.sh
./run.sh
```

### 2. Create Avatar Bundles

Avatar bundles live in `~/.omniavatar/avatars/<name>/`:

```
~/.omniavatar/avatars/
    john/
        metadata.json
        idle/
            idle.mp4      # neutral, mouth-closed clip
        references/       # optional
            front.png
```

**metadata.json:**

```json
{
  "name": "john",
  "fps": 25,
  "resolution": {"width": 512, "height": 512}
}
```

## Options

### Provider Options

| Option | Type | Description |
|--------|------|-------------|
| `endpoint` | string | Unix socket path (default: `/tmp/omniavatar-liveportrait-joyvasa.sock`) |

```go
provider, _ := omniavatar.GetRenderProvider("liveportrait-joyvasa",
    omniavatar.WithExtension("endpoint", "unix:///custom/path.sock"),
)
```

### Generate Request Extensions

| Option | Type | Description |
|--------|------|-------------|
| `seed` | int64 | Random seed for deterministic output |
| `motion_scale` | float32 | Facial movement intensity (default: 1.0) |

```go
job, _ := provider.Generate(ctx, render.GenerateRequest{
    AvatarID: "john",
    AudioURL: audioURL,
    Extensions: map[string]any{
        "seed":         int64(42),
        "motion_scale": float32(1.2),
    },
})
```

## Capabilities

| Capability | Supported |
|------------|-----------|
| `render.Provider` | Yes |
| `render.AudioUploader` | Yes (local:// URLs) |
| `render.AvatarLister` | Yes (via ListAvatars RPC) |

## Performance

On Apple Silicon (M-series):

| Metric | Value |
|--------|-------|
| Resolution | 512×512 |
| Speed | ~5 min for 13.7s output |
| Memory | ~4GB with model loaded |

The bottleneck is Conv3D and grid_sampler_3d falling back to CPU on MPS.
torch 2.5+ may add native support for significant speedup.

## Troubleshooting

### Connection Refused

Ensure the Python server is running:

```bash
ls -la /tmp/omniavatar-liveportrait-joyvasa.sock
```

### Model Loading Fails

The server sets `PYTORCH_ENABLE_MPS_FALLBACK=1` automatically. If you see
Conv3D errors, verify this environment variable is set.

### Out of Memory

The model requires ~4GB. Use `UnloadModel` when not actively rendering:

```go
if lp, ok := provider.(*liveportraitjoyvasa.Provider); ok {
    lp.UnloadModel(ctx)
}
```

## See Also

- [omniavatar-core Local Render Guide](https://github.com/plexusone/omniavatar-core/blob/main/docs/local-render.md)
- [omniavatar-core Provider README](https://github.com/plexusone/omniavatar-core/tree/main/providers/liveportrait-joyvasa)
