package api

import (
	"context"
	fmt "fmt"
	"math/rand"
	"time"
)

type Actor struct {
	position        float64
	failProbability int // expected to be 0-100 as percentage
}

func StartActor(failProbability int) (*Actor, error) {
	if failProbability > 100 || failProbability < 0 {
		return nil, fmt.Errorf("failProbability must be between 0 and 100 but was %d", failProbability)
	}
	return &Actor{
		position:        -1, // to mark not initialized position
		failProbability: failProbability,
	}, nil
}

func (a *Actor) UpdatePosition(ctx context.Context, req *UpdatePositionRequest) (*UpdatePositionResponse, error) {
	rand.Seed(time.Now().UnixNano())
	// range for random Number is 0 - 100
	// check if "failed" attempt
	if a.failProbability >= rand.Intn(100) {
		return nil, fmt.Errorf("i feel like failing")
	} else {
		a.position = req.GetPosition()
		return &UpdatePositionResponse{
			ReachedPosition: req.GetPosition(),
		}, nil
	}
}

func (a *Actor) GetPosition(context context.Context, req *Empty) (*GetPositionResponse, error) {
	rand.Seed(time.Now().UnixNano())
	// range for random Number is 0 - 100
	// check if "failed" attempt
	if a.failProbability >= rand.Intn(100) {
		return nil, fmt.Errorf("i feel like failing")
	} else {
		return &GetPositionResponse{
			Position: a.position,
		}, nil
	}
}
