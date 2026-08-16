package models

// ChatMessageRow is one persisted turn (user or assistant).
type ChatMessageRow struct {
	Role    string          `json:"role"`
	Content string          `json:"content"`
	Sources []SourceRefLite `json:"sources,omitempty"`
}

// SourceRefLite matches AI source cards in the web client.
type SourceRefLite struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	URL   string `json:"url,omitempty"`
}

// ChatSession is a saved conversation (messages trimmed on server).
type ChatSession struct {
	ID        string           `json:"id"`
	Title     string           `json:"title"`
	CreatedAt int64            `json:"created_at"`
	UpdatedAt int64            `json:"updated_at"`
	Messages  []ChatMessageRow `json:"messages"`
}

// ChatSessionBrief is returned from list (no message bodies beyond preview).
type ChatSessionBrief struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Preview   string `json:"preview"`
	UpdatedAt int64  `json:"updated_at"`
}

// PatchChatSessionRequest updates stored turns.
type PatchChatSessionRequest struct {
	Title    string           `json:"title"`
	Messages []ChatMessageRow `json:"messages"`
}
