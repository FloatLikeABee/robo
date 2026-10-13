package grill

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

const (
	PhaseAsking = "asking"
	PhaseReview = "review"

	// MaxPerOwner is the in-memory ceiling.
	// ponytail: dropped on process exit and not shared across processes.
	// Upgrade path: a user_grill table if an interview must survive a restart.
	MaxPerOwner = 32
)

var (
	ErrNotFound       = errors.New("grill not found")
	ErrTitleRequired  = errors.New("title is required")
	ErrAnswerRequired = errors.New("answer is required")
	ErrKind           = errors.New("invalid kind")
	ErrAction         = errors.New("invalid action")
	ErrTooLong        = errors.New("text is too long")
)

// Completer is one model call. Tests pass a function. Production uses the Morph AI client.
type Completer func(ctx context.Context, prompt string) (string, error)

// Turn is one interview step. A skip stores no answer.
type Turn struct {
	Question string
	Answer   string
	Skipped  bool
}

// Session is one interview for one owner.
type Session struct {
	ID          string
	Owner       string
	Kind        string
	Title       string
	Description string
	Questions   []string
	Turns       []Turn
	Index       int
	Phase       string
	Draft       string
	Created     time.Time
}

// View is the JSON the SPA and the MCP tool return.
type View struct {
	ID       string `json:"id"`
	Phase    string `json:"phase"`
	Kind     string `json:"kind"`
	Title    string `json:"title"`
	Question string `json:"question,omitempty"`
	Asked    int    `json:"asked"`
	Planned  int    `json:"planned"`
	Draft    string `json:"draft,omitempty"`
}

// Input is one grill action.
type Input struct {
	Action      string
	ID          string
	Kind        string
	Title       string
	Description string
	Answer      string
}

// Store holds live interviews. It does not write SQLite.
type Store struct {
	mu       sync.Mutex
	sessions map[string]*Session
	order    map[string][]string
}

func NewStore() *Store {
	return &Store{
		sessions: map[string]*Session{},
		order:    map[string][]string{},
	}
}

// Act runs start, answer, skip, or finish for owner.
func (s *Store) Act(ctx context.Context, owner string, in Input, complete Completer) (View, error) {
	if s == nil || strings.TrimSpace(owner) == "" {
		return View{}, ErrNotFound
	}
	switch strings.ToLower(strings.TrimSpace(in.Action)) {
	case "start":
		return s.start(ctx, owner, in, complete)
	case "answer":
		return s.answer(ctx, owner, in, complete)
	case "skip":
		return s.skip(ctx, owner, in, complete)
	case "finish":
		return s.finish(ctx, owner, strings.TrimSpace(in.ID), complete)
	default:
		return View{}, ErrAction
	}
}

func (s *Store) start(ctx context.Context, owner string, in Input, complete Completer) (View, error) {
	kind := strings.ToLower(strings.TrimSpace(in.Kind))
	if !ValidKind(kind) {
		return View{}, ErrKind
	}
	title := strings.Join(strings.Fields(in.Title), " ")
	if title == "" {
		return View{}, ErrTitleRequired
	}
	if overRunes(title, MaxTitleRunes) || overRunes(in.Description, MaxDescriptionRunes) {
		return View{}, ErrTooLong
	}
	description := strings.TrimSpace(in.Description)
	if complete == nil {
		return View{}, errors.New("AI service not configured")
	}
	n := QuestionCap(title)
	raw, err := complete(ctx, PlanPrompt(kind, title, description, n))
	if err != nil {
		return View{}, err
	}
	id, err := newID()
	if err != nil {
		return View{}, err
	}
	sess := &Session{
		ID:          id,
		Owner:       owner,
		Kind:        kind,
		Title:       title,
		Description: description,
		Questions:   ParseQuestions(raw, kind, title, n),
		Phase:       PhaseAsking,
		Created:     time.Now().UTC(),
	}
	s.mu.Lock()
	s.remember(sess)
	view := sess.view()
	s.mu.Unlock()
	return view, nil
}

func (s *Store) answer(ctx context.Context, owner string, in Input, complete Completer) (View, error) {
	s.mu.Lock()
	sess, err := s.get(owner, strings.TrimSpace(in.ID))
	if err != nil {
		s.mu.Unlock()
		return View{}, err
	}
	if sess.Phase == PhaseReview {
		view := sess.view()
		s.mu.Unlock()
		return view, nil
	}
	answer := strings.TrimSpace(in.Answer)
	if answer == "" {
		s.mu.Unlock()
		return View{}, ErrAnswerRequired
	}
	if overRunes(answer, MaxAnswerRunes) {
		s.mu.Unlock()
		return View{}, ErrTooLong
	}
	if IsStopPhrase(answer) {
		id := sess.ID
		s.mu.Unlock()
		return s.finish(ctx, owner, id, complete)
	}
	if sess.Index >= len(sess.Questions) {
		s.mu.Unlock()
		return s.finish(ctx, owner, sess.ID, complete)
	}
	sess.Turns = append(sess.Turns, Turn{Question: sess.Questions[sess.Index], Answer: answer})
	sess.Index++
	done := sess.Index >= len(sess.Questions)
	if !done {
		view := sess.view()
		s.mu.Unlock()
		return view, nil
	}
	s.mu.Unlock()
	return s.finish(ctx, owner, sess.ID, complete)
}

func (s *Store) skip(ctx context.Context, owner string, in Input, complete Completer) (View, error) {
	s.mu.Lock()
	sess, err := s.get(owner, strings.TrimSpace(in.ID))
	if err != nil {
		s.mu.Unlock()
		return View{}, err
	}
	if sess.Phase == PhaseReview {
		view := sess.view()
		s.mu.Unlock()
		return view, nil
	}
	if sess.Index >= len(sess.Questions) {
		s.mu.Unlock()
		return s.finish(ctx, owner, sess.ID, complete)
	}
	sess.Turns = append(sess.Turns, Turn{Question: sess.Questions[sess.Index], Skipped: true})
	sess.Index++
	done := sess.Index >= len(sess.Questions)
	if !done {
		view := sess.view()
		s.mu.Unlock()
		return view, nil
	}
	s.mu.Unlock()
	return s.finish(ctx, owner, sess.ID, complete)
}

func (s *Store) finish(ctx context.Context, owner, id string, complete Completer) (View, error) {
	s.mu.Lock()
	sess, err := s.get(owner, id)
	if err != nil {
		s.mu.Unlock()
		return View{}, err
	}
	if sess.Phase == PhaseReview {
		view := sess.view()
		s.mu.Unlock()
		return view, nil
	}
	snap := *sess
	snap.Turns = append([]Turn(nil), sess.Turns...)
	snap.Questions = append([]string(nil), sess.Questions...)
	s.mu.Unlock()

	draft, err := composeDraft(ctx, snap, complete)
	if err != nil {
		return View{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, err = s.get(owner, id)
	if err != nil {
		return View{}, err
	}
	if sess.Phase == PhaseReview {
		return sess.view(), nil
	}
	sess.Draft = draft
	sess.Phase = PhaseReview
	return sess.view(), nil
}

func composeDraft(ctx context.Context, sess Session, complete Completer) (string, error) {
	if complete == nil {
		return "", errors.New("AI service not configured")
	}
	raw, err := complete(ctx, WritePrompt(sess))
	if err != nil {
		return "", err
	}
	return SanitizeDraft(sess, raw), nil
}

func (s *Store) get(owner, id string) (*Session, error) {
	sess := s.sessions[id]
	if sess == nil || sess.Owner != owner || id == "" {
		return nil, ErrNotFound
	}
	return sess, nil
}

func (s *Store) remember(sess *Session) {
	if s.sessions == nil {
		s.sessions = map[string]*Session{}
		s.order = map[string][]string{}
	}
	s.sessions[sess.ID] = sess
	ids := append(s.order[sess.Owner], sess.ID)
	for len(ids) > MaxPerOwner {
		delete(s.sessions, ids[0])
		ids = ids[1:]
	}
	s.order[sess.Owner] = ids
}

func (s *Session) view() View {
	v := View{
		ID:      s.ID,
		Phase:   s.Phase,
		Kind:    s.Kind,
		Title:   s.Title,
		Planned: len(s.Questions),
	}
	if s.Phase == PhaseReview {
		v.Draft = s.Draft
		v.Asked = len(s.Turns)
		return v
	}
	v.Asked = s.Index + 1
	if s.Index >= 0 && s.Index < len(s.Questions) {
		v.Question = s.Questions[s.Index]
	}
	return v
}

func newID() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
