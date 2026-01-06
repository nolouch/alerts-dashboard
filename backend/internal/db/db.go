package db

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var TiDB *sql.DB

func Init() error {
	var err error

	// Use local database in backend directory
	dbPath := "./alerts_v2.db"
	log.Printf("Connecting to database at: %s", dbPath)

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	log.Println("Database connection established")

	// Run migration to ensure schema is up to date
	log.Println("Running database migration...")
	if err := MigrateDatabase(DB); err != nil {
		log.Printf("Migration error: %v", err)
		return err
	}

	// Initialize TiDB connection for name service
	if err := InitTiDB(); err != nil {
		log.Printf("Warning: TiDB connection failed: %v (name service will be unavailable)", err)
	}

	return nil
}

func InitTiDB() error {
	dsn := os.Getenv("TIDB_DSN")
	if dsn == "" {
		return fmt.Errorf("TIDB_DSN environment variable not set")
	}

	// Register TLS configuration for TiDB Cloud
	err := mysqlDriver.RegisterTLSConfig("tidb", &tls.Config{
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return fmt.Errorf("failed to register TLS config: %w", err)
	}

	TiDB, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	// Set connection pool settings for TiDB
	TiDB.SetMaxOpenConns(20)
	TiDB.SetMaxIdleConns(10)
	TiDB.SetConnMaxLifetime(time.Minute * 5)

	// Test connection
	if err := TiDB.Ping(); err != nil {
		return fmt.Errorf("failed to connect to TiDB: %w", err)
	}

	log.Println("TiDB connection established for name service")
	return nil
}
