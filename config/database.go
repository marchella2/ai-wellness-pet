package config

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"pet-wellness-backend/models"
)

func ConnectDatabase(dsn string) *gorm.DB {
	// PreferSimpleProtocol disables prepared statements (pgx simple query
	// protocol). This avoids "prepared statement ... already exists"
	// (SQLSTATE 42P05) that breaks AutoMigrate's HasTable check, especially
	// when connecting through a transaction-mode pooler (e.g. Supabase).
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("[config] failed to connect database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("[config] failed to get underlying sql.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := db.AutoMigrate(
		&models.User{},
		&models.Pet{},
		&models.DailyLog{},
		&models.ChatMessage{},
	); err != nil {
		log.Fatalf("[config] failed to run auto migration: %v", err)
	}

	log.Println("[config] database connected and migrated successfully")
	return db
}
