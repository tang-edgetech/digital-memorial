package handlers

import (
	"net/http"

	"digital-memorial/api/internal/services"

	"github.com/gin-gonic/gin"
)

func GetSettings(c *gin.Context) {
	settings, err := services.GetAllSettings()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings", "code": "settings_load_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

type updateSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

func UpdateSettingsHandler(c *gin.Context) {
	var req updateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "invalid_request"})
		return
	}

	var updatedBy *uint
	if userID, ok := c.Get("userID"); ok {
		if id, ok := userID.(uint); ok {
			updatedBy = &id
		}
	}

	before, after, err := services.UpdateSettings(req.Settings, updatedBy)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings", "code": "settings_update_failed"})
		return
	}

	services.Log(c, "settings.update", "site_setting", "batch", before, after)
	services.RefreshSetupCompletedFromDB()

	settings, _ := services.GetAllSettings()
	c.JSON(http.StatusOK, gin.H{"settings": settings})
}
