package application

import (
	"D_come/internal/domain/stock"
	"D_come/internal/infrastructure/crawler"
	"D_come/internal/infrastructure/persistence"
	"context"
	"fmt"
	"time"
)

// HAStockService H-A股票服务 - 专门处理H-A股票对的业务逻辑
type HAStockService struct {
	stockRepo      stock.Repository
	crawlerService *CrawlerService
	redisClient    *persistence.RedisClient
}

// HAStockData H-A股票完整数据
type HAStockData struct {
	StockName  string             `json:"stock_name"`
	HStockCode string             `json:"h_stock_code"`
	HStockData *crawler.StockData `json:"h_stock_data"`
	AStockCode string             `json:"a_stock_code"`
	AStockData *crawler.StockData `json:"a_stock_data"`
}

// NewHAStockService 创建H-A股票服务
func NewHAStockService(stockRepo stock.Repository, crawlerService *CrawlerService, redisClient *persistence.RedisClient) *HAStockService {
	return &HAStockService{
		stockRepo:      stockRepo,
		crawlerService: crawlerService,
		redisClient:    redisClient,
	}
}

// GetAllHAStockData 获取所有H-A股票的实时数据
func (s *HAStockService) GetAllHAStockData(ctx context.Context) ([]*HAStockData, error) {
	// 1. 尝试从Redis获取缓存数据
	cacheKey := "ha_stocks:all"
	var cachedResults []*HAStockData
	
	err := s.redisClient.GetStockData(ctx, cacheKey, &cachedResults)
	if err == nil && len(cachedResults) > 0 {
		// 检查缓存数据是否还新鲜（缩短到2分钟）
		if s.isCacheDataFresh(cachedResults) {
			return cachedResults, nil
		}
	}

	// 2. 缓存未命中或数据过期，从数据库获取所有H-A股票对
	stockPairs, err := s.stockRepo.GetAllStockPairs()
	if err != nil {
		return nil, fmt.Errorf("获取股票对失败: %w", err)
	}

	var results []*HAStockData

	// 3. 为每个股票对获取实时数据
	for _, pair := range stockPairs {
		haData, err := s.getHAStockDataByPair(ctx, pair)
		if err != nil {
			// 记录错误但继续处理其他股票
			continue
		}
		results = append(results, haData)
	}

	// 4. 缓存到Redis（缩短到2分钟过期）
	if len(results) > 0 {
		s.cacheToRedis(ctx, cacheKey, results, 2*time.Minute)
	}

	return results, nil
}

// GetHAStockDataByName 根据股票名称获取H-A股票数据
func (s *HAStockService) GetHAStockDataByName(ctx context.Context, stockName string) (*HAStockData, error) {
	// 1. 尝试从Redis获取缓存数据
	cacheKey := fmt.Sprintf("ha_stock:%s", stockName)
	var cachedResult HAStockData
	
	err := s.redisClient.GetStockData(ctx, cacheKey, &cachedResult)
	if err == nil {
		// 检查缓存数据是否还新鲜（2分钟内）
		if s.isSingleStockCacheFresh(&cachedResult) {
			return &cachedResult, nil
		}
	}

	// 2. 缓存未命中或数据过期，从数据库获取股票对信息
	stockPair, err := s.stockRepo.GetStockPairByName(stockName)
	if err != nil {
		return nil, fmt.Errorf("获取股票对失败: %w", err)
	}

	// 3. 获取实时数据
	result, err := s.getHAStockDataByPair(ctx, stockPair)
	if err != nil {
		return nil, err
	}

	// 4. 缓存到Redis（2分钟过期）
	s.cacheToRedis(ctx, cacheKey, result, 2*time.Minute)

	return result, nil
}

// getHAStockDataByPair 根据股票对获取完整的H-A股票数据
func (s *HAStockService) getHAStockDataByPair(ctx context.Context, pair *stock.StockPair) (*HAStockData, error) {
	// 转换股票代码格式 (从 601038.SH 转换为 sh601038)
	aStockCode := s.convertStockCode(pair.AStockCode)
	hStockCode := s.convertStockCode(pair.HStockCode)

	// 并发获取A股和H股数据
	aStockChan := make(chan *crawler.StockData, 1)
	hStockChan := make(chan *crawler.StockData, 1)
	errChan := make(chan error, 2)

	// 获取A股数据 (使用新浪)
	go func() {
		data, err := s.crawlerService.GetStockDataFromSource(ctx, aStockCode, "sina")
		if err != nil {
			errChan <- fmt.Errorf("获取A股数据失败: %w", err)
			return
		}
		aStockChan <- data
	}()

	// 获取H股数据 (使用腾讯)
	go func() {
		data, err := s.crawlerService.GetStockDataFromSource(ctx, hStockCode, "tencent")
		if err != nil {
			errChan <- fmt.Errorf("获取H股数据失败: %w", err)
			return
		}
		hStockChan <- data
	}()

	// 等待结果
	var aStockData, hStockData *crawler.StockData
	var errors []error

	for i := 0; i < 2; i++ {
		select {
		case data := <-aStockChan:
			aStockData = data
		case data := <-hStockChan:
			hStockData = data
		case err := <-errChan:
			errors = append(errors, err)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// 检查是否有错误
	if len(errors) > 0 {
		return nil, fmt.Errorf("获取股票数据时发生错误: %v", errors)
	}

	// 构建结果
	result := &HAStockData{
		StockName:  pair.StockName,
		HStockCode: pair.HStockCode,
		HStockData: hStockData,
		AStockCode: pair.AStockCode,
		AStockData: aStockData,
	}

	return result, nil
}

// cacheToRedis 缓存数据到Redis
func (s *HAStockService) cacheToRedis(ctx context.Context, key string, data interface{}, expiration time.Duration) {
	err := s.redisClient.SetStockData(ctx, key, data, expiration)
	if err != nil {
		// 记录错误但不影响主流程
		fmt.Printf("缓存数据到Redis失败: %v\n", err)
	}
}

// isCacheDataFresh 检查缓存数据是否新鲜
func (s *HAStockService) isCacheDataFresh(data []*HAStockData) bool {
	if len(data) == 0 {
		return false
	}
	
	// 检查第一个股票的更新时间（缩短到2分钟）
	firstStock := data[0]
	if firstStock.AStockData != nil {
		return time.Since(firstStock.AStockData.UpdateTime) < 2*time.Minute
	}
	if firstStock.HStockData != nil {
		return time.Since(firstStock.HStockData.UpdateTime) < 2*time.Minute
	}
	
	return false
}

// isSingleStockCacheFresh 检查单个股票缓存数据是否新鲜
func (s *HAStockService) isSingleStockCacheFresh(data *HAStockData) bool {
	if data.AStockData != nil {
		return time.Since(data.AStockData.UpdateTime) < 1*time.Minute
	}
	if data.HStockData != nil {
		return time.Since(data.HStockData.UpdateTime) < 1*time.Minute
	}
	
	return false
}

// convertStockCode 转换股票代码格式
// 从 "601038.SH" 转换为 "sh601038"
// 从 "00038.HK" 转换为 "hk00038"
func (s *HAStockService) convertStockCode(code string) string {
	if len(code) < 3 {
		return code
	}

	// 查找点号位置
	dotIndex := -1
	for i, char := range code {
		if char == '.' {
			dotIndex = i
			break
		}
	}

	if dotIndex == -1 {
		return code
	}

	// 分离数字部分和市场部分
	number := code[:dotIndex]
	market := code[dotIndex+1:]

	// 转换市场代码
	switch market {
	case "SH":
		return "sh" + number
	case "SZ":
		return "sz" + number
	case "HK":
		return "hk" + number
	default:
		return code
	}
}

// RefreshAllHAStockData 刷新所有H-A股票数据到Redis
func (s *HAStockService) RefreshAllHAStockData(ctx context.Context) error {
	haStockDataList, err := s.GetAllHAStockData(ctx)
	if err != nil {
		return fmt.Errorf("获取H-A股票数据失败: %w", err)
	}

	// 批量更新到Redis
	dataMap := make(map[string]interface{})
	for _, data := range haStockDataList {
		key := fmt.Sprintf("ha_stock:%s", data.StockName)
		dataMap[key] = data
	}

	// 批量存储，5分钟过期
	err = s.redisClient.SetMultipleStockData(ctx, dataMap, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("批量更新Redis失败: %w", err)
	}

	// 同时更新全量缓存
	allDataKey := "ha_stocks:all"
	err = s.redisClient.SetStockData(ctx, allDataKey, haStockDataList, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("更新全量缓存失败: %w", err)
	}

	return nil
}

// ForceRefreshAllHAStockData 强制刷新所有H-A股票数据到Redis（不使用缓存）
func (s *HAStockService) ForceRefreshAllHAStockData(ctx context.Context) error {
	// 1. 从数据库获取所有H-A股票对
	stockPairs, err := s.stockRepo.GetAllStockPairs()
	if err != nil {
		return fmt.Errorf("获取股票对失败: %w", err)
	}

	var results []*HAStockData

	// 2. 为每个股票对强制获取最新实时数据
	for _, pair := range stockPairs {
		haData, err := s.getHAStockDataByPair(ctx, pair)
		if err != nil {
			// 记录错误但继续处理其他股票
			fmt.Printf("获取股票 %s 数据失败: %v\n", pair.StockName, err)
			continue
		}
		results = append(results, haData)
	}

	if len(results) == 0 {
		return fmt.Errorf("没有获取到任何股票数据")
	}

	// 3. 批量更新到Redis
	dataMap := make(map[string]interface{})
	for _, data := range results {
		key := fmt.Sprintf("ha_stock:%s", data.StockName)
		dataMap[key] = data
	}

	// 批量存储，2分钟过期
	err = s.redisClient.SetMultipleStockData(ctx, dataMap, 2*time.Minute)
	if err != nil {
		return fmt.Errorf("批量更新Redis失败: %w", err)
	}

	// 4. 同时更新全量缓存
	allDataKey := "ha_stocks:all"
	err = s.redisClient.SetStockData(ctx, allDataKey, results, 2*time.Minute)
	if err != nil {
		return fmt.Errorf("更新全量缓存失败: %w", err)
	}

	return nil
}

// ClearCache 清除指定股票的缓存
func (s *HAStockService) ClearCache(ctx context.Context, stockName string) error {
	key := fmt.Sprintf("ha_stock:%s", stockName)
	return s.redisClient.DeleteStockData(ctx, key)
}

// ClearAllCache 清除所有缓存
func (s *HAStockService) ClearAllCache(ctx context.Context) error {
	// 清除全量缓存
	allDataKey := "ha_stocks:all"
	err := s.redisClient.DeleteStockData(ctx, allDataKey)
	if err != nil {
		return fmt.Errorf("清除全量缓存失败: %w", err)
	}

	// 获取所有股票名称并清除单个缓存
	stockPairs, err := s.stockRepo.GetAllStockPairs()
	if err != nil {
		return fmt.Errorf("获取股票对失败: %w", err)
	}

	for _, pair := range stockPairs {
		key := fmt.Sprintf("ha_stock:%s", pair.StockName)
		s.redisClient.DeleteStockData(ctx, key) // 忽略单个删除错误
	}

	return nil
}

// GetAllHAStockDataRealTime 获取所有H-A股票的实时数据（不使用缓存）
func (s *HAStockService) GetAllHAStockDataRealTime(ctx context.Context) ([]*HAStockData, error) {
	// 直接从数据库获取所有H-A股票对，不使用缓存
	stockPairs, err := s.stockRepo.GetAllStockPairs()
	if err != nil {
		return nil, fmt.Errorf("获取股票对失败: %w", err)
	}

	var results []*HAStockData

	// 为每个股票对获取实时数据
	for _, pair := range stockPairs {
		haData, err := s.getHAStockDataByPairRealTime(ctx, pair)
		if err != nil {
			// 记录错误但继续处理其他股票
			fmt.Printf("获取股票 %s 实时数据失败: %v\n", pair.StockName, err)
			continue
		}
		results = append(results, haData)
	}

	return results, nil
}

// GetHAStockDataByNameRealTime 根据股票名称获取H-A股票实时数据（不使用缓存）
func (s *HAStockService) GetHAStockDataByNameRealTime(ctx context.Context, stockName string) (*HAStockData, error) {
	// 直接从数据库获取股票对信息，不使用缓存
	stockPair, err := s.stockRepo.GetStockPairByName(stockName)
	if err != nil {
		return nil, fmt.Errorf("获取股票对失败: %w", err)
	}

	// 获取实时数据
	result, err := s.getHAStockDataByPairRealTime(ctx, stockPair)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// getHAStockDataByPairRealTime 根据股票对获取完整的H-A股票实时数据（不使用缓存）
func (s *HAStockService) getHAStockDataByPairRealTime(ctx context.Context, pair *stock.StockPair) (*HAStockData, error) {
	// 转换股票代码格式 (从 601038.SH 转换为 sh601038)
	aStockCode := s.convertStockCode(pair.AStockCode)
	hStockCode := s.convertStockCode(pair.HStockCode)

	// 并发获取A股和H股数据
	aStockChan := make(chan *crawler.StockData, 1)
	hStockChan := make(chan *crawler.StockData, 1)
	errChan := make(chan error, 2)

	// 获取A股数据 (使用新浪)
	go func() {
		data, err := s.crawlerService.GetStockDataFromSourceRealTime(ctx, aStockCode, "sina")
		if err != nil {
			errChan <- fmt.Errorf("获取A股实时数据失败: %w", err)
			return
		}
		aStockChan <- data
	}()

	// 获取H股数据 (使用腾讯)
	go func() {
		data, err := s.crawlerService.GetStockDataFromSourceRealTime(ctx, hStockCode, "tencent")
		if err != nil {
			errChan <- fmt.Errorf("获取H股实时数据失败: %w", err)
			return
		}
		hStockChan <- data
	}()

	// 等待结果
	var aStockData, hStockData *crawler.StockData
	var errors []error

	for i := 0; i < 2; i++ {
		select {
		case data := <-aStockChan:
			aStockData = data
		case data := <-hStockChan:
			hStockData = data
		case err := <-errChan:
			errors = append(errors, err)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// 检查是否有错误
	if len(errors) > 0 {
		return nil, fmt.Errorf("获取股票实时数据时发生错误: %v", errors)
	}

	// 构建结果
	result := &HAStockData{
		StockName:  pair.StockName,
		HStockCode: pair.HStockCode,
		HStockData: hStockData,
		AStockCode: pair.AStockCode,
		AStockData: aStockData,
	}

	return result, nil
}
