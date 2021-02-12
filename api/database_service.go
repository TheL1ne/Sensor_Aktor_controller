package api

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type Database struct {
	db              *sql.DB
	insertStatement *sql.Stmt
	logger          *zap.Logger
}

func NewDB(log *zap.Logger) *Database {
	database, err := sql.Open("sqlite3", "./events.db")
	if err != nil {
		log.Error("could not open database", zap.Error(err))
	}

	statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS events(id INTEGER PRIMARY KEY, time INTEGER, type TEXT, wasEmpty BOOL)")
	if err != nil {
		log.Error("could not prepare table creation", zap.Error(err))
	}
	_, err = statement.Exec()
	if err != nil {
		log.Error("could not execute table creation", zap.Error(err))
	}

	statement, err = database.Prepare("INSERT INTO events (time, type, wasEmpty) VALUES (?, ?, ?)")
	if err != nil {
		log.Error("could not prepare insert statement", zap.Error(err))
	}
	log.Info("DB started")
	return &Database{
		db:              database,
		insertStatement: statement,
		logger:          log,
	}
}

func (db *Database) SaveEvent(ctx context.Context, dbReq *DatabaseRequest) (*Empty, error) {
	db.logger.Info("event Received", zap.Any("Request", dbReq.String()))
	_, err := db.insertStatement.Exec(dbReq.GetTime(), DatabaseRequest_EventType_name[int32(dbReq.Type)], dbReq.GetWasEmpty())
	if err != nil {
		return nil, err
	}
	return &Empty{}, nil
}
