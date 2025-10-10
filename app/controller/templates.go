package controller

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
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

type HandleTemplateCreateResponse struct {
	Resource ResponseHeader                   `json:"resource"`
	Data     CreateTemplateCreateDataResponse `json:"template"`
}

type CreateTemplateCreateDataResponse struct {
	TemplateId int `json:"template_id"`
}

type HandleTemplateCreateRequest struct {
	TemplateName      string `json:"template_name"`
	TemplateContent   string `json:"template_content"`
	TemplateUrlImport string `json:"template_url_import"`
}

type HandleTemplateUpdateResponse struct {
	Resource ResponseHeader             `json:"resource"`
	Data     UpdateTemplateDataResponse `json:"template"`
}

type UpdateTemplateDataResponse struct {
	RowsAffected int `json:"rows_affected"`
}

type HandleTemplateUpdateRequest struct {
	TemplateId        int    `json:"template_id"`
	TemplateName      string `json:"template_name"`
	TemplateContent   string `json:"template_content"`
	TemplateUrlImport string `json:"template_url_import"`
}

type HandleTemplateDeleteRequest struct {
	TemplateId int `json:"template_id"`
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

func (t *Templates) HandleCreate(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)

	response := HandleTemplateCreateResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: CreateTemplateCreateDataResponse{},
	}

	bodyContent, bodyErr := io.ReadAll(r.Body)

	if bodyErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "error on read creation body"

		WriteErrorResponse(w, response, "/templates", response.Resource.Error, http.StatusBadRequest)

		return
	}

	var template HandleTemplateCreateRequest

	jsonErr := json.Unmarshal(bodyContent, &template)

	if jsonErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "some fields can be in invalid format"

		WriteErrorResponse(w, response, "/templates", response.Resource.Error, http.StatusBadRequest)

		return
	}

	template.TemplateName = strings.TrimSpace(template.TemplateName)
	template.TemplateContent = strings.TrimSpace(template.TemplateContent)
	template.TemplateUrlImport = strings.TrimSpace(template.TemplateUrlImport)

	hasAllData := template.TemplateName != "" &&
		template.TemplateContent != "" &&
		template.TemplateUrlImport != ""

	if !hasAllData {
		response.Resource.Ok = false
		response.Resource.Error = "fields template_name, template_content, template_url_import are required"

		WriteErrorResponse(w, response, "/templates", response.Resource.Error, http.StatusBadRequest)

		return
	}

	creationId, creationErr := t.model.Add(model.TemplateAddData{
		TemplateName:      template.TemplateName,
		TemplateContent:   template.TemplateContent,
		TemplateUrlImport: template.TemplateUrlImport,
		UserId:            1,
	})

	if creationErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = creationErr.Error()

		WriteErrorResponse(w, response, "/templates", response.Resource.Error, http.StatusBadRequest)

		return
	}

	response.Data.TemplateId = creationId

	WriteSuccessResponse(w, response)
}

func (t *Templates) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)

	response := HandleTemplateUpdateResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: UpdateTemplateDataResponse{},
	}

	bodyContent, bodyErr := io.ReadAll(r.Body)

	if bodyErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "error on read update body"

		WriteErrorResponse(w, response, "/templates", response.Resource.Error, http.StatusBadRequest)

		return
	}

	var template HandleTemplateUpdateRequest

	jsonErr := json.Unmarshal(bodyContent, &template)

	if jsonErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "some fields can be in invalid format"

		WriteErrorResponse(w, response, "/templates", response.Resource.Error, http.StatusBadRequest)

		return
	}

	template.TemplateName = strings.TrimSpace(template.TemplateName)
	template.TemplateContent = strings.TrimSpace(template.TemplateContent)
	template.TemplateUrlImport = strings.TrimSpace(template.TemplateUrlImport)

	hasAllData := template.TemplateId != 0 &&
		template.TemplateName != "" &&
		template.TemplateContent != "" &&
		template.TemplateUrlImport != ""

	if !hasAllData {
		response.Resource.Ok = false
		response.Resource.Error = "fields template_id, template_name, template_content, template_url_import are required"

		WriteErrorResponse(w, response, "/templates", response.Resource.Error, http.StatusBadRequest)

		return
	}

	templateById, _ := t.model.ById(template.TemplateId)

	if templateById.TemplateId == 0 {
		response.Resource.Ok = false
		response.Resource.Error = "template with id " + strconv.Itoa(template.TemplateId) + " not found"

		WriteErrorResponse(w, response, "/templates", response.Resource.Error, http.StatusBadRequest)

		return
	}

	rowsAffected, updateErr := t.model.Update(model.TemplateUpdateData{
		TemplateId:        templateById.TemplateId,
		TemplateName:      template.TemplateName,
		TemplateContent:   template.TemplateContent,
		TemplateUrlImport: template.TemplateUrlImport,
	})

	if updateErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = updateErr.Error()

		WriteErrorResponse(w, response, "/templates", response.Resource.Error, http.StatusBadRequest)

		return
	}

	response.Data.RowsAffected = rowsAffected

	WriteSuccessResponse(w, response)
}

func (t *Templates) HandleDelete(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)

	response := HandleTemplateUpdateResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: UpdateTemplateDataResponse{},
	}

	bodyContent, bodyErr := io.ReadAll(r.Body)

	if bodyErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "error on read delete body"

		WriteErrorResponse(w, response, "/templates", response.Resource.Error, http.StatusBadRequest)

		return
	}

	var template HandleTemplateDeleteRequest

	jsonErr := json.Unmarshal(bodyContent, &template)

	if jsonErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "some fields can be in invalid format"

		WriteErrorResponse(w, response, "/templates", response.Resource.Error, http.StatusBadRequest)

		return
	}

	hasAllData := template.TemplateId != 0

	if !hasAllData {
		response.Resource.Ok = false
		response.Resource.Error = "fields template_id is required"

		WriteErrorResponse(w, response, "/templates", response.Resource.Error, http.StatusBadRequest)

		return
	}

	templateById, _ := t.model.ById(template.TemplateId)

	if templateById.TemplateId == 0 {
		response.Resource.Ok = false
		response.Resource.Error = "template with id " + strconv.Itoa(template.TemplateId) + " not found"

		WriteErrorResponse(w, response, "/templates", response.Resource.Error, http.StatusBadRequest)

		return
	}

	rowsAffected, updateErr := t.model.Delete(template.TemplateId)

	if updateErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = updateErr.Error()

		WriteErrorResponse(w, response, "/templates", response.Resource.Error, http.StatusBadRequest)

		return
	}

	response.Data.RowsAffected = rowsAffected

	WriteSuccessResponse(w, response)
}
