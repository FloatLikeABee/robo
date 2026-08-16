package handlers

import (
	"net/http"

	"github.com/academi/booki/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (a *API) ListCustomers(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	rows, err := a.DB.Query(`SELECT id, name, email, phone FROM customers WHERE organization_id=? ORDER BY name`, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id int64
		var name, email, phone string
		_ = rows.Scan(&id, &name, &email, &phone)
		list = append(list, gin.H{"id": id, "name": name, "email": email, "phone": phone})
	}
	c.JSON(http.StatusOK, gin.H{"customers": list})
}

func (a *API) CreateCustomer(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	var body struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := a.DB.Exec(`INSERT INTO customers (organization_id, name, email, phone) VALUES (?,?,?,?)`, orgID, body.Name, body.Email, body.Phone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}
