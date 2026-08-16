package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

const tranUserSelectCols = `UserID, LoginID, FirstName, LastName, Email, Phone, Title, Administrator, Deactivated`

func scanTranUserRow(scanner interface{ Scan(dest ...interface{}) error }) (models.TranUser, error) {
	var u models.TranUser
	var login, fn, em, ph, tit sql.NullString
	var admin, deact int
	err := scanner.Scan(&u.ID, &login, &fn, &u.LastName, &em, &ph, &tit, &admin, &deact)
	if err != nil {
		return u, err
	}
	if login.Valid {
		u.LoginID = &login.String
	}
	if fn.Valid {
		u.FirstName = &fn.String
	}
	if em.Valid {
		u.Email = &em.String
	}
	if ph.Valid {
		u.Phone = &ph.String
	}
	if tit.Valid {
		u.Title = &tit.String
	}
	u.Administrator = admin != 0
	u.Deactivated = deact != 0
	return u, nil
}

// ListTranUsers returns users (User table). Query: include_deactivated=1 to include deactivated accounts.
func (h *Handlers) ListTranUsers(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	includeDeactivated := c.Query("include_deactivated") == "1" || strings.EqualFold(c.Query("include_deactivated"), "true")
	q := "SELECT " + tranUserSelectCols + " FROM `User` WHERE 1=1"
	if !includeDeactivated {
		q += " AND Deactivated = 0"
	}
	q += " ORDER BY LastName, FirstName LIMIT 500"
	rows, err := h.TranMySQL.DB.Query(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []models.TranUser
	for rows.Next() {
		u, err := scanTranUserRow(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, u)
	}
	c.JSON(http.StatusOK, list)
}

func tranUserByEmail(h *Handlers, email string) (models.TranUser, error) {
	var u models.TranUser
	if h.TranMySQL == nil {
		return u, errors.New("Tran SQL store not configured")
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return u, errors.New("email required")
	}
	row := h.TranMySQL.DB.QueryRow(
		"SELECT "+tranUserSelectCols+" FROM `User` WHERE Email IS NOT NULL AND LOWER(TRIM(Email)) = LOWER(?) AND Deactivated = 0 LIMIT 1",
		email)
	u, err := scanTranUserRow(row)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return u, err
	}
	row = h.TranMySQL.DB.QueryRow(
		"SELECT "+tranUserSelectCols+" FROM `User` WHERE LoginID IS NOT NULL AND LOWER(TRIM(LoginID)) = LOWER(?) AND Deactivated = 0 LIMIT 1",
		email)
	return scanTranUserRow(row)
}

func deriveNamesFromEmail(email string) (firstName, lastName string) {
	local := strings.TrimSpace(email)
	if i := strings.Index(local, "@"); i > 0 {
		local = local[:i]
	}
	local = strings.NewReplacer(".", " ", "_", " ", "-", " ").Replace(local)
	parts := strings.Fields(local)
	if len(parts) == 0 {
		return "", "User"
	}
	if len(parts) == 1 {
		return "", parts[0]
	}
	return parts[0], strings.Join(parts[1:], " ")
}

// ensureTranUserByAuthEmail returns an active User row for the signed-in email, creating a minimal profile when missing.
func ensureTranUserByAuthEmail(h *Handlers, email string) (models.TranUser, error) {
	u, err := tranUserByEmail(h, email)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return u, err
	}
	firstName, lastName := deriveNamesFromEmail(email)
	loginID := email
	em := email
	var fn interface{}
	if strings.TrimSpace(firstName) != "" {
		fn = firstName
	}
	res, err := h.TranMySQL.DB.Exec(
		"INSERT INTO `User` (Administrator, LoginID, FirstName, LastName, Email, Deactivated) VALUES (0, ?, ?, ?, ?, 0)",
		loginID, fn, lastName, em)
	if err != nil {
		return u, err
	}
	id64, _ := res.LastInsertId()
	row := h.TranMySQL.DB.QueryRow("SELECT "+tranUserSelectCols+" FROM `User` WHERE UserID = ?", id64)
	return scanTranUserRow(row)
}

var allowedTranUserSelfWrite = map[string]string{
	"first_name": "FirstName",
	"last_name":  "LastName",
	"email":      "Email",
	"phone":      "Phone",
}

// GetTranUserMe returns the authenticated user's Morph profile (basic fields only).
func (h *Handlers) GetTranUserMe(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	email := authEmailFromContext(c)
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	u, err := ensureTranUserByAuthEmail(h, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":         u.ID,
		"first_name": u.FirstName,
		"last_name":  u.LastName,
		"email":      u.Email,
		"phone":      u.Phone,
	})
}

// UpdateTranUserMe lets the signed-in user edit their own basic profile fields.
func (h *Handlers) UpdateTranUserMe(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	email := authEmailFromContext(c)
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	current, err := ensureTranUserByAuthEmail(h, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var in map[string]interface{}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if v, ok := in["last_name"]; ok {
		s, ok := v.(string)
		if !ok || strings.TrimSpace(s) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "last_name cannot be empty"})
			return
		}
	}
	var sets []string
	var args []interface{}
	for k, v := range in {
		lk := strings.ToLower(strings.TrimSpace(k))
		col, ok := allowedTranUserSelfWrite[lk]
		if !ok {
			continue
		}
		sets = append(sets, col+" = ?")
		args = append(args, v)
	}
	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}
	args = append(args, current.ID)
	q := "UPDATE `User` SET " + strings.Join(sets, ", ") + " WHERE UserID = ?"
	if _, err := h.TranMySQL.DB.Exec(q, args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	row := h.TranMySQL.DB.QueryRow("SELECT "+tranUserSelectCols+" FROM `User` WHERE UserID = ?", current.ID)
	u, err := scanTranUserRow(row)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"updated": current.ID})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":         u.ID,
		"first_name": u.FirstName,
		"last_name":  u.LastName,
		"email":      u.Email,
		"phone":      u.Phone,
	})
}

func authEmailFromContext(c *gin.Context) string {
	if v, ok := c.Get("auth_email"); ok {
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

// GetTranUser returns one user by UserID.
func (h *Handlers) GetTranUser(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	row := h.TranMySQL.DB.QueryRow("SELECT "+tranUserSelectCols+" FROM `User` WHERE UserID = ?", id)
	u, err := scanTranUserRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}

// CreateTranUser inserts a user (password not set; sign-in integration can follow).
func (h *Handlers) CreateTranUser(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	var in struct {
		LoginID       *string `json:"login_id"`
		FirstName     *string `json:"first_name"`
		LastName      string  `json:"last_name"`
		Email         *string `json:"email"`
		Phone         *string `json:"phone"`
		Title         *string `json:"title"`
		Administrator *bool   `json:"administrator"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if strings.TrimSpace(in.LastName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "last_name is required"})
		return
	}
	admin := 0
	if in.Administrator != nil && *in.Administrator {
		admin = 1
	}
	res, err := h.TranMySQL.DB.Exec(
		"INSERT INTO `User` (Administrator, LoginID, FirstName, LastName, Email, Phone, Title, Deactivated) VALUES (?, ?, ?, ?, ?, ?, ?, 0)",
		admin, in.LoginID, in.FirstName, in.LastName, in.Email, in.Phone, in.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	row := h.TranMySQL.DB.QueryRow("SELECT "+tranUserSelectCols+" FROM `User` WHERE UserID = ?", id64)
	u, err := scanTranUserRow(row)
	if err != nil {
		c.JSON(http.StatusCreated, gin.H{"id": id64})
		return
	}
	syncErr := syncTranUserToUsersPanel(u)
	if syncErr != nil {
		c.JSON(http.StatusCreated, gin.H{
			"user":              u,
			"users_panel_sync":  "failed",
			"users_panel_error": syncErr.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"user":             u,
		"users_panel_sync": "ok",
	})
}

var allowedTranUserWrite = map[string]string{
	"login_id":      "LoginID",
	"first_name":    "FirstName",
	"last_name":     "LastName",
	"email":         "Email",
	"phone":         "Phone",
	"title":         "Title",
	"administrator": "Administrator",
	"deactivated":   "Deactivated",
}

// UpdateTranUser updates profile fields; deactivated toggles soft-offboarding (DeactivatedDate).
func (h *Handlers) UpdateTranUser(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var in map[string]interface{}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if v, ok := in["last_name"]; ok {
		s, ok := v.(string)
		if !ok || strings.TrimSpace(s) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "last_name cannot be empty"})
			return
		}
	}

	var sets []string
	var args []interface{}
	var deactivate *bool
	for k, v := range in {
		lk := strings.ToLower(strings.TrimSpace(k))
		if lk == "id" || lk == "user_id" {
			continue
		}
		col, ok := allowedTranUserWrite[lk]
		if !ok {
			continue
		}
		if col == "Deactivated" {
			if b, ok := v.(bool); ok {
				deactivate = &b
			} else if f, ok := v.(float64); ok {
				b := f != 0
				deactivate = &b
			}
			continue
		}
		if col == "Administrator" {
			var bit int
			if b, ok := v.(bool); ok && b {
				bit = 1
			} else if f, ok := v.(float64); ok && f != 0 {
				bit = 1
			}
			sets = append(sets, "Administrator = ?")
			args = append(args, bit)
			continue
		}
		sets = append(sets, col+" = ?")
		args = append(args, v)
	}
	if deactivate != nil {
		deactVal := 0
		if *deactivate {
			deactVal = 1
		}
		sets = append(sets, "Deactivated = ?")
		args = append(args, deactVal)
		if *deactivate {
			sets = append(sets, "DeactivatedDate = ?")
			args = append(args, time.Now())
		} else {
			sets = append(sets, "DeactivatedDate = NULL")
		}
	}
	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}
	args = append(args, id)
	q := "UPDATE `User` SET " + strings.Join(sets, ", ") + " WHERE UserID = ?"
	_, err = h.TranMySQL.DB.Exec(q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	row := h.TranMySQL.DB.QueryRow("SELECT "+tranUserSelectCols+" FROM `User` WHERE UserID = ?", id)
	u, err := scanTranUserRow(row)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"updated": id})
		return
	}
	c.JSON(http.StatusOK, u)
}

// DeleteTranUser soft-deactivates a user (keeps history and foreign keys).
func (h *Handlers) DeleteTranUser(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	now := time.Now()
	res, err := h.TranMySQL.DB.Exec(
		"UPDATE `User` SET Deactivated = 1, DeactivatedDate = ? WHERE UserID = ? AND Deactivated = 0", now, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found or already deactivated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "deactivated": true})
}

func syncTranUserToUsersPanel(u models.TranUser) error {
	baseURL := strings.TrimSpace(os.Getenv("USERS_PANEL_BASE_URL"))
	if baseURL == "" {
		return nil
	}
	email := strings.TrimSpace(valueOrEmpty(u.Email))
	if email == "" {
		return nil
	}
	username := strings.TrimSpace(valueOrEmpty(u.LoginID))
	if username == "" {
		fullName := strings.TrimSpace(strings.Join([]string{valueOrEmpty(u.FirstName), strings.TrimSpace(u.LastName)}, " "))
		username = strings.ReplaceAll(strings.ToLower(fullName), " ", "_")
	}
	if username == "" {
		username = fmt.Sprintf("tran_user_%d", u.ID)
	}

	registerBody := map[string]string{
		"email":    email,
		"username": username,
		"password": fmt.Sprintf("Tmp!%dPass2026", u.ID),
	}
	raw, _ := json.Marshal(registerBody)
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(baseURL, "/")+"/api/auth/register", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	// Register can fail with conflict when account already exists; treat as synced.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return fmt.Errorf("register failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
