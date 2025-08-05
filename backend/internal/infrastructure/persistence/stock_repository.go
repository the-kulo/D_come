package persistence

import (
	"D_come/internal/domain/stock"
	"gorm.io/gorm"
)

type StockRepository interface {
	GetAll() ([]*StockPair, error)
	GetByName(stockName string) (*StockPair, error)
	
	// 新增方法
	GetAllStockPairs() ([]*stock.StockPair, error)
	GetStockPairByName(name string) (*stock.StockPair, error)
	
	GetAllCustomStocks() ([]*stock.CustomStock, error)
	GetCustomStockByID(id uint) (*stock.CustomStock, error)
	CreateCustomStock(customStock *stock.CustomStock) error
	UpdateCustomStock(customStock *stock.CustomStock) error
	DeleteCustomStock(id uint) error
}

type stockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) StockRepository {
	return &stockRepository{db: db}
}

// 原有方法保持不变
func (s *stockRepository) GetByName(stockName string) (*StockPair, error) {
	var stock StockPair
	err := s.db.Where("stock_name = ?", stockName).First(&stock).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

func (s *stockRepository) GetAll() ([]*StockPair, error) {
	var stocks []*StockPair
	err := s.db.Find(&stocks).Error
	if err != nil {
		return nil, err
	}
	return stocks, nil
}

// 新增领域层方法实现
func (s *stockRepository) GetAllStockPairs() ([]*stock.StockPair, error) {
	var stockPairs []*StockPair
	err := s.db.Find(&stockPairs).Error
	if err != nil {
		return nil, err
	}
	
	// 转换为领域实体
	result := make([]*stock.StockPair, len(stockPairs))
	for i, sp := range stockPairs {
		result[i] = &stock.StockPair{
			ID:         sp.ID,
			StockName:  sp.StockName,
			AStockCode: sp.AStockCode,
			HStockCode: sp.HStockCode,
			CrawlTime:  sp.CrawlTime,
			UpdatedAt:  sp.UpdatedAt,
		}
	}
	return result, nil
}

func (s *stockRepository) GetStockPairByName(name string) (*stock.StockPair, error) {
	var stockPair StockPair
	err := s.db.Where("stock_name = ?", name).First(&stockPair).Error
	if err != nil {
		return nil, err
	}
	
	return &stock.StockPair{
		ID:         stockPair.ID,
		StockName:  stockPair.StockName,
		AStockCode: stockPair.AStockCode,
		HStockCode: stockPair.HStockCode,
		CrawlTime:  stockPair.CrawlTime,
		UpdatedAt:  stockPair.UpdatedAt,
	}, nil
}

func (s *stockRepository) GetAllCustomStocks() ([]*stock.CustomStock, error) {
	var customStocks []*CustomStock
	err := s.db.Find(&customStocks).Error
	if err != nil {
		return nil, err
	}
	
	result := make([]*stock.CustomStock, len(customStocks))
	for i, cs := range customStocks {
		result[i] = &stock.CustomStock{
			ID:         cs.ID,
			CustomName: cs.CustomName,
			CustomCode: cs.CustomCode,
			CrawlTime:  cs.CrawlTime,
			UpdatedAt:  cs.UpdatedAt,
		}
	}
	return result, nil
}

func (s *stockRepository) GetCustomStockByID(id uint) (*stock.CustomStock, error) {
	var customStock CustomStock
	err := s.db.First(&customStock, id).Error
	if err != nil {
		return nil, err
	}
	
	return &stock.CustomStock{
		ID:         customStock.ID,
		CustomName: customStock.CustomName,
		CustomCode: customStock.CustomCode,
		CrawlTime:  customStock.CrawlTime,
		UpdatedAt:  customStock.UpdatedAt,
	}, nil
}

func (s *stockRepository) CreateCustomStock(customStock *stock.CustomStock) error {
	dbCustomStock := &CustomStock{
		CustomName: customStock.CustomName,
		CustomCode: customStock.CustomCode,
		CrawlTime:  customStock.CrawlTime,
		UpdatedAt:  customStock.UpdatedAt,
	}
	
	err := s.db.Create(dbCustomStock).Error
	if err != nil {
		return err
	}
	
	customStock.ID = dbCustomStock.ID
	return nil
}

func (s *stockRepository) UpdateCustomStock(customStock *stock.CustomStock) error {
	dbCustomStock := &CustomStock{
		ID:         customStock.ID,
		CustomName: customStock.CustomName,
		CustomCode: customStock.CustomCode,
		CrawlTime:  customStock.CrawlTime,
		UpdatedAt:  customStock.UpdatedAt,
	}
	
	return s.db.Save(dbCustomStock).Error
}

func (s *stockRepository) DeleteCustomStock(id uint) error {
	return s.db.Delete(&CustomStock{}, id).Error
}
