package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"synk/gateway/app/model"
	"synk/gateway/app/util"
)

type Posts struct {
	model *model.Posts
}

func NewPosts(db *sql.DB) *Posts {
	posts := Posts{model: model.NewPosts(db)}

	return &posts
}

func (p *Posts) HandleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Ok: true,
		Info: map[string]string{
			"show": "show",
		},
	}

	jsonResp, jsonErr := json.Marshal(response)
	if jsonErr != nil {
		util.LogRoute("/posts", "error on response encoding")

		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}
