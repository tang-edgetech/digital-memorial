package main

import (
	"log"

	"digital-memorial/api/internal/config"
	"digital-memorial/api/internal/db"
	"digital-memorial/api/internal/router"
	"digital-memorial/api/internal/services"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	if cfg.IsDBConfigured() {
		gdb, err := gorm.Open(mysql.Open(cfg.MySQLDSN()), &gorm.Config{})
		if err != nil {
			log.Printf("warning: could not connect to configured database: %v", err)
		} else {
			db.Set(gdb)
			services.RefreshSetupCompletedFromDB()
			services.RefreshPermissionsFromDB()
		}
	}

	r := router.New()
	log.Printf("digital-memorial API listening on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}
