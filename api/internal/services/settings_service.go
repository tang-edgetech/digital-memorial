package services

import (
	"strconv"
	"sync/atomic"

	"digital-memorial/api/internal/db"
	"digital-memorial/api/internal/models"

	"gorm.io/gorm"
)

var setupCompleted atomic.Bool

func IsSetupCompleted() bool {
	return setupCompleted.Load()
}

// RefreshSetupCompletedFromDB re-reads the setup_completed flag from the
// database into the in-memory cache the setup-gate middleware checks on every
// request, so that check never needs its own DB round trip.
func RefreshSetupCompletedFromDB() {
	gdb := db.Get()
	if gdb == nil {
		setupCompleted.Store(false)
		return
	}
	var setting models.SiteSetting
	if err := gdb.First(&setting, "`key` = ?", "setup_completed").Error; err != nil {
		setupCompleted.Store(false)
		return
	}
	setupCompleted.Store(setting.Value == "true")
}

func GetAllSettings() (map[string]string, error) {
	gdb := db.Get()
	var rows []models.SiteSetting
	if err := gdb.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]string, len(rows))
	for _, r := range rows {
		result[r.Key] = r.Value
	}
	return result, nil
}

// GetSettingInt reads a single int-valued setting, falling back to def if the
// DB isn't connected yet, the key is missing, or the value doesn't parse.
func GetSettingInt(key string, def int) int {
	gdb := db.Get()
	if gdb == nil {
		return def
	}
	var row models.SiteSetting
	if err := gdb.First(&row, "`key` = ?", key).Error; err != nil {
		return def
	}
	v, err := strconv.Atoi(row.Value)
	if err != nil {
		return def
	}
	return v
}

// UpdateSettings applies a partial set of key/value updates in one
// transaction and returns the before/after snapshots of the changed keys so
// callers can pass them to the audit log. Unknown keys are skipped rather
// than failing the whole batch.
func UpdateSettings(updates map[string]string, updatedBy *uint) (before, after map[string]string, err error) {
	gdb := db.Get()
	before = make(map[string]string)
	after = make(map[string]string)

	err = gdb.Transaction(func(tx *gorm.DB) error {
		for key, value := range updates {
			var row models.SiteSetting
			if e := tx.First(&row, "`key` = ?", key).Error; e != nil {
				continue
			}
			before[key] = row.Value
			if e := tx.Model(&models.SiteSetting{}).Where("`key` = ?", key).Updates(map[string]interface{}{
				"value":      value,
				"updated_by": updatedBy,
			}).Error; e != nil {
				return e
			}
			after[key] = value
		}
		return nil
	})
	return before, after, err
}
