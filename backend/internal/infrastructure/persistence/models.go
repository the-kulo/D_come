package persistence

import "time"

type StockPair struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	StockName  string    `gorm:"column:stock_name" json:"stock_name"`
	AStockCode string    `gorm:"column:a_stock_code" json:"a_stock_code"`
	HStockCode string    `gorm:"column:h_stock_code" json:"h_stock_code"`
	UpdateTime time.Time `gorm:"column:update_at" json:"update_time"`
}

func (StockPair) TableName() string {
	return "stock_pairs"
}
