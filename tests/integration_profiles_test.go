package tests

import (
	"context"
	"database/sql"
	"strconv"
	"synk/gateway/app"
	"synk/gateway/app/model"
	"testing"
)

func setupProfilesDB(t *testing.T) (*sql.DB, int) {
	db, err := app.InitDB(true)
	if err != nil {
		t.Fatalf("db connection failed: %v", err)
	}
	return db, 1
}

func getValidColorId(t *testing.T, db *sql.DB) int {
	var colorId int
	err := db.QueryRow("SELECT color_id FROM color LIMIT 1").Scan(&colorId)
	if err != nil {
		t.Fatalf("Setup failed: No color found in DB. Please seed the 'color' table. Error: %v", err)
	}
	return colorId
}

func createDummyCredential(t *testing.T, db *sql.DB, userId int) int {
	res, err := db.ExecContext(context.Background(),
		`INSERT INTO integration_credential (int_credential_name, int_credential_type, int_credential_config, user_id)
		 VALUES ('Profile Test Cred', 'discord', '{}', ?)`, userId)
	if err != nil {
		t.Fatalf("Setup failed: Could not create dummy credential: %v", err)
	}
	id, _ := res.LastInsertId()
	return int(id)
}

func TestNewIntProfiles(t *testing.T) {
	db, _ := setupProfilesDB(t)
	defer db.Close()

	p := model.NewIntProfiles(db)
	if p == nil {
		t.Error("NewIntProfiles returned nil")
	}
}

func TestIntProfiles_Add(t *testing.T) {
	db, userId := setupProfilesDB(t)
	defer db.Close()
	profileModel := model.NewIntProfiles(db)

	colorId := getValidColorId(t, db)
	credId := createDummyCredential(t, db, userId)
	defer db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", credId)

	input := model.IntProfileAddData{
		IntProfileName: "Test Add Profile",
		ColorId:        colorId,
	}
	credentialsToLink := []int{credId}

	id, err := profileModel.Add(input, credentialsToLink, userId)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if id == 0 {
		t.Fatal("Add returned ID 0")
	}

	defer db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", id)
	defer db.Exec("DELETE FROM integration_group WHERE int_profile_id = ?", id)
}

func TestIntProfiles_Update(t *testing.T) {
	db, userId := setupProfilesDB(t)
	defer db.Close()
	profileModel := model.NewIntProfiles(db)

	colorId := getValidColorId(t, db)
	credId := createDummyCredential(t, db, userId)
	defer db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", credId)

	createInput := model.IntProfileAddData{IntProfileName: "Pre-Update", ColorId: colorId}
	id, _ := profileModel.Add(createInput, []int{}, userId)
	defer db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", id)

	updateInput := model.IntProfileUpdateData{
		IntProfileId:   id,
		IntProfileName: "Post-Update",
		ColorId:        colorId,
	}
	credentialsToLink := []int{credId}

	rows, err := profileModel.Update(updateInput, credentialsToLink, userId)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if rows != 1 {
		t.Errorf("Expected 1 row updated, got %d", rows)
	}

	list, _ := profileModel.List(strconv.Itoa(id), userId)
	if list[0].IntProfileName != "Post-Update" {
		t.Errorf("Update name not persisted. Got %s", list[0].IntProfileName)
	}
}

func TestIntProfiles_Delete(t *testing.T) {
	db, userId := setupProfilesDB(t)
	defer db.Close()
	profileModel := model.NewIntProfiles(db)

	colorId := getValidColorId(t, db)
	input := model.IntProfileAddData{IntProfileName: "To Delete", ColorId: colorId}

	id, err := profileModel.Add(input, []int{}, userId)
	if err != nil {
		t.Fatalf("Setup for Delete failed: %v", err)
	}

	defer db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", id)

	rows, err := profileModel.Delete(id, userId)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if rows != 1 {
		t.Errorf("Expected 1 row deleted, got %d", rows)
	}

	list, _ := profileModel.List(strconv.Itoa(id), userId)
	if len(list) != 0 {
		t.Error("Profile returned after soft delete")
	}
}

func TestIntProfiles_List(t *testing.T) {
	db, userId := setupProfilesDB(t)
	defer db.Close()
	profileModel := model.NewIntProfiles(db)

	colorId := getValidColorId(t, db)
	input := model.IntProfileAddData{IntProfileName: "List Test", ColorId: colorId}

	id, _ := profileModel.Add(input, []int{}, userId)
	defer db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", id)

	list, err := profileModel.List(strconv.Itoa(id), userId)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) == 0 {
		t.Fatal("List returned 0 items")
	}

	item := list[0]
	if item.ColorHex == "" {
		t.Error("Expected ColorHex to be populated (Join check)")
	}
	if item.IntProfileName != "List Test" {
		t.Error("Name mismatch")
	}
}

func TestIntProfiles_BasicList(t *testing.T) {
	db, userId := setupProfilesDB(t)
	defer db.Close()
	profileModel := model.NewIntProfiles(db)

	colorId := getValidColorId(t, db)
	input := model.IntProfileAddData{IntProfileName: "Basic List Test", ColorId: colorId}
	id, _ := profileModel.Add(input, []int{}, userId)
	defer db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", id)

	list, err := profileModel.BasicList(userId)
	if err != nil {
		t.Fatalf("BasicList failed: %v", err)
	}

	found := false
	for _, p := range list {
		if p.IntProfileId == id {
			found = true
			if p.ColorName == "" {
				t.Error("Expected ColorName to be populated in BasicList")
			}
			break
		}
	}
	if !found {
		t.Error("Created profile not found in BasicList")
	}
}

func TestIntProfiles_ById(t *testing.T) {
	db, userId := setupProfilesDB(t)
	defer db.Close()
	profileModel := model.NewIntProfiles(db)

	colorId := getValidColorId(t, db)
	input := model.IntProfileAddData{IntProfileName: "ById Test", ColorId: colorId}
	id, _ := profileModel.Add(input, []int{}, userId)
	defer db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", id)

	data, err := profileModel.ById(id, userId)
	if err != nil {
		t.Fatalf("ById failed: %v", err)
	}

	if data.IntProfileId != id {
		t.Errorf("ById returned wrong ID. Got %d, want %d", data.IntProfileId, id)
	}
}

func TestIntProfileLifecycle(t *testing.T) {
	db, userId := setupProfilesDB(t)
	defer db.Close()
	profileModel := model.NewIntProfiles(db)
	colorId := getValidColorId(t, db)

	newProfile := model.IntProfileAddData{
		IntProfileName: "Lifecycle Test",
		ColorId:        colorId,
	}
	createdId, err := profileModel.Add(newProfile, []int{}, userId)
	if err != nil {
		t.Fatalf("lifecycle: add failed: %v", err)
	}

	list, _ := profileModel.List(strconv.Itoa(createdId), userId)
	if len(list) != 1 {
		t.Errorf("lifecycle: list count mismatch")
	}

	updateData := model.IntProfileUpdateData{
		IntProfileId:   createdId,
		IntProfileName: "Lifecycle Updated",
		ColorId:        colorId,
	}
	profileModel.Update(updateData, []int{}, userId)

	profileModel.Delete(createdId, userId)

	finalList, _ := profileModel.List(strconv.Itoa(createdId), userId)
	if len(finalList) != 0 {
		t.Error("lifecycle: soft delete failed")
	}

	db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", createdId)
}
