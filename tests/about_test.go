package tests

import (
	"synk/gateway/app"
	"synk/gateway/app/model"
	"testing"
)

func TestAboutModelPing(t *testing.T) {
	db, err := app.InitDB(true)

	if err != nil {
		t.Errorf("about: db connection failed [%v]", err.Error())
	}

	aboutModel := model.NewAbout(db)

	if !aboutModel.Ping() {
		t.Errorf("about: ping operation failed")
	}

	db.Close()
}
