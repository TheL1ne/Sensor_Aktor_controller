package api

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Actor struct {
	position     float64
	presentError *ErrorRequest
	database     DatabaseClient
	logger       *zap.Logger
}

func NewActor(dbC DatabaseClient, log *zap.Logger) (*Actor, error) {
	return &Actor{
		position: -1, // to mark not initialized position
		database: dbC,
		logger:   log,
	}, nil
}

func (a *Actor) UpdatePosition(ctx context.Context, req *UpdatePositionRequest) (*UpdatePositionResponse, error) {
	a.logger.Info("Updatepositionrequest received")
	if req == nil {
		err := a.saveEvent(ctx, DatabaseRequest_UpdatePositionRequest, time.Now().Unix(), true)
		if err != nil {
			a.logger.Error("could not save Event to database", zap.Error(err), zap.Any("Request", req))
		}
		return nil, fmt.Errorf("Positionupdate was empty")
	} else {
		err := a.saveEvent(ctx, DatabaseRequest_UpdatePositionRequest, req.GetTime(), false)
		if err != nil {
			a.logger.Error("could not save Event to database", zap.Error(err), zap.Any("Request", req))
		}
	}

	if a.isErrorPresent() {
		switch a.presentError.Type {
		// skipped 2 errors because:
		// can't flood from request
		case Error_missing_packet:
			return nil, nil
		case Error_late:
			time.Sleep(1 * time.Second)

		case Error_empty:
			return &UpdatePositionResponse{}, nil
		}
	}
	a.position = req.GetPosition()
	a.logger.Info("sending PositionResponse...")
	return &UpdatePositionResponse{
		ReachedPosition: req.GetPosition(),
	}, nil
}

func (a *Actor) GetPosition(context context.Context, req *Empty) (*GetPositionResponse, error) {
	a.logger.Info("GetPositionRequest received")
	wasEmpty := false
	if req == nil {
		wasEmpty = true
	}

	err := a.saveEvent(context, DatabaseRequest_Empty, time.Now().Unix(), wasEmpty)
	if err != nil {
		a.logger.Error("could not save Event to database", zap.Error(err), zap.Any("Request", req))
	}

	return &GetPositionResponse{
		Position: a.position,
	}, nil
}

func (a *Actor) SetError(ctx context.Context, req *ErrorRequest) (*Empty, error) {
	a.logger.Info("Errorrequest received")
	err := a.saveEvent(ctx, DatabaseRequest_ErrorRequest, req.GetTime(), false)
	if err != nil {
		a.logger.Error("could not save Event to database", zap.Error(err), zap.Any("Request", req))
	}
	a.presentError = req
	return &Empty{}, nil
}

func (a *Actor) isErrorPresent() bool {
	if a.presentError != nil && (time.Now().Unix() < a.presentError.Time+int64(a.presentError.Milliseconds)) {
		return true
	}
	// reset error state
	if a.presentError != nil && (time.Now().Unix() >= a.presentError.Time+int64(a.presentError.Milliseconds)) {
		a.presentError = nil
	}
	return false
}

func (a *Actor) saveEvent(ctx context.Context, Etype DatabaseRequest_EventType, time int64, wasEmpty bool) error {
	_, err := a.database.SaveEvent(ctx, &DatabaseRequest{
		Time:     time,
		Type:     Etype,
		WasEmpty: wasEmpty,
		Receiver: DatabaseRequest_actor,
	})
	return err
}
