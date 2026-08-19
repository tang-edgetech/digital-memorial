package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"digital-memorial/api/internal/models"
	"digital-memorial/api/internal/services"

	"github.com/gin-gonic/gin"
)

// requesterFromContext loads the current authenticated user's fresh row
// (needed for IsOwner/Role, which aren't worth caching at this request
// volume — same cheap-lookup approach as RequireOwnerOrSuperAdmin).
func requesterFromContext(c *gin.Context) *models.User {
	userID, ok := c.Get("userID")
	if !ok {
		return nil
	}
	id, ok := userID.(uint)
	if !ok {
		return nil
	}
	user, err := services.GetUserByID(id)
	if err != nil {
		return nil
	}
	return user
}

func respondUserServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrUserNotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error(), "code": "not_found"})
	case errors.Is(err, services.ErrForbiddenSelfAction),
		errors.Is(err, services.ErrForbiddenAdminTarget),
		errors.Is(err, services.ErrTransferRequiresPrivilege):
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "forbidden"})
	case errors.Is(err, services.ErrEmailTaken):
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": err.Error(), "code": "email_taken"})
	case errors.Is(err, services.ErrCannotCreateSuperAdmin),
		errors.Is(err, services.ErrOwnerMustTransferFirst),
		errors.Is(err, services.ErrTransferTargetInvalid):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "invalid_request"})
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error", "code": "internal_error"})
	}
}

func ListUsers(c *gin.Context) {
	requesterRole, _ := c.Get("userRole")
	requesterRoleStr, _ := requesterRole.(string)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(services.GetSettingInt("pagination_default", 20))))

	filter := services.UserFilter{
		Role:          c.Query("role"),
		Status:        c.Query("status"),
		Search:        c.Query("search"),
		SortBy:        c.Query("sortBy"),
		SortDir:       c.Query("sortDir"),
		Page:          page,
		PageSize:      pageSize,
		RequesterRole: requesterRoleStr,
	}

	users, total, err := services.ListUsers(filter)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to list users", "code": "list_failed"})
		return
	}

	items := make([]gin.H, 0, len(users))
	for i := range users {
		items = append(items, toUserResponse(&users[i]))
	}
	c.JSON(http.StatusOK, gin.H{"users": items, "total": total, "page": filter.Page, "pageSize": filter.PageSize})
}

type createUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	FullName string `json:"fullName" binding:"required"`
	Role     string `json:"role" binding:"required"`
	IsActive *bool  `json:"isActive"`
}

func CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "invalid_request"})
		return
	}
	if req.Role != string(models.RoleAdmin) && req.Role != string(models.RoleAgent) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "role must be admin or agent", "code": "invalid_role"})
		return
	}
	if minLen := services.GetSettingInt("password_min_length", 8); len(req.Password) < minLen {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("password must be at least %d characters", minLen), "code": "password_too_short"})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	user, err := services.CreateUser(services.CreateUserInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     req.Role,
		IsActive: isActive,
	})
	if err != nil {
		respondUserServiceError(c, err)
		return
	}

	services.Log(c, "user.create", "user", fmt.Sprintf("%d", user.ID), nil, gin.H{"email": user.Email, "role": user.Role})
	c.JSON(http.StatusCreated, gin.H{"user": toUserResponse(user)})
}

func GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id", "code": "invalid_request"})
		return
	}
	requesterRole, _ := c.Get("userRole")
	requesterRoleStr, _ := requesterRole.(string)
	user, svcErr := services.GetUserForViewer(uint(id), models.Role(requesterRoleStr))
	if svcErr != nil {
		respondUserServiceError(c, svcErr)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(user)})
}

type updateUserRequest struct {
	Email    *string `json:"email"`
	FullName *string `json:"fullName"`
	Password *string `json:"password"`
	Role     *string `json:"role"`
}

func UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id", "code": "invalid_request"})
		return
	}
	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "invalid_request"})
		return
	}
	if req.Role != nil && *req.Role != string(models.RoleAdmin) && *req.Role != string(models.RoleAgent) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "role must be admin or agent", "code": "invalid_role"})
		return
	}
	if req.Password != nil && *req.Password != "" {
		if minLen := services.GetSettingInt("password_min_length", 8); len(*req.Password) < minLen {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("password must be at least %d characters", minLen), "code": "password_too_short"})
			return
		}
	}

	requester := requesterFromContext(c)
	if requester == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated", "code": "unauthenticated"})
		return
	}

	before, _ := services.GetUserByID(uint(id))
	user, svcErr := services.UpdateUser(uint(id), services.UpdateUserInput{
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
		Role:     req.Role,
	}, requester)
	if svcErr != nil {
		respondUserServiceError(c, svcErr)
		return
	}

	var beforeSnapshot gin.H
	if before != nil {
		beforeSnapshot = toUserResponse(before)
	}
	services.Log(c, "user.update", "user", fmt.Sprintf("%d", id), beforeSnapshot, toUserResponse(user))
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(user)})
}

type setStatusRequest struct {
	IsActive bool `json:"isActive"`
}

func SetUserStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id", "code": "invalid_request"})
		return
	}
	var req setStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "invalid_request"})
		return
	}

	requester := requesterFromContext(c)
	if requester == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated", "code": "unauthenticated"})
		return
	}

	before, _ := services.GetUserByID(uint(id))
	user, svcErr := services.SetUserStatus(uint(id), req.IsActive, requester)
	if svcErr != nil {
		respondUserServiceError(c, svcErr)
		return
	}

	action := "user.enable"
	if !req.IsActive {
		action = "user.disable"
	}
	var beforeSnapshot gin.H
	if before != nil {
		beforeSnapshot = gin.H{"isActive": before.IsActive}
	}
	services.Log(c, action, "user", fmt.Sprintf("%d", id), beforeSnapshot, gin.H{"isActive": user.IsActive})
	c.JSON(http.StatusOK, gin.H{"user": toUserResponse(user)})
}

func DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id", "code": "invalid_request"})
		return
	}

	requester := requesterFromContext(c)
	if requester == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated", "code": "unauthenticated"})
		return
	}

	before, _ := services.GetUserByID(uint(id))
	if svcErr := services.DeleteUser(uint(id), requester); svcErr != nil {
		respondUserServiceError(c, svcErr)
		return
	}

	var beforeSnapshot gin.H
	if before != nil {
		beforeSnapshot = gin.H{"email": before.Email, "role": before.Role}
	}
	services.Log(c, "user.delete", "user", fmt.Sprintf("%d", id), beforeSnapshot, nil)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// TransferOwnership sets user :id as the new Owner (moving it off whoever
// currently holds it).
func TransferOwnership(c *gin.Context) {
	newOwnerID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id", "code": "invalid_request"})
		return
	}

	requester := requesterFromContext(c)
	if requester == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated", "code": "unauthenticated"})
		return
	}

	if svcErr := services.TransferOwnership(requester, uint(newOwnerID)); svcErr != nil {
		respondUserServiceError(c, svcErr)
		return
	}

	services.Log(c, "user.owner_transfer", "user", fmt.Sprintf("%d", newOwnerID), gin.H{"previousOwnerId": requester.ID}, gin.H{"newOwnerId": newOwnerID})
	c.JSON(http.StatusOK, gin.H{"success": true})
}

type bulkStatusRequest struct {
	IDs      []uint `json:"ids" binding:"required"`
	IsActive bool   `json:"isActive"`
}

func BulkSetStatus(c *gin.Context) {
	var req bulkStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "invalid_request"})
		return
	}
	requester := requesterFromContext(c)
	if requester == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated", "code": "unauthenticated"})
		return
	}

	succeeded, failed := services.BulkSetStatus(req.IDs, req.IsActive, requester)
	action := "user.enable"
	if !req.IsActive {
		action = "user.disable"
	}
	for _, id := range succeeded {
		services.Log(c, action, "user", fmt.Sprintf("%d", id), nil, gin.H{"isActive": req.IsActive})
	}
	c.JSON(http.StatusOK, gin.H{"succeeded": succeeded, "failed": failed})
}

type bulkIDsRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

func BulkDelete(c *gin.Context) {
	var req bulkIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "invalid_request"})
		return
	}
	requester := requesterFromContext(c)
	if requester == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated", "code": "unauthenticated"})
		return
	}

	succeeded, failed := services.BulkDeleteUsers(req.IDs, requester)
	for _, id := range succeeded {
		services.Log(c, "user.delete", "user", fmt.Sprintf("%d", id), nil, nil)
	}
	c.JSON(http.StatusOK, gin.H{"succeeded": succeeded, "failed": failed})
}
