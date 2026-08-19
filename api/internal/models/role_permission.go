package models

import "time"

type RolePermission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Role      string    `gorm:"size:20" json:"role"`
	Module    string    `gorm:"size:50" json:"module"`
	Action    string    `gorm:"size:50" json:"action"`
	Allowed   bool      `json:"allowed"`
	UpdatedAt time.Time `json:"updatedAt"`
	UpdatedBy *uint     `json:"updatedBy"`
}

func (RolePermission) TableName() string { return "role_permissions" }
