package api

import (
	"context"
)

type Actor struct {
	position           float64
	presentError       *ErrorRequest
	lastErrorTriggered bool
	notifyErrorTrigger chan interface{}
}

func StartActor() (*Actor, error) {
	return &Actor{
		position: -1, // to mark not initialized position
	}, nil
}

func (a *Actor) UpdatePosition(ctx context.Context, req *UpdatePositionRequest) (*UpdatePositionResponse, error) {
	a.position = req.GetPosition()
	return &UpdatePositionResponse{
		ReachedPosition: req.GetPosition(),
	}, nil
}

func (a *Actor) GetPosition(context context.Context, req *Empty) (*GetPositionResponse, error) {
	return &GetPositionResponse{
		Position: a.position,
	}, nil
}

func (a *Actor) SetError(ctx context.Context, req *ErrorRequest) (*Empty, error) {
	if a.lastErrorTriggered {
		a.presentError = req
	} else {
		// wait for the previous error to be triggerd at least once
		<-a.notifyErrorTrigger
		a.presentError = req
	}
	return &Empty{}, nil
}
