package db

import (
	"sync/atomic"

	"gorm.io/gorm"
)

var current atomic.Pointer[gorm.DB]

// Set stores the live GORM connection so it can be hot-swapped after
// /api/setup/db configures the database, without restarting the process.
func Set(gdb *gorm.DB) {
	current.Store(gdb)
}

// Get returns the current GORM connection, or nil if the app hasn't been
// through /setup yet.
func Get() *gorm.DB {
	return current.Load()
}
