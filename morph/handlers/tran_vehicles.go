package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

// List / GetVehicle (Asset table essentials for grid).
const vehicleListCols = `ID, asset_tag, description, AssetID, AssetType`

// Aliased for GetVehicleFull / post-create JSON maps.
const vehicleFullSelectCols = `ID AS id, ContractorID AS contractor_id, asset_tag AS asset_tag, description AS description, AssetID AS asset_id, AssetType AS asset_type`

func scanVehicleListRow(rows *sql.Rows) (models.Vehicle, error) {
	var v models.Vehicle
	var assetTag, desc, assetID, atype sql.NullString
	err := rows.Scan(&v.ID, &assetTag, &desc, &assetID, &atype)
	if err != nil {
		return v, err
	}
	if assetTag.Valid {
		s := assetTag.String
		v.AssetTag = &s
	}
	if desc.Valid {
		s := desc.String
		v.Description = &s
	}
	if assetID.Valid {
		s := assetID.String
		v.AssetID = &s
	}
	if atype.Valid {
		s := atype.String
		v.AssetType = &s
	}
	return v, nil
}

func scanVehicleGetRow(row *sql.Row) (models.Vehicle, error) {
	var v models.Vehicle
	var assetTag, desc, assetID, atype sql.NullString
	err := row.Scan(&v.ID, &assetTag, &desc, &assetID, &atype)
	if err != nil {
		return v, err
	}
	if assetTag.Valid {
		s := assetTag.String
		v.AssetTag = &s
	}
	if desc.Valid {
		s := desc.String
		v.Description = &s
	}
	if assetID.Valid {
		s := assetID.String
		v.AssetID = &s
	}
	if atype.Valid {
		s := atype.String
		v.AssetType = &s
	}
	return v, nil
}

// ListVehicles returns vehicles.
func (h *Handlers) ListVehicles(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	query := "SELECT " + vehicleListCols + " FROM Asset WHERE 1=1 ORDER BY ID LIMIT 500"

	rows, err := h.TranMySQL.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []models.Vehicle
	for rows.Next() {
		v, err := scanVehicleListRow(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, v)
	}
	c.JSON(http.StatusOK, list)
}

// GetVehicle returns one vehicle by ID.
func (h *Handlers) GetVehicle(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	row := h.TranMySQL.DB.QueryRow("SELECT "+vehicleListCols+" FROM Asset WHERE ID = ?", id)
	v, err := scanVehicleGetRow(row)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vehicle not found"})
		return
	}
	c.JSON(http.StatusOK, v)
}

// GetVehicleFull returns Vehicle fields used by the admin UI as a map.
func (h *Handlers) GetVehicleFull(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	m, ok, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+vehicleFullSelectCols+" FROM Asset WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "vehicle not found"})
		return
	}
	h.attachEntityDetail(c, entityKeyVehicle, id, m)
	h.attachEntityAttachmentsToRow(c.Request.Context(), "assets", entityAttachmentAsset, id, m)
	c.JSON(http.StatusOK, m)
}

var allowedVehicleWrite = map[string]struct{}{
	"asset_tag": {}, "description": {}, "AssetID": {}, "AssetType": {}, "ContractorID": {},
}

// CreateVehicle inserts a vehicle.
func (h *Handlers) CreateVehicle(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	var in map[string]interface{}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	detailStr, hasDetail, derr := popDetailString(in)
	if derr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": derr.Error()})
		return
	}

	cols := []string{}
	args := []interface{}{}

	for k, v := range in {
		lk := strings.ToLower(k)
		if lk == "id" {
			continue
		}
		col := vehicleJSONToCol(lk, k)
		if col == "" {
			if _, ok := allowedVehicleWrite[k]; ok {
				col = k
			} else {
				continue
			}
		}
		if _, ok := allowedVehicleWrite[col]; !ok {
			continue
		}

		val := v
		switch col {
		case "ContractorID":
			n := intFromAny(in, "contractor_id", "ContractorID")
			val = n
		case "AssetType":
			if s, ok := v.(string); ok {
				t := strings.TrimSpace(s)
				if t == "" {
					val = nil
				} else {
					val = t
				}
			}
		case "asset_tag", "description", "AssetID":
			if s, ok := v.(string); ok {
				t := strings.TrimSpace(s)
				if t == "" {
					val = nil
				} else {
					val = t
				}
			}
		}

		cols = append(cols, col)
		args = append(args, val)
	}

	if len(cols) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to insert"})
		return
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(cols)), ",")
	stmt := "INSERT INTO Asset (" + strings.Join(cols, ",") + ") VALUES (" + placeholders + ")"
	res, err := h.TranMySQL.DB.Exec(stmt, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeyVehicle, int(id64), detailStr)
	}
	m, _, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+vehicleFullSelectCols+" FROM Asset WHERE ID = ?", id64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": id64})
		return
	}
	if m != nil {
		h.attachEntityDetail(c, entityKeyVehicle, int(id64), m)
	}
	c.JSON(http.StatusOK, m)
}

func vehicleJSONToCol(lk, origKey string) string {
	switch lk {
	case "vin", "asset_tag":
		return "asset_tag"
	case "description":
		return "description"
	case "asset_id", "assetid":
		return "AssetID"
	case "asset_type":
		return "AssetType"
	case "contractor_id", "contractorid":
		return "ContractorID"
	}
	if origKey != "" {
		if _, ok := allowedVehicleWrite[origKey]; ok {
			return origKey
		}
	}
	return ""
}

// DeleteVehicle deletes a vehicle by ID.
func (h *Handlers) DeleteVehicle(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	res, err := h.TranMySQL.DB.Exec("DELETE FROM Asset WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "vehicle not found"})
		return
	}
	h.deleteEntityDetailMongo(context.Background(), entityKeyVehicle, id)
	h.purgeEntityAttachments(context.Background(), entityAttachmentAsset, id)
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}
