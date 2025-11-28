package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"synk/gateway/app"
	"synk/gateway/app/controller"
	"testing"
)

func setupPostsControllerDB(t *testing.T) (*sql.DB, int) {
	db, err := app.InitDB(true)
	if err != nil {
		t.Fatalf("db connection failed: %v", err)
	}

	userId, err := createTestUserForPosts(db)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return db, userId
}

func createTestUserForPosts(db *sql.DB) (int, error) {
	var userId int
	userName := "Posts Controller User"
	userEmail := "posts_controller_test@synk.com"
	userPass := "123456"

	insertRes, insertErr := db.ExecContext(
		context.Background(),
		`INSERT INTO synk.user (user_name, user_email, user_pass) VALUES (?, ?, ?)`,
		userName, userEmail, userPass,
	)

	if insertErr != nil {

		var existingId int
		err := db.QueryRow("SELECT user_id FROM user WHERE user_email = ?", userEmail).Scan(&existingId)
		if err == nil {
			return existingId, nil
		}
		return 0, fmt.Errorf("failed to create user: %s", insertErr.Error())
	}

	id, _ := insertRes.LastInsertId()
	userId = int(id)
	return userId, nil
}

func injectPostUserContext(r *http.Request, userId int) *http.Request {
	ctx := context.WithValue(r.Context(), controller.CONTEXT_USER_ID_KEY, userId)
	return r.WithContext(ctx)
}

func createPostDependencies(t *testing.T, db *sql.DB, userId int) (int, int) {

	res, err := db.Exec("INSERT INTO template (template_name, template_content, template_url_import, user_id) VALUES ('Post Ctrl Tpl', 'x', 'x', ?)", userId)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}
	tplId, _ := res.LastInsertId()

	var colorId int
	err = db.QueryRow("SELECT color_id FROM color LIMIT 1").Scan(&colorId)
	if err != nil {
		t.Fatalf("No color found in DB (required for profile): %v", err)
	}

	res, err = db.Exec("INSERT INTO integration_profile (int_profile_name, color_id, user_id) VALUES ('Post Ctrl Profile', ?, ?)", colorId, userId)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}
	profId, _ := res.LastInsertId()

	return int(tplId), int(profId)
}

func TestPosts_HandleCreate(t *testing.T) {
	db, userId := setupPostsControllerDB(t)
	defer db.Close()

	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	tplId, profId := createPostDependencies(t, db, userId)
	defer db.Exec("DELETE FROM template WHERE template_id = ?", tplId)
	defer db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", profId)

	postController := controller.NewPosts(db)

	reqBody := controller.HandlePostCreateRequest{
		PostName:     "Controller Create Test",
		PostContent:  "Content",
		TemplateId:   tplId,
		IntProfileId: profId,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonBody))
	req = injectPostUserContext(req, userId)
	rr := httptest.NewRecorder()

	postController.HandleCreate(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v. Body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var response controller.HandlePostCreateResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if !response.Resource.Ok {
		t.Errorf("API Error: %s", response.Resource.Error)
	}
	if response.Data.PostId == 0 {
		t.Error("Expected valid PostId")
	}

	db.Exec("DELETE FROM post WHERE post_id = ?", response.Data.PostId)
}

func TestPosts_HandleList(t *testing.T) {
	db, userId := setupPostsControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	tplId, profId := createPostDependencies(t, db, userId)
	defer db.Exec("DELETE FROM template WHERE template_id = ?", tplId)
	defer db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", profId)

	res, _ := db.Exec("INSERT INTO post (post_name, post_content, template_id, int_profile_id, user_id) VALUES ('List Test', 'Hidden Content', ?, ?, ?)", tplId, profId, userId)
	postId, _ := res.LastInsertId()
	defer db.Exec("DELETE FROM post WHERE post_id = ?", postId)

	postController := controller.NewPosts(db)

	url := fmt.Sprintf("/posts?post_id=%d&include_content=1", postId)
	req, _ := http.NewRequest("GET", url, nil)
	req = injectPostUserContext(req, userId)
	rr := httptest.NewRecorder()

	postController.HandleList(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var response controller.HandleListResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if len(response.Data) != 1 {
		t.Errorf("Expected 1 post, got %d", len(response.Data))
	} else {
		if response.Data[0].PostContent != "Hidden Content" {
			t.Error("Expected content to be included")
		}
	}
}

func TestPosts_HandleUpdate(t *testing.T) {
	db, userId := setupPostsControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	tplId, profId := createPostDependencies(t, db, userId)
	defer db.Exec("DELETE FROM template WHERE template_id = ?", tplId)
	defer db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", profId)

	res, _ := db.Exec("INSERT INTO post (post_name, post_content, template_id, int_profile_id, user_id) VALUES ('Old Name', 'Old Content', ?, ?, ?)", tplId, profId, userId)
	postId, _ := res.LastInsertId()
	defer db.Exec("DELETE FROM post WHERE post_id = ?", postId)

	postController := controller.NewPosts(db)

	reqBody := controller.HandlePostUpdateRequest{
		PostId:       int(postId),
		PostName:     "Updated Name",
		PostContent:  "Updated Content",
		TemplateId:   tplId,
		IntProfileId: profId,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/posts", bytes.NewBuffer(jsonBody))
	req = injectPostUserContext(req, userId)
	rr := httptest.NewRecorder()

	postController.HandleUpdate(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v. Error: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var response controller.HandlePostUpdateResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response.Data.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", response.Data.RowsAffected)
	}
}

func TestPosts_HandleDelete(t *testing.T) {
	db, userId := setupPostsControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	tplId, profId := createPostDependencies(t, db, userId)
	defer db.Exec("DELETE FROM template WHERE template_id = ?", tplId)
	defer db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", profId)

	res, _ := db.Exec("INSERT INTO post (post_name, post_content, template_id, int_profile_id, user_id) VALUES ('Delete Me', 'x', ?, ?, ?)", tplId, profId, userId)
	postId, _ := res.LastInsertId()

	defer db.Exec("DELETE FROM post WHERE post_id = ?", postId)

	postController := controller.NewPosts(db)

	reqBody := controller.HandlePostDeleteRequest{
		PostId: int(postId),
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("DELETE", "/posts", bytes.NewBuffer(jsonBody))
	req = injectPostUserContext(req, userId)
	rr := httptest.NewRecorder()

	postController.HandleDelete(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var response controller.HandlePostUpdateResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response.Data.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", response.Data.RowsAffected)
	}
}
