package http

import (
	"net/http"
	"strconv"

	"D_come/internal/application/service"

	"github.com/gin-gonic/gin"
)

// CustomStockHandler 自定义股票HTTP处理器
type CustomStockHandler struct {
	customStockService *service.CustomStockService
}

// NewCustomStockHandler 创建自定义股票处理器
func NewCustomStockHandler(customStockService *service.CustomStockService) *CustomStockHandler {
	return &CustomStockHandler{
		customStockService: customStockService,
	}
}

// CreateCustomStockRequest 创建自定义股票请求
type CreateCustomStockRequest struct {
	CustomName string `json:"customName" binding:"required"`
	CustomCode string `json:"customCode" binding:"required"`
}

// UpdateCustomStockRequest 更新自定义股票请求
type UpdateCustomStockRequest struct {
	CustomName string `json:"customName" binding:"required"`
	CustomCode string `json:"customCode" binding:"required"`
}

// GetAllCustomStocks 获取所有自定义股票
func (h *CustomStockHandler) GetAllCustomStocks(c *gin.Context) {
	stocks, err := h.customStockService.GetAllCustomStocks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stocks,
	})
}

// GetCustomStockByID 根据ID获取自定义股票
func (h *CustomStockHandler) GetCustomStockByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的ID",
		})
		return
	}

	stock, stockErr := h.customStockService.GetCustomStockByID(c.Request.Context(), uint(id))
	if stockErr != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": stockErr.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stock,
	})
}

// CreateCustomStock 创建自定义股票
func (h *CustomStockHandler) CreateCustomStock(c *gin.Context) {
	var req CreateCustomStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	stock, createErr := h.customStockService.CreateCustomStock(
		c.Request.Context(),
		req.CustomName,
		req.CustomCode,
	)
	if createErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": createErr.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": stock,
	})
}

// UpdateCustomStock 更新自定义股票
func (h *CustomStockHandler) UpdateCustomStock(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的ID",
		})
		return
	}

	var req UpdateCustomStockRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": bindErr.Error(),
		})
		return
	}

	stock, updateErr := h.customStockService.UpdateCustomStock(
		c.Request.Context(),
		uint(id),
		req.CustomName,
		req.CustomCode,
	)
	if updateErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": updateErr.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stock,
	})
}

// DeleteCustomStock 删除自定义股票
func (h *CustomStockHandler) DeleteCustomStock(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的ID",
		})
		return
	}

	if deleteErr := h.customStockService.DeleteCustomStock(uint(id)); deleteErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": deleteErr.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// RegisterRoutes 注册路由
func (h *CustomStockHandler) RegisterRoutes(router *gin.RouterGroup) {
	customStocks := router.Group("/custom-stocks")
	{
		customStocks.GET("", h.GetAllCustomStocks)
		customStocks.GET("/:id", h.GetCustomStockByID)
		customStocks.POST("", h.CreateCustomStock)
		customStocks.PUT("/:id", h.UpdateCustomStock)
		customStocks.DELETE("/:id", h.DeleteCustomStock)
	}
}
