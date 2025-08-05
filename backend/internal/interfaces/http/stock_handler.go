package http

import (
	"D_come/internal/application"
	"net/http"

	"github.com/gin-gonic/gin"
)

// StockHandler 股票HTTP处理器
type StockHandler struct {
	stockService *application.StockService
}

// NewStockHandler 创建股票处理器
func NewStockHandler(stockService *application.StockService) *StockHandler {
	return &StockHandler{
		stockService: stockService,
	}
}

// GetAllCustomStocks 获取所有自定义股票
func (h *StockHandler) GetAllCustomStocks(c *gin.Context) {
	// 获取所有自定义股票
	customStocks, err := h.stockService.GetAllCustomStocks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取自定义股票失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, customStocks)
}

// CreateCustomStock 创建自定义股票
func (h *StockHandler) CreateCustomStock(c *gin.Context) {
	// TODO: 实现创建自定义股票的逻辑
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "自定义股票创建功能待实现",
	})
}

// GetCustomStockData 获取自定义股票数据
func (h *StockHandler) GetCustomStockData(c *gin.Context) {
	// 获取股票代码参数
	stockCode := c.Query("code")
	if stockCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "股票代码参数不能为空",
		})
		return
	}

	// 获取股票数据
	stockData, err := h.stockService.GetCustomStockData(c.Request.Context(), stockCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取股票数据失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stockData)
}
