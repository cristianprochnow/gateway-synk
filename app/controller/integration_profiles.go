package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"synk/gateway/app/model"
	"synk/gateway/app/util"
)

type IntProfiles struct {
	model *model.IntProfiles
}

type HandleIntProfilesBasicListResponse struct {
	Resource ResponseHeader               `json:"resource"`
	Data     []model.IntProfilesBasicList `json:"int_profiles"`
}

func NewIntProfiles(db *sql.DB) *IntProfiles {
	intProfiles := IntProfiles{model: model.NewIntProfiles(db)}

	return &intProfiles
}

func (ip *IntProfiles) HandleBasicList(w http.ResponseWriter, r *http.Request) {
	EnableCors(w)
	SetJsonContentType(w)

	intProfilesList, intProfilesErr := ip.model.BasicList()

	response := HandleIntProfilesBasicListResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: intProfilesList,
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
