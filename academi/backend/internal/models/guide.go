package models

// GuideSubject is a top-level area (e.g. Mathematics) that contains many guides.
type GuideSubject struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   int64  `json:"created_at"`
	CreatedBy   string `json:"created_by,omitempty"`
}

type Guide struct {
	ID          string      `json:"id"`
	SubjectID   string      `json:"subject_id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Category    string      `json:"category"`
	Icon        string      `json:"icon"`
	Color       string      `json:"color"`
	Steps       []GuideStep `json:"steps"`
	CreatedAt   int64       `json:"created_at"`
	CreatedBy   string      `json:"created_by,omitempty"`
}

type GuideStep struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content,omitempty"`
}

type UserGuideProgress struct {
	UserID         string `json:"user_id"`
	GuideID        string `json:"guide_id"`
	CompletedSteps []int  `json:"completed_steps"`
	StartedAt      int64  `json:"started_at"`
	UpdatedAt      int64  `json:"updated_at"`
}

type CreateGuideRequest struct {
	SubjectID   string      `json:"subject_id" binding:"required"`
	Title       string      `json:"title" binding:"required"`
	Description string      `json:"description" binding:"required"`
	Category    string      `json:"category"`
	Icon        string      `json:"icon"`
	Color       string      `json:"color"`
	Steps       []GuideStep `json:"steps" binding:"required"`
}

type UpdateGuideRequest struct {
	SubjectID   string      `json:"subject_id" binding:"required"`
	Title       string      `json:"title" binding:"required"`
	Description string      `json:"description" binding:"required"`
	Category    string      `json:"category"`
	Icon        string      `json:"icon"`
	Color       string      `json:"color"`
	Steps       []GuideStep `json:"steps" binding:"required"`
}

type CreateSubjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type UpdateSubjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}
