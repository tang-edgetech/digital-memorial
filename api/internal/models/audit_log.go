package models

import "time"

type AuditLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ActorUserID *uint     `gorm:"column:actor_user_id;index" json:"actorUserId"`
	ActorEmail  string    `gorm:"column:actor_email;size:255" json:"actorEmail"`
	Action      string    `gorm:"size:100;index" json:"action"`
	TargetType  string    `gorm:"column:target_type;size:100" json:"targetType"`
	TargetID    string    `gorm:"column:target_id;size:100" json:"targetId"`
	BeforeValue *string   `gorm:"column:before_value;type:json" json:"beforeValue,omitempty"`
	AfterValue  *string   `gorm:"column:after_value;type:json" json:"afterValue,omitempty"`
	IPAddress   string    `gorm:"column:ip_address;size:64" json:"ipAddress"`
	UserAgent   string    `gorm:"column:user_agent;size:255" json:"userAgent"`
	CreatedAt   time.Time `gorm:"index" json:"createdAt"`
}

func (AuditLog) TableName() string { return "audit_logs" }
