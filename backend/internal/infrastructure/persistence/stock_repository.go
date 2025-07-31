package persistence

import "gorm.io/gorm"

type StockRepository interface {
	GetAll() ([]*StockPair, error)
	GetByName(stockName string) (*StockPair, error)
}

type stockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) StockRepository {
	return &stockRepository{db: db}
}

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
