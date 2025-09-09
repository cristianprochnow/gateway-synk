package model

import (
	"database/sql"
)

type About struct {
	db *sql.DB
}

func NewAbout(db *sql.DB) *About {
	about := About{db: db}

	return &about
}

func (a *About) Ping() bool {
	pingErr := a.db.Ping()

	return pingErr == nil
}
