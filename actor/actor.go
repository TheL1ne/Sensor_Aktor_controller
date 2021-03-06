package actor

import (
	"net"

	"github.com/TheL1ne/Sensor_Aktor_controller/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	port = ":9000"
)

func Start() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	// actors address
	lis, err := net.Listen("tcp4", port)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}
	logger.Info("starting actor serving on", zap.String("Port", port))

	// connection to database
	dbconn, err := grpc.Dial("127.0.0.1:9090", grpc.WithInsecure())
	if err != nil {
		logger.Fatal("could not dial to database", zap.Error(err))
	}
	defer dbconn.Close()
	db := api.NewDatabaseClient(dbconn)

	a, err := api.NewActor(db, logger)
	if err != nil {
		logger.Fatal("could not start ActorServer", zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	api.RegisterActorServer(grpcServer, a)

	logger.Info("actor started...")

	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("failed to serve", zap.Error(err))
	}
}
