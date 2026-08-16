package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/academi/backend/internal/config"
	"github.com/academi/backend/internal/database"
	"github.com/academi/backend/internal/models"
)

type Service struct {
	cfg *config.Config
}

func NewService(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func (s *Service) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.cfg.JWT.ExpiryDuration()).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

func (s *Service) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := uuid.New().String()
	now := time.Now().Unix()

	user := models.User{
		ID:           userID,
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: s.hashPassword(req.Password),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	data, _ := json.Marshal(user)
	if err := database.Set([]byte("user:"+userID), data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	if err := database.Set([]byte("user_email:"+req.Email), []byte(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	settings := models.UserSettings{
		UserID:  userID,
		Theme:   "dark",
		AITone:  "casual",
		AIDepth: "intermediate",
		Streak:  0,
	}

	settingsData, _ := json.Marshal(settings)
	database.Set([]byte("settings:"+userID), settingsData)

	token, err := s.generateToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	user.PasswordHash = ""
	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

func (s *Service) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prefer UsersPanel (same auth as Morph / Booki).
	if strings.TrimSpace(s.cfg.UsersPanelBaseURL) != "" {
		sess, code, msg := UsersPanelLogin(c.Request.Context(), s.cfg.UsersPanelBaseURL, req.Email, req.Password)
		if sess == nil {
			c.JSON(code, gin.H{"error": msg})
			return
		}
		user, err := s.ensureUserFromPlatform(sess.Email, sess.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		token, err := s.generateToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}
		user.PasswordHash = ""
		c.JSON(http.StatusOK, models.AuthResponse{Token: token, User: user})
		return
	}

	userIDBytes, err := database.Get([]byte("user_email:" + req.Email))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	userData, err := database.Get([]byte("user:" + string(userIDBytes)))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	var user models.User
	json.Unmarshal(userData, &user)

	if user.PasswordHash != s.hashPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	user.PasswordHash = ""
	c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// DevLogin exchanges a UsersPanel JWT (from MorphAI apps menu) for an Academi JWT.
func (s *Service) DevLogin(c *gin.Context) {
	var body struct {
		Token string `json:"token"`
	}
	_ = c.ShouldBindJSON(&body)
	upToken := strings.TrimSpace(body.Token)
	if upToken == "" {
		upToken = strings.TrimSpace(c.Query("userspanel_token"))
	}
	if upToken == "" {
		h := c.GetHeader("Authorization")
		if strings.HasPrefix(h, "Bearer ") {
			upToken = strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
		}
	}
	if upToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userspanel token required"})
		return
	}
	if strings.TrimSpace(s.cfg.UsersPanelBaseURL) == "" {
		c.JSON(http.StatusBadGateway, gin.H{"error": "USERS_PANEL_BASE_URL not configured"})
		return
	}
	sess, err := UsersPanelSession(c.Request.Context(), s.cfg.UsersPanelBaseURL, upToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid userspanel session"})
		return
	}
	user, err := s.ensureUserFromPlatform(sess.Email, sess.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	token, err := s.generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	user.PasswordHash = ""
	c.JSON(http.StatusOK, models.AuthResponse{Token: token, User: user})
}

func (s *Service) ensureUserFromPlatform(email, username string) (models.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	name := strings.TrimSpace(username)
	if name == "" {
		name = email
	}
	now := time.Now().Unix()

	if userIDBytes, err := database.Get([]byte("user_email:" + email)); err == nil {
		userData, err := database.Get([]byte("user:" + string(userIDBytes)))
		if err == nil {
			var user models.User
			if json.Unmarshal(userData, &user) == nil {
				if user.Name == "" {
					user.Name = name
					user.UpdatedAt = now
					raw, _ := json.Marshal(user)
					_ = database.Set([]byte("user:"+user.ID), raw)
				}
				return user, nil
			}
		}
	}

	userID := uuid.New().String()
	user := models.User{
		ID:           userID,
		Email:        email,
		Name:         name,
		PasswordHash: s.hashPassword(uuid.New().String()),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	raw, _ := json.Marshal(user)
	if err := database.Set([]byte("user:"+userID), raw); err != nil {
		return models.User{}, fmt.Errorf("failed to create user")
	}
	if err := database.Set([]byte("user_email:"+email), []byte(userID)); err != nil {
		return models.User{}, fmt.Errorf("failed to create user")
	}
	settings := models.UserSettings{
		UserID: userID, Theme: "dark", AITone: "casual", AIDepth: "intermediate",
	}
	sd, _ := json.Marshal(settings)
	_ = database.Set([]byte("settings:"+userID), sd)
	return user, nil
}

// MockLogin creates or returns a fixed demo user and JWT (for local / prototype apps).
func (s *Service) MockLogin(c *gin.Context) {
	const mockUserID = "user_mock_dev"
	const mockEmail = "demo@academi.local"

	now := time.Now().Unix()
	data, err := database.Get([]byte("user:" + mockUserID))
	var user models.User
	if err != nil {
		user = models.User{
			ID:        mockUserID,
			Email:     mockEmail,
			Name:      "Study Explorer (demo)",
			CreatedAt: now,
			UpdatedAt: now,
		}
		user.PasswordHash = s.hashPassword("mock-placeholder")
		raw, _ := json.Marshal(user)
		if err := database.Set([]byte("user:"+mockUserID), raw); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to seed demo user"})
			return
		}
		_ = database.Set([]byte("user_email:"+mockEmail), []byte(mockUserID))
		settings := models.UserSettings{
			UserID: mockUserID, Theme: "dark", AITone: "casual", AIDepth: "intermediate",
		}
		sd, _ := json.Marshal(settings)
		_ = database.Set([]byte("settings:"+mockUserID), sd)
	} else {
		_ = json.Unmarshal(data, &user)
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	user.PasswordHash = ""
	c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

func (s *Service) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", s.Register)
		auth.POST("/login", s.Login)
		auth.POST("/dev-login", s.DevLogin)
		auth.POST("/mock", s.MockLogin)
	}
}

func JWTMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}

		tokenStr := authHeader[7:] // Remove "Bearer " prefix

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	return userID.(string)
}
