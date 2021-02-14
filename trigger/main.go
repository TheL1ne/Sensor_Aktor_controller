package main

import (
	"context"
	"flag"
	"time"

	"github.com/TheL1ne/Sensor_Aktor_controller/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	adress         = "127.0.0.1"
	actorPort      = ":8080"
	controllerPort = ":9000"
	sensorPort     = ":8081"

	actorTarget      = "actor"
	controllerTarget = "controller"
	sensorTarget     = "sensor"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	var target string
	var behaviour string
	var duration int

	// flags declaration
	flag.StringVar(&target, "t", "", "Specify target to trigger anomerla behaviour")
	flag.StringVar(&behaviour, "b", "", "Specify anormal behaviour to trigger")
	flag.IntVar(&duration, "d", 0, "Specify duration of errorous behaviour in millisec.")
	flag.Parse()

	if target == "" {
		logger.Fatal("target must be set")
		return
	}
	if behaviour == "" {
		logger.Fatal("bahviour must be set")
		return
	}
	if duration == 0 {
		logger.Fatal("duration must be set")
		return
	}

	var client api.ManipulatableServiceClient

	// parse Target and set Client
	targetAdress := adress
	switch target {
	case actorTarget:
		targetAdress += actorPort
		conn, err := grpc.Dial(targetAdress, grpc.WithInsecure())
		if err != nil {
			logger.Fatal("could not dial to Actor", zap.Error(err))
		}
		defer conn.Close()
		client = api.NewActorClient(conn)
	case controllerTarget:
		targetAdress += controllerPort
		conn, err := grpc.Dial(targetAdress, grpc.WithInsecure())
		if err != nil {
			logger.Fatal("could not dial to Actor", zap.Error(err))
		}
		defer conn.Close()
		client = api.NewControllerClient(conn)
	case sensorTarget:
		targetAdress += sensorPort
		conn, err := grpc.Dial(targetAdress, grpc.WithInsecure())
		if err != nil {
			logger.Fatal("could not dial to Actor", zap.Error(err))
		}
		defer conn.Close()
		client = api.NewSensorClient(conn)
	default:
		logger.Fatal("invalid target", zap.String("Got", target))
		return
	}

	errRequest := api.ErrorRequest{
		Milliseconds: int32(duration),
	}
	// parse behaviour
	switch behaviour {
	case api.Error_name[int32(api.Error_empty)]:
		errRequest.Type = api.Error_empty
	case api.Error_name[int32(api.Error_flood)]:
		errRequest.Type = api.Error_flood
	case api.Error_name[int32(api.Error_late)]:
		errRequest.Type = api.Error_late
	case api.Error_name[int32(api.Error_missing_packet)]:
		errRequest.Type = api.Error_missing_packet
	}
	errRequest.Time = time.Now().Unix()

	ctx := context.Background()
	_, err := client.SetError(ctx, &errRequest)
	if err != nil {
		logger.Fatal("could not set anomalous behaviour", zap.Error(err))
	}
}
