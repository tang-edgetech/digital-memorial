package config

import "fmt"

// MySQLDSN builds a go-sql-driver/mysql DSN for the configured database.
func (c *Config) MySQLDSN() string {
	port := c.DBPort
	if port == "" {
		port = "3306"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, port, c.DBName)
}
