package grpc

import (
	"D_come/internal/application"
	"D_come/internal/infrastructure/persistence"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StockServer struct {
	UnimplementedStockServiceServer
	repo persistence.StockRepository
}

func NewStockServer(repo persistence.StockRepository) *StockServer {
	return &StockServer{repo: repo}
}

func (s *StockServer) GetStockData(ctx context.Context, req *StockRequest) (*StockResponse, error) {
	st, err := s.repo.GetByName(req.StockName)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Stock not found")
	}

	input := &application.CrawlerInput{
		StockName:     st.StockName,
		OriginalACode: st.AStockCode,
		OriginalHCode: st.HStockCode,
	}
	input.Normalize()

	stockData := &StockData{
		StockName:  input.StockName,
		AStockCode: input.OriginalACode,
		HStockCode: input.OriginalHCode,
	}

	return &StockResponse{
		Stocks: []*StockData{stockData},
	}, nil
}

func (s *StockServer) GetAllStockData(req *EmptyRequest, stream StockService_GetAllStockDataServer) error {
	stocks, err := s.repo.GetAll()
	if err != nil {
		return status.Error(codes.Internal, "Failed to fetch stocks")
	}

	for _, st := range stocks {
		input := &application.CrawlerInput{
			StockName:     st.StockName,
			OriginalACode: st.AStockCode,
			OriginalHCode: st.HStockCode,
		}
		input.Normalize()

		stockData := &StockData{
			StockName:  input.StockName,
			AStockCode: input.OriginalACode,
			HStockCode: input.OriginalHCode,
		}

		if err := stream.Send(stockData); err != nil {
			return err
		}
	}
	return nil
}
