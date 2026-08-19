package services

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"digital-memorial/api/internal/db"
	"digital-memorial/api/internal/models"

	"gorm.io/gorm"
)

var (
	ErrCannotCreateSuperAdmin    = errors.New("cannot create a super_admin account via this endpoint")
	ErrEmailTaken                = errors.New("email is already in use")
	ErrUserNotFound              = errors.New("user not found")
	ErrForbiddenSelfAction       = errors.New("you cannot perform this action on your own account")
	ErrForbiddenAdminTarget      = errors.New("only the Owner or a Super Admin can modify another Admin account")
	ErrOwnerMustTransferFirst    = errors.New("transfer ownership to another admin before changing this account's role or removing it")
	ErrTransferTargetInvalid     = errors.New("ownership can only be transferred to an active admin account")
	ErrTransferRequiresPrivilege = errors.New("only the current Owner or a Super Admin can transfer ownership")
)

// ownerMutex serializes the "count admins -> possibly mark owner" sequence in
// CreateUser/UpdateUser and the owner-swap in TransferOwnership, guarding the
// "at most one Owner" invariant against races. Same non-clustered,
// single-instance assumption the rate-limit middleware already makes.
var ownerMutex sync.Mutex

var lastActiveThrottle sync.Map // userID -> time.Time of last DB write

// TouchLastActive updates users.last_active_at, throttled to at most once per
// minute per user so idle-timeout tracing doesn't add a write to every
// authenticated request.
func TouchLastActive(userID uint) {
	now := time.Now()
	if last, ok := lastActiveThrottle.Load(userID); ok {
		if now.Sub(last.(time.Time)) < time.Minute {
			return
		}
	}
	lastActiveThrottle.Store(userID, now)

	gdb := db.Get()
	if gdb == nil {
		return
	}
	gdb.Model(&models.User{}).Where("id = ?", userID).Update("last_active_at", now)
}

func GetUserByEmail(email string) (*models.User, error) {
	gdb := db.Get()
	var user models.User
	if err := gdb.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(id uint) (*models.User, error) {
	gdb := db.Get()
	var user models.User
	if err := gdb.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func RecordFailedLogin(user *models.User, lockoutThreshold int, lockoutDuration time.Duration) {
	gdb := db.Get()
	user.FailedLoginCount++
	updates := map[string]interface{}{"failed_login_count": user.FailedLoginCount}
	if int(user.FailedLoginCount) >= lockoutThreshold {
		updates["locked_until"] = time.Now().Add(lockoutDuration)
	}
	gdb.Model(&models.User{}).Where("id = ?", user.ID).Updates(updates)
}

func RecordSuccessfulLogin(user *models.User) {
	gdb := db.Get()
	now := time.Now()
	gdb.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"failed_login_count": 0,
		"locked_until":       nil,
		"last_login_at":      now,
		"last_active_at":     now,
	})
}

func IsLocked(user *models.User) bool {
	return user.LockedUntil != nil && user.LockedUntil.After(time.Now())
}

// UserFilter drives ListUsers. SortBy is whitelisted against a fixed set of
// columns (never interpolated raw) to avoid SQL injection via query params.
type UserFilter struct {
	Role          string
	Status        string // "active" | "disabled" | ""
	Search        string
	SortBy        string
	SortDir       string
	Page          int
	PageSize      int
	RequesterRole string
}

func ListUsers(f UserFilter) ([]models.User, int64, error) {
	gdb := db.Get()
	query := gdb.Model(&models.User{})

	if f.RequesterRole != string(models.RoleSuperAdmin) {
		query = query.Where("role <> ?", models.RoleSuperAdmin)
	}
	if f.Role != "" {
		query = query.Where("role = ?", f.Role)
	}
	if f.Status == "active" {
		query = query.Where("is_active = ?", true)
	} else if f.Status == "disabled" {
		query = query.Where("is_active = ?", false)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		query = query.Where("full_name LIKE ? OR email LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortColumn := "created_at"
	switch f.SortBy {
	case "fullName":
		sortColumn = "full_name"
	case "email":
		sortColumn = "email"
	case "role":
		sortColumn = "role"
	case "lastActiveAt":
		sortColumn = "last_active_at"
	}
	sortDir := "ASC"
	if strings.EqualFold(f.SortDir, "desc") {
		sortDir = "DESC"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortColumn, sortDir))

	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	query = query.Offset((page - 1) * pageSize).Limit(pageSize)

	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// GetUserForViewer fetches a single user for display, applying the
// super_admin-row-hiding rule (returns ErrUserNotFound rather than leaking
// that the row exists via a 403).
func GetUserForViewer(id uint, requesterRole models.Role) (*models.User, error) {
	user, err := GetUserByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	if err := guardView(requesterRole, user); err != nil {
		return nil, err
	}
	return user, nil
}

// guardView applies the super_admin-row-hiding rule shared by GetUser and
// (indirectly, via query filtering) ListUsers. Returns ErrUserNotFound
// (surfaced as 404, not 403) so a super_admin row's existence isn't leaked to
// a non-super-admin caller.
func guardView(requesterRole models.Role, target *models.User) error {
	if target.Role == models.RoleSuperAdmin && requesterRole != models.RoleSuperAdmin {
		return ErrUserNotFound
	}
	return nil
}

// guardMutation applies the rules shared by UpdateUser/SetUserStatus/DeleteUser:
// a user may never mutate their own account through this admin-management
// path, a super_admin row is invisible (404) to non-super-admins, and only a
// Super Admin or the current Owner may mutate another Admin's account.
func guardMutation(requester, target *models.User) error {
	if requester.ID == target.ID {
		return ErrForbiddenSelfAction
	}
	if err := guardView(requester.Role, target); err != nil {
		return err
	}
	if target.Role == models.RoleAdmin && requester.Role != models.RoleSuperAdmin && !requester.IsOwner {
		return ErrForbiddenAdminTarget
	}
	return nil
}

type CreateUserInput struct {
	Email    string
	Password string
	FullName string
	Role     string // "admin" | "agent" — never "super_admin"
	IsActive bool
}

// CreateUser rejects super_admin at the service layer too (defense in depth
// beyond the frontend role dropdown only offering admin/agent), and marks the
// very first admin-role account ever created as Owner.
func CreateUser(input CreateUserInput) (*models.User, error) {
	if input.Role == string(models.RoleSuperAdmin) {
		return nil, ErrCannotCreateSuperAdmin
	}
	gdb := db.Get()

	hash, err := HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Email:        input.Email,
		PasswordHash: hash,
		FullName:     input.FullName,
		Role:         models.Role(input.Role),
		IsActive:     input.IsActive,
	}

	if input.Role == string(models.RoleAdmin) {
		ownerMutex.Lock()
		var adminCount int64
		gdb.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&adminCount)
		if adminCount == 0 {
			user.IsOwner = true
		}
		err := gdb.Create(&user).Error
		ownerMutex.Unlock()
		if err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				return nil, ErrEmailTaken
			}
			return nil, err
		}
		return &user, nil
	}

	if err := gdb.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return &user, nil
}

type UpdateUserInput struct {
	Email    *string
	FullName *string
	Password *string // nil/empty = unchanged
	Role     *string // "admin" | "agent" — never "super_admin"
}

func UpdateUser(targetID uint, input UpdateUserInput, requester *models.User) (*models.User, error) {
	gdb := db.Get()
	target, err := GetUserByID(targetID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	if err := guardMutation(requester, target); err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if input.Email != nil && *input.Email != "" {
		updates["email"] = *input.Email
	}
	if input.FullName != nil && *input.FullName != "" {
		updates["full_name"] = *input.FullName
	}
	if input.Password != nil && *input.Password != "" {
		hash, err := HashPassword(*input.Password)
		if err != nil {
			return nil, err
		}
		updates["password_hash"] = hash
	}

	if input.Role != nil && *input.Role != "" && *input.Role != string(target.Role) {
		newRole := *input.Role
		if newRole == string(models.RoleSuperAdmin) {
			return nil, ErrCannotCreateSuperAdmin
		}
		if target.IsOwner && newRole != string(models.RoleAdmin) {
			return nil, ErrOwnerMustTransferFirst
		}

		updates["role"] = newRole
		if newRole == string(models.RoleAdmin) {
			ownerMutex.Lock()
			var adminCount int64
			gdb.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&adminCount)
			if adminCount == 0 {
				updates["is_owner"] = true
			}
			ownerMutex.Unlock()
		}
	}

	if len(updates) == 0 {
		return target, nil
	}

	if err := gdb.Model(&models.User{}).Where("id = ?", targetID).Updates(updates).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return GetUserByID(targetID)
}

// SetUserStatus toggles is_active. Re-enabling also clears any stale lockout
// so it doesn't silently persist through what the UI presents as "enable".
func SetUserStatus(targetID uint, active bool, requester *models.User) (*models.User, error) {
	gdb := db.Get()
	target, err := GetUserByID(targetID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	if err := guardMutation(requester, target); err != nil {
		return nil, err
	}

	updates := map[string]interface{}{"is_active": active}
	if active {
		updates["failed_login_count"] = 0
		updates["locked_until"] = nil
	}
	if err := gdb.Model(&models.User{}).Where("id = ?", targetID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return GetUserByID(targetID)
}

// DeleteUser blocks deleting the current Owner outright — ownership must be
// transferred to another admin first, so the system is never left with zero
// Owners.
func DeleteUser(targetID uint, requester *models.User) error {
	target, err := GetUserByID(targetID)
	if err != nil {
		return ErrUserNotFound
	}
	if err := guardMutation(requester, target); err != nil {
		return err
	}
	if target.IsOwner {
		return ErrOwnerMustTransferFirst
	}
	return db.Get().Delete(&models.User{}, targetID).Error
}

// TransferOwnership moves the Owner flag to another active admin. Self-
// transfer by the current Owner is a no-op success rather than an error.
func TransferOwnership(requester *models.User, newOwnerID uint) error {
	if requester.Role != models.RoleSuperAdmin && !requester.IsOwner {
		return ErrTransferRequiresPrivilege
	}
	if requester.IsOwner && requester.ID == newOwnerID {
		return nil
	}

	newOwner, err := GetUserByID(newOwnerID)
	if err != nil || newOwner.Role != models.RoleAdmin || !newOwner.IsActive {
		return ErrTransferTargetInvalid
	}

	gdb := db.Get()
	ownerMutex.Lock()
	defer ownerMutex.Unlock()
	return gdb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("is_owner = ?", true).Update("is_owner", false).Error; err != nil {
			return err
		}
		return tx.Model(&models.User{}).Where("id = ?", newOwnerID).Update("is_owner", true).Error
	})
}

type BulkFailure struct {
	ID     uint   `json:"id"`
	Reason string `json:"reason"`
}

// BulkSetStatus and BulkDeleteUsers process each id independently (one
// call/transaction per row, not one for the whole batch) so a single bad
// target (self, owner, admin-without-privilege, super_admin) doesn't fail
// the entire batch.
func BulkSetStatus(ids []uint, active bool, requester *models.User) (succeeded []uint, failed []BulkFailure) {
	for _, id := range ids {
		if _, err := SetUserStatus(id, active, requester); err != nil {
			failed = append(failed, BulkFailure{ID: id, Reason: err.Error()})
			continue
		}
		succeeded = append(succeeded, id)
	}
	return succeeded, failed
}

func BulkDeleteUsers(ids []uint, requester *models.User) (succeeded []uint, failed []BulkFailure) {
	for _, id := range ids {
		if err := DeleteUser(id, requester); err != nil {
			failed = append(failed, BulkFailure{ID: id, Reason: err.Error()})
			continue
		}
		succeeded = append(succeeded, id)
	}
	return succeeded, failed
}
