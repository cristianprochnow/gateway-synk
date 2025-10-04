package controller

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"synk/gateway/app/model"
)

type Posts struct {
	model           *model.Posts
	templateModel   *model.Templates
	intProfileModel *model.IntProfiles
}

type HandleListResponse struct {
	Resource ResponseHeader    `json:"resource"`
	Data     []model.PostsList `json:"posts"`
}

type HandlePostCreateRequest struct {
	PostName     string `json:"post_name"`
	PostContent  string `json:"post_content"`
	TemplateId   int    `json:"template_id"`
	IntProfileId int    `json:"int_profile_id"`
}

type HandlePostCreateResponse struct {
	Resource ResponseHeader               `json:"resource"`
	Data     CreatePostCreateDataResponse `json:"post"`
}

type CreatePostCreateDataResponse struct {
	PostId int `json:"post_id"`
}

func NewPosts(db *sql.DB) *Posts {
	posts := Posts{
		model:           model.NewPosts(db),
		templateModel:   model.NewTemplates(db),
		intProfileModel: model.NewIntProfiles(db),
	}

	return &posts
}

func (p *Posts) HandleList(w http.ResponseWriter, r *http.Request) {
	EnableCors(w)
	SetJsonContentType(w)

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

		WriteErrorResponse(w, response, "/posts", "error on post fetch", http.StatusInternalServerError)

		return
	}

	WriteSuccessResponse(w, response)
}

func (p *Posts) HandleCreate(w http.ResponseWriter, r *http.Request) {
	EnableCors(w)
	SetJsonContentType(w)

	response := HandlePostCreateResponse{
		Resource: ResponseHeader{
			Ok: true,
		},
		Data: CreatePostCreateDataResponse{},
	}

	bodyContent, bodyErr := io.ReadAll(r.Body)

	if bodyErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "error on read creation body"

		WriteErrorResponse(w, response, "/posts", response.Resource.Error, http.StatusBadRequest)

		return
	}

	var post HandlePostCreateRequest

	jsonErr := json.Unmarshal(bodyContent, &post)

	if jsonErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = "some fields can be in invalid format"

		WriteErrorResponse(w, response, "/posts", response.Resource.Error, http.StatusBadRequest)

		return
	}

	post.PostName = strings.TrimSpace(post.PostName)
	post.PostContent = strings.TrimSpace(post.PostContent)

	hasAllData := post.PostName != "" &&
		post.PostContent != "" &&
		post.TemplateId != 0 &&
		post.IntProfileId != 0

	if !hasAllData {
		response.Resource.Ok = false
		response.Resource.Error = "fields post_name, post_content, template_id, int_profile_id are required"

		WriteErrorResponse(w, response, "/posts", response.Resource.Error, http.StatusBadRequest)

		return
	}

	templateById, _ := p.templateModel.ById(post.TemplateId)

	if templateById.TemplateId == 0 {
		response.Resource.Ok = false
		response.Resource.Error = "template with id " + strconv.Itoa(post.TemplateId) + " not found"

		WriteErrorResponse(w, response, "/posts", response.Resource.Error, http.StatusBadRequest)

		return
	}

	intProfileById, _ := p.intProfileModel.ById(post.IntProfileId)

	if intProfileById.IntProfileId == 0 {
		response.Resource.Ok = false
		response.Resource.Error = "integration profile with id " + strconv.Itoa(post.IntProfileId) + " not found"

		WriteErrorResponse(w, response, "/posts", response.Resource.Error, http.StatusBadRequest)

		return
	}

	creationId, creationErr := p.model.Add(model.PostAddData{
		PostName:     post.PostName,
		PostContent:  post.PostContent,
		TemplateId:   post.TemplateId,
		IntProfileId: post.IntProfileId,
	})

	if creationErr != nil {
		response.Resource.Ok = false
		response.Resource.Error = creationErr.Error()

		WriteErrorResponse(w, response, "/posts", response.Resource.Error, http.StatusBadRequest)

		return
	}

	response.Data.PostId = creationId

	WriteSuccessResponse(w, response)
}
