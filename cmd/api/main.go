package main

import (
	"log"

	"github.com/tobslob/todoApp/internal/env"
)

func main() {
	env.LoadEnv()

	cfg := config{
		addr: env.GetString("ADDR"),
	}

	app := &application{
		config: cfg,
	}

	log.Fatal(app.run(app.mount()))
}
