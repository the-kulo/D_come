package persistence

import "time"

// StockPair H-A股票对模型 - 匹配现有数据库表结构
type StockPair struct {
	ID         uint      `gorm:"primarykey;column:id;type:bigint" json:"id"`
	StockName  string    `gorm:"column:stock_name;type:varchar(50);not null" json:"stock_name"`
	AStockCode string    `gorm:"column:a_stock_code;type:varchar(50);not null" json:"a_stock_code"`
	HStockCode string    `gorm:"column:h_stock_code;type:varchar(50);not null" json:"h_stock_code"`
	CrawlTime  time.Time `gorm:"column:crawl_time;type:datetime(3)" json:"crawl_time"`
	UpdatedAt  time.Time `gorm:"column:updated_at;type:datetime(3)" json:"updated_at"`
}

func (StockPair) TableName() string {
	return "stock_pairs"
}

// CustomStock 自定义股票模型
type CustomStock struct {
	ID         uint      `gorm:"primarykey;column:id;type:bigint" json:"id"`
	CustomName string    `gorm:"column:custom_name;type:varchar(50);not null" json:"customName"`
	CustomCode string    `gorm:"column:custom_code;type:varchar(50);not null" json:"customCode"`
	CrawlTime  time.Time `gorm:"column:crawl_time;type:datetime(3)" json:"crawlTime"`
	UpdatedAt  time.Time `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
}

func (CustomStock) TableName() string {
	return "custom"
}

// LatexFormula Latex公式模型
type LatexFormula struct {
	ID           uint      `gorm:"primarykey;column:id;type:bigint" json:"id"`
	LatexName    string    `gorm:"column:latex_name;type:varchar(50);not null" json:"latex_name"`
	LatexFormula string    `gorm:"column:latex_formula;type:text" json:"latex_formula"`
	CrawlTime    time.Time `gorm:"column:crawl_time;type:datetime(3)" json:"crawl_time"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:datetime(3)" json:"updated_at"`
}

func (LatexFormula) TableName() string {
	return "latex"
}

// StockData Redis缓存的股票数据结构
type StockData struct {
	StockName   string    `json:"stock_name"`
	StockCode   string    `json:"stock_code"`
	StockPrice  float64   `json:"stock_price"`
	ChangeRate  float64   `json:"change_rate"`
	ChangeValue float64   `json:"change_value"`
	Volume      int64     `json:"volume"`
	Amount      float64   `json:"amount"`
	UpdateTime  time.Time `json:"update_time"`
}
