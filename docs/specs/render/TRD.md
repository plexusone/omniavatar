# TRD: OmniAvatar Render — Technical Design

Status: Approved
Date: 2026-07-16
Related: [PRD.md](PRD.md), [PLAN.md](PLAN.md), [ROADMAP.md](ROADMAP.md)

## Package Layout

```
omniavatar-core/                  # interfaces only, stdlib deps
├── live/                         # RENAMED from avatar/ (real-time sessions)
│   ├── provider.go               # live.Provider, live.SessionConfig
│   ├── session.go                # live.Session, live.SessionCallbacks, live.Metrics
│   ├── audio.go                  # live.AudioDestination, live.AudioConfig
│   └── errors.go                 # live sentinel errors, live.ProviderError
├── render/                       # NEW (batch generation)
│   ├── provider.go               # render.Provider, render.AudioUploader
│   ├── job.go                    # render.Job, render.JobState, render.JobStatus, render.Wait
│   ├── request.go                # render.GenerateRequest, render.Background
│   └── errors.go                 # render sentinel errors, render.ProviderError
└── registry/
    └── registry.go               # ProviderConfig + options (shared),
                                  # LiveProviderFactory, RenderProviderFactory

omniavatar/                       # provider implementations
├── registry.go                   # live + render global registries (symmetric API)
├── token.go                      # LiveKit token generation (unchanged)
├── start_options.go              # LiveKitStartOptions (unchanged)
└── providers/
    ├── heygen/                   # live provider + render.go (heygen-go/video)
    ├── tavus/                    # live provider + render.go (tavus-go)
    ├── bithuman/                 # live provider + render.go (bithuman-go), AudioUploader
    └── all/                      # unchanged (imports the three provider packages)
```

## Part 1: Rename `avatar/` → `live/`

Mechanical rename with semantic cleanups:

- `git mv avatar live`; `package avatar` → `package live` in all files.
- Error string prefixes `"avatar: ..."` → `"live: ..."`; `ProviderError.Error()` prefix `"avatar/"` → `"live/"` (Go convention: error strings identify the package).
- All type and function names inside the package are unchanged (`Provider`, `Session`, `AudioDestination`, `SessionConfig`, `DefaultAudioConfig`, ...). Call sites change from `avatar.X` to `live.X`.
- `registry.ProviderFactory` → `registry.LiveProviderFactory` (making room for `RenderProviderFactory`).
- Downstream (`omniavatar`): imports, plus registry API renames for symmetry (see Part 3).

## Part 2: `render/` Package Interfaces

### Provider

```go
// Package render provides provider-agnostic interfaces for asynchronous
// (batch) avatar video generation: submit narration audio or a script,
// poll for completion, download a talking-head video.
//
// This is the offline counterpart to package live, which handles
// real-time streaming avatar sessions.
package render

// Provider generates avatar videos asynchronously.
type Provider interface {
    // Name returns the provider name (e.g., "heygen", "tavus", "bithuman").
    Name() string

    // Generate submits a video generation job. It returns as soon as the
    // provider accepts the job; use Status or Wait to track completion.
    Generate(ctx context.Context, req GenerateRequest) (*Job, error)

    // Status returns the current status of a job.
    Status(ctx context.Context, jobID string) (*JobStatus, error)

    // Download streams the completed video to dst.
    // Returns ErrJobNotCompleted if the job has not completed successfully.
    Download(ctx context.Context, jobID string, dst io.Writer) error
}

// AudioUploader is an optional capability for providers that can host
// local audio files. Callers should feature-detect:
//
//     if up, ok := provider.(render.AudioUploader); ok {
//         url, err = up.UploadAudio(ctx, "narration.mp3", f)
//     }
//
// Providers without hosting support do not implement this interface;
// callers must supply a publicly fetchable GenerateRequest.AudioURL.
type AudioUploader interface {
    // UploadAudio uploads audio content and returns a URL usable as
    // GenerateRequest.AudioURL with the same provider.
    UploadAudio(ctx context.Context, filename string, r io.Reader) (string, error)
}
```

### Request

```go
// GenerateRequest describes an avatar video generation job.
//
// Exactly one of AudioURL or Script must be set. AudioURL is the design
// center: it drives lip-sync from existing narration audio so the avatar
// matches the authoritative audio track exactly.
type GenerateRequest struct {
    // AvatarID identifies the presenter with the provider.
    // HeyGen: avatar_id; Tavus: replica_id; bitHuman: agent_id.
    AvatarID string

    // AudioURL is a fetchable URL to narration audio (.mp3/.wav).
    AudioURL string

    // Script is text for provider TTS (secondary path). Providers that
    // require a voice read it from Extensions (e.g., "voice_id").
    Script string

    // Width, Height are the requested output dimensions (optional;
    // provider defaults apply when zero).
    Width, Height int

    // Background requests a background treatment (optional, best-effort;
    // see provider mapping table).
    Background *Background

    // Title is a human-readable job/video name (optional).
    Title string

    // Extensions holds provider-specific options
    // (e.g., "voice_id", "avatar_style", "test", "fast").
    Extensions map[string]any
}

// Background describes the requested video background.
type Background struct {
    // Type is "color", "image", or "video".
    Type string

    // Value is a hex color or URL depending on Type.
    Value string
}
```

### Job and State

```go
// JobState is the normalized lifecycle state of a generation job.
type JobState string

const (
    JobStatePending    JobState = "pending"
    JobStateProcessing JobState = "processing"
    JobStateCompleted  JobState = "completed"
    JobStateFailed     JobState = "failed"
)

// Terminal reports whether the state is final.
func (s JobState) Terminal() bool

// Job identifies a submitted generation job.
type Job struct {
    ID       string // provider job/video ID
    Provider string // provider name
}

// JobStatus is a point-in-time snapshot of a job.
type JobStatus struct {
    ID        string
    State     JobState
    RawStatus string  // provider-native status string, for logging/debugging
    VideoURL  string  // set when State == JobStateCompleted
    ThumbnailURL string // preview image, when the provider reports one (HeyGen)
    Duration  float64 // video duration in seconds, when reported
    ErrorCode string  // provider error code, when State == JobStateFailed
    ErrorMsg  string  // provider error message, when State == JobStateFailed
}

// Wait polls Status until the job reaches a terminal state, the context
// is cancelled, or an error occurs. Interval <= 0 defaults to 3s.
// A JobStateFailed result returns the final status AND an error wrapping
// ErrJobFailed.
func Wait(ctx context.Context, p Provider, jobID string, interval time.Duration) (*JobStatus, error)
```

### Errors

```go
var (
    // ErrAudioUploadUnsupported: provider cannot host audio; supply AudioURL.
    ErrAudioUploadUnsupported = errors.New("render: audio upload unsupported")

    // ErrInvalidRequest: request failed validation before submission
    // (e.g., neither or both of AudioURL/Script set, missing AvatarID).
    ErrInvalidRequest = errors.New("render: invalid request")

    // ErrJobNotFound: the provider does not recognize the job ID.
    ErrJobNotFound = errors.New("render: job not found")

    // ErrJobFailed: the job reached JobStateFailed.
    ErrJobFailed = errors.New("render: job failed")

    // ErrJobNotCompleted: Download called before successful completion.
    ErrJobNotCompleted = errors.New("render: job not completed")
)

// ProviderError mirrors live.ProviderError with a "render/" prefix.
type ProviderError struct {
    Provider string
    Op       string
    Err      error
}
```

## Part 3: Registry Symmetry

`omniavatar-core/registry` keeps the shared `ProviderConfig` / `ProviderOption` / `With*` helpers and gains one factory type per surface:

```go
type LiveProviderFactory   func(config ProviderConfig) (live.Provider, error)
type RenderProviderFactory func(config ProviderConfig) (render.Provider, error)
```

`omniavatar` (root package) hosts two independent registries with a symmetric API. The old `*AvatarProvider*` names are removed (breaking, v0.x):

| Old (removed) | Live surface | Render surface |
|---|---|---|
| `RegisterAvatarProvider` | `RegisterLiveProvider` | `RegisterRenderProvider` |
| `GetAvatarProvider` | `GetLiveProvider` | `GetRenderProvider` |
| `ListAvatarProviders` | `ListLiveProviders` | `ListRenderProviders` |
| `HasAvatarProvider` | `HasLiveProvider` | `HasRenderProvider` |
| `GetAvatarProviderPriority` | `GetLiveProviderPriority` | `GetRenderProviderPriority` |

`PriorityThin` / `PriorityThick` apply to both registries. Provider packages register both surfaces in their `register.go` `init()`.

## Part 4: Provider Mappings

### Normalized state mapping

| `render.JobState` | HeyGen | Tavus | bitHuman |
|---|---|---|---|
| `pending` | `pending` | `queued` | `pending` |
| `processing` | `processing` | `generating` | `processing` |
| `completed` | `completed` | `ready` | `completed` |
| `failed` | `failed` | `error`, `deleted` | `failed` |

`RawStatus` always carries the provider-native string.

### HeyGen (`providers/heygen/render.go`)

- SDK: `heygen-go` `heygen.NewClient` + `video.NewClient` (base URL `api.heygen.com`; note this is the **HeyGen** API key, distinct from the LiveAvatar key used by the live provider).
- Generate: one `video.VideoInput` with `Character{Type: "avatar"|"talking_photo", AvatarID|TalkingPhotoID, AvatarStyle}` and `VoiceInput{Type: "audio", AudioURL}` (or `Type: "text", InputText, VoiceID`). `Dimension` from Width/Height; `Background` maps natively (color/image/video); `Test` from extension.
- Status: `GetStatus(videoID)` → `Video{Status, VideoURL, Duration, Error}`.
- Download: HTTP GET of `VideoURL` (time-limited signed URL) streamed to `dst`; status is re-fetched immediately before download.
- **Implements `AudioUploader`** via `heygen-go/asset` (`upload.heygen.com/v1/asset`, raw-body POST with the file's MIME type; MP3/`audio/mpeg` is the documented audio asset type). The upload host is separate from the API host and overridable via the `upload_base_url` extension.
- Extensions: `talking_photo_id`, `avatar_style`, `voice_id`, `test`, `upload_base_url`.

### Tavus (`providers/tavus/render.go`)

- SDK: `tavus-go` (ogen-generated). `CreateVideo(&api.CreateVideoRequest{ReplicaID, AudioURL|Script, VideoName, ...})` → `CreateVideoResponse{VideoID}`; `GetVideo(api.GetVideoParams{...})` → `Video{Status, DownloadURL, StreamURL, HostedURL}`.
- `AvatarID` maps to `replica_id`.
- Background: `background_url` (website recording) / `background_source_url` (video file) via `Background{Type: "video"|"url"}` best-effort; no color backgrounds.
- Download: prefer `DownloadURL`, fall back to `StreamURL`.
- `AudioUploader`: **not implemented** (Tavus has no upload API). Callers get `AudioURL`-only behavior; the capability check fails cleanly.
- Extensions: `fast`, `transparent_background`, `callback_url`.
- ogen note: responses are sum types; implementations type-switch on the success variant and convert typed error variants into `render.ProviderError` wrapping core sentinels.

### bitHuman (`providers/bithuman/render.go`)

- SDK: `bithuman-go` (ogen-generated). `CreateVideo(&api.CreateVideoRequest{AgentID, AudioURL|Text+VoiceID})` → `VideoJob`; `GetVideo(api.GetVideoParams{...})` → `VideoJob{Status, VideoURL, DurationSeconds, Error}`.
- `AvatarID` maps to `agent_id`.
- **Implements `AudioUploader`**: `UploadFile(&api.UploadFileRequest{Base64, Filename, ContentType})` → `File{URL}`. This provider validates the capability interface design.
- Width/Height/Background: not supported by the API; documented as ignored.
- Extensions: `voice_id`.

## Part 5: Consumer Flow (videoascode, informative)

```go
p, err := omniavatar.GetRenderProvider("bithuman",
    omniavatar.WithAPIKey(os.Getenv("BITHUMAN_API_KEY")))

audioURL := cfg.AudioURL
if up, ok := p.(render.AudioUploader); ok && audioURL == "" {
    audioURL, err = up.UploadAudio(ctx, "narration.mp3", f)
}

job, err := p.Generate(ctx, render.GenerateRequest{
    AvatarID: cfg.AvatarID,
    AudioURL: audioURL,
})
status, err := render.Wait(ctx, p, job.ID, 5*time.Second)
err = p.Download(ctx, job.ID, presenterFile)
```

Circle masking and overlay composition remain in `videoascode` (FFmpeg), per the internal ideation notes (IDEATION_CHAT_PRESENTATION.md, untracked).

## Testing

- `omniavatar-core/render`: unit tests for `JobState.Terminal`, `Wait` (fake provider: completion, failure, context cancellation, default interval), request validation helper if exported.
- `omniavatar`: registry tests for both surfaces (registration, priority override, get/list/has); per-provider state-mapping unit tests. Live API calls are out of scope for unit tests.

## Breaking Changes (release notes)

- `omniavatar-core` v0.2.0: package `avatar` renamed to `live`; `registry.ProviderFactory` renamed to `registry.LiveProviderFactory`; new package `render`.
- `omniavatar` v0.2.0: all `*AvatarProvider*` registry functions renamed to `*LiveProvider*`; new `*RenderProvider*` registry API; new render providers for heygen, tavus, bithuman.
