package service

import (
	"context"
	"fmt"
	"time"

	"D_come/internal/domain/stock"
	"D_come/internal/infrastructure/crawler"
)

// CustomStockService 自定义股票服务
type CustomStockService struct {
	stockRepo      stock.Repository
	crawlerManager crawler.CrawlerManager
}

// NewCustomStockService 创建自定义股票服务
func NewCustomStockService(stockRepo stock.Repository, crawlerManager crawler.CrawlerManager) *CustomStockService {
	return &CustomStockService{
		stockRepo:      stockRepo,
		crawlerManager: crawlerManager,
	}
}

// CustomStockWithData 带有实时数据的自定义股票
type CustomStockWithData struct {
	ID            uint      `json:"id"`
	CustomName    string    `json:"customName"`
	CustomCode    string    `json:"customCode"`
	CrawlTime     time.Time `json:"crawlTime"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Price         float64   `json:"price"`
	ChangeRate    float64   `json:"changeRate"`
	ChangeValue   float64   `json:"changeValue"`
	Volume        int64     `json:"volume"`
	Amount        float64   `json:"amount"`
	FormulaResult float64   `json:"formulaResult,omitempty"`
}

// GetAllCustomStocks 获取所有自定义股票及其实时数据
func (s *CustomStockService) GetAllCustomStocks(ctx context.Context) ([]*CustomStockWithData, error) {
	customStocks, err := s.stockRepo.GetAllCustomStocks()
	if err != nil {
		return nil, fmt.Errorf("获取自定义股票失败: %w", err)
	}

	var result []*CustomStockWithData
	for _, customStock := range customStocks {
		stockWithData := &CustomStockWithData{
			ID:         customStock.ID,
			CustomName: customStock.CustomName,
			CustomCode: customStock.CustomCode,
			CrawlTime:  customStock.CrawlTime,
			UpdatedAt:  customStock.UpdatedAt,
		}

		// 获取实时股票数据
		if stockData, err := s.crawlerManager.GetStockDataRealTime(ctx, customStock.CustomCode); err == nil {
			stockWithData.Price = stockData.Price
			stockWithData.ChangeRate = stockData.Change
			stockWithData.ChangeValue = stockData.ChangeValue
			stockWithData.Volume = stockData.Volume
			stockWithData.Amount = stockData.Amount
		}

		result = append(result, stockWithData)
	}

	return result, nil
}

// GetCustomStockByID 根据ID获取自定义股票
func (s *CustomStockService) GetCustomStockByID(ctx context.Context, id uint) (*CustomStockWithData, error) {
	customStock, err := s.stockRepo.GetCustomStockByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取自定义股票失败: %w", err)
	}

	stockWithData := &CustomStockWithData{
		ID:         customStock.ID,
		CustomName: customStock.CustomName,
		CustomCode: customStock.CustomCode,
		CrawlTime:  customStock.CrawlTime,
		UpdatedAt:  customStock.UpdatedAt,
	}

	// 获取实时股票数据
	if stockData, err := s.crawlerManager.GetStockDataRealTime(ctx, customStock.CustomCode); err == nil {
		stockWithData.Price = stockData.Price
		stockWithData.ChangeRate = stockData.Change
		stockWithData.ChangeValue = stockData.ChangeValue
		stockWithData.Volume = stockData.Volume
		stockWithData.Amount = stockData.Amount
	}

	return stockWithData, nil
}

// CreateCustomStock 创建自定义股票
func (s *CustomStockService) CreateCustomStock(ctx context.Context, name, code string) (*stock.CustomStock, error) {
	// 验证股票代码格式
	if err := s.validateStockCode(ctx, code); err != nil {
		return nil, fmt.Errorf("股票代码验证失败: %w", err)
	}

	customStock := &stock.CustomStock{
		CustomName: name,
		CustomCode: code,
		CrawlTime:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.stockRepo.CreateCustomStock(customStock); err != nil {
		return nil, fmt.Errorf("创建自定义股票失败: %w", err)
	}

	return customStock, nil
}

// UpdateCustomStock 更新自定义股票
func (s *CustomStockService) UpdateCustomStock(ctx context.Context, id uint, name, code string) (*stock.CustomStock, error) {
	// 验证股票代码格式
	if err := s.validateStockCode(ctx, code); err != nil {
		return nil, fmt.Errorf("股票代码验证失败: %w", err)
	}

	customStock, err := s.stockRepo.GetCustomStockByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取自定义股票失败: %w", err)
	}

	customStock.CustomName = name
	customStock.CustomCode = code
	customStock.UpdatedAt = time.Now()

	if err := s.stockRepo.UpdateCustomStock(customStock); err != nil {
		return nil, fmt.Errorf("更新自定义股票失败: %w", err)
	}

	return customStock, nil
}

// DeleteCustomStock 删除自定义股票
func (s *CustomStockService) DeleteCustomStock(id uint) error {
	if err := s.stockRepo.DeleteCustomStock(id); err != nil {
		return fmt.Errorf("删除自定义股票失败: %w", err)
	}
	return nil
}

// validateStockCode 验证股票代码是否有效
func (s *CustomStockService) validateStockCode(ctx context.Context, code string) error {
	// 尝试获取股票数据来验证代码是否有效
	_, err := s.crawlerManager.GetStockDataRealTime(ctx, code)
	if err != nil {
		return fmt.Errorf("无效的股票代码: %s", code)
	}
	return nil
}
