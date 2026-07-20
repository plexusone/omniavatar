// Command render-basic generates a talking-head video from narration
// audio using any registered render provider.
//
// Usage:
//
//	export HEYGEN_API_KEY=...       # or TAVUS_API_KEY / BITHUMAN_API_KEY
//	go run . -provider heygen -avatar <avatar-id> -audio narration.mp3 -out presenter.mp4
//
// The -audio flag accepts either a local file (uploaded automatically
// when the provider supports it) or an https:// URL.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/plexusone/omniavatar"
	"github.com/plexusone/omniavatar-core/render"
	_ "github.com/plexusone/omniavatar/providers/all"
)

func main() {
	provider := flag.String("provider", "heygen", "render provider: heygen, tavus, or bithuman")
	avatarID := flag.String("avatar", "", "avatar identity (heygen avatar_id / tavus replica_id / bithuman agent_id)")
	audio := flag.String("audio", "", "narration audio: local file path or https:// URL")
	out := flag.String("out", "presenter.mp4", "output video path")
	interval := flag.Duration("interval", 5*time.Second, "status poll interval")
	flag.Parse()

	if *avatarID == "" || *audio == "" {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(context.Background(), *provider, *avatarID, *audio, *out, *interval); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, providerName, avatarID, audio, out string, interval time.Duration) error {
	apiKeyEnv := map[string]string{
		"heygen":   "HEYGEN_API_KEY",
		"tavus":    "TAVUS_API_KEY",
		"bithuman": "BITHUMAN_API_KEY",
	}[providerName]
	if apiKeyEnv == "" {
		return fmt.Errorf("unknown provider %q (available: %v)", providerName, omniavatar.ListRenderProviders())
	}

	provider, err := omniavatar.GetRenderProvider(providerName,
		omniavatar.WithAPIKey(os.Getenv(apiKeyEnv)))
	if err != nil {
		return err
	}

	audioURL, err := resolveAudio(ctx, provider, audio)
	if err != nil {
		return err
	}

	log.Printf("submitting generation job to %s", provider.Name())
	job, err := provider.Generate(ctx, render.GenerateRequest{
		AvatarID: avatarID,
		AudioURL: audioURL,
	})
	if err != nil {
		return err
	}
	log.Printf("job %s submitted; polling every %s", job.ID, interval)

	status, err := render.Wait(ctx, provider, job.ID, interval)
	if err != nil {
		return err
	}
	log.Printf("video ready (%.1fs)", status.Duration)
	if status.ThumbnailURL != "" {
		log.Printf("thumbnail: %s", status.ThumbnailURL)
	}

	f, err := os.Create(out) //nolint:gosec // G304: out is an operator-supplied CLI flag
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if err := provider.Download(ctx, job.ID, f); err != nil {
		return err
	}
	log.Printf("wrote %s", out)
	return nil
}

// resolveAudio returns a fetchable URL for the narration audio, uploading
// local files when the provider supports hosting.
func resolveAudio(ctx context.Context, provider render.Provider, audio string) (string, error) {
	if strings.HasPrefix(audio, "http://") || strings.HasPrefix(audio, "https://") {
		return audio, nil
	}

	uploader, ok := provider.(render.AudioUploader)
	if !ok {
		return "", fmt.Errorf("%w: provider %s cannot host local files; pass -audio as an https:// URL",
			render.ErrAudioUploadUnsupported, provider.Name())
	}

	f, err := os.Open(audio) //nolint:gosec // G304: audio is an operator-supplied CLI flag
	if err != nil {
		return "", err
	}
	defer func() {
		// Read-only file; close errors after a successful read are unactionable.
		_ = f.Close() //nolint:errcheck // see above
	}()

	log.Printf("uploading %s via %s", audio, provider.Name())
	return uploader.UploadAudio(ctx, audio, f)
}
