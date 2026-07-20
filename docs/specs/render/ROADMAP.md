# ROADMAP: OmniAvatar Render

Status: Living document
Date: 2026-07-16
Related: [PRD.md](PRD.md), [TRD.md](TRD.md), [PLAN.md](PLAN.md)

## v0.2.0 (current — see PLAN.md)

- Rename `omniavatar-core/avatar` → `omniavatar-core/live`.
- New `omniavatar-core/render` interfaces (Generate/Status/Download, `AudioUploader` capability, `Wait` helper).
- Render providers: HeyGen, Tavus, bitHuman.
- Audio upload: bitHuman (file API) and HeyGen (new `heygen-go/asset` package wrapping `upload.heygen.com/v1/asset`) implement `AudioUploader`; Tavus remains URL-only (no upload API).
- Symmetric registry API: `*LiveProvider*` / `*RenderProvider*`.

## v0.3.x — Audio delivery completeness

- Evaluate a generic S3/GCS-presigned-URL uploader helper for providers without hosting (Tavus), likely as a separate utility package so `omniavatar-core` stays dependency-free.
- Evaluate WAV support for HeyGen uploads (only MP3/`audio/mpeg` is documented as an audio asset type today); fall back to transcoding guidance in videoascode if HeyGen keeps the restriction.

## videoascode integration (UC1) — landed with v0.2.0 (videoascode v0.7.0)

Work lands in `grokify/videoascode` (consumer), tracked here for coordination:

- [x] Concatenate per-slide narration (including pause gaps) into one normalized MP3 before generation (`pkg/avatar.ConcatManifestAudio`).
- [x] `vac avatar generate` → `presenter.mp4`, cached by hash of (audio content + avatar config), matching the existing TTS-cache philosophy.
- [x] `vac avatar compose` — FFmpeg circular mask + corner overlay with optional border ring; narration audio remains the authoritative track (never the avatar video's audio).
- [x] Orchestrator integration: `--avatar-*` flags on `vac slides video` run the overlay as an integrated final stage (`orchestrator.AvatarConfig`). Upload-capable providers only (heygen, bithuman); Tavus stays on the decoupled `--audio-url` flow.
- [ ] Slide-level overrides (`visible: false`) driven by the timing manifest (fade avatar in/out).

## v0.5.x — Additional providers

- **D-ID** (`Create a talk` API: photo + audio/text) — validates the interface against a photo-first provider; likely requires a `d-id-go` SDK repo first.
- Local/self-hosted renderer (e.g., open-source lip-sync models) for private deployments — exercises the interface without any network provider.

## Later / investigating

- **HeyGen lipsync-precision path (Path B)**: neutral presenter source video + replacement audio for frame-accurate sync and full visual control (framing, wardrobe, camera). Requires wrapping the lipsync endpoint in heygen-go and adding a `SourceVideoURL` field or extension to `GenerateRequest`.
- **Webhook/callback completion**: all three providers support callbacks (HeyGen `callback_id`, Tavus `callback_url`, bitHuman webhooks). Add an optional capability interface so long jobs don't require polling.
- **Avatar normalization helpers**: standardize provider output (square crop, fixed FPS/codec/duration alignment) — possibly in videoascode rather than here, to keep omniavatar free of FFmpeg dependencies.
- **Transparent output**: Tavus supports `transparent_background` (WebM, fast mode); evaluate alpha-channel passthrough in `GenerateRequest.Background` once a second provider supports it.
- **Multi-language batches**: one render job per language from per-language narration, reusing videoascode's JSON transcript support.

## Non-goals (revisit only with new evidence)

- Real-time use of render providers (LiveAvatar/CVI already cover this via the `live` surface).
- Provider-side full-presentation rendering (slides stay a local, deterministic FFmpeg composition).
