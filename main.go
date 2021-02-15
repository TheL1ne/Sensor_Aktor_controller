package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/TheL1ne/Sensor_Aktor_controller/actor"
	"github.com/TheL1ne/Sensor_Aktor_controller/controller"
	"github.com/TheL1ne/Sensor_Aktor_controller/database"
	"github.com/TheL1ne/Sensor_Aktor_controller/sensor"
)

func main() {
	// OS signals to wait for
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go database.Start()
	go actor.Start()
	go controller.Start()
	go sensor.Start()
	log.Println("all started...")
	<-sigs
}
