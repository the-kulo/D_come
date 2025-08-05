package http

import (
	"D_come/internal/application"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HAStockHandler H-A股票HTTP处理器
type HAStockHandler struct {
	haStockService *application.HAStockService
}

// NewHAStockHandler 创建H-A股票处理器
func NewHAStockHandler(haStockService *application.HAStockService) *HAStockHandler {
	return &HAStockHandler{
		haStockService: haStockService,
	}
}

// GetAllHAStocks 获取所有H-A股票数据
func (h *HAStockHandler) GetAllHAStocks(c *gin.Context) {
	// 获取H-A股票数据（带Redis缓存）
	haStockDataList, err := h.haStockService.GetAllHAStockData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取股票数据失败: " + err.Error(),
		})
		return
	}

	// 转换为响应格式
	var response []HAStockResponse
	for _, data := range haStockDataList {
		haResponse := HAStockResponse{
			StockName:  data.StockName,
			HStockCode: data.HStockCode,
			AStockCode: data.AStockCode,
			CacheHit:   true, // 这里可以根据实际情况设置
		}

		// 转换H股数据和更新时间
		if data.HStockData != nil {
			haResponse.HStockData = &StockInfo{
				Name:        data.HStockData.Name,
				Code:        data.HStockData.Code,
				Price:       data.HStockData.Price,
				Change:      data.HStockData.Change,
				ChangeValue: data.HStockData.ChangeValue,
				Volume:      data.HStockData.Volume,
				Amount:      data.HStockData.Amount,
			}
			haResponse.HUpdateTime = data.HStockData.UpdateTime.Format("2006-01-02 15:04:05")
		}

		// 转换A股数据和更新时间
		if data.AStockData != nil {
			haResponse.AStockData = &StockInfo{
				Name:        data.AStockData.Name,
				Code:        data.AStockData.Code,
				Price:       data.AStockData.Price,
				Change:      data.AStockData.Change,
				ChangeValue: data.AStockData.ChangeValue,
				Volume:      data.AStockData.Volume,
				Amount:      data.AStockData.Amount,
			}
			haResponse.AUpdateTime = data.AStockData.UpdateTime.Format("2006-01-02 15:04:05")
		}

		response = append(response, haResponse)
	}

	c.JSON(http.StatusOK, response)
}

// GetHAStockByName 根据股票名称获取H-A股票数据
func (h *HAStockHandler) GetHAStockByName(c *gin.Context) {
	// 获取股票名称参数
	stockName := c.Query("name")
	if stockName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "股票名称参数不能为空",
		})
		return
	}

	// 获取指定股票数据（带Redis缓存）
	haStockData, err := h.haStockService.GetHAStockDataByName(c.Request.Context(), stockName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取股票数据失败: " + err.Error(),
		})
		return
	}

	// 转换为响应格式
	response := HAStockResponse{
		StockName:  haStockData.StockName,
		HStockCode: haStockData.HStockCode,
		AStockCode: haStockData.AStockCode,
		CacheHit:   true, // 这里可以根据实际情况设置
	}

	// 转换H股数据和更新时间
	if haStockData.HStockData != nil {
		response.HStockData = &StockInfo{
			Name:        haStockData.HStockData.Name,
			Code:        haStockData.HStockData.Code,
			Price:       haStockData.HStockData.Price,
			Change:      haStockData.HStockData.Change,
			ChangeValue: haStockData.HStockData.ChangeValue,
			Volume:      haStockData.HStockData.Volume,
			Amount:      haStockData.HStockData.Amount,
		}
		response.HUpdateTime = haStockData.HStockData.UpdateTime.Format("2006-01-02 15:04:05")
	}

	// 转换A股数据和更新时间
	if haStockData.AStockData != nil {
		response.AStockData = &StockInfo{
			Name:        haStockData.AStockData.Name,
			Code:        haStockData.AStockData.Code,
			Price:       haStockData.AStockData.Price,
			Change:      haStockData.AStockData.Change,
			ChangeValue: haStockData.AStockData.ChangeValue,
			Volume:      haStockData.AStockData.Volume,
			Amount:      haStockData.AStockData.Amount,
		}
		response.AUpdateTime = haStockData.AStockData.UpdateTime.Format("2006-01-02 15:04:05")
	}

	c.JSON(http.StatusOK, response)
}

// RefreshHAStocks 刷新所有H-A股票数据到Redis
func (h *HAStockHandler) RefreshHAStocks(c *gin.Context) {
	// 先清除所有缓存
	err := h.haStockService.ClearAllCache(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "清除缓存失败: " + err.Error(),
		})
		return
	}

	// 强制刷新所有H-A股票数据到Redis
	err = h.haStockService.ForceRefreshAllHAStockData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "刷新股票数据失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "股票数据已成功强制刷新到Redis缓存",
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

// ClearCache 清除缓存
func (h *HAStockHandler) ClearCache(c *gin.Context) {
	// 获取股票名称参数（可选）
	stockName := c.Query("name")

	var err error
	var message string

	if stockName != "" {
		// 清除指定股票的缓存
		err = h.haStockService.ClearCache(c.Request.Context(), stockName)
		message = "指定股票缓存已清除"
	} else {
		// 清除所有缓存
		err = h.haStockService.ClearAllCache(c.Request.Context())
		message = "所有股票缓存已清除"
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "清除缓存失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}

// GetAllHAStocksRealTime 获取所有H-A股票实时数据（不使用缓存）
func (h *HAStockHandler) GetAllHAStocksRealTime(c *gin.Context) {
	// 设置不缓存的响应头
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// 获取H-A股票实时数据（不使用Redis缓存）
	haStockDataList, err := h.haStockService.GetAllHAStockDataRealTime(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取实时股票数据失败: " + err.Error(),
		})
		return
	}

	// 转换为响应格式
	var response []HAStockResponse
	for _, data := range haStockDataList {
		haResponse := HAStockResponse{
			StockName:  data.StockName,
			HStockCode: data.HStockCode,
			AStockCode: data.AStockCode,
			CacheHit:   false, // 实时数据不使用缓存
		}

		// 转换H股数据和更新时间
		if data.HStockData != nil {
			haResponse.HStockData = &StockInfo{
				Name:        data.HStockData.Name,
				Code:        data.HStockData.Code,
				Price:       data.HStockData.Price,
				Change:      data.HStockData.Change,
				ChangeValue: data.HStockData.ChangeValue,
				Volume:      data.HStockData.Volume,
				Amount:      data.HStockData.Amount,
			}
			haResponse.HUpdateTime = data.HStockData.UpdateTime.Format("2006-01-02 15:04:05")
		}

		// 转换A股数据和更新时间
		if data.AStockData != nil {
			haResponse.AStockData = &StockInfo{
				Name:        data.AStockData.Name,
				Code:        data.AStockData.Code,
				Price:       data.AStockData.Price,
				Change:      data.AStockData.Change,
				ChangeValue: data.AStockData.ChangeValue,
				Volume:      data.AStockData.Volume,
				Amount:      data.AStockData.Amount,
			}
			haResponse.AUpdateTime = data.AStockData.UpdateTime.Format("2006-01-02 15:04:05")
		}

		response = append(response, haResponse)
	}

	c.JSON(http.StatusOK, response)
}

// GetHAStockByNameRealTime 根据股票名称获取H-A股票实时数据（不使用缓存）
func (h *HAStockHandler) GetHAStockByNameRealTime(c *gin.Context) {
	// 设置不缓存的响应头
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// 获取股票名称参数
	stockName := c.Query("name")
	if stockName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "股票名称参数不能为空",
		})
		return
	}

	// 获取指定股票实时数据（不使用Redis缓存）
	haStockData, err := h.haStockService.GetHAStockDataByNameRealTime(c.Request.Context(), stockName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取实时股票数据失败: " + err.Error(),
		})
		return
	}

	// 转换为响应格式
	response := HAStockResponse{
		StockName:  haStockData.StockName,
		HStockCode: haStockData.HStockCode,
		AStockCode: haStockData.AStockCode,
		CacheHit:   false, // 实时数据不使用缓存
	}

	// 转换H股数据和更新时间
	if haStockData.HStockData != nil {
		response.HStockData = &StockInfo{
			Name:        haStockData.HStockData.Name,
			Code:        haStockData.HStockData.Code,
			Price:       haStockData.HStockData.Price,
			Change:      haStockData.HStockData.Change,
			ChangeValue: haStockData.HStockData.ChangeValue,
			Volume:      haStockData.HStockData.Volume,
			Amount:      haStockData.HStockData.Amount,
		}
		response.HUpdateTime = haStockData.HStockData.UpdateTime.Format("2006-01-02 15:04:05")
	}

	// 转换A股数据和更新时间
	if haStockData.AStockData != nil {
		response.AStockData = &StockInfo{
			Name:        haStockData.AStockData.Name,
			Code:        haStockData.AStockData.Code,
			Price:       haStockData.AStockData.Price,
			Change:      haStockData.AStockData.Change,
			ChangeValue: haStockData.AStockData.ChangeValue,
			Volume:      haStockData.AStockData.Volume,
			Amount:      haStockData.AStockData.Amount,
		}
		response.AUpdateTime = haStockData.AStockData.UpdateTime.Format("2006-01-02 15:04:05")
	}

	c.JSON(http.StatusOK, response)
}
