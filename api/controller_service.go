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
	counter      int
}

func NewController(actor ActorClient, dbClient DatabaseClient, log *zap.Logger) (*Controller, error) {
	if actor == nil {
		return nil, fmt.Errorf("Actor must be set")
	}
	c := Controller{
		values:       []float64{},
		actor:        actor,
		database:     dbClient,
		presentError: nil,
		logger:       log,
		counter:      0,
	}

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// running for flooding occurences
		for {
			if c.presentError != nil && c.presentError.Type == Error_flood {
				c.logger.Info("start flooding")
				for c.isErrorPresent() {
					_, err := c.actor.UpdatePosition(ctx, &UpdatePositionRequest{
						Position: 3.14159,
					})
					if err != nil {
						c.logger.Error("could not Update Position", zap.Error(err))
					}
				}
			}
			time.Sleep(time.Millisecond)
		}
	}()

	return &c, nil
}

func (c *Controller) UpdateMeasurement(ctx context.Context, mes *Measurement) (*Empty, error) {
	c.logger.Info("UpdateMeasurement received")
	// early return because of empty update
	if mes == nil {
		return nil, fmt.Errorf("measurement update was empty")
	}
	if len(c.values) == 0 {
		c.values = []float64{mes.GetValue()}
	} else {
		c.values = append(c.values, mes.GetValue())
	}
	// counter how many measurements lead to actor update
	if c.counter%10 == 0 {
		if c.isErrorPresent() {
			switch c.presentError.Type {
			case Error_missing_packet:
				return nil, nil
			case Error_empty:
				_, err := c.actor.UpdatePosition(ctx, &UpdatePositionRequest{})
				if err != nil {
					c.logger.Error("error after empty position update", zap.Error(err))
				}
			case Error_late:
				time.Sleep(time.Second)
			}
		}
		_, err := c.actor.UpdatePosition(ctx, &UpdatePositionRequest{
			Position: 3.14159,
		})
		if err != nil {
			c.logger.Error("could not Update actor position", zap.Error(err))
		}
	}
	c.counter++
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
	c.logger.Info("GetHistory received")
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
	c.logger.Info("SetError received")
	c.saveAnomaly(ctx, req)
	c.presentError = req
	return &Empty{}, nil
}

func (c *Controller) isErrorPresent() bool {
	if c.presentError != nil && (c.presentError.Time <= 0 || time.Now().Unix() <= c.presentError.Time+int64(c.presentError.Milliseconds/1000)) {
		return true
	}
	// reset error state
	if c.presentError != nil && (time.Now().Unix() > c.presentError.Time+int64(c.presentError.Milliseconds/1000)) {
		c.presentError = nil
	}
	return false
}

func (c *Controller) saveAnomaly(ctx context.Context, req *ErrorRequest) error {
	_, err := c.database.SaveAnomaly(ctx, &DatabaseRequest{
		Time:         req.Time,
		Type:         req.Type,
		Receiver:     DatabaseRequest_controller,
		Milliseconds: int64(req.Milliseconds),
	})
	return err
}
