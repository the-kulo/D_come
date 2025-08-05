package application

import (
	"D_come/internal/infrastructure/crawler"
	"context"
	"fmt"
)

// CrawlerService 爬虫服务
type CrawlerService struct {
	converter *StockCodeConverter
	manager   crawler.CrawlerManager
}

// NewCrawlerService 创建爬虫服务
func NewCrawlerService(manager crawler.CrawlerManager) *CrawlerService {
	return &CrawlerService{
		converter: NewStockCodeConverter(),
		manager:   manager,
	}
}

// GetStockData 获取股票数据（智能选择数据源）
func (s *CrawlerService) GetStockData(ctx context.Context, code string) (*crawler.StockData, error) {
	// 解析股票代码
	stockCode, err := s.converter.ParseStockCode(code)
	if err != nil {
		return nil, fmt.Errorf("invalid stock code: %w", err)
	}

	// 根据README要求：港股用腾讯，A股用新浪
	source := s.selectSource(stockCode.Region)

	// 转换代码格式
	convertedCode, err := s.converter.ConvertForSource(code, source)
	if err != nil {
		return nil, fmt.Errorf("failed to convert code: %w", err)
	}

	// 获取数据
	return s.manager.GetStockDataFromSource(ctx, convertedCode, source)
}

// GetStockDataFromSource 从指定数据源获取股票数据
func (s *CrawlerService) GetStockDataFromSource(ctx context.Context, code, source string) (*crawler.StockData, error) {
	convertedCode, err := s.converter.ConvertForSource(code, source)
	if err != nil {
		return nil, fmt.Errorf("failed to convert code: %w", err)
	}

	return s.manager.GetStockDataFromSource(ctx, convertedCode, source)
}

// GetMultipleStockData 批量获取股票数据
func (s *CrawlerService) GetMultipleStockData(ctx context.Context, codes []string) ([]*crawler.StockData, error) {
	var convertedCodes []string

	for _, code := range codes {
		stockCode, err := s.converter.ParseStockCode(code)
		if err != nil {
			continue // 跳过无效代码
		}

		source := s.selectSource(stockCode.Region)
		convertedCode, err := s.converter.ConvertForSource(code, source)
		if err != nil {
			continue // 跳过转换失败的代码
		}

		convertedCodes = append(convertedCodes, convertedCode)
	}

	return s.manager.GetMultipleStockData(ctx, convertedCodes)
}

// selectSource 根据地区选择数据源
func (s *CrawlerService) selectSource(region string) string {
	switch region {
	case "hk":
		return "tencent" // 港股用腾讯
	case "sh", "sz":
		return "sina" // A股用新浪
	default:
		return "sina" // 默认用新浪
	}
}

// IsValidStockCode 验证股票代码
func (s *CrawlerService) IsValidStockCode(code string) bool {
	return s.converter.IsValidStockCode(code)
}

// GetStockDataRealTime 获取股票实时数据（不使用缓存）
func (s *CrawlerService) GetStockDataRealTime(ctx context.Context, code string) (*crawler.StockData, error) {
	// 解析股票代码
	stockCode, err := s.converter.ParseStockCode(code)
	if err != nil {
		return nil, fmt.Errorf("invalid stock code: %w", err)
	}

	// 根据README要求：港股用腾讯，A股用新浪
	source := s.selectSource(stockCode.Region)

	// 转换代码格式
	convertedCode, err := s.converter.ConvertForSource(code, source)
	if err != nil {
		return nil, fmt.Errorf("failed to convert code: %w", err)
	}

	// 获取实时数据（不使用缓存）
	return s.manager.GetStockDataFromSourceRealTime(ctx, convertedCode, source)
}

// GetStockDataFromSourceRealTime 从指定数据源获取股票实时数据（不使用缓存）
func (s *CrawlerService) GetStockDataFromSourceRealTime(ctx context.Context, code, source string) (*crawler.StockData, error) {
	convertedCode, err := s.converter.ConvertForSource(code, source)
	if err != nil {
		return nil, fmt.Errorf("failed to convert code: %w", err)
	}

	return s.manager.GetStockDataFromSourceRealTime(ctx, convertedCode, source)
}

// GetMultipleStockDataRealTime 批量获取股票实时数据（不使用缓存）
func (s *CrawlerService) GetMultipleStockDataRealTime(ctx context.Context, codes []string) ([]*crawler.StockData, error) {
	var convertedCodes []string

	for _, code := range codes {
		stockCode, err := s.converter.ParseStockCode(code)
		if err != nil {
			continue // 跳过无效代码
		}

		source := s.selectSource(stockCode.Region)
		convertedCode, err := s.converter.ConvertForSource(code, source)
		if err != nil {
			continue // 跳过转换失败的代码
		}

		convertedCodes = append(convertedCodes, convertedCode)
	}

	return s.manager.GetMultipleStockDataRealTime(ctx, convertedCodes)
}
