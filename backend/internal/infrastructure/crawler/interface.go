package crawler

import (
	"context"
	"time"
)

// StockData 股票数据结构
type StockData struct {
	Name        string  `json:"name"`         // 股票名称
	Code        string  `json:"code"`         // 股票代码  
	Price       float64 `json:"price"`        // 股票价格
	Change      float64 `json:"change"`       // 涨跌幅
	ChangeValue float64 `json:"change_value"` // 涨跌量
	Volume      int64   `json:"volume"`       // 总手
	Amount      float64 `json:"amount"`       // 总金额
	UpdateTime  time.Time `json:"update_time"` // 更新时间
}

// Crawler 爬虫接口
type Crawler interface {
	// GetStockData 获取单个股票数据
	GetStockData(ctx context.Context, code string) (*StockData, error)
	
	// GetStockDataRealTime 获取单个股票实时数据（不使用缓存）
	GetStockDataRealTime(ctx context.Context, code string) (*StockData, error)
	
	// GetMultipleStockData 批量获取股票数据
	GetMultipleStockData(ctx context.Context, codes []string) ([]*StockData, error)
	
	// GetMultipleStockDataRealTime 批量获取股票实时数据（不使用缓存）
	GetMultipleStockDataRealTime(ctx context.Context, codes []string) ([]*StockData, error)
	
	// GetSourceName 获取数据源名称
	GetSourceName() string
}

// CrawlerManager 爬虫管理器接口
type CrawlerManager interface {
	// GetStockData 智能获取股票数据（自动选择最佳爬虫）
	GetStockData(ctx context.Context, code string) (*StockData, error)
	
	// GetStockDataRealTime 智能获取股票实时数据（不使用缓存）
	GetStockDataRealTime(ctx context.Context, code string) (*StockData, error)
	
	// GetStockDataFromSource 从指定数据源获取股票数据
	GetStockDataFromSource(ctx context.Context, code, source string) (*StockData, error)
	
	// GetStockDataFromSourceRealTime 从指定数据源获取股票实时数据（不使用缓存）
	GetStockDataFromSourceRealTime(ctx context.Context, code, source string) (*StockData, error)
	
	// GetMultipleStockData 批量获取股票数据
	GetMultipleStockData(ctx context.Context, codes []string) ([]*StockData, error)
	
	// GetMultipleStockDataRealTime 批量获取股票实时数据（不使用缓存）
	GetMultipleStockDataRealTime(ctx context.Context, codes []string) ([]*StockData, error)
}