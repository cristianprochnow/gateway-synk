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

type HandleTemplateListResponse struct {
	Resource ResponseHeader        `json:"resource"`
	Data     []model.TemplatesList `json:"templates"`
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

func (t *Templates) HandleList(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)

	templateId := r.URL.Query().Get("template_id")
	includeContent := r.URL.Query().Get("include_content")

	templateList, templateErr := t.model.List(templateId, includeContent == "1")

	response := HandleTemplateListResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: templateList,
	}

	if templateErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = templateErr.Error()

		WriteErrorResponse(w, response, "/templates", "error on template fetch", http.StatusInternalServerError)

		return
	}

	WriteSuccessResponse(w, response)
}
