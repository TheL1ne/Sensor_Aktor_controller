package api

import (
	"context"
	fmt "fmt"
	"time"
)

type Controller struct {
	values             []float64
	actor              ActorClient
	presentError       *ErrorRequest
	lastErrorTriggered bool
	notifyErrorTrigger chan interface{}
}

func StartController(actor ActorClient) (*Controller, error) {
	if actor == nil {
		return nil, fmt.Errorf("Actor must be set")
	}
	c := Controller{
		values:             []float64{},
		actor:              actor,
		presentError:       nil,
		lastErrorTriggered: false,
	}

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// running for flooding occurences
		if c.presentError != nil && c.presentError.Type == Error_flood {
			for c.isErrorPresent() {
				c.actor.UpdatePosition(ctx, &UpdatePositionRequest{
					Position: 3.14159,
				})
			}
		}
		time.Sleep(time.Millisecond)
	}()

	return &c, nil
}

func (c *Controller) UpdateMeasurement(ctx context.Context, mes *Measurement) (*Empty, error) {
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
		c.actor.UpdatePosition(ctx, &UpdatePositionRequest{
			Position: 3.14159,
		})
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
	// reduce history to last 100 values
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
	if c.lastErrorTriggered {
		c.presentError = req
	} else {
		// wait for the previous error to be triggerd at least once
		<-c.notifyErrorTrigger
		c.presentError = req
	}
	return &Empty{}, nil
}

func (c *Controller) isErrorPresent() bool {
	if c.presentError != nil && (time.Now().Unix() < c.presentError.Time+int64(c.presentError.Milliseconds) || c.lastErrorTriggered == false) {
		return true
	}
	// reset error state
	if c.presentError != nil && (time.Now().Unix() >= c.presentError.Time+int64(c.presentError.Milliseconds) && c.lastErrorTriggered == true) {
		c.presentError = nil
	}
	return false
}
