package app

import (
	"database/sql"
	"os"
	"synk/gateway/app/util"

	"github.com/go-sql-driver/mysql"
)

func InitDB(testing bool) (*sql.DB, error) {
	cfg := mysql.NewConfig()

	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	address := os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT")

	cfg.User = user
	cfg.Passwd = pass
	cfg.Net = "tcp"
	cfg.Addr = address
	cfg.DBName = "synk"

	util.Log("connecting do database")

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		util.Log("error when connecting on db: " + err.Error())

		return nil, err
	}

	pingErr := db.Ping()
	if pingErr != nil {
		util.Log("error when ping on db: " + pingErr.Error())

		return nil, pingErr
	}

	return db, nil
}
