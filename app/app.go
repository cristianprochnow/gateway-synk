package app

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"synk/gateway/app/util"

	"github.com/getsentry/sentry-go"
)

type Service struct {
	DB *sql.DB
}

func InitSentry() error {
	dsn := os.Getenv("SENTRY_DSN")

	if dsn == "" {
		return errors.New("sentry DSN config is missing")
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		TracesSampleRate: 1.0,
	})

	if err != nil {
		return errors.New("sentry setup failed: " + err.Error())
	}

	return nil
}

func Run() {
	util.Log("starting app")

	db, dbErr := InitDB(false)

	if dbErr != nil {
		log.Fatal(dbErr)
	}

	sentryErr := InitSentry()

	if sentryErr != nil {
		log.Fatal(sentryErr)
	}

	Router(&Service{DB: db})
}
