package api

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// represents a Sensor which has an Interval of milliseconds until new update
type Sensor struct {
	intervall          int64
	controller         ControllerClient
	presentError       *ErrorRequest
	lastErrorTriggered bool
	notifyErrorTrigger chan interface{}
}

func NewSensor(intervall int64, con ControllerClient) (*Sensor, error) {
	if intervall < 0 {
		return nil, fmt.Errorf("intervall must be positiv but was %d", intervall)
	}
	if con == nil {
		return nil, fmt.Errorf("controller was not set")
	}
	return &Sensor{
		intervall:          intervall,
		controller:         con,
		presentError:       nil,
		lastErrorTriggered: false,
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
				s.communicate(ctx)
			}
		}
	}()

	return done
}

func (s *Sensor) SetError(ctx context.Context, req *ErrorRequest) (*Empty, error) {
	if s.lastErrorTriggered {
		s.presentError = req
	} else {
		// wait for the previous error to be triggerd at least once
		<-s.notifyErrorTrigger
		s.presentError = req
	}
	return &Empty{}, nil
}

func (s *Sensor) communicate(ctx context.Context) {
	if s.isErrorPresent() {
		switch s.presentError.Type {
		case Error_missing_packet:
			s.errorWasTriggered()
			return
		case Error_late:
			time.Sleep(time.Duration(s.intervall*2) * time.Millisecond) // sleep double the typical sending interval
			s.errorWasTriggered()
			// here sending normally afterwards -> no return
		case Error_empty:
			// send empty package
			_, err := s.controller.UpdateMeasurement(ctx, nil)
			if err != nil {
				zap.L().Error("Nil Update to controller failed", zap.Error(err))
			}
			s.errorWasTriggered()
			return
		case Error_flood:
			// send everything continuously to all connected devices
			for s.isErrorPresent() {
				_, err := s.controller.UpdateMeasurement(ctx, &Measurement{
					Value: 49.5,
					Time:  time.Now().Unix(),
					Unit:  Unit_degree_celsius,
				})
				if err != nil {
					zap.L().Error("Nil Update to controller failed", zap.Error(err))
				}
			}
			s.errorWasTriggered()
		}
	}
	// normal sending from here on
	_, err := s.controller.UpdateMeasurement(ctx, &Measurement{
		Value: 49.5,
		Time:  time.Now().Unix(),
		Unit:  Unit_degree_celsius,
	})
	if err != nil {
		zap.L().Error("Update to controller failed", zap.Error(err))
	}
}

func (s *Sensor) isErrorPresent() bool {
	if s.presentError != nil && (time.Now().Unix() < s.presentError.Time+int64(s.presentError.Milliseconds) || s.lastErrorTriggered == false) {
		return true
	}
	// reset error state
	if s.presentError != nil && (time.Now().Unix() >= s.presentError.Time+int64(s.presentError.Milliseconds) && s.lastErrorTriggered == true) {
		s.presentError = nil
	}
	return false
}

func (s *Sensor) errorWasTriggered() {
	s.lastErrorTriggered = true
}
