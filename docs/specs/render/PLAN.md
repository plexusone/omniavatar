# PLAN: Rename `avatar/` → `live/` and Add `render/` Across 3 Providers

Status: Implemented (2026-07-16); pending release sequencing below
Date: 2026-07-16
Related: [PRD.md](PRD.md), [TRD.md](TRD.md), [ROADMAP.md](ROADMAP.md)

## Scope

Three repos change in lockstep:

1. `plexusone/omniavatar-core` — rename `avatar/` → `live/`, add `render/`, extend `registry/`.
2. `plexusone/heygen-go` — add `asset` upload package (prerequisite for HeyGen `AudioUploader`); fix retry body rewind.
3. `plexusone/omniavatar` — adopt the rename, make the registry API symmetric, add render providers for heygen, tavus, bithuman (heygen and bithuman with audio upload).

Also landed alongside this plan (see [ROADMAP.md](ROADMAP.md)): consumer integration in `grokify/videoascode` v0.7.0 (`vac avatar generate` / `vac avatar compose`), which validated the render interfaces from the consumer side before tagging.

Out of scope (see [ROADMAP.md](ROADMAP.md)): D-ID, webhooks, orchestrator-integrated avatar pipeline in videoascode.

## Phase 1: omniavatar-core

| # | Task | Notes |
|---|------|-------|
| 1.1 | `git mv avatar live`; `package avatar` → `package live` | Types/functions keep their names |
| 1.2 | Error prefixes `"avatar:"` → `"live:"`, `ProviderError` prefix `"avatar/"` → `"live/"` | |
| 1.3 | New `render/` package: `provider.go`, `request.go`, `job.go`, `errors.go` | Interfaces per TRD Part 2 |
| 1.4 | `registry/`: `ProviderFactory` → `LiveProviderFactory`; add `RenderProviderFactory` | Shared `ProviderConfig` unchanged |
| 1.5 | Unit tests: `render` (`Wait`, `Terminal`, validation) | Fake provider, no network |
| 1.6 | Update `doc.go`, `README.md` | Document both surfaces |
| 1.7 | Release notes `docs/releases/v0.2.0.md`; update `CHANGELOG.json` | Breaking-change notes per TRD |

## Phase 1b: heygen-go

| # | Task | Notes |
|---|------|-------|
| 1b.1 | `asset` package: `Upload` against `upload.heygen.com/v1/asset` | Raw-body POST with file MIME type |
| 1b.2 | `heygen.Client.RequestURL` raw-body/absolute-URL request helper | Shared by Request (JSON) and asset upload |
| 1b.3 | Fix retry body rewind via `GetBody` | 429-retried POSTs previously resent an empty body |
| 1b.4 | `Client.Asset` facade field; README, changelog, release notes | |

## Phase 2: omniavatar

| # | Task | Notes |
|---|------|-------|
| 2.1 | Update all imports `omniavatar-core/avatar` → `.../live` | registry.go, providers/* |
| 2.2 | Registry symmetry: rename `*AvatarProvider*` → `*LiveProvider*`; add `*RenderProvider*` | Per TRD Part 3 |
| 2.3 | `providers/bithuman/render.go` — render provider + `AudioUploader` | First: proves the upload capability |
| 2.4 | `providers/tavus/render.go` — render provider (URL-only audio) | Proves the no-upload path |
| 2.5 | `providers/heygen/render.go` — render provider via `heygen-go/video` | Includes `AudioUploader` via `heygen-go/asset` (added in heygen-go alongside this work) |
| 2.6 | Register render providers in each `register.go` `init()` | Both surfaces per provider |
| 2.7 | Unit tests: registry surfaces, per-provider state mapping | |
| 2.8 | Update `doc.go`, `README.md` | Quick-start for both surfaces |
| 2.9 | Release notes `docs/releases/v0.2.0.md`; update `CHANGELOG.json` | |

## Phase 3: Verification

| # | Check | Command |
|---|-------|---------|
| 3.1 | Local cross-module build | `go.work` (untracked) spanning both repos |
| 3.2 | Format | `gofmt -l .` clean in both repos |
| 3.3 | Vet + tests | `go vet ./...`, `go test ./...` in both repos |
| 3.4 | Lint | `golangci-lint run` in both repos |
| 3.5 | No local `replace` directives in either `go.mod` | Pre-push checklist |

## Sequencing and Release

1. Land + tag `omniavatar-core v0.2.0` (rename + render interfaces) and `heygen-go v0.2.0` (asset upload) — independent, any order.
2. Bump `omniavatar` to `omniavatar-core v0.2.0` and `heygen-go v0.2.0`, land providers, tag `omniavatar v0.2.0`.
3. Bump `videoascode` go.mod from the v0.1.0 placeholders to `omniavatar v0.2.0` / `omniavatar-core v0.2.0`, run `go mod tidy`, tag `videoascode v0.7.0`.
4. Per house rules: push commits, wait for CI green, then tag.

Local development uses `go.work` files (not committed) spanning the repos instead of `replace` directives. Note: `videoascode` CI will not pass until step 2's tags exist, because its go.mod pins v0.1.0 while the code needs the workspace-supplied v0.2.0 interfaces.

## Commit Plan (conventional commits)

omniavatar-core:

```
refactor!: rename avatar package to live
feat: add render package for batch avatar video generation
feat(registry): add RenderProviderFactory, rename ProviderFactory to LiveProviderFactory
test: add render package unit tests
docs: update README and doc.go for live/render surfaces
docs: add v0.2.0 release notes and changelog
```

heygen-go:

```
fix(heygen): rewind request body on retry
feat(heygen): add RequestURL for raw-body absolute-URL requests
feat(asset): add asset upload client for upload.heygen.com
docs: add v0.2.0 release notes and changelog
```

omniavatar:

```
refactor!: adopt omniavatar-core live package rename; symmetric registry API
feat(bithuman): add render provider with audio upload support
feat(tavus): add render provider
feat(heygen): add render provider with audio upload support
test: add registry and render provider unit tests
docs: update README and doc.go for live/render surfaces
docs: add v0.2.0 release notes and changelog; add render specs
docs: add MkDocs site with PlexusOne theme
ci: add Docs workflow deploying via mkdocs gh-deploy
```

videoascode:

```
chore(deps): add omniavatar and omniavatar-core dependencies
feat(avatar): add narration concat and presenter generation via OmniAvatar
feat(video): add circular avatar overlay compositor
feat(cli): add vac avatar generate and compose commands
test: add avatar and overlay unit tests
docs: document avatar presenter overlay; update changelog
```

## Risks

| Risk | Mitigation |
|------|------------|
| Interface churn after freezing | Three concurrent implementations before tagging (rule of three) |
| ogen sum-type responses awkward to consume | Type-switch on success variant; typed error variants → `render.ProviderError` |
| HeyGen signed `VideoURL` expiry | `Download` re-fetches status immediately before GET |
| Provider status enums drift | `RawStatus` passthrough; unknown states map to `processing` (non-terminal, safe for pollers) |
