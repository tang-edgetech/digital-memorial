package handlers

import (
	"net/http"

	"digital-memorial/api/internal/services"

	"github.com/gin-gonic/gin"
)

func GetPermissions(c *gin.Context) {
	rows, err := services.GetAllPermissions()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load permissions", "code": "permissions_load_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"permissions": rows, "registry": services.PermissionRegistry})
}

type updatePermissionsRequest struct {
	Updates []services.PermissionUpdate `json:"updates" binding:"required"`
}

func UpdatePermissions(c *gin.Context) {
	var req updatePermissionsRequest
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

	before, after, err := services.UpdatePermissions(req.Updates, updatedBy)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to update permissions", "code": "permissions_update_failed"})
		return
	}

	services.Log(c, "permissions.update", "role_permission", "batch", before, after)
	services.RefreshPermissionsFromDB()

	rows, _ := services.GetAllPermissions()
	c.JSON(http.StatusOK, gin.H{"permissions": rows})
}
