package morphai

// Message is a chat turn for DashScope text-generation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Content part types for multimodal (vision) messages.
const (
	ContentPartText     = "text"
	ContentPartImageURL = "image_url"
)

// ImageURL carries an image reference for a multimodal message. URL accepts an
// https URL or a base64 data URL.
type ImageURL struct {
	URL string `json:"url"`
}

// ContentPart is one element of a multimodal message's content array. Exactly
// one of Text or ImageURL is set, matching Type.
type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// TextPart builds a text element of a multimodal message.
func TextPart(text string) ContentPart {
	return ContentPart{Type: ContentPartText, Text: text}
}

// ImagePart builds an image element from an https or data URL.
func ImagePart(url string) ContentPart {
	return ContentPart{Type: ContentPartImageURL, ImageURL: &ImageURL{URL: url}}
}

// ImageDataPart builds an image element from raw bytes and a MIME type,
// encoding them as a base64 data URL.
func ImageDataPart(mimeType string, raw []byte) ContentPart {
	return ImagePart(DataURL(mimeType, raw))
}

// MultiMessage is a chat turn whose content is an array of parts, as used by
// OpenAI-compatible vision endpoints.
type MultiMessage struct {
	Role    string        `json:"role"`
	Content []ContentPart `json:"content"`
}

// UserMultiMessage builds a user turn from content parts.
func UserMultiMessage(parts ...ContentPart) MultiMessage {
	return MultiMessage{Role: "user", Content: parts}
}
