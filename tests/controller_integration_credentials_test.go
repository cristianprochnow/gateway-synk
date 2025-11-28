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
	"synk/gateway/app/model"
	"testing"
)

func setupCredControllerDB(t *testing.T) (*sql.DB, int) {
	db, err := app.InitDB(true)
	if err != nil {
		t.Fatalf("db connection failed: %v", err)
	}

	userId, err := createTestUserForCreds(db)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return db, userId
}

func createTestUserForCreds(db *sql.DB) (int, error) {
	var userId int
	userName := "Creds Controller Test User"
	userEmail := "creds_controller@synk.com"
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

func injectCredUserContext(r *http.Request, userId int) *http.Request {
	ctx := context.WithValue(r.Context(), controller.CONTEXT_USER_ID_KEY, userId)
	return r.WithContext(ctx)
}

func TestIntCredentials_HandleCreate(t *testing.T) {
	db, userId := setupCredControllerDB(t)
	defer db.Close()

	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	credsController := controller.NewIntCredentials(db)

	//

	reqBody := controller.HandleIntCredentialCreateRequest{
		IntCredentialName:   "Controller Create Test",
		IntCredentialType:   model.Discord,
		IntCredentialConfig: `{"token": "abc"}`,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/int_credentials", bytes.NewBuffer(jsonBody))
	req = injectCredUserContext(req, userId)
	rr := httptest.NewRecorder()

	credsController.HandleCreate(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v. Body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var response controller.HandleIntCredentialCreateResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if !response.Resource.Ok {
		t.Errorf("API Error: %s", response.Resource.Error)
	}
	if response.Data.IntCredentialId == 0 {
		t.Error("Expected valid IntCredentialId, got 0")
	}

	db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", response.Data.IntCredentialId)
}

func TestIntCredentials_HandleBasicList(t *testing.T) {
	db, userId := setupCredControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	credsController := controller.NewIntCredentials(db)

	res, _ := db.Exec("INSERT INTO integration_credential (int_credential_name, int_credential_type, int_credential_config, user_id) VALUES ('BasicList Item', 'discord', '{}', ?)", userId)
	id, _ := res.LastInsertId()
	defer db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", id)

	req, _ := http.NewRequest("GET", "/int_credentials/basic", nil)
	req = injectCredUserContext(req, userId)
	rr := httptest.NewRecorder()

	credsController.HandleBasicList(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var response controller.HandleIntCredentialsBasicListResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if len(response.Data) == 0 {
		t.Error("expected list to contain items")
	}
}

func TestIntCredentials_HandleList(t *testing.T) {
	db, userId := setupCredControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	credsController := controller.NewIntCredentials(db)

	res, _ := db.Exec("INSERT INTO integration_credential (int_credential_name, int_credential_type, int_credential_config, user_id) VALUES ('DetailedList Item', 'telegram', '{\"secret\":1}', ?)", userId)
	id, _ := res.LastInsertId()
	defer db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", id)

	url := fmt.Sprintf("/int_credentials?int_credential_id=%d&include_config=1", id)
	req, _ := http.NewRequest("GET", url, nil)
	req = injectCredUserContext(req, userId)
	rr := httptest.NewRecorder()

	credsController.HandleList(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var response controller.HandleIntCredentialListResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if len(response.Data) != 1 {
		t.Errorf("expected 1 item, got %d", len(response.Data))
	} else {
		if response.Data[0].IntCredentialConfig == "" {
			t.Error("Expected config to be included")
		}
	}
}

func TestIntCredentials_HandleUpdate(t *testing.T) {
	db, userId := setupCredControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	credsController := controller.NewIntCredentials(db)

	res, _ := db.Exec("INSERT INTO integration_credential (int_credential_name, int_credential_type, int_credential_config, user_id) VALUES ('Old Name', 'discord', '{}', ?)", userId)
	id, _ := res.LastInsertId()
	defer db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", id)

	reqBody := controller.HandleIntCredentialUpdateRequest{
		IntCredentialId:     int(id),
		IntCredentialName:   "Updated Name",
		IntCredentialType:   model.Telegram,
		IntCredentialConfig: `{"new":1}`,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/int_credentials", bytes.NewBuffer(jsonBody))
	req = injectCredUserContext(req, userId)
	rr := httptest.NewRecorder()

	credsController.HandleUpdate(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v. Error: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var response controller.HandleIntCredentialUpdateResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response.Data.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", response.Data.RowsAffected)
	}
}

func TestIntCredentials_HandleDelete(t *testing.T) {
	db, userId := setupCredControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	credsController := controller.NewIntCredentials(db)

	res, _ := db.Exec("INSERT INTO integration_credential (int_credential_name, int_credential_type, int_credential_config, user_id) VALUES ('To Delete', 'discord', '{}', ?)", userId)
	id, _ := res.LastInsertId()

	defer db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", id)

	reqBody := controller.HandleIntCredentialDeleteRequest{
		IntCredentialId: int(id),
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("DELETE", "/int_credentials", bytes.NewBuffer(jsonBody))
	req = injectCredUserContext(req, userId)
	rr := httptest.NewRecorder()

	credsController.HandleDelete(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var response controller.HandleIntCredentialUpdateResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response.Data.RowsAffected != 1 {
		t.Errorf("expected 1 row affected (deleted), got %d", response.Data.RowsAffected)
	}
}
