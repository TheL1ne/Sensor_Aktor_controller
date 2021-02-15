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
	database, err := sql.Open("sqlite3", "./anomaly.db")
	if err != nil {
		log.Error("could not open database", zap.Error(err))
	}

	statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS anomalies(id INTEGER PRIMARY KEY, time INTEGER, type TEXT, receiver STRING, milliseconds INTEGER)")
	if err != nil {
		log.Error("could not prepare table creation", zap.Error(err))
	}
	_, err = statement.Exec()
	if err != nil {
		log.Error("could not execute table creation", zap.Error(err))
	}

	statement, err = database.Prepare("INSERT INTO anomalies (time, type, receiver, milliseconds) VALUES (?, ?, ?, ?)")
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

func (db *Database) SaveAnomaly(ctx context.Context, dbReq *DatabaseRequest) (*Empty, error) {
	db.logger.Info("anomaly Received", zap.String("Type", dbReq.Type.String()), zap.String("receiver", dbReq.Receiver.String()))
	_, err := db.insertStatement.Exec(dbReq.GetTime(), Error_name[int32(dbReq.GetType())], dbReq.GetReceiver(), dbReq.GetMilliseconds())
	if err != nil {
		return nil, err
	}
	return &Empty{}, nil
}
