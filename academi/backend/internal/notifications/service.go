package notifications

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/academi/backend/internal/database"
	"github.com/academi/backend/internal/models"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CreateNotification(c *gin.Context) {
	var req models.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notif := models.Notification{
		ID:          uuid.New().String(),
		UserID:      req.UserID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Timestamp:   time.Now().Format("2006-01-02T15:04:05Z"),
		Icon:        req.Icon,
		Read:        false,
		CreatedAt:   time.Now().Unix(),
	}

	data, _ := json.Marshal(notif)
	if err := database.Set([]byte("notif:"+notif.UserID+":"+notif.ID), data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	c.JSON(http.StatusCreated, notif)
}

func (s *Service) GetUserNotifications(c *gin.Context) {
	userID := c.Param("userId")
	var notifs []models.Notification

	database.Iterate([]byte("notif:"+userID+":"), func(key, value []byte) error {
		var notif models.Notification
		if err := json.Unmarshal(value, &notif); err == nil {
			notifs = append(notifs, notif)
		}
		return nil
	})

	unread := 0
	for _, n := range notifs {
		if !n.Read {
			unread++
		}
	}

	resp := models.NotificationResponse{
		Notifications: notifs,
		UnreadCount:   unread,
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Service) MarkRead(c *gin.Context) {
	userID := c.Param("userId")
	notifID := c.Param("id")

	data, err := database.Get([]byte("notif:" + userID + ":" + notifID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	var notif models.Notification
	json.Unmarshal(data, &notif)
	notif.Read = true

	ndata, _ := json.Marshal(notif)
	database.Set([]byte("notif:"+userID+":"+notifID), ndata)

	c.JSON(http.StatusOK, notif)
}

func (s *Service) RegisterRoutes(r *gin.RouterGroup) {
	notifs := r.Group("/notifications")
	{
		notifs.POST("", s.CreateNotification)
		notifs.GET("/user/:userId", s.GetUserNotifications)
		notifs.POST("/:userId/:id/read", s.MarkRead)
	}
}
