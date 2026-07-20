package omniavatar

import (
	"context"
	"io"
	"slices"
	"testing"

	"github.com/plexusone/omniavatar-core/live"
	"github.com/plexusone/omniavatar-core/render"
)

type fakeLiveProvider struct{ name string }

func (f *fakeLiveProvider) Name() string { return f.name }
func (f *fakeLiveProvider) CreateSession(_ live.SessionConfig) (live.Session, error) {
	return nil, live.ErrInvalidConfig
}

type fakeRenderProvider struct{ name string }

func (f *fakeRenderProvider) Name() string { return f.name }
func (f *fakeRenderProvider) Generate(_ context.Context, _ render.GenerateRequest) (*render.Job, error) {
	return nil, render.ErrInvalidRequest
}
func (f *fakeRenderProvider) Status(_ context.Context, _ string) (*render.JobStatus, error) {
	return nil, render.ErrJobNotFound
}
func (f *fakeRenderProvider) Download(_ context.Context, _ string, _ io.Writer) error {
	return render.ErrJobNotCompleted
}

func TestLiveRegistry(t *testing.T) {
	name := "test-live"
	RegisterLiveProvider(name, func(_ ProviderConfig) (live.Provider, error) {
		return &fakeLiveProvider{name: name}, nil
	}, PriorityThin)

	if !HasLiveProvider(name) {
		t.Fatalf("HasLiveProvider(%q) = false, want true", name)
	}
	if !slices.Contains(ListLiveProviders(), name) {
		t.Errorf("ListLiveProviders() missing %q", name)
	}
	if got := GetLiveProviderPriority(name); got != PriorityThin {
		t.Errorf("GetLiveProviderPriority(%q) = %d, want %d", name, got, PriorityThin)
	}

	p, err := GetLiveProvider(name)
	if err != nil {
		t.Fatalf("GetLiveProvider(%q) error = %v", name, err)
	}
	if p.Name() != name {
		t.Errorf("provider.Name() = %q, want %q", p.Name(), name)
	}

	if _, err := GetLiveProvider("nonexistent-live"); err == nil {
		t.Error("GetLiveProvider(nonexistent) error = nil, want error")
	}
}

func TestRenderRegistry(t *testing.T) {
	name := "test-render"
	RegisterRenderProvider(name, func(_ ProviderConfig) (render.Provider, error) {
		return &fakeRenderProvider{name: name}, nil
	}, PriorityThin)

	if !HasRenderProvider(name) {
		t.Fatalf("HasRenderProvider(%q) = false, want true", name)
	}
	if !slices.Contains(ListRenderProviders(), name) {
		t.Errorf("ListRenderProviders() missing %q", name)
	}
	if got := GetRenderProviderPriority(name); got != PriorityThin {
		t.Errorf("GetRenderProviderPriority(%q) = %d, want %d", name, got, PriorityThin)
	}

	p, err := GetRenderProvider(name)
	if err != nil {
		t.Fatalf("GetRenderProvider(%q) error = %v", name, err)
	}
	if p.Name() != name {
		t.Errorf("provider.Name() = %q, want %q", p.Name(), name)
	}

	if _, err := GetRenderProvider("nonexistent-render"); err == nil {
		t.Error("GetRenderProvider(nonexistent) error = nil, want error")
	}
}

func TestRegistriesAreIndependent(t *testing.T) {
	name := "test-live-only"
	RegisterLiveProvider(name, func(_ ProviderConfig) (live.Provider, error) {
		return &fakeLiveProvider{name: name}, nil
	}, PriorityThin)

	if HasRenderProvider(name) {
		t.Errorf("HasRenderProvider(%q) = true, want false (live-only registration)", name)
	}
}

func TestRenderPriorityOverride(t *testing.T) {
	name := "test-render-priority"
	RegisterRenderProvider(name, func(_ ProviderConfig) (render.Provider, error) {
		return &fakeRenderProvider{name: "thin"}, nil
	}, PriorityThin)
	RegisterRenderProvider(name, func(_ ProviderConfig) (render.Provider, error) {
		return &fakeRenderProvider{name: "thick"}, nil
	}, PriorityThick)

	p, err := GetRenderProvider(name)
	if err != nil {
		t.Fatalf("GetRenderProvider(%q) error = %v", name, err)
	}
	if p.Name() != "thick" {
		t.Errorf("provider.Name() = %q, want %q (thick overrides thin)", p.Name(), "thick")
	}

	// Lower priority must not override higher.
	RegisterRenderProvider(name, func(_ ProviderConfig) (render.Provider, error) {
		return &fakeRenderProvider{name: "thin-again"}, nil
	}, PriorityThin)

	p, err = GetRenderProvider(name)
	if err != nil {
		t.Fatalf("GetRenderProvider(%q) error = %v", name, err)
	}
	if p.Name() != "thick" {
		t.Errorf("provider.Name() = %q, want %q (thin must not override thick)", p.Name(), "thick")
	}
}
