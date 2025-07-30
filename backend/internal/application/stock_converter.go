package application

import (
	"reflect"
	"strings"
)

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
