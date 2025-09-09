package app

import (
	"database/sql"
	"log"
	"synk/gateway/app/util"
)

type Service struct {
	DB *sql.DB
}

func Run() {
	util.Log("starting app")

	db, dbErr := InitDB(false)

	if dbErr != nil {
		log.Fatal(dbErr)
	}

	Router(&Service{DB: db})
}
