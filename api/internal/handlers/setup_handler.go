package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"

	"digital-memorial/api/internal/config"
	"digital-memorial/api/internal/db"
	"digital-memorial/api/internal/models"
	"digital-memorial/api/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var validDBNamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

func SetupStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"setupCompleted": services.IsSetupCompleted()})
}

type dbConfigRequest struct {
	Host     string `json:"host" binding:"required"`
	Port     string `json:"port"`
	User     string `json:"user" binding:"required"`
	Password string `json:"password"`
	DBName   string `json:"dbName" binding:"required"`
}

// SetupDB validates the submitted DB credentials by connecting, runs the
// golang-migrate schema migrations, then persists the credentials to .env and
// hot-swaps the live GORM connection so no restart is needed.
func SetupDB(c *gin.Context) {
	if services.IsSetupCompleted() {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "setup already completed", "code": "setup_already_completed"})
		return
	}

	var req dbConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "invalid_request"})
		return
	}
	if req.Port == "" {
		req.Port = "3306"
	}
	if !validDBNamePattern.MatchString(req.DBName) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "database name may only contain letters, numbers, and underscores", "code": "invalid_db_name"})
		return
	}

	serverDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		req.User, req.Password, req.Host, req.Port)
	serverDB, err := sql.Open("mysql", serverDSN)
	if err != nil || serverDB.Ping() != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "could not connect to the database server with the given credentials", "code": "db_connect_failed"})
		return
	}
	defer serverDB.Close()

	// CREATE DATABASE can't use a placeholder for the identifier; req.DBName
	// is validated against validDBNamePattern above so this is safe to embed.
	createStmt := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", req.DBName)
	if _, err := serverDB.Exec(createStmt); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create database: " + err.Error(), "code": "db_create_failed"})
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		req.User, req.Password, req.Host, req.Port, req.DBName)

	// A separate multiStatements-enabled DSN just for running migrations —
	// go-sql-driver disables multi-statement execution by default (it's a
	// footgun for user-facing queries), but migration files legitimately
	// contain multiple statements (e.g. CREATE TABLE + seed INSERTs).
	migrationDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		req.User, req.Password, req.Host, req.Port, req.DBName)

	migrationDB, err := sql.Open("mysql", migrationDSN)
	if err != nil || migrationDB.Ping() != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "could not connect to the created database", "code": "db_connect_failed"})
		return
	}
	defer migrationDB.Close()

	if err := db.RunMigrations(migrationDB, config.MigrationsPath()); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to run migrations: " + err.Error(), "code": "migration_failed"})
		return
	}

	if _, err := config.Save(config.DBParams{
		Host: req.Host, Port: req.Port, User: req.User, Password: req.Password, Name: req.DBName,
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to persist configuration", "code": "config_save_failed"})
		return
	}

	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to open ORM connection", "code": "orm_connect_failed"})
		return
	}
	db.Set(gdb)
	services.RefreshSetupCompletedFromDB()
	services.RefreshPermissionsFromDB()

	c.JSON(http.StatusOK, gin.H{"success": true})
}

type createAdminRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"fullName" binding:"required"`
}

func SetupAdmin(c *gin.Context) {
	if services.IsSetupCompleted() {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "setup already completed", "code": "setup_already_completed"})
		return
	}
	gdb := db.Get()
	if gdb == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "database not configured yet", "code": "db_not_configured"})
		return
	}

	var req createAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "invalid_request"})
		return
	}

	var count int64
	gdb.Model(&models.User{}).Where("role = ?", models.RoleSuperAdmin).Count(&count)
	if count > 0 {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "a super admin already exists", "code": "super_admin_exists"})
		return
	}

	hash, err := services.HashPassword(req.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password", "code": "hash_failed"})
		return
	}

	user := models.User{
		Email:        req.Email,
		PasswordHash: hash,
		FullName:     req.FullName,
		Role:         models.RoleSuperAdmin,
		IsActive:     true,
	}
	if err := gdb.Create(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create super admin: " + err.Error(), "code": "create_failed"})
		return
	}

	c.Set("userID", user.ID)
	c.Set("userEmail", user.Email)
	services.Log(c, "setup.super_admin_created", "user", fmt.Sprintf("%d", user.ID), nil, gin.H{"email": user.Email})

	c.JSON(http.StatusOK, gin.H{"success": true})
}

type initialSettingsRequest struct {
	SiteTitle string `json:"siteTitle" binding:"required"`
	LogoPath  string `json:"logoPath"`
}

func SetupSettings(c *gin.Context) {
	if services.IsSetupCompleted() {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "setup already completed", "code": "setup_already_completed"})
		return
	}
	gdb := db.Get()
	if gdb == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "database not configured yet", "code": "db_not_configured"})
		return
	}

	var req initialSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "invalid_request"})
		return
	}

	updates := map[string]string{
		"site_title":      req.SiteTitle,
		"setup_completed": "true",
	}
	if req.LogoPath != "" {
		updates["logo_path"] = req.LogoPath
	}

	if _, _, err := services.UpdateSettings(updates, nil); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings", "code": "settings_save_failed"})
		return
	}

	services.RefreshSetupCompletedFromDB()
	services.Log(c, "setup.completed", "site_setting", "setup_completed", nil, updates)

	c.JSON(http.StatusOK, gin.H{"success": true})
}
