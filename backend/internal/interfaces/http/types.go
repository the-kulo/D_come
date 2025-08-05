package http

// HAStockResponse H-A股票响应数据
type HAStockResponse struct {
	StockName    string     `json:"stockName"`
	HStockCode   string     `json:"hStockCode"`
	HStockData   *StockInfo `json:"hStockData"`
	HUpdateTime  string     `json:"hUpdateTime"`  // H股更新时间
	AStockCode   string     `json:"aStockCode"`
	AStockData   *StockInfo `json:"aStockData"`
	AUpdateTime  string     `json:"aUpdateTime"`  // A股更新时间
	CacheHit     bool       `json:"cacheHit"`     // 是否命中缓存
}

// StockInfo 股票信息
type StockInfo struct {
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Price       float64 `json:"price"`
	Change      float64 `json:"change"`
	ChangeValue float64 `json:"changeValue"`
	Volume      int64   `json:"volume"`
	Amount      float64 `json:"amount"`
}