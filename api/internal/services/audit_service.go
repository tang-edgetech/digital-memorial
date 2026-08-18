package services

import (
	"encoding/json"
	"log"

	"digital-memorial/api/internal/db"
	"digital-memorial/api/internal/models"

	"github.com/gin-gonic/gin"
)

// Log records an audit trail entry for a mutation or auth event. Actor
// info is read from the Gin context (set by the auth middleware, or by
// handlers directly for pre-session events like login attempts). Best-effort:
// a failure to write the audit row is logged to stderr and never propagated,
// so a broken audit insert can never block a legitimate user action.
func Log(c *gin.Context, action, targetType, targetID string, before, after any) {
	gdb := db.Get()
	if gdb == nil {
		return
	}

	entry := models.AuditLog{
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	}

	if userID, ok := c.Get("userID"); ok {
		if id, ok := userID.(uint); ok {
			entry.ActorUserID = &id
		}
	}
	if email, ok := c.Get("userEmail"); ok {
		if s, ok := email.(string); ok {
			entry.ActorEmail = s
		}
	}

	if before != nil {
		if b, err := json.Marshal(before); err == nil {
			s := string(b)
			entry.BeforeValue = &s
		}
	}
	if after != nil {
		if b, err := json.Marshal(after); err == nil {
			s := string(b)
			entry.AfterValue = &s
		}
	}

	if err := gdb.Create(&entry).Error; err != nil {
		log.Printf("audit log write failed: %v", err)
	}
}
