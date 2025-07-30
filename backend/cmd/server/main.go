package main

import (
	"D_come/internal/application"
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
	}

	for _, stock := range stockPairs {
		input := application.CrawlerInput{
			StockName:     stock.StockName,
			OriginalACode: stock.AStockCode,
			OriginalHCode: stock.HStockCode,
		}

		input.Normalize()

		fmt.Printf("股票名称: %s, 转换前 A 股: %s, 转换后 A 股: %s, 转换前 H 股: %s, 转换后 H 股: %s\n",
			input.StockName, stock.AStockCode, input.OriginalACode, stock.HStockCode, input.OriginalHCode)
	}
}
