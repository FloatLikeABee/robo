package models

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	AvatarURL    string `json:"avatar_url"`
	PasswordHash string `json:"-"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

type UserSettings struct {
	UserID    string `json:"user_id"`
	Theme     string `json:"theme"`
	AITone    string `json:"ai_tone"`
	AIDepth   string `json:"ai_depth"`
	Streak    int    `json:"streak"`
	Contribs  int    `json:"contributions"`
	SavedDocs int    `json:"saved_docs"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
