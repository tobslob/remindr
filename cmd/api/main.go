package main

import (
	"fmt"
	"log"

	"github.com/tobslob/remindr/cmd/tokens"
	"github.com/tobslob/remindr/internal/db"
	"github.com/tobslob/remindr/internal/env"
	"github.com/tobslob/remindr/internal/store"
)

func main() {
	env.LoadEnv()

	cfg := config{
		addr: env.GetString("ADDR"),
		db: env.GetString("DB_ADDR"),
		dbMaxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS"),
		dbMaxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS"),
		dbMaxIdleTime:  env.GetString("DB_MAX_IDLE_TIME"),
		TokenSecretKey: env.GetString("TOKEN_SECRET_KEY"),
	}

	dbConn, err := db.New(cfg.db, cfg.dbMaxOpenConns, cfg.dbMaxIdleConns, cfg.dbMaxIdleTime)
	if err != nil {
		panic(fmt.Sprintf("Error connecting to database: %v", err))
	}
	defer dbConn.Close()

	log.Println("database connection pool established")


	tokenMaker, err := tokens.NewJWTMaker(cfg.TokenSecretKey)
	if err != nil {
		panic(fmt.Sprintf("cannot create token maker: %v", err))
	}

	app := &application{
		config: cfg,
		store: store.NewStorage(dbConn),
		tokenMaker: tokenMaker,
	}

	log.Fatal(app.run(app.mount()))
}
