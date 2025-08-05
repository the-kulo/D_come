package application

import (
	"D_come/internal/domain/stock"
	"D_come/internal/infrastructure/crawler"
	"context"
	"time"
)

type StockService struct {
	stockRepo      stock.Repository
	crawlerService *CrawlerService
	haStockService *HAStockService
}

func NewStockService(stockRepo stock.Repository, crawlerService *CrawlerService) *StockService {
	service := &StockService{
		stockRepo:      stockRepo,
		crawlerService: crawlerService,
	}

	// 创建H-A股票服务
	service.haStockService = NewHAStockService(stockRepo, crawlerService, nil)

	return service
}

// ===== H-A股票相关方法 =====

// GetAllStockPairs 获取所有H-A股票对
func (s *StockService) GetAllStockPairs(ctx context.Context) ([]*stock.StockPair, error) {
	return s.stockRepo.GetAllStockPairs()
}

// GetStockPairByName 根据名称获取股票对
func (s *StockService) GetStockPairByName(ctx context.Context, name string) (*stock.StockPair, error) {
	return s.stockRepo.GetStockPairByName(name)
}

// GetAllHAStockData 获取所有H-A股票的实时数据
func (s *StockService) GetAllHAStockData(ctx context.Context) ([]*HAStockData, error) {
	return s.haStockService.GetAllHAStockData(ctx)
}

// GetHAStockDataByName 根据股票名称获取H-A股票实时数据
func (s *StockService) GetHAStockDataByName(ctx context.Context, stockName string) (*HAStockData, error) {
	return s.haStockService.GetHAStockDataByName(ctx, stockName)
}

// RefreshAllHAStockData 刷新所有H-A股票数据
func (s *StockService) RefreshAllHAStockData(ctx context.Context) error {
	return s.haStockService.RefreshAllHAStockData(ctx)
}

// ===== 自定义股票相关方法 =====

// GetAllCustomStocks 获取所有自定义股票
func (s *StockService) GetAllCustomStocks(ctx context.Context) ([]*stock.CustomStock, error) {
	return s.stockRepo.GetAllCustomStocks()
}

// CreateCustomStock 创建自定义股票
func (s *StockService) CreateCustomStock(ctx context.Context, customStock *stock.CustomStock) error {
	customStock.CrawlTime = time.Now()
	customStock.UpdatedAt = time.Now()
	return s.stockRepo.CreateCustomStock(customStock)
}

// UpdateCustomStock 更新自定义股票
func (s *StockService) UpdateCustomStock(ctx context.Context, customStock *stock.CustomStock) error {
	customStock.UpdatedAt = time.Now()
	return s.stockRepo.UpdateCustomStock(customStock)
}

// DeleteCustomStock 删除自定义股票
func (s *StockService) DeleteCustomStock(ctx context.Context, id uint) error {
	return s.stockRepo.DeleteCustomStock(id)
}

// GetCustomStockData 获取自定义股票的实时数据
func (s *StockService) GetCustomStockData(ctx context.Context, code string) (*crawler.StockData, error) {
	return s.crawlerService.GetStockData(ctx, code)
}

// ===== 工具方法 =====

// ConvertToGRPCStockData 转换为gRPC数据格式
func (s *StockService) ConvertToGRPCStockData(stockPair *stock.StockPair) *crawler.StockData {
	// 这里需要从Redis获取实时股票数据
	// 暂时返回基础数据
	return &crawler.StockData{
		Name: stockPair.StockName,
		Code: stockPair.AStockCode,
		// 其他字段需要从爬虫获取
	}
}
