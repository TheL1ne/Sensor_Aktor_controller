package main

import (
	"log"
	"net"

	"github.com/TheL1ne/Sensor_Aktor_controller/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	// controllers address
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// TODO: change probablity of failure here
	s, err := api.StartActor(0)
	if err != nil {
		zap.L().Fatal("could not start ActorServer", zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	api.RegisterActorServer(grpcServer, s)

	if err := grpcServer.Serve(lis); err != nil {
		zap.L().Fatal("failed to serve", zap.Error(err))
	}
}
