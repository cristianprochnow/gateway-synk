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

type HandleListResponse struct {
	Resource ResponseHeader    `json:"resource"`
	Data     []model.PostsList `json:"posts"`
}

func NewPosts(db *sql.DB) *Posts {
	posts := Posts{model: model.NewPosts(db)}

	return &posts
}

func (p *Posts) HandleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	postList, postErr := p.model.List()

	response := HandleListResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: postList,
	}

	if postErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = postErr.Error()
	}

	jsonResp, jsonErr := json.Marshal(response)
	if jsonErr != nil {
		util.LogRoute("/posts", "error on response encoding")

		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}
