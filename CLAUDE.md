# CLAUDE.md

Agent guidelines for omniavatar (batteries-included package).

## Project Overview

Batteries-included AI avatar package with provider implementations. Re-exports
`omniavatar-core` interfaces and adds:

- **Provider registration** — `RegisterLiveProvider`, `RegisterRenderProvider`
- **LiveKit integration** — `GenerateAvatarToken`, `LiveKitStartOptions`
- **Provider adapters** — HeyGen, Tavus, bitHuman, LivePortrait+JoyVASA (local)

## Architecture

```
omniavatar-core/           # Interfaces (no provider deps)
├── live/                  # Session interface
├── render/                # Provider interface
└── registry/              # Factory types

omniavatar/                # This package (batteries)
├── registry.go            # Global registries
├── providers/
│   ├── heygen/            # Live + render (SDK render adapter)
│   ├── tavus/             # Live + render
│   ├── bithuman/          # Live + render
│   ├── liveportrait-joyvasa/  # Render only (local)
│   └── all/               # Convenience import
└── token.go               # LiveKit token generation
```

## Provider Registration Pattern

Providers register in `init()`:

```go
// providers/liveportrait-joyvasa/register.go
func init() {
    omniavatar.RegisterRenderProvider("liveportrait-joyvasa",
        lp.NewFromConfig, omniavatar.PriorityThick)
}
```

Include in `providers/all/all.go` for auto-discovery.

## Key Files

| Path | Purpose |
|------|---------|
| `registry.go` | Live + render registries, priority system |
| `token.go` | LiveKit token generation |
| `start_options.go` | `LiveKitStartOptions` struct |
| `providers/*/register.go` | Provider registration |
| `providers/all/all.go` | Convenience import |

## Adding a New Provider

1. Create `providers/<name>/register.go`
2. Implement `init()` calling `RegisterLiveProvider` and/or `RegisterRenderProvider`
3. Add import to `providers/all/all.go`
4. Update README with provider docs

For local providers (no API key), the factory should work with empty `cfg.APIKey`.

## Testing

```bash
go test -v ./...
golangci-lint run
```

## Dependencies

- `omniavatar-core` — interfaces and local provider implementation
- Provider SDKs — `heygen-go`, `tavus-go`, `bithuman-go`
- `livekit/server-sdk-go` — LiveKit integration

## Common Tasks

| Task | Command |
|------|---------|
| Build | `go build ./...` |
| Test | `go test -v ./...` |
| Lint | `golangci-lint run` |
| Update core | `go get github.com/plexusone/omniavatar-core@latest` |

## Consumer Usage

```go
import (
    "github.com/plexusone/omniavatar"
    _ "github.com/plexusone/omniavatar/providers/all"
)

// Cloud provider (needs API key)
provider, _ := omniavatar.GetRenderProvider("heygen",
    omniavatar.WithAPIKey(os.Getenv("HEYGEN_API_KEY")))

// Local provider (no API key)
provider, _ := omniavatar.GetRenderProvider("liveportrait-joyvasa")
```
