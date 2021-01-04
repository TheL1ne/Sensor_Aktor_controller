package api

import (
	"context"
	fmt "fmt"
)

type Controller struct {
	values             []float64
	actor              *ActorClient
	presentError       *ErrorRequest
	lastErrorTriggered bool
	notifyErrorTrigger chan interface{}
}

func StartController(actor *ActorClient) (*Controller, error) {
	if actor == nil {
		return nil, fmt.Errorf("Actor must be set")
	}
	return &Controller{
		values:             []float64{},
		actor:              actor,
		presentError:       nil,
		lastErrorTriggered: false,
	}, nil
}

func (c *Controller) UpdateMeasurement(ctx context.Context, mes *Measurement) (*Empty, error) {
	if len(c.values) == 0 {
		c.values = []float64{mes.GetValue()}
	} else {
		c.values = append(c.values, mes.GetValue())
	}
	return &Empty{}, nil
}

func (c *Controller) GetHistory(ctx context.Context, req *GetHistoryRequest) (*GetHistoryResponse, error) {
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
