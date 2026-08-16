package models

// DefaultChatSessionID is the session ID used when the client does not send one.
const DefaultChatSessionID = "default"

type ChatRequest struct {
	Message   string `json:"message,omitempty"`
	SessionID string `json:"session_id,omitempty"` // Optional; empty means default session
	// AgentID selects a Morph AI assistant persona (e.g. general, image-generator).
	AgentID string `json:"agent_id,omitempty"`
	// SkillIDs optionally loads full skill bodies into the assistant system context.
	SkillIDs []string `json:"skill_ids,omitempty"`
}

// MorphAIAgent is a selectable chat assistant persona stored in Badger.
type MorphAIAgent struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Instructions  string `json:"instructions"` // Appended to base Morph instructions for this agent
	SystemDefined bool   `json:"system_defined"`
	SortOrder     int    `json:"sort_order"`
	Enabled       bool   `json:"enabled"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

// ChatImage is an inline image returned by chat (e.g. image-generator agent).
type ChatImage struct {
	ContentType string `json:"content_type"`
	Base64      string `json:"base64"`
	Alt         string `json:"alt,omitempty"`
}

// ChatSession is a conversation session (default or user-created).
type ChatSession struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// StoredChatMessage is one message in a session (user or assistant), stored in DB.
type StoredChatMessage struct {
	Role             string                        `json:"role"` // "user" | "assistant" | "error"
	Content          string                        `json:"content"`
	SQL              string                        `json:"sql,omitempty"`
	ConfirmationCard *RegistrationConfirmationCard `json:"confirmation_card,omitempty"`
	ProposedForm     *ProposedFormCard             `json:"proposed_form,omitempty"`
	ResearchContent  string                        `json:"research_content,omitempty"`
	Images           []ChatImage                   `json:"images,omitempty"`
	Timestamp        string                        `json:"timestamp"`
}

type ChatResponse struct {
	Response         string                        `json:"response"`
	SQL              string                        `json:"sql,omitempty"`
	FormJSON         string                        `json:"form_json,omitempty"`
	ConfirmationCard *RegistrationConfirmationCard `json:"confirmation_card,omitempty"`
	ProposedForm     *ProposedFormCard             `json:"proposed_form,omitempty"`
	ResearchContent  string                        `json:"research_content,omitempty"`
	Images           []ChatImage                   `json:"images,omitempty"`
}

// ProposedFormCard is sent when a form is generated from document upload; user must confirm before saving.
type ProposedFormCard struct {
	FormTemplate FormTemplate `json:"form_template"`
}

// RegistrationConfirmationCard is sent so the chat UI can show a review card before submitting.
type RegistrationConfirmationCard struct {
	FormName string                 `json:"form_name"`
	UserType string                 `json:"user_type"`
	Answers  map[string]interface{} `json:"answers"`
	Fields   []FormField            `json:"fields"` // name + label for display
}

type SQLFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type ChatHistory struct {
	Message   string `json:"message"`
	Response  string `json:"response"`
	Timestamp string `json:"timestamp"`
}

// Complaint flow models (legacy DB keys may still exist)
type ComplaintState struct {
	ConversationID string                 `json:"conversation_id"`
	Step           string                 `json:"step"` // "start", "dialogue", "waiting_complaint", "executing", "complete"
	ComplaintText  string                 `json:"complaint_text,omitempty"`
	DialogueResult map[string]interface{} `json:"dialogue_result,omitempty"`
	InitialData    map[string]interface{} `json:"initial_data,omitempty"`
	ExchangeCount  int                    `json:"exchange_count"`          // Track number of exchanges
	LastResponse   string                 `json:"last_response,omitempty"` // Store last AI response
}

// Voice recognition models
type VoiceProfile struct {
	UserID       string   `json:"user_id"`
	Name         string   `json:"name"`
	VoiceSamples []string `json:"voice_samples"` // Base64 encoded audio samples or file paths
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

type VoiceRegistrationRequest struct {
	Name        string `json:"name" binding:"required"`
	AudioData   string `json:"audio_data" binding:"required"` // Base64 encoded audio
	AudioFormat string `json:"audio_format"`                  // "wav", "mp3", "webm", etc.
}

type VoiceRecognitionRequest struct {
	AudioData   string `json:"audio_data" binding:"required"` // Base64 encoded audio
	AudioFormat string `json:"audio_format"`                  // "wav", "mp3", "webm", etc.
}

type VoiceRecognitionResponse struct {
	Recognized bool   `json:"recognized"`
	UserID     string `json:"user_id,omitempty"`
	Name       string `json:"name,omitempty"`
	Transcript string `json:"transcript,omitempty"`
	Intent     string `json:"intent,omitempty"` // "attendance", "punch_in", etc.
	Message    string `json:"message"`
}

// Form system models
type FormField struct {
	Name        string   `json:"name"`              // Field identifier (e.g., "name", "age")
	Label       string   `json:"label"`             // Display label (e.g., "Full Name")
	Type        string   `json:"type"`              // Field type: Quick Sheets only support "text"
	Required    bool     `json:"required"`          // Whether field is required
	Placeholder string   `json:"placeholder"`       // Placeholder text
	Options     []string `json:"options,omitempty"` // Unused for Quick Sheets (text-only)
}

type FormTemplate struct {
	ID          string      `json:"id"`          // Unique identifier
	Name        string      `json:"name"`        // Form name (e.g., "Student Registration Form")
	Description string      `json:"description"` // Form description
	UserType    string      `json:"user_type"`   // "student" or "staff"
	Fields      []FormField `json:"fields"`      // Form fields
	CreatedAt   string      `json:"created_at"`  // Creation timestamp
	UpdatedAt   string      `json:"updated_at"`  // Last update timestamp
	CreatedBy   string      `json:"created_by"`  // User who created the form
}

type FormAnswer struct {
	ID           string                 `json:"id"`                      // Unique identifier
	FormID       string                 `json:"form_id"`                 // Reference to FormTemplate
	FormName     string                 `json:"form_name"`               // Form name (denormalized for easy access)
	UserID       string                 `json:"user_id"`                 // Student or staff ID
	UserType     string                 `json:"user_type"`               // "student" or "staff"
	AssignmentID string                 `json:"assignment_id,omitempty"` // Linked assignment when submitted from Collective Sheets
	Answers      map[string]interface{} `json:"answers"`                 // Field name -> answer value
	SubmittedAt  string                 `json:"submitted_at"`            // Submission timestamp
	SubmittedBy  string                 `json:"submitted_by"`            // User who submitted
}

type FormAssignment struct {
	ID                string `json:"id"`
	FormID            string `json:"form_id"`
	FormName          string `json:"form_name"`
	FormUserType      string `json:"form_user_type"`
	AssigneeUserID    string `json:"assignee_user_id"`
	AssigneeUserType  string `json:"assignee_user_type"`
	AssigneeName      string `json:"assignee_name,omitempty"`
	AssignedBy        string `json:"assigned_by"`
	AssignedAt        string `json:"assigned_at"`
	Status            string `json:"status"` // pending | completed
	CompletedAt       string `json:"completed_at,omitempty"`
	CompletedAnswerID string `json:"completed_answer_id,omitempty"`
}

type FormNotification struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	CreatedAt    string `json:"created_at"`
	Status       string `json:"status"` // pending | completed
	FormID       string `json:"form_id"`
	FormName     string `json:"form_name"`
	AssignmentID string `json:"assignment_id"`
}

// RegistrationFlowState holds state for the "register a student" (or similar) chat flow
type RegConvTurn struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"` // message content
}

type RegistrationState struct {
	ConversationID      string                 `json:"conversation_id"`      // unique session id
	Step                string                 `json:"step"`                 // "selecting_form" | "gathering_fields" | "complete"
	FormID              string                 `json:"form_id,omitempty"`    // chosen form template id (internal, not shown to AI)
	FormName            string                 `json:"form_name,omitempty"`  // form name for context
	UserType            string                 `json:"user_type,omitempty"`  // student | staff from form
	GatheredAnswers     map[string]interface{} `json:"gathered_answers"`     // field name -> value so far
	ConversationHistory []RegConvTurn          `json:"conversation_history"` // full chat history for this session
	LastAIResponse      string                 `json:"last_ai_response,omitempty"`
	ExchangeCount       int                    `json:"exchange_count"`
	CreatedAt           string                 `json:"created_at,omitempty"`
}
