package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"database/sql"

	"github.com/academi/booki/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	DB                *sql.DB
	RefreshCache      *gocache.Cache
	Secret            string
	Access            time.Duration
	RefreshTTL        time.Duration
	UsersPanelBaseURL string
}

func (s *Service) Register(c *gin.Context) {
	var body struct {
		OrganizationName string `json:"organization_name" binding:"required"`
		Name             string `json:"name" binding:"required"`
		Email            string `json:"email" binding:"required,email"`
		Password         string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash failed"})
		return
	}
	tx, err := s.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO organizations (name) VALUES (?)`, body.OrganizationName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("org: %v", err)})
		return
	}
	orgID, _ := res.LastInsertId()

	if err := seedChartOfAccounts(tx, orgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("chart: %v", err)})
		return
	}

	ures, err := tx.Exec(`INSERT INTO users (organization_id, name, email, password_hash, role) VALUES (?,?,?,?, 'owner')`,
		orgID, body.Name, body.Email, string(hash))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("user: %v", err)})
		return
	}
	userID, _ := ures.LastInsertId()

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	access, refresh, err := s.issueTokens(userID, orgID, "owner")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"user_id": userID, "organization_id": orgID,
		"access_token": access, "refresh_token": refresh,
		"token_type": "Bearer", "expires_in": int(s.Access.Seconds()),
	})
}

func (s *Service) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, status, msg := usersPanelLogin(c.Request.Context(), s.UsersPanelBaseURL, body.Email, body.Password)
	if session == nil {
		if status == http.StatusUnauthorized && strings.EqualFold(strings.TrimSpace(msg), "unauthorized") {
			msg = "Invalid email or password"
		}
		c.JSON(status, gin.H{"error": msg})
		return
	}
	if !hasBookiAccess(session.Roles, session.Permissions) {
		c.JSON(http.StatusForbidden, gin.H{"error": "booki access is restricted by admin policy"})
		return
	}

	displayName := session.Username
	if displayName == "" {
		displayName = session.Email
	}
	userID, orgID, bookiRole, err := EnsurePlatformIdentity(s.DB, session.Email, displayName, session.Roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	access, refresh, err := s.issueTokens(userID, orgID, bookiRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         session.Token,
		"access_token":  access,
		"refresh_token": refresh,
		"token_type":    "Bearer",
		"expires_in":    int(s.Access.Seconds()),
		"user": gin.H{
			"email":    session.Email,
			"username": session.Username,
			"roles":    session.Roles,
		},
		"permissions": session.Permissions,
	})
}

// PlatformSession exchanges an existing UsersPanel JWT for Booki ledger tokens (shared SSO cookie).
func (s *Service) PlatformSession(c *gin.Context) {
	token := bearerToken(c.GetHeader("Authorization"))
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
		return
	}
	session, err := usersPanelSession(c.Request.Context(), s.UsersPanelBaseURL, token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}
	if !hasBookiAccess(session.Roles, session.Permissions) {
		c.JSON(http.StatusForbidden, gin.H{"error": "booki access is restricted by admin policy"})
		return
	}
	displayName := session.Username
	if displayName == "" {
		displayName = session.Email
	}
	userID, orgID, bookiRole, err := EnsurePlatformIdentity(s.DB, session.Email, displayName, session.Roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	access, refresh, err := s.issueTokens(userID, orgID, bookiRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":         session.Token,
		"access_token":  access,
		"refresh_token": refresh,
		"token_type":    "Bearer",
		"expires_in":    int(s.Access.Seconds()),
		"user": gin.H{
			"email":    session.Email,
			"username": session.Username,
			"roles":    session.Roles,
		},
		"permissions": session.Permissions,
	})
}

func (s *Service) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if s.RefreshCache == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh store unavailable"})
		return
	}
	raw, ok := s.RefreshCache.Get("refresh:" + body.RefreshToken)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh"})
		return
	}
	val, _ := raw.([]byte)
	var pl struct {
		UserID int64  `json:"user_id"`
		OrgID  int64  `json:"org_id"`
		Role   string `json:"role"`
	}
	if len(val) == 0 || json.Unmarshal(val, &pl) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh"})
		return
	}
	s.RefreshCache.Delete("refresh:" + body.RefreshToken)
	access, refresh, err := s.issueTokens(pl.UserID, pl.OrgID, pl.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token": access, "refresh_token": refresh,
		"token_type": "Bearer", "expires_in": int(s.Access.Seconds()),
	})
}

func (s *Service) Logout(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.RefreshToken != "" && s.RefreshCache != nil {
		s.RefreshCache.Delete("refresh:" + body.RefreshToken)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Service) Me(c *gin.Context) {
	uid := middleware.GetUserID(c)
	orgID := middleware.GetOrgID(c)
	var name, email, role string
	_ = s.DB.QueryRow(`SELECT name, email, role FROM users WHERE id=?`, uid).Scan(&name, &email, &role)
	var oname, country, currency string
	_ = s.DB.QueryRow(`SELECT name, country, currency FROM organizations WHERE id=?`, orgID).Scan(&oname, &country, &currency)
	c.JSON(http.StatusOK, gin.H{
		"user":           gin.H{"id": uid, "name": name, "email": email, "role": role},
		"organization": gin.H{"id": orgID, "name": oname, "country": country, "currency": currency},
	})
}

// DevLogin exchanges the shared UsersPanel session cookie for Booki tokens (development only).
func (s *Service) DevLogin(c *gin.Context) {
	token := bearerToken(c.GetHeader("Authorization"))
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "sign in via UsersPanel first, or POST /auth/login with platform credentials"})
		return
	}
	session, err := usersPanelSession(c.Request.Context(), s.UsersPanelBaseURL, token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}
	if !hasBookiAccess(session.Roles, session.Permissions) {
		c.JSON(http.StatusForbidden, gin.H{"error": "booki access is restricted by admin policy"})
		return
	}
	displayName := session.Username
	if displayName == "" {
		displayName = session.Email
	}
	userID, orgID, bookiRole, err := EnsurePlatformIdentity(s.DB, session.Email, displayName, session.Roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	access, refresh, err := s.issueTokens(userID, orgID, bookiRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":         session.Token,
		"access_token":  access,
		"refresh_token": refresh,
		"token_type":    "Bearer",
		"expires_in":    int(s.Access.Seconds()),
	})
}

func (s *Service) issueTokens(userID, orgID int64, role string) (access, refresh string, err error) {
	now := time.Now()
	claims := &middleware.Claims{
		UserID: userID,
		OrgID:  orgID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.Access)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	access, err = at.SignedString([]byte(s.Secret))
	if err != nil {
		return "", "", err
	}
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	refresh = hex.EncodeToString(b)
	if s.RefreshCache != nil {
		pl, _ := json.Marshal(map[string]any{"user_id": userID, "org_id": orgID, "role": role})
		s.RefreshCache.Set("refresh:"+refresh, pl, s.RefreshTTL)
	}
	return access, refresh, nil
}

func seedChartOfAccounts(tx *sql.Tx, orgID int64) error {
	rows := []struct {
		code, name, typ string
	}{
		{"1000", "Cash", "asset"},
		{"1010", "Bank", "asset"},
		{"1200", "Inventory", "asset"},
		{"1500", "Equipment", "asset"},
		{"1100", "Accounts Receivable", "asset"},
		{"2000", "Loans Payable", "liability"},
		{"2100", "Accounts Payable", "liability"},
		{"2200", "Tax Payable", "liability"},
		{"3000", "Owner Equity", "equity"},
		{"3100", "Retained Earnings", "equity"},
		{"4000", "Sales Revenue", "revenue"},
		{"4100", "Service Revenue", "revenue"},
		{"5000", "Utilities Expense", "expense"},
		{"5100", "Salary Expense", "expense"},
		{"5200", "Rent Expense", "expense"},
		{"5300", "Office Expense", "expense"},
	}
	for _, r := range rows {
		if _, err := tx.Exec(`INSERT INTO accounts (organization_id, code, name, type, is_system) VALUES (?,?,?,?,1)`,
			orgID, r.code, r.name, r.typ); err != nil {
			return err
		}
	}
	return nil
}
