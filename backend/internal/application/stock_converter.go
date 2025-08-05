package application

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// StockCode 股票代码结构
type StockCode struct {
	Region string `json:"region"` // hk, sh, sz
	Number string `json:"number"` // 数字部分
}

// StockCodeConverter 股票代码转换器
type StockCodeConverter struct {
	// 使用反射配置不同数据源的转换规则
	converters map[string]interface{}
}

// 新浪财经转换规则
type SinaConverter struct {
	HK string `format:"hk%s"`     // hk00038
	SH string `format:"sh%s"`     // sh601038  
	SZ string `format:"sz%s"`     // sz000001
}

// 腾讯股票转换规则  
type TencentConverter struct {
	HK string `format:"hk%s"`     // hk00038
	SH string `format:"sh%s"`     // sh601038
	SZ string `format:"sz%s"`     // sz000001
}

// NewStockCodeConverter 创建转换器
func NewStockCodeConverter() *StockCodeConverter {
	return &StockCodeConverter{
		converters: map[string]interface{}{
			"sina":    &SinaConverter{},
			"tencent": &TencentConverter{},
		},
	}
}

// ParseStockCode 解析股票代码
func (c *StockCodeConverter) ParseStockCode(code string) (*StockCode, error) {
	// 使用正则匹配 地区+数字 格式
	re := regexp.MustCompile(`^([a-zA-Z]+)(\d+)$`)
	matches := re.FindStringSubmatch(strings.ToLower(code))
	
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid stock code format: %s", code)
	}
	
	return &StockCode{
		Region: matches[1],
		Number: matches[2],
	}, nil
}

// ConvertForSource 为指定数据源转换股票代码
func (c *StockCodeConverter) ConvertForSource(code, source string) (string, error) {
	stockCode, err := c.ParseStockCode(code)
	if err != nil {
		return "", err
	}
	
	converter, exists := c.converters[source]
	if !exists {
		return "", fmt.Errorf("unsupported source: %s", source)
	}
	
	return c.convertUsingReflection(stockCode, converter)
}

// convertUsingReflection 使用反射进行转换
func (c *StockCodeConverter) convertUsingReflection(stockCode *StockCode, converter interface{}) (string, error) {
	v := reflect.ValueOf(converter).Elem()
	t := v.Type()
	
	// 查找匹配的地区字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if strings.ToLower(field.Name) == stockCode.Region {
			format := field.Tag.Get("format")
			if format == "" {
				return "", fmt.Errorf("no format found for region: %s", stockCode.Region)
			}
			return fmt.Sprintf(format, stockCode.Number), nil
		}
	}
	
	return "", fmt.Errorf("unsupported region: %s", stockCode.Region)
}

// GetSupportedRegions 获取支持的地区列表
func (c *StockCodeConverter) GetSupportedRegions(source string) []string {
	converter, exists := c.converters[source]
	if !exists {
		return nil
	}
	
	var regions []string
	t := reflect.TypeOf(converter).Elem()
	
	for i := 0; i < t.NumField(); i++ {
		regions = append(regions, strings.ToLower(t.Field(i).Name))
	}
	
	return regions
}

// IsValidStockCode 验证股票代码格式
func (c *StockCodeConverter) IsValidStockCode(code string) bool {
	_, err := c.ParseStockCode(code)
	return err == nil
}

type CrawlerInput struct {
	StockName     string `json:"stock_name"`
	OriginalACode string `json:"a_stock_code" transform:"true"`
	OriginalHCode string `json:"h_stock_code" transform:"true"`
}

func (s *CrawlerInput) Normalize() {
	v := reflect.ValueOf(s).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("transform") == "true" && field.Type.Kind() == reflect.String {
			parts := strings.Split(v.Field(i).String(), ".")
			if len(parts) == 2 {
				v.Field(i).SetString(parts[1] + parts[0])
			}
		}
	}

}
