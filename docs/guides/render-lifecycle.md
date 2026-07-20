# Render Job Lifecycle

How render jobs progress, how to wait for them, and how to build
reliable pipelines on top.

## States

`render.JobState` normalizes each provider's native statuses:

| `JobState` | HeyGen | Tavus | bitHuman | Terminal |
|------------|--------|-------|----------|----------|
| `pending` | `pending` | `queued` | `pending` | no |
| `processing` | `processing` | `generating` | `processing` | no |
| `completed` | `completed` | `ready` | `completed` | yes |
| `failed` | `failed` | `error`, `deleted` | `failed` | yes |

`JobStatus.RawStatus` always preserves the provider-native string for
logging. Unknown provider states map to `processing` (non-terminal), so
pollers keep waiting rather than aborting on a new vendor status.

## Waiting

`render.Wait` polls until a terminal state, honoring context
cancellation:

```go
status, err := render.Wait(ctx, provider, job.ID, 5*time.Second)
if errors.Is(err, render.ErrJobFailed) {
    // status is non-nil here: inspect status.ErrorCode / status.ErrorMsg
}
```

On failure, `Wait` returns **both** the final status and an error
wrapping `render.ErrJobFailed`, so `errors.Is` works while the error
details stay inspectable.

## Errors

| Sentinel | Meaning |
|----------|---------|
| `ErrInvalidRequest` | Request failed validation (missing `AvatarID`, neither/both of `AudioURL`/`Script`) |
| `ErrAudioUploadUnsupported` | Provider cannot host audio files |
| `ErrJobNotFound` | Provider doesn't recognize the job ID |
| `ErrJobFailed` | Job reached `failed` |
| `ErrJobNotCompleted` | `Download` called before successful completion |

Provider errors are wrapped in `render.ProviderError` with the provider
name and operation (`render/heygen: generate: ...`).

## Downloading

```go
f, _ := os.Create("presenter.mp4")
defer f.Close()
err := provider.Download(ctx, job.ID, f)
```

`Download` streams to any `io.Writer` and re-checks status first — some
providers (HeyGen) return time-limited signed URLs, so the URL is
fetched fresh at download time. `JobStatus.VideoURL` is therefore best
treated as informational; use `Download` for the actual bytes.

## Caching Guidance for Consumers

Job IDs, states, and errors are plain serializable values so consumers
can cache results. The recommended cache key is a hash of:

- the narration audio **content** (not path),
- the provider name,
- the avatar ID,
- any request extensions.

[videoascode](https://github.com/grokify/videoascode) implements exactly
this for its `vac avatar generate` command — unchanged narration makes
regeneration free.

## Thumbnails and Duration

`JobStatus.Duration` (seconds) and `JobStatus.ThumbnailURL` are
populated when the provider reports them (HeyGen reports both; others
vary). Useful for progress UI without downloading the video.
