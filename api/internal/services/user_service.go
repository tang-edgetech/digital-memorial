package services

import (
	"sync"
	"time"

	"digital-memorial/api/internal/db"
	"digital-memorial/api/internal/models"
)

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
