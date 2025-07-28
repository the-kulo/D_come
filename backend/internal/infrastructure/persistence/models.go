package persistence

import "time"

type StorePair struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	StockName  string    `gorm:"column:stock_name"`
	AStockCode string    `gorm:"column:a_stock_code"`
	HStockCode string    `gorm:"column:h_stock_code"`
	UpdateTime time.Time `gorm:"column:update_at"`
}
