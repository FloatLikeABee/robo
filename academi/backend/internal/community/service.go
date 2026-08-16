package community

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/academi/backend/internal/auth"
	"github.com/academi/backend/internal/database"
	"github.com/academi/backend/internal/models"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func authorDisplayName(userID string) string {
	if userID == "" {
		return "Anonymous"
	}
	data, err := database.Get([]byte("user:" + userID))
	if err != nil {
		return "Member"
	}
	var u models.User
	if json.Unmarshal(data, &u) != nil || strings.TrimSpace(u.Name) == "" {
		return "Member"
	}
	return u.Name
}

func docExists(docID string) bool {
	if docID == "" {
		return false
	}
	_, err := database.Get([]byte("doc:" + docID))
	return err == nil
}

func docTitleFallback(docID string) string {
	data, err := database.Get([]byte("doc:" + docID))
	if err != nil {
		return "Document"
	}
	var d struct {
		Title string `json:"title"`
	}
	if json.Unmarshal(data, &d) != nil || strings.TrimSpace(d.Title) == "" {
		return "Document"
	}
	return d.Title
}

func (s *Service) CreatePost(c *gin.Context) {
	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	content := strings.TrimSpace(req.Content)
	docID := strings.TrimSpace(req.DocID)
	if content == "" && docID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content or doc_id is required"})
		return
	}

	if docID != "" && !docExists(docID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	userID := auth.GetUserID(c)
	docTitle := strings.TrimSpace(req.DocTitle)
	if docTitle == "" && docID != "" {
		docTitle = docTitleFallback(docID)
	}

	finalContent := content
	tags := append([]string{}, req.Tags...)
	if docID != "" {
		if finalContent == "" {
			finalContent = "📎 Shared a document: **" + docTitle + "**"
		} else {
			finalContent += "\n\n—\n📎 **" + docTitle + "**"
		}
		hasDocTag := false
		for _, t := range tags {
			if strings.EqualFold(t, "#docs") || strings.EqualFold(t, "#published") {
				hasDocTag = true
				break
			}
		}
		if !hasDocTag {
			tags = append(tags, "#docs", "#Published")
		}
	}

	post := models.CommunityPost{
		ID:         uuid.New().String(),
		AuthorID:   userID,
		AuthorName: authorDisplayName(userID),
		Content:    finalContent,
		Tags:       tags,
		DocID:      docID,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
		Upvotes:    0,
		Downvotes:  0,
		Comments:   0,
		AIVerified: false,
	}

	data, _ := json.Marshal(post)
	if err := database.Set([]byte("post:"+post.ID), data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	if err := database.Set([]byte("user:"+userID+":posts:"+post.ID), []byte("1")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link post to user"})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func (s *Service) GetPost(c *gin.Context) {
	postID := c.Param("id")

	data, err := database.Get([]byte("post:" + postID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var post models.CommunityPost
	json.Unmarshal(data, &post)

	c.JSON(http.StatusOK, post)
}

func (s *Service) ListPosts(c *gin.Context) {
	var posts []models.CommunityPost

	_ = database.Iterate([]byte("post:"), func(key, value []byte) error {
		var post models.CommunityPost
		if err := json.Unmarshal(value, &post); err == nil {
			if post.AuthorName == "" {
				post.AuthorName = authorDisplayName(post.AuthorID)
			}
			posts = append(posts, post)
		}
		return nil
	})

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt > posts[j].CreatedAt
	})

	c.JSON(http.StatusOK, posts)
}

func (s *Service) ListComments(c *gin.Context) {
	postID := c.Param("id")
	prefix := []byte("comm:" + postID + ":")

	var comments []models.Comment
	_ = database.Iterate(prefix, func(key, value []byte) error {
		var cm models.Comment
		if json.Unmarshal(value, &cm) == nil {
			if cm.AuthorName == "" {
				cm.AuthorName = authorDisplayName(cm.AuthorID)
			}
			comments = append(comments, cm)
		}
		return nil
	})
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt < comments[j].CreatedAt
	})
	c.JSON(http.StatusOK, comments)
}

func (s *Service) CreateComment(c *gin.Context) {
	postID := c.Param("id")
	_, err := database.Get([]byte("post:" + postID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := auth.GetUserID(c)
	cm := models.Comment{
		ID:         uuid.New().String(),
		PostID:     postID,
		AuthorID:   userID,
		AuthorName: authorDisplayName(userID),
		Content:    strings.TrimSpace(req.Content),
		CreatedAt:  time.Now().Unix(),
	}
	if cm.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	ckey := []byte("comm:" + postID + ":" + cm.ID)
	raw, _ := json.Marshal(cm)
	if err := database.Set(ckey, raw); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save comment"})
		return
	}

	pdata, _ := database.Get([]byte("post:" + postID))
	var post models.CommunityPost
	if json.Unmarshal(pdata, &post) == nil {
		post.Comments++
		post.UpdatedAt = time.Now().Unix()
		out, _ := json.Marshal(post)
		_ = database.Set([]byte("post:"+postID), out)
	}

	c.JSON(http.StatusCreated, cm)
}

func (s *Service) VotePost(c *gin.Context) {
	postID := c.Param("id")

	data, err := database.Get([]byte("post:" + postID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var post models.CommunityPost
	json.Unmarshal(data, &post)

	var vote struct {
		Type string `json:"type" binding:"required,oneof=up down"`
	}

	if err := c.ShouldBindJSON(&vote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if vote.Type == "up" {
		post.Upvotes++
	} else {
		post.Downvotes++
	}

	post.UpdatedAt = time.Now().Unix()
	data, _ = json.Marshal(post)
	database.Set([]byte("post:"+postID), data)

	c.JSON(http.StatusOK, gin.H{
		"upvotes":   post.Upvotes,
		"downvotes": post.Downvotes,
		"score":     post.Upvotes - post.Downvotes,
	})
}

func (s *Service) SearchPosts(c *gin.Context) {
	query := strings.ToLower(c.Query("q"))
	tag := c.Query("tag")

	var posts []models.CommunityPost

	_ = database.Iterate([]byte("post:"), func(key, value []byte) error {
		var post models.CommunityPost
		if json.Unmarshal(value, &post) != nil {
			return nil
		}

		match := true
		if query != "" {
			match = match && (strings.Contains(strings.ToLower(post.Content), query) ||
				containsSliceIgnoreCase(post.Tags, query))
		}
		if tag != "" {
			match = match && containsSlice(post.Tags, tag)
		}

		if match {
			if post.AuthorName == "" {
				post.AuthorName = authorDisplayName(post.AuthorID)
			}
			posts = append(posts, post)
		}
		return nil
	})

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt > posts[j].CreatedAt
	})
	c.JSON(http.StatusOK, posts)
}

func containsSliceIgnoreCase(slice []string, query string) bool {
	q := strings.ToLower(query)
	for _, item := range slice {
		if strings.Contains(strings.ToLower(item), q) {
			return true
		}
	}
	return false
}

func containsSlice(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (s *Service) RegisterRoutes(r *gin.RouterGroup) {
	posts := r.Group("/posts")
	{
		posts.POST("", s.CreatePost)
		posts.GET("", s.ListPosts)
		posts.GET("/search", s.SearchPosts)
		posts.GET("/:id/comments", s.ListComments)
		posts.POST("/:id/comments", s.CreateComment)
		posts.GET("/:id", s.GetPost)
		posts.POST("/:id/vote", s.VotePost)
	}
}
