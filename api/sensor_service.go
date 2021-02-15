package api

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// represents a Sensor which has an Interval of milliseconds until new update
type Sensor struct {
	intervall    int64
	controller   ControllerClient
	presentError *ErrorRequest
	database     DatabaseClient
	logger       *zap.Logger
}

func NewSensor(intervall int64, con ControllerClient, dbClient DatabaseClient, logger *zap.Logger) (*Sensor, error) {
	if intervall < 0 {
		return nil, fmt.Errorf("intervall must be positiv but was %d", intervall)
	}
	if con == nil {
		return nil, fmt.Errorf("controller was not set")
	}
	return &Sensor{
		intervall:    intervall,
		controller:   con,
		presentError: nil,
		database:     dbClient,
		logger:       logger,
	}, nil
}

func (s *Sensor) StartSensor() chan bool {
	s.logger.Info("Sensor Started")
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
	s.logger.Info("SetError received")
	err := s.saveEvent(ctx, req)
	if err != nil {
		s.logger.Error("could not save ErrorEvent", zap.Error(err))
	}
	s.presentError = req
	return &Empty{}, nil
}

func (s *Sensor) communicate(ctx context.Context) {
	if s.isErrorPresent() {
		s.logger.Info("errorous comminucation", zap.String("Error", s.presentError.String()))
		switch s.presentError.Type {
		case Error_missing_packet:
			return
		case Error_late:
			time.Sleep(time.Duration(s.intervall*10) * time.Millisecond) // sleep double the typical sending interval
			// here sending normally afterwards -> no return
		case Error_empty:
			// skip as nil marshalling does not work
			s.logger.Warn("Skipping empty packets as nil marshalling does not work")
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
					s.logger.Error("flooding event to controller failed", zap.Error(err))
				}
			}
		}
	}
	// normal sending from here on
	_, err := s.controller.UpdateMeasurement(ctx, &Measurement{
		Value: 49.5,
		Time:  time.Now().Unix(),
		Unit:  Unit_degree_celsius,
	})
	if err != nil {
		s.logger.Error("Update to controller failed", zap.Error(err))
	}
}

func (s *Sensor) isErrorPresent() bool {
	if s.presentError != nil && (s.presentError.Time <= 0 || time.Now().Unix() <= s.presentError.Time+int64(s.presentError.Milliseconds/1000)) {
		return true
	}
	// reset error state
	if s.presentError != nil && (time.Now().Unix() > s.presentError.Time+int64(s.presentError.Milliseconds/1000)) {
		s.presentError = nil
	}
	return false
}

func (s *Sensor) saveEvent(ctx context.Context, req *ErrorRequest) error {
	_, err := s.database.SaveAnomaly(ctx, &DatabaseRequest{
		Time:         req.Time,
		Type:         req.Type,
		Receiver:     DatabaseRequest_sensor,
		Milliseconds: int64(req.Milliseconds),
	})
	return err
}
