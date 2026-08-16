package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// UpdateVehicle updates vehicle fields by ID (same writable columns as create).
func (h *Handlers) UpdateVehicle(c *gin.Context) {
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
	detailStr, hasDetail, derr := popDetailString(in)
	if derr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": derr.Error()})
		return
	}
	allowed := map[string]struct{}{
		"asset_tag": {}, "description": {}, "AssetID": {}, "AssetType": {}, "ContractorID": {},
	}
	var sets []string
	var args []interface{}
	for k, v := range in {
		lk := strings.ToLower(k)
		if lk == "id" {
			continue
		}
		if strings.HasSuffix(lk, "_id") && lk != "asset_id" && lk != "contractor_id" {
			continue
		}
		col := k
		switch lk {
		case "vin", "asset_tag":
			col = "asset_tag"
		case "description":
			col = "description"
		case "asset_id", "assetid":
			col = "AssetID"
		case "asset_type":
			col = "AssetType"
		case "contractor_id", "contractorid":
			col = "ContractorID"
			switch t := v.(type) {
			case float64:
				v = int(t)
			case string:
				if n, e := strconv.Atoi(strings.TrimSpace(t)); e == nil {
					v = n
				}
			}
		}
		if _, ok := allowed[col]; !ok {
			continue
		}
		if col == "asset_tag" || col == "description" || col == "AssetID" || col == "AssetType" {
			if s, ok := v.(string); ok {
				t := strings.TrimSpace(s)
				if t == "" {
					v = nil
				} else {
					v = t
				}
			}
		}
		sets = append(sets, col+" = ?")
		args = append(args, v)
	}
	if len(sets) == 0 {
		m, ok, _ := querySingleRowMap(h.TranMySQL.DB, "SELECT "+vehicleFullSelectCols+" FROM Asset WHERE ID = ?", id)
		if !ok || m == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vehicle not found"})
			return
		}
		if !hasDetail {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}
		_ = h.savePoppedDetail(c, entityKeyVehicle, id, detailStr)
		h.attachEntityDetail(c, entityKeyVehicle, id, m)
		c.JSON(http.StatusOK, m)
		return
	}
	args = append(args, id)
	_, err = h.TranMySQL.DB.Exec("UPDATE Asset SET "+strings.Join(sets, ", ")+" WHERE ID = ?", args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeyVehicle, id, detailStr)
	}
	m, _, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+vehicleFullSelectCols+" FROM Asset WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"updated": id})
		return
	}
	if m != nil {
		h.attachEntityDetail(c, entityKeyVehicle, id, m)
	}
	c.JSON(http.StatusOK, m)
}

// UpdateTrip updates trip fields by ID.
func (h *Handlers) UpdateTrip(c *gin.Context) {
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
	detailStr, hasDetail, derr := popDetailString(in)
	if derr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": derr.Error()})
		return
	}
	allowed := map[string]struct{}{
		"Name": {}, "ActivityType": {}, "start_date": {}, "end_date": {}, "location": {}, "GUID": {}, "description": {},
	}
	var sets []string
	var args []interface{}
	for k, v := range in {
		lk := strings.ToLower(k)
		if lk == "id" || lk == "db_id" {
			continue
		}
		if strings.HasSuffix(lk, "_id") {
			continue
		}
		col := k
		argVal := v
		switch lk {
		case "name":
			col = "Name"
		case "activity_type":
			col = "ActivityType"
		case "start_date":
			col = "start_date"
			argVal = normalizeTranDateTimeColVal(v)
		case "end_date":
			col = "end_date"
			argVal = normalizeTranDateTimeColVal(v)
		case "location":
			col = "location"
			norm, e := normalizeLocationValue(v)
			if e != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location json"})
				return
			}
			argVal = norm
		case "guid":
			col = "GUID"
		case "description":
			col = "description"
			if s, ok := v.(string); ok {
				t := strings.TrimSpace(s)
				if t == "" {
					argVal = nil
				} else {
					argVal = t
				}
			}
		default:
			continue
		}
		if _, ok := allowed[col]; !ok {
			continue
		}
		sets = append(sets, col+" = ?")
		args = append(args, argVal)
	}
	if len(sets) == 0 {
		m, ok, _ := querySingleRowMap(h.TranMySQL.DB, "SELECT "+tripFullSelectCols+" FROM Activity WHERE ID = ?", id)
		if !ok || m == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
			return
		}
		if !hasDetail {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}
		_ = h.savePoppedDetail(c, entityKeyTrip, id, detailStr)
		h.attachEntityDetail(c, entityKeyTrip, id, m)
		c.JSON(http.StatusOK, m)
		return
	}
	args = append(args, id)
	_, err = h.TranMySQL.DB.Exec("UPDATE Activity SET "+strings.Join(sets, ", ")+" WHERE ID = ?", args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeyTrip, id, detailStr)
	}
	m, _, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+tripFullSelectCols+" FROM Activity WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"updated": id})
		return
	}
	if m != nil {
		h.attachEntityDetail(c, entityKeyTrip, id, m)
	}
	c.JSON(http.StatusOK, m)
}
