package main

import (
	"log"

	"D_come/internal/application"
	"D_come/internal/application/service"
	"D_come/internal/config"
	"D_come/internal/infrastructure/crawler"
	"D_come/internal/infrastructure/persistence"
	httphandler "D_come/internal/interfaces/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库连接
	database, err := persistence.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer database.Close()

	// 初始化Redis客户端
	redisClient, err := persistence.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Redis连接失败: %v", err)
	}
	defer redisClient.Close()

	// 初始化爬虫管理器
	crawlerManager := crawler.NewCrawlerManager()

	// 初始化爬虫服务
	crawlerService := application.NewCrawlerService(crawlerManager)

	// 初始化仓储（使用MySQL实现）
	stockRepo := persistence.NewStockRepository(database.DB)

	// 初始化H-A股票服务（带Redis缓存）
	haStockService := application.NewHAStockService(stockRepo, crawlerService, redisClient)

	// 初始化自定义股票服务
	customStockService := service.NewCustomStockService(stockRepo, crawlerService)

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由器
	router := gin.Default()

	// 添加CORS中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 初始化处理器
	haStockHandler := httphandler.NewHAStockHandler(haStockService)
	customStockHandler := httphandler.NewCustomStockHandler(customStockService)

	// 设置路由
	api := router.Group("/api")
	{
		// H-A股票相关路由
		stocks := api.Group("/stocks")
		{
			stocks.GET("/ha-pairs", haStockHandler.GetAllHAStocks)
			stocks.GET("/ha-pairs/realtime", haStockHandler.GetAllHAStocksRealTime)
			stocks.GET("/ha-pair", haStockHandler.GetHAStockByName)
			stocks.GET("/ha-pair/realtime", haStockHandler.GetHAStockByNameRealTime)
			stocks.POST("/refresh", haStockHandler.RefreshHAStocks)
			stocks.DELETE("/clear-cache", haStockHandler.ClearCache)
		}

		// 自定义股票相关路由
		customStockHandler.RegisterRoutes(api)
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	log.Println("数据库连接成功")
	log.Println("Redis连接成功")
	log.Println("服务器启动成功，监听端口 :8080")
	log.Println("H-A股票API: http://localhost:8080/api/stocks/ha-pairs")
	log.Println("自定义股票API: http://localhost:8080/api/custom-stocks")
	log.Println("健康检查: http://localhost:8080/health")

	// 启动服务器
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
