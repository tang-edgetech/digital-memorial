package services

// PermissionRegistry is the canonical list of module/action pairs the
// permission matrix covers. The role_permissions table is the source of
// truth for `allowed` values; this registry is what lets EffectivePermissions
// synthesize a full grid (including for super_admin, which has no rows) and
// what the matrix editor UI iterates over.
var PermissionRegistry = map[string][]string{
	"users":    {"view", "create", "edit", "delete", "enable_disable"},
	"settings": {"edit"},
}
