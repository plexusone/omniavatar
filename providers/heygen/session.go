package heygen

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	lksdk "github.com/livekit/server-sdk-go/v2"

	"github.com/plexusone/heygen-go/liveavatar"

	"github.com/plexusone/omniavatar"
	"github.com/plexusone/omniavatar-core/live"
)

// SessionConfig configures a HeyGen LiveAvatar session.
type SessionConfig struct {
	// Client is the LiveAvatar API client.
	// Required.
	Client *liveavatar.Client

	// AvatarID is the UUID of the avatar to use.
	// Required.
	AvatarID string

	// Sandbox enables sandbox mode.
	Sandbox bool

	// VideoQuality sets the avatar video quality.
	VideoQuality liveavatar.VideoQuality

	// AudioConfig configures the audio format.
	AudioConfig live.AudioConfig
}

// Session implements live.Session for HeyGen LiveAvatar.
type Session struct {
	client       *liveavatar.Client
	avatarID     string
	sandbox      bool
	videoQuality liveavatar.VideoQuality
	audioConfig  live.AudioConfig

	// Session identity and state
	identity     string
	sessionID    string
	sessionToken string

	// Room reference (set during Start)
	room *lksdk.Room

	// Audio output (set externally or through opts)
	audioOutput live.AudioDestination

	// Callbacks
	callbacks *live.SessionCallbacks

	// State tracking
	started   bool
	startTime time.Time

	// Participant tracking
	participantJoined chan struct{}
	participantLeft   chan struct{}

	mu     sync.Mutex
	closed bool
}

// NewSession creates a new HeyGen LiveAvatar session.
func NewSession(cfg SessionConfig) (*Session, error) {
	if cfg.Client == nil {
		return nil, live.ErrInvalidConfig
	}
	if cfg.AvatarID == "" {
		return nil, live.ErrInvalidConfig
	}

	audioConfig := cfg.AudioConfig
	if audioConfig.SampleRate == 0 {
		audioConfig = live.DefaultAudioConfig()
	}

	videoQuality := cfg.VideoQuality
	if videoQuality == "" {
		videoQuality = liveavatar.QualityHigh
	}

	// Generate a unique avatar identity
	identity := fmt.Sprintf("heygen-avatar-%s", uuid.New().String()[:8])

	return &Session{
		client:            cfg.Client,
		avatarID:          cfg.AvatarID,
		sandbox:           cfg.Sandbox,
		videoQuality:      videoQuality,
		audioConfig:       audioConfig,
		identity:          identity,
		participantJoined: make(chan struct{}),
		participantLeft:   make(chan struct{}),
	}, nil
}

// Identity returns the avatar participant identity.
func (s *Session) Identity() string {
	return s.identity
}

// Provider returns the provider name.
func (s *Session) Provider() string {
	return "heygen"
}

// Start initializes the HeyGen LiveAvatar session.
//
// The opts parameter must be *omniavatar.LiveKitStartOptions.
func (s *Session) Start(ctx context.Context, opts any) error {
	lkOpts, ok := opts.(*omniavatar.LiveKitStartOptions)
	if !ok {
		return fmt.Errorf("heygen: expected *omniavatar.LiveKitStartOptions, got %T", opts)
	}

	if err := lkOpts.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return live.ErrSessionAlreadyStarted
	}
	s.mu.Unlock()

	// Store room reference and callbacks
	s.room = lkOpts.Room
	s.callbacks = lkOpts.Callbacks

	// Use provided audio destination if available
	if lkOpts.AudioDestination != nil {
		s.audioOutput = lkOpts.AudioDestination
	}

	// Generate token for avatar to join
	token, err := omniavatar.GenerateAvatarToken(omniavatar.TokenOptions{
		APIKey:          lkOpts.LiveKitAPIKey,
		APISecret:       lkOpts.LiveKitAPISecret,
		RoomName:        lkOpts.Room.Name(),
		AvatarIdentity:  s.identity,
		AvatarName:      "HeyGen Avatar",
		PublishOnBehalf: lkOpts.AgentIdentity,
		TTL:             10 * time.Minute,
		Metadata:        s.buildMetadata(lkOpts.AgentIdentity),
	})
	if err != nil {
		return fmt.Errorf("failed to generate avatar token: %w", err)
	}

	// Create session with LiveAvatar (LITE mode)
	sessionResp, err := s.client.NewSession(ctx, &liveavatar.NewSessionRequest{
		Mode:         "LITE",
		AvatarID:     s.avatarID,
		IsSandbox:    s.sandbox,
		VideoQuality: s.videoQuality,
		LiveKitConfig: &liveavatar.LiveKitConfig{
			LiveKitURL:         lkOpts.LiveKitURL,
			LiveKitRoom:        lkOpts.Room.Name(),
			LiveKitClientToken: token,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create LiveAvatar session: %w", err)
	}

	s.mu.Lock()
	s.sessionID = sessionResp.SessionID
	s.sessionToken = sessionResp.SessionToken
	s.started = true
	s.startTime = time.Now()
	s.mu.Unlock()

	// Start the session
	_, err = s.client.StartSession(ctx, sessionResp.SessionID, sessionResp.SessionToken)
	if err != nil {
		return fmt.Errorf("failed to start LiveAvatar session: %w", err)
	}

	return nil
}

// WaitForJoin blocks until the avatar participant joins the room.
func (s *Session) WaitForJoin(ctx context.Context, timeout time.Duration) error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return live.ErrSessionNotStarted
	}
	room := s.room
	s.mu.Unlock()

	if room == nil {
		return live.ErrSessionNotStarted
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Poll for participant join
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		// Check if avatar has joined
		if p := room.GetParticipantByIdentity(s.identity); p != nil {
			s.emitJoinMetrics()

			// Notify via callback if set
			if s.callbacks != nil && s.callbacks.OnAvatarJoined != nil {
				s.callbacks.OnAvatarJoined(p.Identity())
			}

			// Signal join for any waiters
			select {
			case <-s.participantJoined:
				// Already closed
			default:
				close(s.participantJoined)
			}

			return nil
		}

		select {
		case <-ticker.C:
			// Continue polling
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return live.ErrAvatarJoinTimeout
			}
			return ctx.Err()
		}
	}
}

// AudioOutput returns the audio destination for streaming TTS audio.
func (s *Session) AudioOutput() live.AudioDestination {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.audioOutput
}

// SetAudioOutput sets the audio destination.
func (s *Session) SetAudioOutput(out live.AudioDestination) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.audioOutput = out
}

// Close disconnects the avatar and cleans up resources.
func (s *Session) Close(ctx context.Context) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	sessionID := s.sessionID
	sessionToken := s.sessionToken
	audioOut := s.audioOutput
	s.mu.Unlock()

	var errs []error

	// Close audio output
	if audioOut != nil {
		if err := audioOut.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close audio output: %w", err))
		}
	}

	// Stop session with LiveAvatar
	if sessionID != "" && sessionToken != "" {
		if err := s.client.StopSession(ctx, sessionID, sessionToken, liveavatar.StopReasonSessionEnded); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop LiveAvatar session: %w", err))
		}
	}

	// Close channels
	select {
	case <-s.participantLeft:
		// Already closed
	default:
		close(s.participantLeft)
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// SetCallbacks registers event callbacks for the session.
func (s *Session) SetCallbacks(callbacks *live.SessionCallbacks) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callbacks = callbacks
}

// EmitPlaybackStarted emits a playback started event.
func (s *Session) EmitPlaybackStarted() {
	s.mu.Lock()
	cb := s.callbacks
	s.mu.Unlock()

	if cb != nil && cb.OnPlaybackStarted != nil {
		cb.OnPlaybackStarted()
	}
}

// EmitPlaybackFinished emits a playback finished event.
func (s *Session) EmitPlaybackFinished(position float64, interrupted bool) {
	s.mu.Lock()
	cb := s.callbacks
	s.mu.Unlock()

	if cb != nil && cb.OnPlaybackFinished != nil {
		cb.OnPlaybackFinished(position, interrupted)
	}
}

// EmitError emits an error event.
func (s *Session) EmitError(err error) {
	s.mu.Lock()
	cb := s.callbacks
	s.mu.Unlock()

	if cb != nil && cb.OnError != nil {
		cb.OnError(err)
	}
}

// buildMetadata creates the avatar participant metadata.
func (s *Session) buildMetadata(agentIdentity string) string {
	meta := omniavatar.DefaultAvatarMetadata("heygen", agentIdentity)
	data, _ := json.Marshal(meta)
	return string(data)
}

// emitJoinMetrics emits metrics about avatar join latency.
func (s *Session) emitJoinMetrics() {
	// Metrics emission could be added here if needed
}

// SessionID returns the LiveAvatar session ID.
func (s *Session) SessionID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessionID
}

// Room returns the LiveKit room reference.
func (s *Session) Room() *lksdk.Room {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.room
}

// Verify interface compliance at compile time.
var _ live.Session = (*Session)(nil)
