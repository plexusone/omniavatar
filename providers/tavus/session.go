package tavus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	lksdk "github.com/livekit/server-sdk-go/v2"

	"github.com/plexusone/omniavatar"
	"github.com/plexusone/omniavatar-core/avatar"
)

// SessionConfig configures a Tavus avatar session.
type SessionConfig struct {
	// Client is the Tavus API client.
	// Required.
	Client *Client

	// PalID is the PAL (Personalized AI Likeness) to use.
	// Required.
	PalID string

	// FaceID is an optional face override.
	FaceID string

	// AudioConfig configures the audio format.
	AudioConfig avatar.AudioConfig
}

// Session implements avatar.Session for Tavus avatars.
type Session struct {
	client      *Client
	palID       string
	faceID      string
	audioConfig avatar.AudioConfig

	// Session identity and state
	identity       string
	conversationID string

	// Room reference (set during Start)
	room *lksdk.Room

	// Audio output (set externally or through opts)
	audioOutput avatar.AudioDestination

	// Callbacks
	callbacks *avatar.SessionCallbacks

	// State tracking
	started   bool
	startTime time.Time

	// Participant tracking
	participantJoined chan struct{}
	participantLeft   chan struct{}

	mu     sync.Mutex
	closed bool
}

// NewSession creates a new Tavus avatar session.
func NewSession(cfg SessionConfig) (*Session, error) {
	if cfg.Client == nil {
		return nil, avatar.ErrInvalidConfig
	}

	palID := cfg.PalID
	if palID == "" {
		palID = DefaultPalID
	}

	audioConfig := cfg.AudioConfig
	if audioConfig.SampleRate == 0 {
		audioConfig = avatar.DefaultAudioConfig()
	}

	// Generate a unique avatar identity
	identity := fmt.Sprintf("tavus-avatar-%s", uuid.New().String()[:8])

	return &Session{
		client:            cfg.Client,
		palID:             palID,
		faceID:            cfg.FaceID,
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
	return "tavus"
}

// Start initializes the Tavus avatar session.
//
// The opts parameter must be *omniavatar.LiveKitStartOptions.
func (s *Session) Start(ctx context.Context, opts any) error {
	lkOpts, ok := opts.(*omniavatar.LiveKitStartOptions)
	if !ok {
		return fmt.Errorf("tavus: expected *omniavatar.LiveKitStartOptions, got %T", opts)
	}

	if err := lkOpts.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return avatar.ErrSessionAlreadyStarted
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
		AvatarName:      "Tavus Avatar",
		PublishOnBehalf: lkOpts.AgentIdentity,
		TTL:             10 * time.Minute,
		Metadata:        s.buildMetadata(lkOpts.AgentIdentity),
	})
	if err != nil {
		return fmt.Errorf("failed to generate avatar token: %w", err)
	}

	// Create conversation with Tavus
	conversationResp, err := s.client.CreateConversation(ctx, CreateConversationRequest{
		PalID:            s.palID,
		FaceID:           s.faceID,
		LiveKitURL:       lkOpts.LiveKitURL,
		LiveKitToken:     token,
		ConversationName: fmt.Sprintf("lk_conversation_%s", uuid.New().String()[:8]),
	})
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}

	s.mu.Lock()
	s.conversationID = conversationResp.ConversationID
	s.started = true
	s.startTime = time.Now()
	s.mu.Unlock()

	return nil
}

// WaitForJoin blocks until the avatar participant joins the room.
func (s *Session) WaitForJoin(ctx context.Context, timeout time.Duration) error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return avatar.ErrSessionNotStarted
	}
	room := s.room
	s.mu.Unlock()

	if room == nil {
		return avatar.ErrSessionNotStarted
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Poll for participant join
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		// Check if avatar has joined
		if p := room.GetParticipantByIdentity(s.identity); p != nil {
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
				return avatar.ErrAvatarJoinTimeout
			}
			return ctx.Err()
		}
	}
}

// AudioOutput returns the audio destination for streaming TTS audio.
func (s *Session) AudioOutput() avatar.AudioDestination {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.audioOutput
}

// SetAudioOutput sets the audio destination.
func (s *Session) SetAudioOutput(out avatar.AudioDestination) {
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
	conversationID := s.conversationID
	audioOut := s.audioOutput
	s.mu.Unlock()

	var errs []error

	// Close audio output
	if audioOut != nil {
		if err := audioOut.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close audio output: %w", err))
		}
	}

	// End conversation with Tavus
	if conversationID != "" {
		if err := s.client.EndConversation(ctx, conversationID); err != nil {
			errs = append(errs, fmt.Errorf("failed to end conversation: %w", err))
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
func (s *Session) SetCallbacks(callbacks *avatar.SessionCallbacks) {
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
	meta := omniavatar.DefaultAvatarMetadata("tavus", agentIdentity)
	data, _ := json.Marshal(meta)
	return string(data)
}

// ConversationID returns the Tavus conversation ID.
func (s *Session) ConversationID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conversationID
}

// Room returns the LiveKit room reference.
func (s *Session) Room() *lksdk.Room {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.room
}

// Verify interface compliance at compile time.
var _ avatar.Session = (*Session)(nil)
