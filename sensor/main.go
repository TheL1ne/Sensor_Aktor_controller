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

var (
	// intervall of reports in Milliseconds
	intervall = int64(500)
	// in percent
	failProbability = 0
	port            = ":8081"
)

func main() {
	// OS signals to wait for
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// sensors address
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}
	logger.Info("starting sensor serving on", zap.String("Port", port))

	// dial controller
	conn, err := grpc.Dial("controller:9000", grpc.WithInsecure())
	if err != nil {
		logger.Fatal("could not dial to Actor", zap.Error(err))
	}
	defer conn.Close()
	controller := api.NewControllerClient(conn)

	// connection to database
	dbconn, err := grpc.Dial("database:9090", grpc.WithInsecure())
	if err != nil {
		logger.Fatal("could not dial to database", zap.Error(err))
	}
	defer dbconn.Close()
	db := api.NewDatabaseClient(dbconn)

	sensor, err := api.NewSensor(intervall, controller, db, logger)
	done := sensor.StartSensor()
	defer close(done)

	grpcServer := grpc.NewServer()
	api.RegisterSensorServer(grpcServer, sensor)

	logger.Info("started sensor, waiting for Signal...")
	// serving
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("failed to serve", zap.Error(err))
	}
}
