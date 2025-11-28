package tests

import (
	"context"
	"database/sql"
	"strconv"
	"synk/gateway/app"
	"synk/gateway/app/model"
	"testing"
)

func setupCredsDB(t *testing.T) (*sql.DB, int) {
	db, err := app.InitDB(true)
	if err != nil {
		t.Fatalf("db connection failed: %v", err)
	}

	return db, 1
}

func TestSocialPlatform_IsValid(t *testing.T) {
	valid := model.Discord
	if !valid.IsValid() {
		t.Error("Expected Discord to be valid")
	}

	valid2 := model.Telegram
	if !valid2.IsValid() {
		t.Error("Expected Telegram to be valid")
	}

	invalid := model.SocialPlatform("myspace")
	if invalid.IsValid() {
		t.Error("Expected myspace to be invalid")
	}
}

func TestNewIntCredentials(t *testing.T) {
	db, _ := setupCredsDB(t)
	defer db.Close()

	creds := model.NewIntCredentials(db)
	if creds == nil {
		t.Error("NewIntCredentials returned nil")
	}
}

func TestIntCredentials_Add(t *testing.T) {
	db, userId := setupCredsDB(t)
	defer db.Close()
	credsModel := model.NewIntCredentials(db)

	input := model.IntCredentialAddData{
		IntCredentialName:   "Test Add Bot",
		IntCredentialType:   model.Discord,
		IntCredentialConfig: "{}",
	}

	id, err := credsModel.Add(input, userId)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if id == 0 {
		t.Fatal("Add returned ID 0")
	}

	db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", id)
}

func TestIntCredentials_Update(t *testing.T) {
	db, userId := setupCredsDB(t)
	defer db.Close()
	credsModel := model.NewIntCredentials(db)

	input := model.IntCredentialAddData{
		IntCredentialName:   "Pre-Update Name",
		IntCredentialType:   model.Discord,
		IntCredentialConfig: "{}",
	}
	id, err := credsModel.Add(input, userId)
	if err != nil {
		t.Fatalf("Setup for Update failed (Add): %v", err)
	}
	defer db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", id)

	updateData := model.IntCredentialUpdateData{
		IntCredentialId:     id,
		IntCredentialName:   "Post-Update Name",
		IntCredentialType:   model.Telegram,
		IntCredentialConfig: `{"new": "val"}`,
	}

	rows, err := credsModel.Update(updateData, userId)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if rows != 1 {
		t.Errorf("Expected 1 row updated, got %d", rows)
	}

	list, _ := credsModel.List(strconv.Itoa(id), true, userId)
	if len(list) > 0 {
		if list[0].IntCredentialName != "Post-Update Name" {
			t.Error("Update name not persisted")
		}
	} else {
		t.Error("Could not retrieve updated item")
	}
}

func TestIntCredentials_Delete(t *testing.T) {
	db, userId := setupCredsDB(t)
	defer db.Close()
	credsModel := model.NewIntCredentials(db)

	input := model.IntCredentialAddData{
		IntCredentialName:   "To Delete",
		IntCredentialType:   model.Discord,
		IntCredentialConfig: "{}",
	}

	id, err := credsModel.Add(input, userId)
	if err != nil {
		t.Fatalf("Setup for Delete failed (Add): %v", err)
	}
	if id == 0 {
		t.Fatal("Setup for Delete failed: Returned ID is 0")
	}

	defer db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", id)

	rows, err := credsModel.Delete(id, userId)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if rows != 1 {
		t.Errorf("Expected 1 row deleted, got %d. (Did the ID exist?)", rows)
	}

	list, _ := credsModel.List(strconv.Itoa(id), false, userId)
	if len(list) != 0 {
		t.Error("Item still returned after delete")
	}
}

func TestIntCredentials_List(t *testing.T) {
	db, userId := setupCredsDB(t)
	defer db.Close()
	credsModel := model.NewIntCredentials(db)

	input := model.IntCredentialAddData{
		IntCredentialName:   "List Test",
		IntCredentialType:   model.Discord,
		IntCredentialConfig: `{"token":"123"}`,
	}
	id, err := credsModel.Add(input, userId)
	if err != nil {
		t.Fatalf("Setup for List failed: %v", err)
	}
	defer db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", id)

	listWithConfig, err := credsModel.List(strconv.Itoa(id), true, userId)
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(listWithConfig) == 0 || listWithConfig[0].IntCredentialConfig == "" {
		t.Error("Expected config to be present")
	}

	listNoConfig, _ := credsModel.List(strconv.Itoa(id), false, userId)
	if len(listNoConfig) > 0 && listNoConfig[0].IntCredentialConfig != "" {
		t.Error("Expected config to be empty string")
	}

	listAll, _ := credsModel.List("", false, userId)
	if len(listAll) == 0 {
		t.Error("Expected list all to return items")
	}
}

func TestIntCredentials_BasicList(t *testing.T) {
	db, userId := setupCredsDB(t)
	defer db.Close()
	credsModel := model.NewIntCredentials(db)

	input := model.IntCredentialAddData{
		IntCredentialName:   "Basic List Test",
		IntCredentialType:   model.Discord,
		IntCredentialConfig: "{}",
	}
	id, err := credsModel.Add(input, userId)
	if err != nil {
		t.Fatalf("Setup for BasicList failed (Add): %v", err)
	}
	defer db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", id)

	list, err := credsModel.BasicList(userId)
	if err != nil {
		t.Fatalf("BasicList failed: %v", err)
	}

	found := false
	for _, item := range list {
		if item.IntCredentialId == id {
			found = true
			if item.IntCredentialName != "Basic List Test" {
				t.Error("BasicList returned wrong name")
			}
			break
		}
	}
	if !found {
		t.Errorf("Created item (ID %d) not found in BasicList", id)
	}
}

func TestIntCredentials_BasicListByProfile(t *testing.T) {
	db, userId := setupCredsDB(t)
	defer db.Close()
	credsModel := model.NewIntCredentials(db)

	credInput := model.IntCredentialAddData{
		IntCredentialName:   "Linked Cred",
		IntCredentialType:   model.Discord,
		IntCredentialConfig: "{}",
	}
	credId, err := credsModel.Add(credInput, userId)
	if err != nil {
		t.Fatalf("Setup failed: Could not create credential: %v", err)
	}
	defer db.Exec("DELETE FROM integration_credential WHERE int_credential_id = ?", credId)

	var colorId int
	err = db.QueryRow("SELECT color_id FROM color LIMIT 1").Scan(&colorId)
	if err != nil {
		t.Skip("Skipping BasicListByProfile: No color found in DB")
	}

	res, err := db.ExecContext(context.Background(),
		"INSERT INTO integration_profile (int_profile_name, color_id, user_id) VALUES ('Test Profile', ?, ?)",
		colorId, userId)
	if err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}
	profileId64, _ := res.LastInsertId()
	profileId := int(profileId64)
	defer db.Exec("DELETE FROM integration_profile WHERE int_profile_id = ?", profileId)

	_, err = db.Exec("INSERT INTO integration_group (int_profile_id, int_credential_id) VALUES (?, ?)", profileId, credId)
	if err != nil {
		t.Fatalf("Failed to link group (Profile %d, Cred %d): %v", profileId, credId, err)
	}
	defer db.Exec("DELETE FROM integration_group WHERE int_profile_id = ?", profileId)

	list, err := credsModel.BasicListByProfile(profileId, userId)
	if err != nil {
		t.Fatalf("BasicListByProfile failed: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("Expected 1 credential linked to profile, got %d", len(list))
	} else {
		if list[0].IntCredentialId != credId {
			t.Errorf("Returned wrong credential ID. Got %d, want %d", list[0].IntCredentialId, credId)
		}
	}
}

func TestIntCredentialsLifecycle(t *testing.T) {

	db, err := app.InitDB(true)
	if err != nil {
		t.Fatalf("credentials: db connection failed [%v]", err.Error())
	}
	defer db.Close()

	credsModel := model.NewIntCredentials(db)

	testUserId := 1

	newCred := model.IntCredentialAddData{
		IntCredentialName:   "Test Integration Bot",
		IntCredentialType:   model.Discord,
		IntCredentialConfig: `{"token": "123-test-token"}`,
	}

	createdId, err := credsModel.Add(newCred, testUserId)
	if err != nil {
		t.Fatalf("credentials: add failed: %v", err)
	}

	if createdId == 0 {
		t.Fatalf("credentials: add returned 0 ID")
	}

	t.Logf("Created Credential ID: %d", createdId)

	list, err := credsModel.List(strconv.Itoa(createdId), true, testUserId)
	if err != nil {
		t.Errorf("credentials: list failed: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("credentials: expected 1 item, got %d", len(list))
	} else {

		item := list[0]
		if item.IntCredentialName != newCred.IntCredentialName {
			t.Errorf("credentials: name mismatch. Got %s, want %s", item.IntCredentialName, newCred.IntCredentialName)
		}
		if item.IntCredentialConfig == "" {
			t.Errorf("credentials: config should not be empty when includeConfig is true")
		}
	}

	updateData := model.IntCredentialUpdateData{
		IntCredentialId:     createdId,
		IntCredentialName:   "Updated Bot Name",
		IntCredentialType:   model.Telegram,
		IntCredentialConfig: `{"token": "456-new-token"}`,
	}

	rowsAffected, err := credsModel.Update(updateData, testUserId)
	if err != nil {
		t.Errorf("credentials: update failed: %v", err)
	}
	if rowsAffected != 1 {
		t.Errorf("credentials: expected 1 row updated, got %d", rowsAffected)
	}

	updatedList, _ := credsModel.List(strconv.Itoa(createdId), false, testUserId)
	if len(updatedList) > 0 {
		if updatedList[0].IntCredentialName != "Updated Bot Name" {
			t.Errorf("credentials: update name not persisted")
		}
		if updatedList[0].IntCredentialType != string(model.Telegram) {
			t.Errorf("credentials: update type not persisted")
		}

		if updatedList[0].IntCredentialConfig != "" {
			t.Errorf("credentials: expected empty config when includeConfig=false, got %s", updatedList[0].IntCredentialConfig)
		}
	}

	delRows, err := credsModel.Delete(createdId, testUserId)
	if err != nil {
		t.Errorf("credentials: delete failed: %v", err)
	}
	if delRows != 1 {
		t.Errorf("credentials: expected 1 row deleted, got %d", delRows)
	}

	finalList, _ := credsModel.List(strconv.Itoa(createdId), false, testUserId)
	if len(finalList) != 0 {
		t.Errorf("credentials: item should be deleted but was returned in list")
	}
}

func TestIntCredentialsBasicList(t *testing.T) {
	db, err := app.InitDB(true)
	if err != nil {
		t.Fatalf("credentials: db connection failed [%v]", err.Error())
	}
	defer db.Close()

	credsModel := model.NewIntCredentials(db)

	list, err := credsModel.BasicList(1)
	if err != nil {
		t.Errorf("credentials: basic list failed: %v", err)
	}

	t.Logf("Basic List returned %d items", len(list))
}
