package main

import (
	"D_come/internal/config"
	"D_come/internal/infrastructure/grpc"
	"D_come/internal/infrastructure/persistence"
	"fmt"
	"log"
	"net"

	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	db, err := persistence.NewDatabase(&cfg.Database)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	stockRepo := persistence.NewStockRepository(db.DB)

	grpcServer := googlegrpc.NewServer()

	stockServer := grpc.NewStockServer(stockRepo)
	grpc.RegisterStockServiceServer(grpcServer, stockServer)

	reflection.Register(grpcServer)

	port := cfg.Server.Port
	if port == 0 {
		port = 50051
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("无法监听端口: %v", err)
	}

	log.Printf("gRPC服务器正在监听端口 %d", port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("服务器运行失败: %v", err)
	}
}
