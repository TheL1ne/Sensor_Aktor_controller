package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/TheL1ne/Sensor_Aktor_controller/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	// OS signals to wait for
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("starting DB")

	// databases own address
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	db := api.NewDB(logger)

	grpcServer := grpc.NewServer()
	api.RegisterDatabaseServer(grpcServer, db)

	logger.Info("DB is serving...")

	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("failed to serve", zap.Error(err))
	}
}
