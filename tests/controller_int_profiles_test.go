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

func setupProfileControllerDB(t *testing.T) (*sql.DB, int) {
	db, err := app.InitDB(true)
	if err != nil {
		t.Fatalf("db connection failed: %v", err)
	}

	userId, err := createTestUserForProfiles(db)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return db, userId
}

func createTestUserForProfiles(db *sql.DB) (int, error) {
	var userId int

	userName := "Profile Controller User"
	userEmail := "profile_controller@synk.com"
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

func injectProfileUserContext(r *http.Request, userId int) *http.Request {
	ctx := context.WithValue(r.Context(), controller.CONTEXT_USER_ID_KEY, userId)
	return r.WithContext(ctx)
}

func getValidColorForProfile(t *testing.T, db *sql.DB) int {
	var colorId int
	err := db.QueryRow("SELECT color_id FROM color LIMIT 1").Scan(&colorId)
	if err != nil {
		t.Fatalf("Setup failed: No color found in DB. %v", err)
	}
	return colorId
}

func createDummyCredentialForProfile(t *testing.T, db *sql.DB, userId int) int {
	res, err := db.ExecContext(context.Background(),
		`INSERT INTO integration_credential (int_credential_name, int_credential_type, int_credential_config, user_id)
		 VALUES ('Profile Test Cred', 'discord', '{}', ?)`, userId)
	if err != nil {
		t.Fatalf("Setup failed: Could not create dummy credential: %v", err)
	}
	id, _ := res.LastInsertId()
	return int(id)
}

func TestIntProfiles_HandleCreate(t *testing.T) {
	db, userId := setupProfileControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	colorId := getValidColorForProfile(t, db)
	credId := createDummyCredentialForProfile(t, db, userId)
	defer db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", credId)

	pController := controller.NewIntProfiles(db)

	reqBody := controller.HandleIntProfileCreateRequest{
		IntProfileName:  "Controller Create Profile",
		ColorId:         colorId,
		CredentialsList: []int{credId},
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/int_profiles", bytes.NewBuffer(jsonBody))
	req = injectProfileUserContext(req, userId)
	rr := httptest.NewRecorder()

	pController.HandleCreate(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v. Body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var response controller.HandleIntProfileCreateResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if !response.Resource.Ok {
		t.Errorf("API Error: %s", response.Resource.Error)
	}
	if response.Data.IntProfileId == 0 {
		t.Error("Expected valid IntProfileId")
	}

	db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", response.Data.IntProfileId)
	db.Exec("DELETE FROM integration_group WHERE int_profile_id = ?", response.Data.IntProfileId)
}

func TestIntProfiles_HandleBasicList(t *testing.T) {
	db, userId := setupProfileControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	pController := controller.NewIntProfiles(db)
	colorId := getValidColorForProfile(t, db)

	res, _ := db.Exec("INSERT INTO integration_profile (int_profile_name, color_id, user_id) VALUES ('BasicList Item', ?, ?)", colorId, userId)
	id, _ := res.LastInsertId()
	defer db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", id)

	req, _ := http.NewRequest("GET", "/int_profiles/basic", nil)
	req = injectProfileUserContext(req, userId)
	rr := httptest.NewRecorder()

	pController.HandleBasicList(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var response controller.HandleIntProfilesBasicListResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if len(response.Data) == 0 {
		t.Error("expected list to contain items")
	}
}

func TestIntProfiles_HandleList(t *testing.T) {
	db, userId := setupProfileControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	pController := controller.NewIntProfiles(db)

	colorId := getValidColorForProfile(t, db)
	credId := createDummyCredentialForProfile(t, db, userId)
	defer db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", credId)

	res, _ := db.Exec("INSERT INTO integration_profile (int_profile_name, color_id, user_id) VALUES ('DetailedList Item', ?, ?)", colorId, userId)
	profId, _ := res.LastInsertId()
	defer db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", profId)

	db.Exec("INSERT INTO integration_group (int_profile_id, int_credential_id) VALUES (?, ?)", profId, credId)
	defer db.Exec("DELETE FROM integration_group WHERE int_profile_id = ?", profId)

	req, _ := http.NewRequest("GET", "/int_profiles", nil)
	req = injectProfileUserContext(req, userId)
	rr := httptest.NewRecorder()

	pController.HandleList(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var response controller.HandleIntProfileListResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	found := false
	for _, p := range response.Data {
		if p.IntProfileId == int(profId) {
			found = true

			if len(p.Credentials) == 0 {
				t.Error("Expected credentials to be populated in Detailed List")
			}
		}
	}
	if !found {
		t.Error("Created profile not found in list")
	}
}

func TestIntProfiles_HandleUpdate(t *testing.T) {
	db, userId := setupProfileControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	pController := controller.NewIntProfiles(db)

	colorId := getValidColorForProfile(t, db)
	credId := createDummyCredentialForProfile(t, db, userId)
	defer db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", credId)

	res, _ := db.Exec("INSERT INTO integration_profile (int_profile_name, color_id, user_id) VALUES ('Old Name', ?, ?)", colorId, userId)
	profId, _ := res.LastInsertId()
	defer db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", profId)

	reqBody := controller.HandleIntProfileUpdateRequest{
		IntProfileId:    int(profId),
		IntProfileName:  "Updated Name",
		ColorId:         colorId,
		CredentialsList: []int{credId},
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/int_profiles", bytes.NewBuffer(jsonBody))
	req = injectProfileUserContext(req, userId)
	rr := httptest.NewRecorder()

	pController.HandleUpdate(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v. Error: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var response controller.HandleIntProfileUpdateResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response.Data.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", response.Data.RowsAffected)
	}
}

func TestIntProfiles_HandleDelete(t *testing.T) {
	db, userId := setupProfileControllerDB(t)
	defer db.Close()
	defer db.Exec("DELETE FROM user WHERE user_id = ?", userId)

	pController := controller.NewIntProfiles(db)
	colorId := getValidColorForProfile(t, db)

	res, _ := db.Exec("INSERT INTO integration_profile (int_profile_name, color_id, user_id) VALUES ('To Delete', ?, ?)", colorId, userId)
	profId, _ := res.LastInsertId()

	defer db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", profId)

	reqBody := controller.HandleIntProfileDeleteRequest{
		IntProfileId: int(profId),
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("DELETE", "/int_profiles", bytes.NewBuffer(jsonBody))
	req = injectProfileUserContext(req, userId)
	rr := httptest.NewRecorder()

	pController.HandleDelete(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var response controller.HandleIntProfileUpdateResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response.Data.RowsAffected != 1 {
		t.Errorf("expected 1 row affected (deleted), got %d", response.Data.RowsAffected)
	}
}
