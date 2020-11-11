package api

import (
	"context"
	fmt "fmt"
)

type Controller struct {
	values []float64
	actor  *Actor
}

func StartController(actor *Actor) (*Controller, error) {
	if actor == nil {
		return nil, fmt.Errorf("Actor must be set")
	}
	return &Controller{
		values: []float64{},
		actor:  actor,
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
