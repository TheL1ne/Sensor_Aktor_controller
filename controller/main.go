package main

import (
	"net"

	"github.com/TheL1ne/Sensor_Aktor_controller/api"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"google.golang.org/grpc"
)

var (
	actorAddress = pflag.String("actor-address", "localhost:8080", "The address for reaching the actor service")
)

func main() {
	pflag.Parse()

	zap.L().Info("Starting Controller")

	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		zap.L().Fatal("Starting TCP-Listener failed", zap.Error(err))
	}

	conn, err := grpc.Dial(":8080", grpc.WithInsecure())
	if err != nil {
		zap.L().Fatal("could not dial to Actor", zap.Error(err))
	}
	defer conn.Close()

	actor := api.NewActorClient(conn)

	controller, err := api.StartController(&actor)
	if err != nil {
		zap.L().Fatal("could not start ControllerServer", zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	api.RegisterControllerServer(grpcServer, controller)

	zap.L().Info("controller starts serving")
	if err = grpcServer.Serve(lis); err != nil {
		zap.L().Fatal("failed to serve", zap.Error(err))
	}
}
