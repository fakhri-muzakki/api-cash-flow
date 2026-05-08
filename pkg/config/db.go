package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDB() *sql.DB {
	dsn := buildDSN()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	// Serverless-friendly
	// Supaya gak terlalu banyak koneksi — setiap function instance punya pool sendiri
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute) // tutup koneksi lama agar tidak menumpuk
	db.SetConnMaxIdleTime(1 * time.Minute) // tutup koneksi idle lebih agresif

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	log.Println("database connected")
	return db
}

// buildDSN — prioritaskan DATABASE_URL kalau ada (format Neon/Supabase)
// fallback ke individual env vars untuk development lokal
func buildDSN() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)
}
