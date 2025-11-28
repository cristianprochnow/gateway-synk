package tests

import (
	"synk/gateway/app"
	"synk/gateway/app/model"
	"testing"
)

func TestColorsList(t *testing.T) {
	db, err := app.InitDB(true)
	if err != nil {
		t.Fatalf("colors: db connection failed [%v]", err.Error())
	}
	defer db.Close()

	colorsModel := model.NewColors(db)

	allColors, err := colorsModel.List(0)
	if err != nil {
		t.Errorf("colors: list all operation failed: %v", err)
	}

	if len(allColors) > 0 {
		targetColor := allColors[0]

		specificList, err := colorsModel.List(targetColor.ColorId)

		if err != nil {
			t.Errorf("colors: list specific id failed: %v", err)
		}

		if len(specificList) != 1 {
			t.Errorf("colors: expected 1 result for specific id, got %d", len(specificList))
		}

		if specificList[0].ColorId != targetColor.ColorId {
			t.Errorf("colors: returned ID mismatch. Expected %d, got %d", targetColor.ColorId, specificList[0].ColorId)
		}

		if specificList[0].ColorName == "" || specificList[0].ColorHex == "" {
			t.Errorf("colors: returned empty name or hex for id %d", targetColor.ColorId)
		}

	} else {
		t.Log("colors: warning - 'color' table is empty, skipping specific ID verification")
	}
}

func TestColorsListNotFound(t *testing.T) {
	db, err := app.InitDB(true)
	if err != nil {
		t.Fatalf("colors: db connection failed [%v]", err.Error())
	}
	defer db.Close()

	colorsModel := model.NewColors(db)

	list, err := colorsModel.List(-1)

	if err != nil {
		t.Errorf("colors: list should not error on non-existent ID, got: %v", err)
	}

	if len(list) != 0 {
		t.Errorf("colors: expected empty list for invalid ID, got %d items", len(list))
	}
}
