package main

import (
	"fmt"
	"log"

	"github.com/tobslob/todoApp/internal/db"
	"github.com/tobslob/todoApp/internal/env"
	"github.com/tobslob/todoApp/internal/store"
)

func main() {
	env.LoadEnv()

	cfg := config{
		addr: env.GetString("ADDR"),
		db: env.GetString("DB_ADDR"),
		dbMaxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS"),
		dbMaxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS"),
		dbMaxIdleTime:  env.GetString("DB_MAX_IDLE_TIME"),
	}

	dbConn, err := db.New(cfg.db, cfg.dbMaxOpenConns, cfg.dbMaxIdleConns, cfg.dbMaxIdleTime)
	if err != nil {
		panic(fmt.Sprintf("Error connecting to database: %v", err))
	}
	defer dbConn.Close()

	log.Println("database connection pool established")

	app := &application{
		config: cfg,
		store: store.NewStorage(dbConn),
	}

	log.Fatal(app.run(app.mount()))
}
