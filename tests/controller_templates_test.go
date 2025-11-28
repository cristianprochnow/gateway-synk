package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"synk/gateway/app"
	"synk/gateway/app/controller"
	"testing"
)

func setupControllerDB(t *testing.T) (*sql.DB, int) {
	db, err := app.InitDB(true)
	if err != nil {
		t.Fatalf("db connection failed: %v", err)
	}

	userId, err := createTestUser(db)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return db, userId
}

func createTestUser(db *sql.DB) (int, error) {
	var userId int

	userName := "Controller Test User"
	userEmail := "test_controller@synk.com"
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

	id, exception := insertRes.LastInsertId()
	if exception != nil {
		return 0, fmt.Errorf("failed to get last insert id: %s", exception.Error())
	}
	userId = int(id)
	return userId, nil
}

func injectUserContext(r *http.Request, userId int) *http.Request {

	ctx := context.WithValue(r.Context(), controller.CONTEXT_USER_ID_KEY, userId)
	return r.WithContext(ctx)
}

func TestTemplates_HandleCreate(t *testing.T) {
	db, userId := setupControllerDB(t)
	defer db.Close()

	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	tmplController := controller.NewTemplates(db)

	reqBody := controller.HandleTemplateCreateRequest{
		TemplateName:      "Controller Test",
		TemplateContent:   "<h1>Hello</h1>",
		TemplateUrlImport: "http://localhost",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/templates", bytes.NewBuffer(jsonBody))
	req = injectUserContext(req, userId)
	rr := httptest.NewRecorder()

	tmplController.HandleCreate(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s",
			status, http.StatusOK, rr.Body.String())
	}

	var response controller.HandleTemplateCreateResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if !response.Resource.Ok {
		t.Errorf("handler returned error: %s", response.Resource.Error)
	}
	if response.Data.TemplateId == 0 {
		t.Error("handler did not return a valid template ID")
	}

	db.Exec("DELETE FROM template WHERE template_id = ?", response.Data.TemplateId)
}

func TestTemplates_HandleBasicList(t *testing.T) {
	db, userId := setupControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	tmplController := controller.NewTemplates(db)

	res, _ := db.Exec("INSERT INTO template (template_name, template_content, template_url_import, user_id) VALUES ('BasicList Item', 'x', 'x', ?)", userId)
	tID, _ := res.LastInsertId()
	defer db.Exec("DELETE FROM template WHERE template_id = ?", tID)

	req, _ := http.NewRequest("GET", "/templates/basic", nil)
	req = injectUserContext(req, userId)
	rr := httptest.NewRecorder()

	tmplController.HandleBasicList(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var response controller.HandleTemplateBasicListResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if len(response.Data) == 0 {
		t.Error("expected list to contain items")
	}
}

func TestTemplates_HandleList(t *testing.T) {
	db, userId := setupControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	tmplController := controller.NewTemplates(db)

	res, _ := db.Exec("INSERT INTO template (template_name, template_content, template_url_import, user_id) VALUES ('DetailedList Item', 'SecretContent', 'x', ?)", userId)
	tID, _ := res.LastInsertId()
	defer db.Exec("DELETE FROM template WHERE template_id = ?", tID)

	req, _ := http.NewRequest("GET", "/templates?include_content=1&template_id="+strconv.Itoa(int(tID)), nil)
	req = injectUserContext(req, userId)
	rr := httptest.NewRecorder()

	tmplController.HandleList(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var response controller.HandleTemplateListResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if len(response.Data) != 1 {
		t.Errorf("expected 1 item, got %d", len(response.Data))
	} else {
		if response.Data[0].TemplateContent != "SecretContent" {
			t.Error("content was not included or incorrect")
		}
	}
}

func TestTemplates_HandleUpdate(t *testing.T) {
	db, userId := setupControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	tmplController := controller.NewTemplates(db)

	res, _ := db.Exec("INSERT INTO template (template_name, template_content, template_url_import, user_id) VALUES ('Old Name', 'Old Content', 'x', ?)", userId)
	tID, _ := res.LastInsertId()
	defer db.Exec("DELETE FROM template WHERE template_id = ?", tID)

	reqBody := controller.HandleTemplateUpdateRequest{
		TemplateId:        int(tID),
		TemplateName:      "New Name",
		TemplateContent:   "New Content",
		TemplateUrlImport: "http://new.com",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/templates", bytes.NewBuffer(jsonBody))
	req = injectUserContext(req, userId)
	rr := httptest.NewRecorder()

	tmplController.HandleUpdate(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v. Error: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var response controller.HandleTemplateUpdateResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response.Data.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", response.Data.RowsAffected)
	}
}

func TestTemplates_HandleDelete(t *testing.T) {
	db, userId := setupControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	tmplController := controller.NewTemplates(db)

	res, _ := db.Exec("INSERT INTO template (template_name, template_content, template_url_import, user_id) VALUES ('To Delete', 'x', 'x', ?)", userId)
	tID, _ := res.LastInsertId()

	defer db.Exec("DELETE FROM template WHERE template_id = ?", tID)

	reqBody := controller.HandleTemplateDeleteRequest{
		TemplateId: int(tID),
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("DELETE", "/templates", bytes.NewBuffer(jsonBody))
	req = injectUserContext(req, userId)
	rr := httptest.NewRecorder()

	tmplController.HandleDelete(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var response controller.HandleTemplateUpdateResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response.Data.RowsAffected != 1 {
		t.Errorf("expected 1 row affected (deleted), got %d", response.Data.RowsAffected)
	}
}
