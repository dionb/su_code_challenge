package main

import (
	"fmt"
	"log"
	"os"
	"sumup_challenge/api"
	"sumup_challenge/storage"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	db := connectDB()

	store, err := storage.NewPGAccountStore(db)

	server, err := api.NewServer(db, store)
	if err != nil {
		log.Fatal(err)
	}

	port := getEnv("PORT", "8080")
	if err := server.Router.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func connectDB() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "appuser"),
		getEnv("DB_PASSWORD", "changeme"),
		getEnv("DB_NAME", "appdb"),
		getEnv("DB_SSLMODE", "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("connecting to db: {}", err)
	}
	return db
}
