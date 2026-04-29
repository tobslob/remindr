package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tobslob/remindr/cmd/tokens"
	"github.com/tobslob/remindr/internal/db"
	"github.com/tobslob/remindr/internal/env"
	"github.com/tobslob/remindr/internal/reminder"
	"github.com/tobslob/remindr/internal/store"
)

func main() {
	env.LoadEnv()

	cfg := config{
		addr:           env.GetString("ADDR"),
		db:             env.GetString("DB_ADDR"),
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

	storage := store.NewStorage(dbConn)
	app := &application{
		config:     cfg,
		store:      storage,
		tokenMaker: tokenMaker,
	}

	reminderService := reminder.NewService(storage.Reminders, reminder.NewLogSender(log.Default()), reminder.ServiceConfig{})
	reminderService.Start(context.Background())

	if err := app.run(app.mount()); err != nil {
		reminderService.Stop()
		log.Fatal(err)
	}
}
