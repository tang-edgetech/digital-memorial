package services

import (
	"sync/atomic"

	"digital-memorial/api/internal/db"
	"digital-memorial/api/internal/models"

	"gorm.io/gorm"
)

type permissionMatrix map[string]map[string]map[string]bool // role -> module -> action -> allowed

var matrixCache atomic.Value

// RefreshPermissionsFromDB reloads the role_permissions table into the
// in-memory cache HasPermission checks on every request — mirrors
// settings_service.go's setupCompleted caching pattern. Call at boot and
// after every successful permission-matrix update; no re-login is needed for
// a change to take effect.
func RefreshPermissionsFromDB() {
	gdb := db.Get()
	matrix := permissionMatrix{}
	if gdb == nil {
		matrixCache.Store(matrix)
		return
	}
	var rows []models.RolePermission
	if err := gdb.Find(&rows).Error; err != nil {
		matrixCache.Store(matrix)
		return
	}
	for _, r := range rows {
		if matrix[r.Role] == nil {
			matrix[r.Role] = map[string]map[string]bool{}
		}
		if matrix[r.Role][r.Module] == nil {
			matrix[r.Role][r.Module] = map[string]bool{}
		}
		matrix[r.Role][r.Module][r.Action] = r.Allowed
	}
	matrixCache.Store(matrix)
}

func getMatrix() permissionMatrix {
	m, _ := matrixCache.Load().(permissionMatrix)
	if m == nil {
		return permissionMatrix{}
	}
	return m
}

// HasPermission reports whether role may perform action on module.
// super_admin is hardcoded full-access — it never has rows in the table, to
// avoid a self-referential escalation path through the matrix editor.
func HasPermission(role, module, action string) bool {
	if role == "super_admin" {
		return true
	}
	if roleMap, ok := getMatrix()[role]; ok {
		if moduleMap, ok := roleMap[module]; ok {
			return moduleMap[action]
		}
	}
	return false
}

// EffectivePermissions returns the full module->action->allowed grid for
// role, synthesizing all-true for super_admin.
func EffectivePermissions(role string) map[string]map[string]bool {
	result := make(map[string]map[string]bool, len(PermissionRegistry))
	for module, actions := range PermissionRegistry {
		result[module] = make(map[string]bool, len(actions))
		for _, action := range actions {
			result[module][action] = HasPermission(role, module, action)
		}
	}
	return result
}

func GetAllPermissions() ([]models.RolePermission, error) {
	gdb := db.Get()
	var rows []models.RolePermission
	if err := gdb.Order("role, module, action").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

type PermissionUpdate struct {
	Role    string `json:"role"`
	Module  string `json:"module"`
	Action  string `json:"action"`
	Allowed bool   `json:"allowed"`
}

// UpdatePermissions applies a batch of role/module/action/allowed updates in
// one transaction, returning before/after snapshots for the audit log.
// Unknown role/module/action combinations are skipped rather than failing
// the whole batch.
func UpdatePermissions(updates []PermissionUpdate, updatedBy *uint) (before, after []models.RolePermission, err error) {
	gdb := db.Get()
	err = gdb.Transaction(func(tx *gorm.DB) error {
		for _, u := range updates {
			var row models.RolePermission
			if e := tx.Where("role = ? AND module = ? AND action = ?", u.Role, u.Module, u.Action).First(&row).Error; e != nil {
				continue
			}
			before = append(before, row)
			if e := tx.Model(&models.RolePermission{}).Where("id = ?", row.ID).Updates(map[string]interface{}{
				"allowed":    u.Allowed,
				"updated_by": updatedBy,
			}).Error; e != nil {
				return e
			}
			row.Allowed = u.Allowed
			after = append(after, row)
		}
		return nil
	})
	return before, after, err
}
