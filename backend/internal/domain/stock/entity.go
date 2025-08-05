package stock

import "time"

// StockPair 港股A股对实体
type StockPair struct {
	ID         uint
	StockName  string
	AStockCode string
	HStockCode string
	CrawlTime  time.Time
	UpdatedAt  time.Time
}

// CustomStock 自定义股票实体
type CustomStock struct {
	ID         uint
	CustomName string
	CustomCode string
	CrawlTime  time.Time
	UpdatedAt  time.Time
}

// StockData 股票数据实体
type StockData struct {
	StockName   string
	StockCode   string
	StockPrice  float64
	ChangeRate  float64
	ChangeValue float64
	Volume      int64
	Amount      float64
	UpdateTime  time.Time
}

// Repository 股票仓储接口
type Repository interface {
	GetAllStockPairs() ([]*StockPair, error)
	GetStockPairByName(name string) (*StockPair, error)
	
	GetAllCustomStocks() ([]*CustomStock, error)
	GetCustomStockByID(id uint) (*CustomStock, error)
	CreateCustomStock(stock *CustomStock) error
	UpdateCustomStock(stock *CustomStock) error
	DeleteCustomStock(id uint) error
}