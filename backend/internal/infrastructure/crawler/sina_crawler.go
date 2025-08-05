package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// SinaCrawler 爬取新浪财经股票数据
// 支持单次及批量查询，并自动处理 GB18030 转 UTF-8
type SinaCrawler struct {
	client  *http.Client
	baseURL string
	headers http.Header
}

// NewSinaCrawler 创建默认配置的爬虫
func NewSinaCrawler() *SinaCrawler {
	headers := make(http.Header)
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	headers.Set("Accept", "*/*")
	headers.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	headers.Set("Accept-Encoding", "identity")
	headers.Set("Referer", "https://finance.sina.com.cn/")
	headers.Set("Connection", "keep-alive")

	return &SinaCrawler{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: "https://hq.sinajs.cn/list=",
		headers: headers,
	}
}

// GetSourceName 返回数据源名称
func (s *SinaCrawler) GetSourceName() string {
	return "新浪财经"
}

// GetStockData 查询单只股票数据
func (s *SinaCrawler) GetStockData(ctx context.Context, code string) (*StockData, error) {
	raw, err := s.fetch(ctx, s.baseURL+code)
	if err != nil {
		return nil, err
	}
	return s.parseStockData(raw, code)
}

// GetMultipleStockData 批量查询
func (s *SinaCrawler) GetMultipleStockData(ctx context.Context, codes []string) ([]*StockData, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	raw, err := s.fetch(ctx, s.baseURL+strings.Join(codes, ","))
	if err != nil {
		return nil, err
	}

	var results []*StockData
	for i, line := range strings.Split(raw, "\n") {
		if i < len(codes) && strings.TrimSpace(line) != "" {
			if data, err := s.parseStockData(line, codes[i]); err == nil {
				results = append(results, data)
			}
		}
	}
	return results, nil
}

// fetch 发起 HTTP 请求并返回 UTF-8 编码的响应体
func (s *SinaCrawler) fetch(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header = s.headers.Clone()

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 先按 GB18030 解码，再读全
	reader := transform.NewReader(resp.Body, simplifiedchinese.GB18030.NewDecoder())
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("读取&解码失败: %w", err)
	}
	return string(data), nil
}

// parseStockData 解析单条数据，提取双引号中的 CSV 内容
func (s *SinaCrawler) parseStockData(line, code string) (*StockData, error) {
	// 格式: var hq_str_{code}="...";
	prefix := fmt.Sprintf("var hq_str_%s=\"", code)
	suffix := "\";"

	// 去除前后缀
	raw := strings.TrimPrefix(line, prefix)
	raw = strings.TrimSuffix(raw, suffix)

	fields := strings.Split(raw, ",")
	if len(fields) < 5 {
		return nil, fmt.Errorf("字段数不足(%d) for %s", len(fields), code)
	}

	name := fields[0]
	price := parseFloat(fields, 3)
	prev := parseFloat(fields, 2)
	vol := parseInt(fields, 8)
	amt := parseFloat(fields, 9) // 直接使用原始数字，不进行单位转换

	// 解析更新时间 - 新浪API通常在字段30和31包含日期和时间
	updateTime := time.Now() // 默认使用当前时间
	if len(fields) > 31 {
		dateStr := strings.TrimSpace(fields[30]) // 日期字段
		timeStr := strings.TrimSpace(fields[31]) // 时间字段
		
		if dateStr != "" && timeStr != "" {
			// 组合日期和时间
			dateTimeStr := dateStr + " " + timeStr
			// 尝试解析时间
			formats := []string{
				"2006-01-02 15:04:05",
				"2006/01/02 15:04:05",
			}
			for _, format := range formats {
				if parsedTime, err := time.Parse(format, dateTimeStr); err == nil {
					updateTime = parsedTime
					break
				}
			}
		}
	}

	return &StockData{
		Name:        name,
		Code:        code,
		Price:       price,
		ChangeValue: price - prev,
		Change:      pct(price, prev),
		Volume:      vol,
		Amount:      amt, // 直接使用原始数字
		UpdateTime:  updateTime,
	}, nil
}

func parseFloat(fields []string, idx int) float64 {
	if idx < len(fields) {
		if v, err := strconv.ParseFloat(fields[idx], 64); err == nil {
			return v
		}
	}
	return 0
}

func parseInt(fields []string, idx int) int64 {
	if idx < len(fields) {
		if v, err := strconv.ParseInt(fields[idx], 10, 64); err == nil {
			return v
		}
	}
	return 0
}

func pct(curr, prev float64) float64 {
	if prev != 0 {
		return (curr - prev) / prev * 100
	}
	return 0
}

// GetStockDataRealTime 获取股票实时数据（不使用缓存）
func (s *SinaCrawler) GetStockDataRealTime(ctx context.Context, code string) (*StockData, error) {
	return s.GetStockData(ctx, code)
}

// GetMultipleStockDataRealTime 批量获取股票实时数据（不使用缓存）
func (s *SinaCrawler) GetMultipleStockDataRealTime(ctx context.Context, codes []string) ([]*StockData, error) {
	return s.GetMultipleStockData(ctx, codes)
}
