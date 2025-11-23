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

type IntProfiles struct {
	model              *model.IntProfiles
	ColorModel         *model.Colors
	IntCredentialModel *model.IntCredentials
}

type HandleIntProfilesBasicListResponse struct {
	Resource ResponseHeader               `json:"resource"`
	Data     []model.IntProfilesBasicList `json:"int_profiles"`
}

type HandleIntProfileListResponse struct {
	Resource ResponseHeader         `json:"resource"`
	Data     []model.IntProfileList `json:"int_profiles"`
}

type HandleIntProfileCreateResponse struct {
	Resource ResponseHeader                     `json:"resource"`
	Data     CreateIntProfileCreateDataResponse `json:"int_profile"`
}

type CreateIntProfileCreateDataResponse struct {
	IntProfileId int `json:"int_profile_id"`
}

type HandleIntProfileCreateRequest struct {
	IntProfileName  string `json:"int_profile_name"`
	ColorId         int    `json:"color_id"`
	CredentialsList []int  `json:"credentials"`
}

type HandleIntProfileUpdateResponse struct {
	Resource ResponseHeader               `json:"resource"`
	Data     UpdateIntProfileDataResponse `json:"int_profile"`
}

type UpdateIntProfileDataResponse struct {
	RowsAffected int `json:"rows_affected"`
}

type HandleIntProfileUpdateRequest struct {
	IntProfileId    int    `json:"int_profile_id"`
	IntProfileName  string `json:"int_profile_name"`
	ColorId         int    `json:"color_id"`
	CredentialsList []int  `json:"credentials"`
}

type HandleIntProfileDeleteRequest struct {
	IntProfileId int `json:"int_profile_id"`
}

func NewIntProfiles(db *sql.DB) *IntProfiles {
	intProfiles := IntProfiles{
		model:              model.NewIntProfiles(db),
		ColorModel:         model.NewColors(db),
		IntCredentialModel: model.NewIntCredentials(db),
	}

	return &intProfiles
}

func (ip *IntProfiles) HandleBasicList(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)

	response := HandleIntProfilesBasicListResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: []model.IntProfilesBasicList{},
	}

	ctxUserId := r.Context().Value(CONTEXT_USER_ID_KEY).(int)

	if ctxUserId == 0 {
		response.Resource.Ok = false
		response.Resource.Error = "reference to user not found in context"

		WriteErrorResponse(w, response, "/int_profiles/basic", response.Resource.Error, http.StatusInternalServerError)

		return
	}

	intProfilesList, intProfilesErr := ip.model.BasicList(ctxUserId)

	if intProfilesErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = intProfilesErr.Error()
	}

	response.Data = intProfilesList

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

func (ip *IntProfiles) HandleList(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)

	response := HandleIntProfileListResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: []model.IntProfileList{},
	}

	ctxUserId := r.Context().Value(CONTEXT_USER_ID_KEY).(int)

	if ctxUserId == 0 {
		response.Resource.Ok = false
		response.Resource.Error = "reference to user not found in context"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusInternalServerError)

		return
	}

	intProfileId := r.URL.Query().Get("int_profile_id")

	intProfileList, templateErr := ip.model.List(intProfileId, ctxUserId)

	serializeProfileList := []model.IntProfileList{}

	for _, intProfileItem := range intProfileList {
		itemCredentialsList, _ := ip.IntCredentialModel.BasicListByProfile(intProfileItem.IntProfileId, ctxUserId)

		intProfileItem.Credentials = itemCredentialsList

		serializeProfileList = append(serializeProfileList, intProfileItem)
	}

	response.Data = serializeProfileList

	if templateErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = templateErr.Error()

		WriteErrorResponse(w, response, "/int_profiles", "error on integration profiles fetch", http.StatusInternalServerError)

		return
	}

	WriteSuccessResponse(w, response)
}

func (ip *IntProfiles) HandleCreate(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)

	response := HandleIntProfileCreateResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: CreateIntProfileCreateDataResponse{},
	}

	ctxUserId := r.Context().Value(CONTEXT_USER_ID_KEY).(int)

	if ctxUserId == 0 {
		response.Resource.Ok = false
		response.Resource.Error = "reference to user not found in context"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusInternalServerError)

		return
	}

	bodyContent, bodyErr := io.ReadAll(r.Body)

	if bodyErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "error on read creation body"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	var intProfile HandleIntProfileCreateRequest

	jsonErr := json.Unmarshal(bodyContent, &intProfile)

	if jsonErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "some fields can be in invalid format"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	intProfile.IntProfileName = strings.TrimSpace(intProfile.IntProfileName)

	hasAllData := intProfile.IntProfileName != "" &&
		intProfile.ColorId != 0 &&
		len(intProfile.CredentialsList) > 0

	if !hasAllData {
		response.Resource.Ok = false
		response.Resource.Error = "fields int_profile_name, color_id and credentials are required"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	colorsById, _ := ip.ColorModel.List(intProfile.ColorId)

	if len(colorsById) == 0 {
		response.Resource.Ok = false
		response.Resource.Error = "color with id " + strconv.Itoa(intProfile.ColorId) + " not found"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	allCredentialsExists := true

	for _, credentialId := range intProfile.CredentialsList {
		credentialSearchResult, credentialSearchError := ip.IntCredentialModel.List(strconv.Itoa(credentialId), false, ctxUserId)

		if credentialSearchError != nil || len(credentialSearchResult) == 0 {
			allCredentialsExists = false

			break
		}
	}

	if !allCredentialsExists {
		response.Resource.Ok = false
		response.Resource.Error = "not all credential IDs from list are valid or exists"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	creationId, creationErr := ip.model.Add(model.IntProfileAddData{
		IntProfileName: intProfile.IntProfileName,
		ColorId:        intProfile.ColorId,
	}, intProfile.CredentialsList, ctxUserId)

	if creationErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = creationErr.Error()

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	response.Data.IntProfileId = creationId

	WriteSuccessResponse(w, response)
}

func (ip *IntProfiles) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)

	response := HandleIntProfileUpdateResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: UpdateIntProfileDataResponse{},
	}

	ctxUserId := r.Context().Value(CONTEXT_USER_ID_KEY).(int)

	if ctxUserId == 0 {
		response.Resource.Ok = false
		response.Resource.Error = "reference to user not found in context"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusInternalServerError)

		return
	}

	bodyContent, bodyErr := io.ReadAll(r.Body)

	if bodyErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "error on read update body"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	var intProfile HandleIntProfileUpdateRequest

	jsonErr := json.Unmarshal(bodyContent, &intProfile)

	if jsonErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "some fields can be in invalid format"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	intProfile.IntProfileName = strings.TrimSpace(intProfile.IntProfileName)

	hasAllData := intProfile.IntProfileId != 0 &&
		intProfile.IntProfileName != "" &&
		intProfile.ColorId != 0 &&
		len(intProfile.CredentialsList) > 0

	if !hasAllData {
		response.Resource.Ok = false
		response.Resource.Error = "fields int_profile_id, int_profile_name, color_id and credentials are required"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	colorsById, _ := ip.ColorModel.List(intProfile.ColorId)

	if len(colorsById) == 0 {
		response.Resource.Ok = false
		response.Resource.Error = "color with id " + strconv.Itoa(intProfile.ColorId) + " not found"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	intProfileById, _ := ip.model.ById(intProfile.IntProfileId, ctxUserId)

	if intProfileById.IntProfileId == 0 {
		response.Resource.Ok = false
		response.Resource.Error = "integration profile with id " + strconv.Itoa(intProfile.IntProfileId) + " not found"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	allCredentialsExists := true

	for _, credentialId := range intProfile.CredentialsList {
		credentialSearchResult, credentialSearchError := ip.IntCredentialModel.List(strconv.Itoa(credentialId), false, ctxUserId)

		if credentialSearchError != nil || len(credentialSearchResult) == 0 {
			allCredentialsExists = false

			break
		}
	}

	if !allCredentialsExists {
		response.Resource.Ok = false
		response.Resource.Error = "not all credential IDs from list are valid or exists"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	rowsAffected, updateErr := ip.model.Update(model.IntProfileUpdateData{
		IntProfileId:   intProfile.IntProfileId,
		IntProfileName: intProfile.IntProfileName,
		ColorId:        intProfile.ColorId,
	}, intProfile.CredentialsList, ctxUserId)

	if updateErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = updateErr.Error()

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	response.Data.RowsAffected = rowsAffected

	WriteSuccessResponse(w, response)
}

func (ip *IntProfiles) HandleDelete(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)

	response := HandleIntProfileUpdateResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: UpdateIntProfileDataResponse{},
	}

	ctxUserId := r.Context().Value(CONTEXT_USER_ID_KEY).(int)

	if ctxUserId == 0 {
		response.Resource.Ok = false
		response.Resource.Error = "reference to user not found in context"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusInternalServerError)

		return
	}

	bodyContent, bodyErr := io.ReadAll(r.Body)

	if bodyErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "error on read delete body"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	var intProfile HandleIntProfileDeleteRequest

	jsonErr := json.Unmarshal(bodyContent, &intProfile)

	if jsonErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "some fields can be in invalid format"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	hasAllData := intProfile.IntProfileId != 0

	if !hasAllData {
		response.Resource.Ok = false
		response.Resource.Error = "fields int_profile_id is required"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	intProfileById, _ := ip.model.ById(intProfile.IntProfileId, ctxUserId)

	if intProfileById.IntProfileId == 0 {
		response.Resource.Ok = false
		response.Resource.Error = "integration profile with id " + strconv.Itoa(intProfile.IntProfileId) + " not found"

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	rowsAffected, updateErr := ip.model.Delete(intProfile.IntProfileId, ctxUserId)

	if updateErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = updateErr.Error()

		WriteErrorResponse(w, response, "/int_profiles", response.Resource.Error, http.StatusBadRequest)

		return
	}

	response.Data.RowsAffected = rowsAffected

	WriteSuccessResponse(w, response)
}
