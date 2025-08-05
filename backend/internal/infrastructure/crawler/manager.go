package crawler

import (
	"context"
	"fmt"
)

// CrawlerManagerImpl 爬虫管理器实现
type CrawlerManagerImpl struct {
	crawlers map[string]Crawler
}

// NewCrawlerManager 创建爬虫管理器
func NewCrawlerManager() *CrawlerManagerImpl {
	manager := &CrawlerManagerImpl{
		crawlers: make(map[string]Crawler),
	}
	
	// 注册默认爬虫
	manager.RegisterCrawler("sina", NewSinaCrawler())
	manager.RegisterCrawler("tencent", NewTencentCrawler())
	
	return manager
}

// RegisterCrawler 注册爬虫
func (m *CrawlerManagerImpl) RegisterCrawler(name string, crawler Crawler) {
	m.crawlers[name] = crawler
}

// GetStockData 智能获取股票数据（自动选择最佳爬虫）
func (m *CrawlerManagerImpl) GetStockData(ctx context.Context, code string) (*StockData, error) {
	// 根据代码格式选择爬虫
	if len(code) >= 2 {
		prefix := code[:2]
		switch prefix {
		case "hk":
			return m.GetStockDataFromSource(ctx, code, "tencent")
		case "sh", "sz":
			return m.GetStockDataFromSource(ctx, code, "sina")
		}
	}
	
	// 默认使用新浪
	return m.GetStockDataFromSource(ctx, code, "sina")
}

// GetStockDataFromSource 从指定数据源获取股票数据
func (m *CrawlerManagerImpl) GetStockDataFromSource(ctx context.Context, code, source string) (*StockData, error) {
	crawler, exists := m.crawlers[source]
	if !exists {
		return nil, fmt.Errorf("爬虫不存在: %s", source)
	}
	
	return crawler.GetStockData(ctx, code)
}

// GetMultipleStockData 批量获取股票数据
func (m *CrawlerManagerImpl) GetMultipleStockData(ctx context.Context, codes []string) ([]*StockData, error) {
	var results []*StockData
	
	// 按数据源分组
	sinaGroup := make([]string, 0)
	tencentGroup := make([]string, 0)
	
	for _, code := range codes {
		if len(code) >= 2 {
			prefix := code[:2]
			switch prefix {
			case "hk":
				tencentGroup = append(tencentGroup, code)
			case "sh", "sz":
				sinaGroup = append(sinaGroup, code)
			default:
				sinaGroup = append(sinaGroup, code)
			}
		}
	}
	
	// 批量获取新浪数据
	if len(sinaGroup) > 0 {
		if sinaCrawler, exists := m.crawlers["sina"]; exists {
			sinaResults, err := sinaCrawler.GetMultipleStockData(ctx, sinaGroup)
			if err == nil {
				results = append(results, sinaResults...)
			}
		}
	}
	
	// 批量获取腾讯数据
	if len(tencentGroup) > 0 {
		if tencentCrawler, exists := m.crawlers["tencent"]; exists {
			tencentResults, err := tencentCrawler.GetMultipleStockData(ctx, tencentGroup)
			if err == nil {
				results = append(results, tencentResults...)
			}
		}
	}
	
	return results, nil
}

// GetStockDataRealTime 智能获取股票实时数据（不使用缓存）
func (m *CrawlerManagerImpl) GetStockDataRealTime(ctx context.Context, code string) (*StockData, error) {
	// 根据代码格式选择爬虫
	if len(code) >= 2 {
		prefix := code[:2]
		switch prefix {
		case "hk":
			return m.GetStockDataFromSourceRealTime(ctx, code, "tencent")
		case "sh", "sz":
			return m.GetStockDataFromSourceRealTime(ctx, code, "sina")
		}
	}
	
	// 默认使用新浪
	return m.GetStockDataFromSourceRealTime(ctx, code, "sina")
}

// GetStockDataFromSourceRealTime 从指定数据源获取股票实时数据（不使用缓存）
func (m *CrawlerManagerImpl) GetStockDataFromSourceRealTime(ctx context.Context, code, source string) (*StockData, error) {
	crawler, exists := m.crawlers[source]
	if !exists {
		return nil, fmt.Errorf("爬虫不存在: %s", source)
	}
	
	return crawler.GetStockDataRealTime(ctx, code)
}

// GetMultipleStockDataRealTime 批量获取股票实时数据（不使用缓存）
func (m *CrawlerManagerImpl) GetMultipleStockDataRealTime(ctx context.Context, codes []string) ([]*StockData, error) {
	var results []*StockData
	
	// 按数据源分组
	sinaGroup := make([]string, 0)
	tencentGroup := make([]string, 0)
	
	for _, code := range codes {
		if len(code) >= 2 {
			prefix := code[:2]
			switch prefix {
			case "hk":
				tencentGroup = append(tencentGroup, code)
			case "sh", "sz":
				sinaGroup = append(sinaGroup, code)
			default:
				sinaGroup = append(sinaGroup, code)
			}
		}
	}
	
	// 批量获取新浪实时数据
	if len(sinaGroup) > 0 {
		if sinaCrawler, exists := m.crawlers["sina"]; exists {
			sinaResults, err := sinaCrawler.GetMultipleStockDataRealTime(ctx, sinaGroup)
			if err == nil {
				results = append(results, sinaResults...)
			}
		}
	}
	
	// 批量获取腾讯实时数据
	if len(tencentGroup) > 0 {
		if tencentCrawler, exists := m.crawlers["tencent"]; exists {
			tencentResults, err := tencentCrawler.GetMultipleStockDataRealTime(ctx, tencentGroup)
			if err == nil {
				results = append(results, tencentResults...)
			}
		}
	}
	
	return results, nil
}