package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"github.com/example/currency-converter-api/internal/config"
	"github.com/example/currency-converter-api/internal/db"
	"github.com/example/currency-converter-api/internal/models"
	"github.com/example/currency-converter-api/internal/rates"
	"github.com/example/currency-converter-api/internal/router"
	"github.com/gin-gonic/gin"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	dbConn, err := db.Connect(dbConfig(cfg))
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	if err := dbConn.AutoMigrate(&models.User{}, &models.Session{}, &models.RateSnapshot{}); err != nil {
		log.Fatalf("DB migration error: %v", err)
	}

	provider := rates.NewProvider(cfg.RatesProvider)
	svc := rates.NewService(dbConn, provider)

	// Initial rates refresh
	go func() {
		if err := svc.RefreshAll("USD"); err != nil {
			log.Printf("Initial rates refresh failed: %v", err)
		}
	}()

	interval := time.Duration(cfg.RatesRefreshIntervalHours) * time.Hour
	if interval <= 0 {
		interval = 6 * time.Hour
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			if err := svc.RefreshAll("USD"); err != nil {
				log.Printf("Scheduled rates refresh failed: %v", err)
			}
		}
	}()

	g := gin.New()
	g.Use(gin.Logger(), gin.Recovery())
	router.Register(g, dbConn, svc, cfg)

	addr := ":" + strconv.Itoa(cfg.AppPort)
	log.Printf("Currency Converter API listening on %s (env=%s)", addr, os.Getenv("APP_ENV"))
	if err := g.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func dbConfig(cfg config.Config) db.Config {
	return db.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Name:     cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}
}
