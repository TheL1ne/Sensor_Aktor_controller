package main

import (
	"net"

	"github.com/TheL1ne/Sensor_Aktor_controller/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	// controllers address
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}
	logger.Info("starting actor serving on :8080")

	// connection to database
	dbconn, err := grpc.Dial(":9090", grpc.WithInsecure())
	if err != nil {
		zap.L().Fatal("could not dial to database", zap.Error(err))
	}
	defer dbconn.Close()
	db := api.NewDatabaseClient(dbconn)

	s, err := api.StartActor(db)
	if err != nil {
		logger.Fatal("could not start ActorServer", zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	api.RegisterActorServer(grpcServer, s)

	logger.Info("actor started...")

	if err := grpcServer.Serve(lis); err != nil {
		zap.L().Fatal("failed to serve", zap.Error(err))
	}
}
