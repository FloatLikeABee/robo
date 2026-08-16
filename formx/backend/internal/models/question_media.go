package models

// Question prompt attachment limits align with typical web forms: ~10 MB covers high-res JPEGs/PINg
// without encouraging huge uploads; ~100 MB fits short cellphone clips (often under 60 s at moderate bitrate).
const (
	MaxQuestionPromptImageBytes = 10 * 1024 * 1024  // 10 MiB
	MaxQuestionPromptVideoBytes = 100 * 1024 * 1024 // 100 MiB
)

const (
	PromptMediaKindImage = "image"
	PromptMediaKindVideo = "video"
)

// QuestionPromptMedia is optional media shown beside the question on the public form (image XOR video).
type QuestionPromptMedia struct {
	Kind         string `json:"kind"`          // "image" | "video"
	RelativePath string `json:"relative_path"` // path under uploads dir; served at GET /uploads/{relative_path}
}
