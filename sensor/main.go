package main

import (
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
)

func main() {
	// OS signals to wait for
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// dial controller
	conn, err := grpc.Dial(":9000", grpc.WithInsecure())
	if err != nil {
		logger.Fatal("could not dial to Actor", zap.Error(err))
	}
	defer conn.Close()
	controller := api.NewControllerClient(conn)

	// connection to database
	dbconn, err := grpc.Dial(":9090", grpc.WithInsecure())
	if err != nil {
		logger.Fatal("could not dial to database", zap.Error(err))
	}
	defer dbconn.Close()
	db := api.NewDatabaseClient(dbconn)

	sensor, err := api.NewSensor(intervall, controller, db, logger)
	done := sensor.StartSensor()
	defer close(done)
	logger.Info("started sensor, waiting for Signal...")
	// waiting for killing
	<-sigs
	logger.Info("exiting")
}
