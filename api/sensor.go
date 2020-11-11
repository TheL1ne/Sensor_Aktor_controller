package api

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// represents a Sensor which has an Interval of milliseconds until new update
// and a failing probability 0 to 100 percent
type Sensor struct {
	intervall       int
	failProbability int
	controller      ControllerClient
}

func NewSensor(intervall int, failProbability int, con ControllerClient) (*Sensor, error) {
	if failProbability < 0 || failProbability > 100 {
		return nil, fmt.Errorf("intervall must be between 0 and 100 but was %d", failProbability)
	}
	if intervall < 0 {
		return nil, fmt.Errorf("intervall must be positiv but was %d", intervall)
	}
	if con == nil {
		return nil, fmt.Errorf("controller was not set")
	}
	return &Sensor{
		failProbability: failProbability,
		intervall:       intervall,
		controller:      con,
	}, nil
}

func (s *Sensor) StartSensor() chan bool {
	done := make(chan bool)
	ticker := time.NewTicker(time.Duration(s.intervall) * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()
		for {
			select {
			// kill loop to not leave zombie routines
			case <-done:
				return
			case <-ticker.C:
				rand.Seed(time.Now().UnixNano())
				// range for random Number is 0 - 100
				// check if "failed" attempt
				if s.failProbability >= rand.Intn(100) {
					// skip report when failed
					continue
				}
				_, err := s.controller.UpdateMeasurement(ctx, &Measurement{
					// TODO: switch temperature occasionally
					Value: 49.5,
					Time:  time.Now().Unix(),
					Unit:  "°C",
				})
				if err != nil {
					zap.L().Error("Update to controller failed", zap.Error(err))
				}
			}
		}
	}()

	return done
}
