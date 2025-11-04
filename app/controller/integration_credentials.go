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

type IntCredentials struct {
	model      *model.IntCredentials
	ColorModel *model.Colors
}

type HandleIntCredentialsBasicListResponse struct {
	Resource ResponseHeader                  `json:"resource"`
	Data     []model.IntCredentialsBasicList `json:"int_credentials"`
}

type HandleIntCredentialListResponse struct {
	Resource ResponseHeader            `json:"resource"`
	Data     []model.IntCredentialList `json:"int_credentials"`
}

type HandleIntCredentialCreateResponse struct {
	Resource ResponseHeader                        `json:"resource"`
	Data     CreateIntCredentialCreateDataResponse `json:"int_credential"`
}

type CreateIntCredentialCreateDataResponse struct {
	IntCredentialId int `json:"int_credential_id"`
}

type HandleIntCredentialCreateRequest struct {
	IntCredentialName   string               `json:"int_credential_name"`
	IntCredentialType   model.SocialPlatform `json:"int_credential_type"`
	IntCredentialConfig string               `json:"int_credential_config"`
}

type HandleIntCredentialUpdateResponse struct {
	Resource ResponseHeader                  `json:"resource"`
	Data     UpdateIntCredentialDataResponse `json:"int_credential"`
}

type UpdateIntCredentialDataResponse struct {
	RowsAffected int `json:"rows_affected"`
}

type HandleIntCredentialUpdateRequest struct {
	IntCredentialId     int                  `json:"int_credential_id"`
	IntCredentialName   string               `json:"int_credential_name"`
	IntCredentialType   model.SocialPlatform `json:"int_credential_type"`
	IntCredentialConfig string               `json:"int_credential_config"`
}

type HandleIntCredentialDeleteRequest struct {
	IntCredentialId int `json:"int_credential_id"`
}

func NewIntCredentials(db *sql.DB) *IntCredentials {
	intCredentials := IntCredentials{
		model:      model.NewIntCredentials(db),
		ColorModel: model.NewColors(db),
	}

	return &intCredentials
}

func (ic *IntCredentials) HandleBasicList(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)

	intCredentialsList, intProfilesErr := ic.model.BasicList()

	response := HandleIntCredentialsBasicListResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: intCredentialsList,
	}

	if intProfilesErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = intProfilesErr.Error()
	}

	jsonResp, jsonErr := json.Marshal(response)
	if jsonErr != nil {
		util.LogRoute("/int_profiles/basic", "error on response encoding")

		return
	}

	w.WriteHeader(http.StatusOK)

	_, writeErr := w.Write(jsonResp)
	if writeErr != nil {
		util.LogRoute("/int_profiles/basic", "error on response log")
	}
}

func (ic *IntCredentials) HandleList(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)

	intCredentialId := r.URL.Query().Get("int_credential_id")
	includeConfig := r.URL.Query().Get("include_config")

	intCredentialList, intCrendentialErr := ic.model.List(intCredentialId, includeConfig == "1")

	response := HandleIntCredentialListResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: intCredentialList,
	}

	if intCrendentialErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = intCrendentialErr.Error()

		WriteErrorResponse(w, response, "/int_credentials", "error on integration credentials fetch", http.StatusInternalServerError)

		return
	}

	WriteSuccessResponse(w, response)
}

func (ic *IntCredentials) HandleCreate(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)

	response := HandleIntCredentialCreateResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: CreateIntCredentialCreateDataResponse{},
	}

	bodyContent, bodyErr := io.ReadAll(r.Body)

	if bodyErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "error on read creation body"

		WriteErrorResponse(w, response, "/int_credentials", response.Resource.Error, http.StatusBadRequest)

		return
	}

	var intCredential HandleIntCredentialCreateRequest

	jsonErr := json.Unmarshal(bodyContent, &intCredential)

	if jsonErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "some fields can be in invalid format"

		WriteErrorResponse(w, response, "/int_credentials", response.Resource.Error, http.StatusBadRequest)

		return
	}

	intCredential.IntCredentialName = strings.TrimSpace(intCredential.IntCredentialName)

	hasAllData := intCredential.IntCredentialName != "" &&
		intCredential.IntCredentialType.IsValid() &&
		intCredential.IntCredentialConfig != ""

	if !hasAllData {
		response.Resource.Ok = false
		response.Resource.Error = "fields int_credential_name, int_credential_type, int_credential_config are required"

		WriteErrorResponse(w, response, "/int_credentials", response.Resource.Error, http.StatusBadRequest)

		return
	}

	creationId, creationErr := ic.model.Add(model.IntCredentialAddData{
		IntCredentialName:   intCredential.IntCredentialName,
		IntCredentialType:   intCredential.IntCredentialType,
		IntCredentialConfig: intCredential.IntCredentialConfig,
	})

	if creationErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = creationErr.Error()

		WriteErrorResponse(w, response, "/int_credentials", response.Resource.Error, http.StatusBadRequest)

		return
	}

	response.Data.IntCredentialId = creationId

	WriteSuccessResponse(w, response)
}

func (ic *IntCredentials) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)

	response := HandleIntCredentialUpdateResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: UpdateIntCredentialDataResponse{},
	}

	bodyContent, bodyErr := io.ReadAll(r.Body)

	if bodyErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "error on read update body"

		WriteErrorResponse(w, response, "/int_credentials", response.Resource.Error, http.StatusBadRequest)

		return
	}

	var intCredential HandleIntCredentialUpdateRequest

	jsonErr := json.Unmarshal(bodyContent, &intCredential)

	if jsonErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "some fields can be in invalid format"

		WriteErrorResponse(w, response, "/int_credentials", response.Resource.Error, http.StatusBadRequest)

		return
	}

	intCredential.IntCredentialName = strings.TrimSpace(intCredential.IntCredentialName)

	hasAllData := intCredential.IntCredentialId != 0 &&
		intCredential.IntCredentialName != "" &&
		intCredential.IntCredentialType.IsValid() &&
		intCredential.IntCredentialConfig != ""

	if !hasAllData {
		response.Resource.Ok = false
		response.Resource.Error = "fields int_credential_id, int_credential_name, int_credential_type, int_credential_config are required"

		WriteErrorResponse(w, response, "/int_credentials", response.Resource.Error, http.StatusBadRequest)

		return
	}

	intCredentialById, _ := ic.model.List(strconv.Itoa(intCredential.IntCredentialId), false)

	if len(intCredentialById) == 0 {
		response.Resource.Ok = false
		response.Resource.Error = "integration credential with id " + strconv.Itoa(intCredential.IntCredentialId) + " not found"

		WriteErrorResponse(w, response, "/int_credentials", response.Resource.Error, http.StatusBadRequest)

		return
	}

	rowsAffected, updateErr := ic.model.Update(model.IntCredentialUpdateData{
		IntCredentialId:     intCredential.IntCredentialId,
		IntCredentialName:   intCredential.IntCredentialName,
		IntCredentialType:   intCredential.IntCredentialType,
		IntCredentialConfig: intCredential.IntCredentialConfig,
	})

	if updateErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = updateErr.Error()

		WriteErrorResponse(w, response, "/int_credentials", response.Resource.Error, http.StatusBadRequest)

		return
	}

	response.Data.RowsAffected = rowsAffected

	WriteSuccessResponse(w, response)
}

func (ic *IntCredentials) HandleDelete(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)

	response := HandleIntCredentialUpdateResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: UpdateIntCredentialDataResponse{},
	}

	bodyContent, bodyErr := io.ReadAll(r.Body)

	if bodyErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "error on read delete body"

		WriteErrorResponse(w, response, "/int_credentials", response.Resource.Error, http.StatusBadRequest)

		return
	}

	var intCredential HandleIntCredentialDeleteRequest

	jsonErr := json.Unmarshal(bodyContent, &intCredential)

	if jsonErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "some fields can be in invalid format"

		WriteErrorResponse(w, response, "/int_credentials", response.Resource.Error, http.StatusBadRequest)

		return
	}

	hasAllData := intCredential.IntCredentialId != 0

	if !hasAllData {
		response.Resource.Ok = false
		response.Resource.Error = "fields int_credential_id is required"

		WriteErrorResponse(w, response, "/int_credentials", response.Resource.Error, http.StatusBadRequest)

		return
	}

	intCredentialById, _ := ic.model.List(strconv.Itoa(intCredential.IntCredentialId), false)

	if len(intCredentialById) == 0 {
		response.Resource.Ok = false
		response.Resource.Error = "integration credential with id " + strconv.Itoa(intCredential.IntCredentialId) + " not found"

		WriteErrorResponse(w, response, "/int_credentials", response.Resource.Error, http.StatusBadRequest)

		return
	}

	rowsAffected, updateErr := ic.model.Delete(intCredential.IntCredentialId)

	if updateErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = updateErr.Error()

		WriteErrorResponse(w, response, "/int_credentials", response.Resource.Error, http.StatusBadRequest)

		return
	}

	response.Data.RowsAffected = rowsAffected

	WriteSuccessResponse(w, response)
}
