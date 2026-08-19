package models

import "time"

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleAgent      Role = "agent"
)

type User struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	Email            string     `gorm:"uniqueIndex;size:255" json:"email"`
	PasswordHash     string     `gorm:"size:255" json:"-"`
	FullName         string     `gorm:"size:255" json:"fullName"`
	Role             Role       `gorm:"type:enum('super_admin','admin','agent');default:agent" json:"role"`
	IsOwner          bool       `gorm:"column:is_owner;default:false" json:"isOwner"`
	IsActive         bool       `gorm:"default:true" json:"isActive"`
	ThemePreference  string     `gorm:"column:theme_preference;type:enum('light','dark');default:light" json:"themePreference"`
	FailedLoginCount uint       `gorm:"column:failed_login_count;default:0" json:"-"`
	LockedUntil      *time.Time `gorm:"column:locked_until" json:"-"`
	LastLoginAt      *time.Time `gorm:"column:last_login_at" json:"lastLoginAt"`
	LastActiveAt     *time.Time `gorm:"column:last_active_at" json:"lastActiveAt"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

func (User) TableName() string { return "users" }
