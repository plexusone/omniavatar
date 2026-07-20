# Render Quick Start

Batch avatar video generation: submit narration audio, poll for
completion, download a talking-head MP4.

The design center is **audio-driven generation** — the avatar's lip-sync
comes from the exact audio you supply, so pauses, pronunciation, and
timing match your final production audio. A text `Script` with provider
TTS is supported as a secondary path.

## Generate a Video

```go
import (
    "github.com/plexusone/omniavatar"
    "github.com/plexusone/omniavatar-core/render"
    _ "github.com/plexusone/omniavatar/providers/all"
)

func main() {
    provider, err := omniavatar.GetRenderProvider("bithuman",
        omniavatar.WithAPIKey(os.Getenv("BITHUMAN_API_KEY")))
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
    log.Printf("video ready: %.1fs", status.Duration)

    out, _ := os.Create("presenter.mp4")
    defer out.Close()
    if err := provider.Download(ctx, job.ID, out); err != nil {
        log.Fatal(err)
    }
}
```

## Job Lifecycle

```
1. Get Provider    → omniavatar.GetRenderProvider("heygen", opts...)
2. Upload Audio    → provider.(render.AudioUploader).UploadAudio(...)  [optional]
3. Generate        → provider.Generate(ctx, render.GenerateRequest{...})
4. Wait            → render.Wait(ctx, provider, job.ID, interval)
5. Download        → provider.Download(ctx, job.ID, dst)
```

## Audio Delivery

All providers consume audio **by URL**. Hosting support varies, which is
why upload is a feature-detected capability rather than part of the core
interface:

| Provider | Local file upload |
|----------|-------------------|
| HeyGen | Yes — asset API (MP3/`audio/mpeg` is the documented type) |
| bitHuman | Yes — file API |
| Tavus | No — supply a publicly fetchable `.wav`/`.mp3` URL |

```go
if up, ok := provider.(render.AudioUploader); ok {
    audioURL, err = up.UploadAudio(ctx, "narration.mp3", f)
} else {
    // provider cannot host files; audioURL must be supplied externally
}
```

## Avatar Identity

`GenerateRequest.AvatarID` maps to each provider's identity concept:

| Provider | AvatarID means |
|----------|----------------|
| HeyGen | `avatar_id` (or `talking_photo_id` via extension) |
| Tavus | `replica_id` |
| bitHuman | `agent_id` |

## Next Steps

- [Render Job Lifecycle](../guides/render-lifecycle.md) — states, errors, caching guidance
- [Provider pages](../providers/heygen.md) — per-provider render options
- `examples/render-basic` in the repo — runnable CLI covering this whole flow
