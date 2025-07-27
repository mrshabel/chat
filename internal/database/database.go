package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/mrshabel/chat/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

type DB struct {
	*sql.DB
}

var (
	dbInstance *DB
)

func New(cfg *config.Config) (*DB, error) {
	// reuse Connection
	if dbInstance != nil {
		return dbInstance, nil
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.DbUsername, cfg.DbPassword, cfg.DbHost, cfg.DbPort, cfg.Db)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}
	// ping the database
	if err := db.Ping(); err != nil {
		return nil, err
	}
	log.Println("database connected successfully")

	// run simple migration
	if err = migrate(db); err != nil {
		log.Println("failed to perform db migration", err)
	}

	// setup connection pool
	db.SetMaxOpenConns(30)
	dbInstance = &DB{
		db,
	}
	return dbInstance, nil
}
