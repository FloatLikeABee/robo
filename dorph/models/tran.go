package models

import (
	"time"
)

// District (MySQL tran)
type District struct {
	ID          int     `json:"id"`
	DistrictID  int     `json:"district_id"`
	District    string  `json:"district"`
	Name        *string `json:"name"`
	Description *string `json:"description,omitempty"`
}

// School models one row in MySQL `facility`. REST: /api/tran/facilities.
type School struct {
	ID           int     `json:"id"`
	FacilityCode string  `json:"facility_code"`
	Name         *string `json:"name"`
	DistrictID   *int    `json:"district_id"`
	// FacilityType is a configurable code (see platform dictionaries: facility_type)
	FacilityType *string `json:"facility_type"`
	Location     *string `json:"location,omitempty"`
	Description  *string `json:"description,omitempty"`
}

// Student (MySQL `member` table). Large detail JSON is stored in MongoDB. REST: /api/tran/members.
type Student struct {
	ID              int        `json:"id"`
	LastName        *string    `json:"last_name"`
	FirstName       *string    `json:"first_name"`
	MiddleName      *string    `json:"middle_name"`
	Dob             *time.Time `json:"dob"`
	EntryDate       *time.Time `json:"entry_date"`
	Facility        *string    `json:"facility"`
	Gender          *int       `json:"gender"`
	Email           *string    `json:"email"`
	ParticipantType *string    `json:"participant_type"`
	Description     *string    `json:"description,omitempty"`
}

// Staff (MySQL `employee` table; Activity relations use employee.id). REST: /api/tran/employees.
type Staff struct {
	ID              int     `json:"id"`
	LastName        string  `json:"last_name"`
	FirstName       *string `json:"first_name"`
	MiddleName      *string `json:"middle_name"`
	Email           *string `json:"email"`
	PhoneNumber     *string `json:"phone_number"`
	ActiveFlag      bool    `json:"active_flag"`
	EmployType      *string `json:"employ_type"`
	Description     *string `json:"description,omitempty"`
	FacilityID      *int    `json:"facility_id,omitempty"`
	FacilityDisplay *string `json:"facility_display,omitempty"`
}

// Contact (MySQL tran) — contact info for schools (personnel), students (parents), etc.
type Contact struct {
	ID          int     `json:"id"`
	LastName    string  `json:"last_name"`
	FirstName   *string `json:"first_name"`
	Email       *string `json:"email"`
	Phone       *string `json:"phone"`
	Mobile      *string `json:"mobile"`
	Description *string `json:"description,omitempty"`
}

// RecordContact links an entity (student, school, trip, vehicle, staff, district) to a contact.
// EntityType: "student" | "school" | "trip" | "vehicle" | "staff" | "district"
// Relationship: e.g. "Parent", "Guardian", "Principal", "Emergency contact"
type RecordContact struct {
	ID           int     `json:"id"`
	EntityType   string  `json:"entity_type"`
	RecordID     int     `json:"record_id"`
	ContactID    int     `json:"contact_id"`
	Relationship *string `json:"relationship"`
	IsPrimary    bool    `json:"is_primary"`
}

// Vehicle (MySQL `Asset` table). REST: /api/tran/assets.
type Vehicle struct {
	ID           int     `json:"id"`
	AssetTag     *string `json:"asset_tag"`
	Description  *string `json:"description,omitempty"`
	AssetID      *string `json:"asset_id,omitempty"`
	AssetType    *string `json:"asset_type"`
	ContractorID *int    `json:"contractor_id,omitempty"`
}

// Trip (MySQL `Activity` table). REST: /api/tran/activities.
type Trip struct {
	ID           int        `json:"id"`
	Name         string     `json:"name"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Location     *string    `json:"location"`
	ActivityType *string    `json:"activity_type"`
	Description  *string    `json:"description,omitempty"`
}

// CaseTask (MySQL `CaseTask` table). REST: /api/tran/case-tasks.
type CaseTask struct {
	ID           int        `json:"id"`
	Title        string     `json:"title"`
	Description  *string    `json:"description,omitempty"`
	StartAt      *time.Time `json:"start_at,omitempty"`
	EndAt        *time.Time `json:"end_at,omitempty"`
	Location     *string    `json:"location,omitempty"`
	AssigneeType string     `json:"assignee_type"`
	AssigneeID   int        `json:"assignee_id"`
	AssigneeName *string    `json:"assignee_name,omitempty"`
	CreatedOn    *time.Time `json:"created_on,omitempty"`
	LastUpdated  *time.Time `json:"last_updated,omitempty"`
}

// StoryPost (MySQL `StoryPost` table). REST: /api/tran/story-posts.
type StoryPost struct {
	ID           int        `json:"id"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	AuthorUserID int        `json:"author_user_id"`
	AuthorName   *string    `json:"author_name,omitempty"`
	CreatedOn    *time.Time `json:"created_on,omitempty"`
	LastUpdated  *time.Time `json:"last_updated,omitempty"`
	CommentCount int        `json:"comment_count,omitempty"`
}

type CaseTaskAttachment struct {
	ID           int        `json:"id"`
	CaseTaskID   int        `json:"case_task_id"`
	OriginalName string     `json:"original_name"`
	MimeType     *string    `json:"mime_type,omitempty"`
	SizeBytes    int64      `json:"size_bytes"`
	DownloadURL  string     `json:"download_url"`
	CreatedOn    *time.Time `json:"created_on,omitempty"`
}

// TranUser (User table — accounts, tools dropdowns, note targeting)
type TranUser struct {
	ID            int     `json:"id"`
	LoginID       *string `json:"login_id"`
	FirstName     *string `json:"first_name"`
	LastName      string  `json:"last_name"`
	Email         *string `json:"email"`
	Phone         *string `json:"phone"`
	Title         *string `json:"title"`
	Administrator bool    `json:"administrator"`
	Deactivated   bool    `json:"deactivated"`
}

// ToolNote (tool_note) - private notes and public messages
type ToolNote struct {
	ID           int        `json:"id"`
	AuthorUserID int        `json:"author_user_id"`
	TargetType   string     `json:"target_type"` // self, user, public
	TargetUserID *int       `json:"target_user_id"`
	IsPrivate    bool       `json:"is_private"`
	Title        *string    `json:"title"`
	Body         *string    `json:"body"`
	ReadAt       *time.Time `json:"read_at"`
	CreatedOn    *time.Time `json:"created_on"`
	LastUpdated  *time.Time `json:"last_updated"`
}

// UserNoteTodo (user_note_todo) - personal notes and TODOs for a User.UserID.
type UserNoteTodo struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	ItemType    string     `json:"item_type"` // note | todo
	Title       *string    `json:"title"`
	Body        *string    `json:"body"`
	Completed   bool       `json:"completed"`
	DeadlineAt  *time.Time `json:"deadline_at,omitempty"`
	CreatedOn   *time.Time `json:"created_on"`
	LastUpdated *time.Time `json:"last_updated"`
}

// GenericData (generic_data) — imported CSV/JSON/PDF material. Extended payload in Mongo entity_details.
type GenericData struct {
	ID             int        `json:"id"`
	Title          string     `json:"title"`
	SourceType     string     `json:"source_type"`
	SourceFilename *string    `json:"source_filename,omitempty"`
	RecordCount    int        `json:"record_count"`
	Description    *string    `json:"description,omitempty"`
	AIAnalysis     *string    `json:"ai_analysis,omitempty"`
	CreatedOn      *time.Time `json:"created_on,omitempty"`
	LastUpdated    *time.Time `json:"last_updated,omitempty"`
}

// Comment (comment) - comments on students, staff, schools, trips. Supports threads via ParentID.
type Comment struct {
	ID           int        `json:"id"`
	EntityType   string     `json:"entity_type"`
	RecordID     int        `json:"record_id"`
	ParentID     *int       `json:"parent_id"`
	AuthorUserID *int       `json:"author_user_id"`
	Body         string     `json:"body"`
	CreatedOn    *time.Time `json:"created_on"`
	LastUpdated  *time.Time `json:"last_updated"`
	Replies      []Comment  `json:"replies,omitempty"`
}
