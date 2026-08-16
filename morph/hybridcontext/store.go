package hybridcontext

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	"github.com/google/uuid"
)

// SourceKind identifies where a document came from.
type SourceKind string

const (
	SourceFile          SourceKind = "file"
	SourcePaste         SourceKind = "paste"
	SourceSharpSchema   SourceKind = "sharpreport_schema"
	SourceTranFormEvents  SourceKind = "tranform_events_info"
	SourceBookiLedger   SourceKind = "booki_ledger"
)

const (
	maxSessionChunks   = 320
	maxChunkRunes      = 1000
	chunkOverlapRunes  = 80
	maxDocRunes        = 400_000
	defaultMaxRetrieve = 14_000
)

// Chunk is one searchable slice of a source document.
type Chunk struct {
	ID    string     `json:"id"`
	Kind  SourceKind `json:"kind"`
	Label string     `json:"label"`
	Text  string     `json:"text"`
}

type sessionState struct {
	chunks   []Chunk
	attached bool
}

// Store holds per-user, per-chat-session hybrid context (temporary "RAG" corpora).
type Store struct {
	mu sync.RWMutex
	m  map[string]*sessionState
}

func sessionKey(userID, sessionID string) string {
	return strings.TrimSpace(userID) + "\x00" + strings.TrimSpace(sessionID)
}

func NewStore() *Store {
	return &Store{m: make(map[string]*sessionState)}
}

func (s *Store) HasContent(userID, sessionID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st := s.m[sessionKey(userID, sessionID)]
	return st != nil && len(st.chunks) > 0
}

func (s *Store) Clear(userID, sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, sessionKey(userID, sessionID))
}

// SetAttached marks HybridContext as an active chat reference (Bring to conversation).
func (s *Store) SetAttached(userID, sessionID string, attached bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := sessionKey(userID, sessionID)
	st := s.m[key]
	if st == nil || len(st.chunks) == 0 {
		return false
	}
	st.attached = attached
	return true
}

// IsAttached reports whether Bring to conversation is active for this session.
func (s *Store) IsAttached(userID, sessionID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st := s.m[sessionKey(userID, sessionID)]
	return st != nil && st.attached && len(st.chunks) > 0
}

// AttachmentTitle returns a short comma-separated list of source labels for the UI.
func (s *Store) AttachmentTitle(userID, sessionID string) string {
	summaries := s.Summary(userID, sessionID)
	if len(summaries) == 0 {
		return ""
	}
	labels := make([]string, 0, len(summaries))
	for _, sum := range summaries {
		if l := strings.TrimSpace(sum.Label); l != "" {
			labels = append(labels, l)
		}
	}
	if len(labels) == 0 {
		return "HybridContext sources"
	}
	if len(labels) == 1 {
		return labels[0]
	}
	if len(labels) == 2 {
		return labels[0] + ", " + labels[1]
	}
	return fmt.Sprintf("%s, %s +%d more", labels[0], labels[1], len(labels)-2)
}

// AddDocument chunks fullText and appends. Returns number of chunks added.
func (s *Store) AddDocument(userID, sessionID string, kind SourceKind, label, fullText string) int {
	fullText = strings.TrimSpace(fullText)
	if fullText == "" {
		return 0
	}
	r := []rune(fullText)
	if len(r) > maxDocRunes {
		r = r[:maxDocRunes]
		fullText = string(r) + "\n…(document truncated for HybridContext)"
	}
	parts := chunkText(fullText, maxChunkRunes, chunkOverlapRunes)

	s.mu.Lock()
	defer s.mu.Unlock()
	key := sessionKey(userID, sessionID)
	st := s.m[key]
	if st == nil {
		st = &sessionState{}
		s.m[key] = st
	}
	added := 0
	for _, p := range parts {
		if strings.TrimSpace(p) == "" {
			continue
		}
		if len(st.chunks) >= maxSessionChunks {
			break
		}
		st.chunks = append(st.chunks, Chunk{
			ID:    uuid.NewString(),
			Kind:  kind,
			Label: label,
			Text:  p,
		})
		added++
	}
	return added
}

// SourceSummary is one grouped source for the UI.
type SourceSummary struct {
	Kind       string `json:"kind"`
	Label      string `json:"label"`
	ChunkCount int    `json:"chunk_count"`
}

// Summary returns counts grouped by kind+label for the UI.
func (s *Store) Summary(userID, sessionID string) []SourceSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st := s.m[sessionKey(userID, sessionID)]
	if st == nil || len(st.chunks) == 0 {
		return nil
	}
	type key struct {
		kind  SourceKind
		label string
	}
	counts := make(map[key]int)
	order := make([]key, 0)
	for _, c := range st.chunks {
		k := key{kind: c.Kind, label: c.Label}
		if counts[k] == 0 {
			order = append(order, k)
		}
		counts[k]++
	}
	out := make([]SourceSummary, 0, len(order))
	for _, k := range order {
		out = append(out, SourceSummary{
			Kind:       string(k.kind),
			Label:      k.label,
			ChunkCount: counts[k],
		})
	}
	return out
}

// RemoveSource deletes all chunks matching kind+label. Returns number of chunks removed.
func (s *Store) RemoveSource(userID, sessionID string, kind SourceKind, label string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := sessionKey(userID, sessionID)
	st := s.m[key]
	if st == nil || len(st.chunks) == 0 {
		return 0
	}
	filtered := st.chunks[:0]
	removed := 0
	for _, c := range st.chunks {
		if c.Kind == kind && c.Label == label {
			removed++
			continue
		}
		filtered = append(filtered, c)
	}
	if removed == 0 {
		return 0
	}
	if len(filtered) == 0 {
		delete(s.m, key)
	} else {
		st.chunks = filtered
		if len(st.chunks) == 0 {
			st.attached = false
		}
	}
	return removed
}

// RetrieveIfRelevant returns excerpts when the query overlaps chunk content.
// When allowBroad is true (user explicitly referenced HybridContext), falls back to top chunks if overlap is weak.
func (s *Store) RetrieveIfRelevant(userID, sessionID, userQuery string, maxRunes int, allowBroad bool) string {
	if maxRunes <= 0 {
		maxRunes = defaultMaxRetrieve
	}
	s.mu.RLock()
	st := s.m[sessionKey(userID, sessionID)]
	if st == nil || len(st.chunks) == 0 {
		s.mu.RUnlock()
		return ""
	}
	chunks := append([]Chunk(nil), st.chunks...)
	s.mu.RUnlock()

	qTokens := tokenize(userQuery)
	type scored struct {
		c     Chunk
		score int
	}
	list := make([]scored, 0, len(chunks))
	for _, c := range chunks {
		toks := tokenize(c.Text)
		score := 0
		for _, q := range qTokens {
			for _, t := range toks {
				if t == q {
					score++
				}
			}
		}
		list = append(list, scored{c: c, score: score})
	}
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].score > list[i].score {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	var b strings.Builder
	total := 0
	for _, sc := range list {
		if sc.score == 0 {
			continue
		}
		block := fmt.Sprintf("[%s | %s]\n%s\n\n", sc.c.Kind, sc.c.Label, sc.c.Text)
		r := []rune(block)
		if total+len(r) > maxRunes {
			remain := maxRunes - total
			if remain <= 0 {
				break
			}
			b.WriteString(string(r[:remain]))
			break
		}
		b.WriteString(block)
		total += len(r)
	}
	result := strings.TrimSpace(b.String())
	if result != "" {
		return result
	}
	if allowBroad {
		broadMax := maxRunes
		if broadMax > 8000 {
			broadMax = 8000
		}
		return concatTopChunks(chunks, broadMax)
	}
	return ""
}

// RetrieveForQuery scores chunks by simple token overlap with the user message.
func (s *Store) RetrieveForQuery(userID, sessionID, userQuery string, maxRunes int) string {
	if maxRunes <= 0 {
		maxRunes = defaultMaxRetrieve
	}
	s.mu.RLock()
	st := s.m[sessionKey(userID, sessionID)]
	if st == nil || len(st.chunks) == 0 {
		s.mu.RUnlock()
		return ""
	}
	chunks := append([]Chunk(nil), st.chunks...)
	s.mu.RUnlock()

	qTokens := tokenize(userQuery)
	if len(qTokens) == 0 {
		// No query tokens: return beginning of corpus (still useful for "continue" style prompts).
		return concatTopChunks(chunks, maxRunes)
	}

	type scored struct {
		c     Chunk
		score int
	}
	list := make([]scored, 0, len(chunks))
	for _, c := range chunks {
		toks := tokenize(c.Text)
		score := 0
		for _, q := range qTokens {
			for _, t := range toks {
				if t == q {
					score++
				}
			}
		}
		list = append(list, scored{c: c, score: score})
	}
	// Sort by score desc, stable by chunk order for ties
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].score > list[i].score {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	var b strings.Builder
	total := 0
	for _, sc := range list {
		if sc.score == 0 {
			continue
		}
		block := fmt.Sprintf("[%s | %s]\n%s\n\n", sc.c.Kind, sc.c.Label, sc.c.Text)
		r := []rune(block)
		if total+len(r) > maxRunes {
			remain := maxRunes - total
			if remain <= 0 {
				break
			}
			b.WriteString(string(r[:remain]))
			break
		}
		b.WriteString(block)
		total += len(r)
	}
	result := strings.TrimSpace(b.String())
	if result != "" {
		return result
	}
	return concatTopChunks(chunks, maxRunes)
}

func concatTopChunks(chunks []Chunk, maxRunes int) string {
	var b strings.Builder
	total := 0
	for _, c := range chunks {
		block := fmt.Sprintf("[%s | %s]\n%s\n\n", c.Kind, c.Label, c.Text)
		r := []rune(block)
		if total+len(r) > maxRunes {
			remain := maxRunes - total
			if remain <= 0 {
				break
			}
			b.WriteString(string(r[:remain]))
			break
		}
		b.WriteString(block)
		total += len(r)
	}
	return strings.TrimSpace(b.String())
}

// FullCorpusText joins all stored chunks in order (for "bring to conversation"), capped at maxRunes.
func (s *Store) FullCorpusText(userID, sessionID string, maxRunes int) string {
	if maxRunes <= 0 {
		maxRunes = defaultMaxRetrieve
	}
	s.mu.RLock()
	st := s.m[sessionKey(userID, sessionID)]
	if st == nil || len(st.chunks) == 0 {
		s.mu.RUnlock()
		return ""
	}
	chunks := append([]Chunk(nil), st.chunks...)
	s.mu.RUnlock()
	return concatTopChunks(chunks, maxRunes)
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	var cur []rune
	var out []string
	flush := func() {
		if len(cur) >= 3 {
			out = append(out, string(cur))
		}
		cur = cur[:0]
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			cur = append(cur, r)
		} else {
			flush()
		}
	}
	flush()
	return out
}

func chunkText(s string, size, overlap int) []string {
	if size <= 0 {
		return nil
	}
	r := []rune(s)
	if len(r) <= size {
		return []string{s}
	}
	var parts []string
	start := 0
	for start < len(r) {
		end := start + size
		if end > len(r) {
			end = len(r)
		}
		parts = append(parts, string(r[start:end]))
		if end >= len(r) {
			break
		}
		next := end - overlap
		if next <= start {
			next = start + size/2
		}
		start = next
	}
	return parts
}
