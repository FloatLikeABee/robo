package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/academi/booki/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (a *API) ListAssets(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	rows, err := a.DB.Query(`SELECT id, asset_tag, name, category, purchase_value, current_value, depreciation_method, useful_life_months, location, status
		FROM assets WHERE organization_id=? ORDER BY asset_tag`, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id int64
		var tag, name, cat, dm, loc, status string
		var pv, cv float64
		var life int
		_ = rows.Scan(&id, &tag, &name, &cat, &pv, &cv, &dm, &life, &loc, &status)
		list = append(list, gin.H{
			"id": id, "asset_tag": tag, "name": name, "category": cat,
			"purchase_value": pv, "current_value": cv, "depreciation_method": dm,
			"useful_life_months": life, "location": loc, "status": status,
		})
	}
	c.JSON(http.StatusOK, gin.H{"assets": list})
}

func (a *API) CreateAsset(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	var body struct {
		AssetTag            string  `json:"asset_tag" binding:"required"`
		Name                string  `json:"name" binding:"required"`
		Category            string  `json:"category"`
		PurchaseValue       float64 `json:"purchase_value"`
		CurrentValue        float64 `json:"current_value"`
		DepreciationMethod  string  `json:"depreciation_method"`
		UsefulLifeMonths    int     `json:"useful_life_months"`
		Location            string  `json:"location"`
		Status              string  `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.DepreciationMethod == "" {
		body.DepreciationMethod = "straight_line"
	}
	if body.Status == "" {
		body.Status = "active"
	}
	if body.CurrentValue == 0 && body.PurchaseValue > 0 {
		body.CurrentValue = body.PurchaseValue
	}
	res, err := a.DB.Exec(`INSERT INTO assets (organization_id, asset_tag, name, category, purchase_value, current_value, depreciation_method, useful_life_months, location, status)
		VALUES (?,?,?,?,?,?,?,?,?,?)`, orgID, body.AssetTag, body.Name, body.Category, body.PurchaseValue, body.CurrentValue, body.DepreciationMethod, body.UsefulLifeMonths, body.Location, body.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func morphAssetSelectSQL() string {
	return `SELECT asset_tag, description, AssetID, AssetType FROM Asset ORDER BY ID`
}

func nullStringValue(v sql.NullString) string {
	if v.Valid {
		return strings.TrimSpace(v.String)
	}
	return ""
}

// SyncMorphAssets upserts MorphData Asset rows into Booki assets for the current organization.
func (a *API) SyncMorphAssets(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	rows, err := a.DB.Query(morphAssetSelectSQL())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "MorphData Asset table unavailable: " + err.Error()})
		return
	}
	defer rows.Close()

	created := 0
	updated := 0
	skipped := 0

	for rows.Next() {
		var tag, desc, assetID, assetType sql.NullString
		if err := rows.Scan(&tag, &desc, &assetID, &assetType); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		assetTag := nullStringValue(tag)
		if assetTag == "" {
			assetTag = nullStringValue(assetID)
		}
		if assetTag == "" {
			skipped++
			continue
		}
		name := nullStringValue(desc)
		if name == "" {
			name = assetTag
		}
		category := nullStringValue(assetType)
		location := nullStringValue(assetID)

		var existingID int64
		err := a.DB.QueryRow(
			`SELECT id FROM assets WHERE organization_id=? AND asset_tag=? LIMIT 1`,
			orgID, assetTag,
		).Scan(&existingID)
		if err == sql.ErrNoRows {
			_, err = a.DB.Exec(`INSERT INTO assets (organization_id, asset_tag, name, category, purchase_value, current_value, depreciation_method, useful_life_months, location, status)
				VALUES (?,?,?,?,0,0,'straight_line',0,?,'active')`, orgID, assetTag, name, category, location)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			created++
			continue
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		_, err = a.DB.Exec(
			`UPDATE assets SET name=?, category=?, location=? WHERE id=? AND organization_id=?`,
			name, category, location, existingID, orgID,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updated++
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"created": created,
		"updated": updated,
		"skipped": skipped,
		"total":   created + updated,
	})
}
