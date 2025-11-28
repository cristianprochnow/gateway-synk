package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"synk/gateway/app"
	"synk/gateway/app/controller"
	"testing"
)

func TestAbout_HandleAbout(t *testing.T) {

	db, err := app.InitDB(true)
	if err != nil {
		t.Fatalf("db connection failed: %v", err)
	}
	defer db.Close()

	aboutController := controller.NewAbout(db)

	originalPort := os.Getenv("PORT")
	os.Setenv("PORT", "9999")
	defer os.Setenv("PORT", originalPort)

	req, _ := http.NewRequest("GET", "/about", nil)
	rr := httptest.NewRecorder()

	aboutController.HandleAbout(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	type AboutInfo struct {
		ServerPort string `json:"server_port"`
		AppPort    string `json:"app_port"`
		DbWorking  bool   `json:"db_working"`
	}
	type AboutResponse struct {
		Ok   bool      `json:"ok"`
		Info AboutInfo `json:"info"`
	}

	var response AboutResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if !response.Ok {
		t.Error("expected 'ok' to be true")
	}

	if !response.Info.DbWorking {
		t.Error("expected 'db_working' to be true (since we connected to a real DB)")
	}

	if response.Info.AppPort != "9999" {
		t.Errorf("expected 'app_port' to be '9999', got '%s'", response.Info.AppPort)
	}
}
