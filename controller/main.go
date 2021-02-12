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
	logger.Info("Starting Controller")

	// controllers own address
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		logger.Fatal("Starting TCP-Listener failed", zap.Error(err))
	}

	// connection to actor
	conn, err := grpc.Dial("actor:8080", grpc.WithInsecure())
	if err != nil {
		logger.Fatal("could not dial to Actor", zap.Error(err))
	}
	defer conn.Close()
	actor := api.NewActorClient(conn)

	// connection to database
	dbconn, err := grpc.Dial("database:9090", grpc.WithInsecure())
	if err != nil {
		logger.Fatal("could not dial to database", zap.Error(err))
	}
	defer dbconn.Close()
	db := api.NewDatabaseClient(dbconn)

	controller, err := api.NewController(actor, db, logger)
	if err != nil {
		logger.Fatal("could not start ControllerServer", zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	api.RegisterControllerServer(grpcServer, controller)

	logger.Info("controller starts serving")
	if err = grpcServer.Serve(lis); err != nil {
		logger.Fatal("failed to serve", zap.Error(err))
	}
}
