package database

import (
	"fmt"

	"github.com/leetsecure/qryptic-controller/internal/config"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/utils/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() error {
	log := logger.Default()
	var err error

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName, config.DBSslMode)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Errorf("Could not connect to the database: %v", err)
		return err
	}
	return nil
}

func AutomigrateDatabase() error {
	err := DB.AutoMigrate(&models.User{},
		&models.VpnGateway{},
		&models.Client{},
		&models.IPPool{},
		&models.AdminConfiguration{},
		&models.SSOConfig{},
		&models.Auth{},
	)
	if err != nil {
		return err
	}
	return nil
}
