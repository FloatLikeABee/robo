package models

type CommunityPost struct {
	ID         string   `json:"id"`
	AuthorID   string   `json:"author_id"`
	AuthorName string   `json:"author_name"`
	Content    string   `json:"content"`
	Tags       []string `json:"tags"`
	DocID      string   `json:"doc_id,omitempty"`
	CreatedAt  int64    `json:"created_at"`
	UpdatedAt  int64    `json:"updated_at"`
	Upvotes    int      `json:"upvotes"`
	Downvotes  int      `json:"downvotes"`
	Comments   int      `json:"comments"`
	AISummary  string   `json:"ai_summary,omitempty"`
	AIVerified bool     `json:"ai_verified"`
}

type Comment struct {
	ID         string `json:"id"`
	PostID     string `json:"post_id"`
	AuthorID   string `json:"author_id"`
	AuthorName string `json:"author_name"`
	Content    string `json:"content"`
	CreatedAt  int64  `json:"created_at"`
}

type CreatePostRequest struct {
	Content  string   `json:"content"`
	Tags     []string `json:"tags"`
	DocID    string   `json:"doc_id"`
	DocTitle string   `json:"doc_title"`
}

type UpdatePostRequest struct {
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}
