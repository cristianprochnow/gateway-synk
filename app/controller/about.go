package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"synk/gateway/app/model"
	"synk/gateway/app/util"
)

type AboutResponse struct {
	ServerPort string `json:"server_port"`
	AppPort    string `json:"app_port"`
	DbWorking  bool   `json:"db_working"`
}

type About struct {
	model *model.About
}

func NewAbout(db *sql.DB) *About {
	about := About{model: model.NewAbout(db)}

	return &about
}

func (a *About) HandleAbout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	isDbWorking := a.model.Ping()

	response := Response{
		Ok: true,
		Info: AboutResponse{
			AppPort:    os.Getenv("PORT"),
			ServerPort: "8080",
			DbWorking:  isDbWorking,
		},
	}

	jsonResp, jsonErr := json.Marshal(response)
	if jsonErr != nil {
		util.LogRoute("/about", "error on response encoding")

		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}
