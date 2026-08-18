package handlers

import (
	"fmt"
	"net/http"
	"time"

	"digital-memorial/api/internal/middleware"
	"digital-memorial/api/internal/models"
	"digital-memorial/api/internal/services"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "invalid_request"})
		return
	}

	user, err := services.GetUserByEmail(req.Email)
	if err != nil {
		services.Log(c, "auth.login_failed", "user", req.Email, nil, gin.H{"reason": "unknown_email"})
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password", "code": "invalid_credentials"})
		return
	}

	c.Set("userID", user.ID)
	c.Set("userEmail", user.Email)

	if services.IsLocked(user) {
		services.Log(c, "auth.login_failed", "user", fmt.Sprintf("%d", user.ID), nil, gin.H{"reason": "locked"})
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "account is temporarily locked, try again later", "code": "account_locked"})
		return
	}

	if !user.IsActive {
		services.Log(c, "auth.login_failed", "user", fmt.Sprintf("%d", user.ID), nil, gin.H{"reason": "disabled"})
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "account is disabled", "code": "account_disabled"})
		return
	}

	if !services.CheckPassword(user.PasswordHash, req.Password) {
		lockoutThreshold := services.GetSettingInt("lockout_threshold", 5)
		lockoutMinutes := services.GetSettingInt("lockout_duration_minutes", 15)
		services.RecordFailedLogin(user, lockoutThreshold, time.Duration(lockoutMinutes)*time.Minute)
		services.Log(c, "auth.login_failed", "user", fmt.Sprintf("%d", user.ID), nil, gin.H{"reason": "bad_password"})
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password", "code": "invalid_credentials"})
		return
	}

	services.RecordSuccessfulLogin(user)

	timeoutMinutes := services.GetSettingInt("session_timeout_minutes", 120)
	ttl := time.Duration(timeoutMinutes) * time.Minute
	token, expiresAt, err := services.IssueToken(user.ID, user.Email, string(user.Role), ttl)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to issue session", "code": "token_issue_failed"})
		return
	}
	middleware.SetSessionCookie(c, token, expiresAt)

	if csrfToken, err := services.GenerateRandomToken(24); err == nil {
		middleware.SetCSRFCookie(c, csrfToken, int(time.Until(expiresAt).Seconds()))
	}

	services.Log(c, "auth.login", "user", fmt.Sprintf("%d", user.ID), nil, nil)

	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(user)})
}

// Logout is intentionally reachable without a valid (non-expired) session so
// an idle-expired client can still call it. It best-effort identifies the
// actor from the cookie even if the token already expired, and distinguishes
// a manual logout from an idle-triggered one for the audit trail.
func Logout(c *gin.Context) {
	cookie, cookieErr := c.Cookie(middleware.SessionCookieName)
	middleware.ClearAuthCookies(c)

	if cookieErr != nil || cookie == "" {
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	claims, wasExpired, err := services.ParseTokenAllowExpired(cookie)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	c.Set("userID", claims.UserID)
	c.Set("userEmail", claims.Email)

	action := "auth.logout"
	if wasExpired {
		action = "auth.idle_logout"
	}
	services.Log(c, action, "user", fmt.Sprintf("%d", claims.UserID), nil, nil)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func Me(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, ok := userID.(uint)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated", "code": "unauthenticated"})
		return
	}
	user, err := services.GetUserByID(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated", "code": "unauthenticated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(user)})
}

func toUserResponse(u *models.User) gin.H {
	return gin.H{
		"id":              u.ID,
		"email":           u.Email,
		"fullName":        u.FullName,
		"role":            u.Role,
		"themePreference": u.ThemePreference,
	}
}
