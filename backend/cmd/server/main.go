package main

import (
	"D_come/internal/config"
	"D_come/internal/infrastructure/persistence"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	db, err := persistence.NewDatabase(&cfg.Database)
	if err != nil {
		panic(err)
	}
	defer db.Close()
}
