package api

import (
	"context"
	fmt "fmt"
	"time"

	"go.uber.org/zap"
)

type Controller struct {
	values       []float64
	actor        ActorClient
	database     DatabaseClient
	presentError *ErrorRequest
	logger       *zap.Logger
}

func StartController(actor ActorClient, dbClient DatabaseClient, log *zap.Logger) (*Controller, error) {
	if actor == nil {
		return nil, fmt.Errorf("Actor must be set")
	}
	c := Controller{
		values:       []float64{},
		actor:        actor,
		database:     dbClient,
		presentError: nil,
		logger:       log,
	}

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// running for flooding occurences
		for {
			if c.presentError != nil && c.presentError.Type == Error_flood {
				for c.isErrorPresent() {
					resp, err := c.actor.UpdatePosition(ctx, &UpdatePositionRequest{
						Position: 3.14159,
					})
					if err != nil {
						c.logger.Error("could not Update Position", zap.Error(err))
					} else {
						wasEmpty := false
						if resp.GetTime() == 0 {
							wasEmpty = true
						}
						err := c.saveEvent(ctx, DatabaseRequest_UpdatePositionResponse, time.Now().Unix(), wasEmpty)
						if err != nil {
							c.logger.Error("could not save UpdatePositionResponse to DB", zap.Error(err))
						}
					}
				}
			}
			time.Sleep(time.Millisecond)
		}
	}()

	return &c, nil
}

func (c *Controller) UpdateMeasurement(ctx context.Context, mes *Measurement) (*Empty, error) {
	// early return because of empty update
	if mes == nil {
		err := c.saveEvent(ctx, DatabaseRequest_measurement, time.Now().Unix(), true)
		if err != nil {
			c.logger.Error("could not save Measurementupdate", zap.Error(err))
		}
		return nil, fmt.Errorf("measurement update was empty")
	}
	// save Event
	wasEmpty := false
	if mes.GetTime() == 0 {
		wasEmpty = true // struct is present but not initialized
	}
	err := c.saveEvent(ctx, DatabaseRequest_measurement, time.Now().Unix(), wasEmpty)
	if err != nil {
		c.logger.Error("could not save Measurementupdate", zap.Error(err))
	}
	if len(c.values) == 0 {
		c.values = []float64{mes.GetValue()}
	} else {
		c.values = append(c.values, mes.GetValue())
	}
	// counter how many measurements lead to actor update
	i := 0
	if i%10 == 0 {
		i = 0
		if c.isErrorPresent() {
			switch c.presentError.Type {
			case Error_missing_packet:
				c.actor.UpdatePosition(ctx, nil)
			case Error_empty:
				c.actor.UpdatePosition(ctx, &UpdatePositionRequest{})
			case Error_late:
				time.Sleep(time.Second)
			}
		}
		resp, err := c.actor.UpdatePosition(ctx, &UpdatePositionRequest{
			Position: 3.14159,
		})
		if err != nil {
			c.logger.Error("could not Update actor position", zap.Error(err))
		}
		wasEmpty = false
		if resp.GetTime() == 0 {
			wasEmpty = true
		}
		c.saveEvent(ctx, DatabaseRequest_UpdatePositionResponse, time.Now().Unix(), false)
	} else {
		i++
	}
	if c.isErrorPresent() {
		switch c.presentError.Type {
		case Error_missing_packet:
			return nil, nil
		case Error_empty:
			// just a normal return as we always send an empty message here
		case Error_late:
			time.Sleep(time.Second)
		}
	}
	return &Empty{}, nil
}

func (c *Controller) GetHistory(ctx context.Context, req *GetHistoryRequest) (*GetHistoryResponse, error) {
	// save Event
	err := c.saveEvent(ctx, DatabaseRequest_historyRequest, time.Now().Unix(), false)
	if err != nil {
		c.logger.Error("could not save GetHistoryRequest", zap.Error(err))
	}

	// reduce history to last 1000 values
	historyLength := 1000
	if len(c.values) > historyLength {
		c.values = c.values[len(c.values)-historyLength:]
	}
	if c.isErrorPresent() {
		switch c.presentError.Type {
		case Error_missing_packet:
			return nil, nil
		case Error_late:
			// sleep a second which should be way later then normally
			time.Sleep(time.Second)
		case Error_empty:
			return &GetHistoryResponse{}, nil
		}
	}
	return &GetHistoryResponse{
		Values: c.values,
	}, nil
}

func (c *Controller) SetError(ctx context.Context, req *ErrorRequest) (*Empty, error) {
	c.presentError = req
	return &Empty{}, nil
}

func (c *Controller) isErrorPresent() bool {
	if c.presentError != nil && (time.Now().Unix() < c.presentError.Time+int64(c.presentError.Milliseconds)) {
		return true
	}
	// reset error state
	if c.presentError != nil && (time.Now().Unix() >= c.presentError.Time+int64(c.presentError.Milliseconds)) {
		c.presentError = nil
	}
	return false
}

func (c *Controller) saveEvent(ctx context.Context, Etype DatabaseRequest_EventType, time int64, wasEmpty bool) error {
	_, err := c.database.SaveEvent(ctx, &DatabaseRequest{
		Time:     time,
		Type:     Etype,
		WasEmpty: wasEmpty,
		Receiver: DatabaseRequest_actor,
	})
	return err
}
