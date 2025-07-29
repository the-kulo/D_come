package main

import (
	"D_come/internal/config"
	"D_come/internal/infrastructure/persistence"
	"fmt"
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

	stockRepo := persistence.NewStockRepository(db.DB)

	stockPairs, err := stockRepo.GetAll()
	if err != nil {
		panic(err)
	} else {
		for _, stock := range stockPairs {
			fmt.Println(stock.StockName, stock.AStockCode, stock.HStockCode)
		}
	}
}
