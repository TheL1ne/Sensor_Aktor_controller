package sensor

import (
	"net"

	"github.com/TheL1ne/Sensor_Aktor_controller/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	// intervall of reports in Milliseconds
	intervall = int64(500)
	port      = ":9020"
)

func Start() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// sensors address
	lis, err := net.Listen("tcp4", port)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}
	logger.Info("starting sensor serving on", zap.String("Port", port))

	// dial controller
	conn, err := grpc.Dial("127.0.0.1:9010", grpc.WithInsecure())
	if err != nil {
		logger.Fatal("could not dial to Actor", zap.Error(err))
	}
	defer conn.Close()
	controller := api.NewControllerClient(conn)

	// connection to database
	dbconn, err := grpc.Dial("127.0.0.1:9090", grpc.WithInsecure())
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
