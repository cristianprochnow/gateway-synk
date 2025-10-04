package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"synk/gateway/app/model"
	"synk/gateway/app/util"
)

type Templates struct {
	model *model.Templates
}

type HandleTemplateBasicListResponse struct {
	Resource ResponseHeader             `json:"resource"`
	Data     []model.TemplatesBasicList `json:"templates"`
}

func NewTemplates(db *sql.DB) *Templates {
	templates := Templates{model: model.NewTemplates(db)}

	return &templates
}

func (t *Templates) HandleBasicList(w http.ResponseWriter, r *http.Request) {
	EnableCors(w)
	SetJsonContentType(w)

	templatesList, templatesErr := t.model.BasicList()

	response := HandleTemplateBasicListResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: templatesList,
	}

	if templatesErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = templatesErr.Error()
	}

	jsonResp, jsonErr := json.Marshal(response)
	if jsonErr != nil {
		util.LogRoute("/templates/basic", "error on response encoding")

		return
	}

	w.WriteHeader(http.StatusOK)

	_, writeErr := w.Write(jsonResp)
	if writeErr != nil {
		util.LogRoute("/templates/basic", "error on response log")
	}
}
