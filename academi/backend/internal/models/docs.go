package models

type Document struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	UploaderID      string   `json:"uploader_id"`
	Type            string   `json:"type"`
	MimeType        string   `json:"mime_type"`
	Size            string   `json:"size"`
	Thumbnail       string   `json:"thumbnail"`
	Tags            []string `json:"tags"`
	AISummary       string   `json:"ai_summary"`
	Content         string   `json:"content"`
	StoredFilename  string   `json:"stored_filename,omitempty"`
	SourceFilename  string   `json:"source_filename,omitempty"`
	CreatedAt       int64    `json:"created_at"`
}

// LearnableTypes are document types the UI may offer “Help you learn” for.
var LearnableTypes = map[string]bool{
	"markdown": true, "text": true, "pdf": true, "image": true,
}

type CreateDocRequest struct {
	Title     string   `json:"title"`
	Type      string   `json:"type"`
	Size      string   `json:"size"`
	Thumbnail string   `json:"thumbnail"`
	Tags      []string `json:"tags"`
	Content   string   `json:"content"`
	AISummary string   `json:"ai_summary"`
}

type SearchResponse struct {
	Results []Document `json:"results"`
	Total   int        `json:"total"`
	Query   string     `json:"query"`
}
