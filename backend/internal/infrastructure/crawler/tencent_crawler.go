package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// TencentCrawler 腾讯股票爬虫
type TencentCrawler struct {
	client  *http.Client
	baseURL string
}

// NewTencentCrawler 创建腾讯股票爬虫
func NewTencentCrawler() *TencentCrawler {
	return &TencentCrawler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		// 使用官方文档推荐的接口地址（注意：腾讯接口使用http而不是https）
		baseURL: "http://qt.gtimg.cn/q=",
	}
}

// GetSourceName 获取数据源名称
func (t *TencentCrawler) GetSourceName() string {
	return "腾讯股票"
}

// GetStockData 获取股票数据 - 接收已转换的代码
func (t *TencentCrawler) GetStockData(ctx context.Context, convertedStockCode string) (*StockData, error) {
	// 构建请求URL
	url := t.baseURL + convertedStockCode

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头，确保正确处理中文编码
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "http://gu.qq.com/")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Charset", "utf-8,gbk,gb2312;q=0.7,*;q=0.7")
	req.Header.Set("Connection", "keep-alive")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 验证响应内容
	responseStr := string(body)
	if responseStr == "" || !strings.Contains(responseStr, "\"") {
		return nil, fmt.Errorf("接口返回空数据或格式错误")
	}

	// 解析数据
	return t.parseStockData(responseStr, convertedStockCode)
}

// GetMultipleStockData 批量获取股票数据
func (t *TencentCrawler) GetMultipleStockData(ctx context.Context, codes []string) ([]*StockData, error) {
	var results []*StockData

	if len(codes) == 0 {
		return results, nil
	}

	// 构建批量请求URL
	url := t.baseURL + strings.Join(codes, ",")

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "http://gu.qq.com/")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Charset", "utf-8,gbk,gb2312;q=0.7,*;q=0.7")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析批量数据
	lines := strings.Split(string(body), "\n")
	for i, line := range lines {
		if i < len(codes) && strings.TrimSpace(line) != "" {
			if data, err := t.parseStockData(line, codes[i]); err == nil {
				results = append(results, data)
			}
		}
	}

	return results, nil
}

// parseStockData 解析股票数据
func (t *TencentCrawler) parseStockData(data, originalCode string) (*StockData, error) {
	// 提取引号内的数据
	start := strings.Index(data, "\"")
	end := strings.LastIndex(data, "\"")
	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("数据格式错误")
	}

	content := data[start+1 : end]
	fields := strings.Split(content, "~")

	// 添加调试日志
	fmt.Printf("腾讯API原始数据 [%s]: %s\n", originalCode, data)
	fmt.Printf("解析后字段数量: %d\n", len(fields))
	if len(fields) > 10 {
		fmt.Printf("前10个字段: %v\n", fields[:10])
	}
	if len(fields) > 40 {
		fmt.Printf("字段30-40: %v\n", fields[30:40])
	}

	// 简化验证：只要有基本的几个字段就可以
	if len(fields) < 10 {
		return nil, fmt.Errorf("数据字段不足，无法解析基本信息")
	}

	// 安全解析各个字段
	stockName := ""
	if len(fields) > 1 {
		stockName = strings.TrimSpace(fields[1])
	}

	// 当前价格 (字段3)
	currentPrice := 0.0
	if len(fields) > 3 {
		if price, err := strconv.ParseFloat(fields[3], 64); err == nil {
			currentPrice = price
		}
	}

	// 昨收 (字段4)
	yesterdayClose := 0.0
	if len(fields) > 4 {
		if price, err := strconv.ParseFloat(fields[4], 64); err == nil {
			yesterdayClose = price
		}
	}

	// 成交量处理 - 修复解析逻辑，先解析为浮点数再转换为整数
	volume := int64(0)
	volumeFields := []int{6, 36, 28} // 尝试多个可能的字段位置
	for _, fieldIndex := range volumeFields {
		if len(fields) > fieldIndex {
			// 先尝试解析为浮点数，然后转换为整数
			if volFloat, err := strconv.ParseFloat(fields[fieldIndex], 64); err == nil && volFloat > 0 {
				volume = int64(volFloat)
				fmt.Printf("成交量从字段%d获取: %.0f -> %d\n", fieldIndex, volFloat, volume)
				break
			}
		}
	}

	// 涨跌量和涨跌幅 (字段31, 32)
	changeValue := 0.0
	changeRate := 0.0
	if len(fields) > 31 {
		if change, err := strconv.ParseFloat(fields[31], 64); err == nil {
			changeValue = change
		}
	}
	if len(fields) > 32 {
		if rate, err := strconv.ParseFloat(fields[32], 64); err == nil {
			changeRate = rate
		}
	}

	// 如果没有直接的涨跌数据，则计算
	if changeValue == 0 && changeRate == 0 && yesterdayClose != 0 {
		changeValue = currentPrice - yesterdayClose
		changeRate = (changeValue / yesterdayClose) * 100
	}

	// 成交额处理 - 直接使用原始数字，不进行单位转换
	amount := 0.0
	amountFields := []int{37, 38} // 腾讯API中成交额字段
	for _, fieldIndex := range amountFields {
		if len(fields) > fieldIndex {
			if amt, err := strconv.ParseFloat(fields[fieldIndex], 64); err == nil && amt > 0 {
				// 直接使用原始数字，不进行任何转换
				amount = amt
				fmt.Printf("成交额从字段%d获取: %.2f (原始数字)\n", fieldIndex, amt)
				break
			}
		}
	}

	// 更新时间处理 - 使用API返回的时间（字段30）
	updateTime := time.Now() // 默认使用当前时间
	if len(fields) > 30 {
		timeStr := strings.TrimSpace(fields[30])
		if timeStr != "" && timeStr != "0" {
			// 腾讯API时间格式通常是 "2025/01/08 15:30:00"
			if parsedTime, err := time.Parse("2006/01/02 15:04:05", timeStr); err == nil {
				updateTime = parsedTime
				fmt.Printf("更新时间从字段30获取: %s\n", timeStr)
			} else {
				// 尝试其他可能的时间格式
				formats := []string{
					"2006-01-02 15:04:05",
					"2006/01/02 15:04:05",
					"15:04:05",
				}
				for _, format := range formats {
					if parsedTime, err := time.Parse(format, timeStr); err == nil {
						// 如果只有时间没有日期，使用今天的日期
						if format == "15:04:05" {
							now := time.Now()
							updateTime = time.Date(now.Year(), now.Month(), now.Day(), 
								parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), 0, now.Location())
						} else {
							updateTime = parsedTime
						}
						fmt.Printf("更新时间解析成功: %s -> %s\n", timeStr, updateTime.Format("2006-01-02 15:04:05"))
						break
					}
				}
			}
		}
	}

	fmt.Printf("最终解析结果 - 股票: %s, 价格: %.2f, 成交量: %d, 成交额: %.2f, 更新时间: %s\n", 
		stockName, currentPrice, volume, amount, updateTime.Format("2006-01-02 15:04:05"))

	return &StockData{
		Name:        stockName,
		Code:        originalCode,
		Price:       currentPrice,
		Change:      changeRate,
		ChangeValue: changeValue,
		Volume:      volume,
		Amount:      amount,
		UpdateTime:  updateTime,
	}, nil
}

// GetStockDataRealTime 获取股票实时数据（不使用缓存）
func (t *TencentCrawler) GetStockDataRealTime(ctx context.Context, convertedStockCode string) (*StockData, error) {
	// 构建请求URL，添加时间戳防止缓存
	timestamp := time.Now().UnixNano()
	url := fmt.Sprintf("%s%s&_t=%d", t.baseURL, convertedStockCode, timestamp)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头，确保正确处理中文编码，并禁用缓存
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "http://gu.qq.com/")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Charset", "utf-8,gbk,gb2312;q=0.7,*;q=0.7")
	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")
	req.Header.Set("Connection", "keep-alive")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 验证响应内容
	responseStr := string(body)
	if responseStr == "" || !strings.Contains(responseStr, "\"") {
		return nil, fmt.Errorf("接口返回空数据或格式错误")
	}

	// 解析数据
	return t.parseStockData(responseStr, convertedStockCode)
}

// GetMultipleStockDataRealTime 批量获取股票实时数据（不使用缓存）
func (t *TencentCrawler) GetMultipleStockDataRealTime(ctx context.Context, codes []string) ([]*StockData, error) {
	var results []*StockData

	if len(codes) == 0 {
		return results, nil
	}

	// 构建批量请求URL，添加时间戳防止缓存
	timestamp := time.Now().UnixNano()
	url := fmt.Sprintf("%s%s&_t=%d", t.baseURL, strings.Join(codes, ","), timestamp)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头，禁用缓存
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "http://gu.qq.com/")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Charset", "utf-8,gbk,gb2312;q=0.7,*;q=0.7")
	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")
	req.Header.Set("Connection", "keep-alive")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析批量数据
	lines := strings.Split(string(body), "\n")
	for i, line := range lines {
		if i < len(codes) && strings.TrimSpace(line) != "" {
			if data, err := t.parseStockData(line, codes[i]); err == nil {
				results = append(results, data)
			}
		}
	}

	return results, nil
}
