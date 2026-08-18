package models

import "time"

type SiteSetting struct {
	Key       string    `gorm:"primaryKey;column:key;size:100" json:"key"`
	Value     string    `gorm:"column:value;type:text" json:"value"`
	ValueType string    `gorm:"column:value_type;type:enum('string','int','bool','json');default:string" json:"valueType"`
	UpdatedAt time.Time `json:"updatedAt"`
	UpdatedBy *uint     `json:"updatedBy"`
}

func (SiteSetting) TableName() string { return "site_settings" }
