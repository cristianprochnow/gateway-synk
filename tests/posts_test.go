package tests

import (
	"context"
	"database/sql"
	"strconv"
	"synk/gateway/app"
	"synk/gateway/app/model"
	"testing"
)

func setupPostsDB(t *testing.T) (*sql.DB, int) {
	db, err := app.InitDB(true)
	if err != nil {
		t.Fatalf("db connection failed: %v", err)
	}
	return db, 1
}

func createDummyTemplate(t *testing.T, db *sql.DB, userId int) int {
	res, err := db.ExecContext(context.Background(),
		`INSERT INTO template (template_name, template_content, template_url_import, user_id)
		 VALUES ('Post Test Template', '<p>Test</p>', '', ?)`, userId)
	if err != nil {
		t.Fatalf("Setup failed: Could not create dummy template: %v", err)
	}
	id, _ := res.LastInsertId()
	return int(id)
}

func getValidProfileId(t *testing.T, db *sql.DB) int {
	var id int
	err := db.QueryRow("SELECT int_profile_id FROM integration_profile WHERE deleted_at IS NULL LIMIT 1").Scan(&id)
	if err != nil {
		t.Fatalf("Setup failed: valid 'integration_profile' required. %v", err)
	}
	return id
}

func getValidCredentialId(t *testing.T, db *sql.DB) int {
	var id int
	err := db.QueryRow("SELECT int_credential_id FROM integration_credential WHERE deleted_at IS NULL LIMIT 1").Scan(&id)
	if err != nil {
		t.Fatalf("Setup failed: valid 'integration_credential' required. %v", err)
	}
	return id
}

func TestNewPosts(t *testing.T) {
	db, _ := setupPostsDB(t)
	defer db.Close()

	p := model.NewPosts(db)
	if p == nil {
		t.Error("NewPosts returned nil")
	}
}

func TestPosts_Add(t *testing.T) {
	db, userId := setupPostsDB(t)
	defer db.Close()
	postsModel := model.NewPosts(db)

	tplId := createDummyTemplate(t, db, userId)
	defer db.Exec("DELETE FROM template WHERE template_id = ?", tplId)
	profileId := getValidProfileId(t, db)

	input := model.PostAddData{
		PostName:     "Test Add Post",
		PostContent:  "Content",
		TemplateId:   tplId,
		IntProfileId: profileId,
	}

	id, err := postsModel.Add(input, userId)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if id == 0 {
		t.Fatal("Add returned ID 0")
	}

	defer db.Exec("DELETE FROM post WHERE post_id = ?", id)
}

func TestPosts_Update(t *testing.T) {
	db, userId := setupPostsDB(t)
	defer db.Close()
	postsModel := model.NewPosts(db)

	tplId := createDummyTemplate(t, db, userId)
	defer db.Exec("DELETE FROM template WHERE template_id = ?", tplId)
	profileId := getValidProfileId(t, db)

	createInput := model.PostAddData{
		PostName:     "Pre-Update",
		TemplateId:   tplId,
		IntProfileId: profileId,
	}
	id, err := postsModel.Add(createInput, userId)
	if err != nil {
		t.Fatalf("Setup failed (Add): %v", err)
	}
	defer db.Exec("DELETE FROM post WHERE post_id = ?", id)

	updateInput := model.PostUpdateData{
		PostId:       id,
		PostName:     "Post-Update",
		PostContent:  "New Content",
		TemplateId:   tplId,
		IntProfileId: profileId,
	}

	rows, err := postsModel.Update(updateInput, userId)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if rows != 1 {
		t.Errorf("Expected 1 row updated, got %d", rows)
	}

	list, _ := postsModel.List(strconv.Itoa(id), true, userId)
	if len(list) > 0 {
		if list[0].PostName != "Post-Update" {
			t.Errorf("Name not updated. Got %s", list[0].PostName)
		}
		if list[0].PostContent != "New Content" {
			t.Errorf("Content not updated. Got %s", list[0].PostContent)
		}
	}
}

func TestPosts_Delete(t *testing.T) {
	db, userId := setupPostsDB(t)
	defer db.Close()
	postsModel := model.NewPosts(db)

	tplId := createDummyTemplate(t, db, userId)
	defer db.Exec("DELETE FROM template WHERE template_id = ?", tplId)
	profileId := getValidProfileId(t, db)

	createInput := model.PostAddData{PostName: "To Delete", TemplateId: tplId, IntProfileId: profileId}
	id, _ := postsModel.Add(createInput, userId)

	defer db.Exec("DELETE FROM post WHERE post_id = ?", id)

	rows, err := postsModel.Delete(id, userId)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if rows != 1 {
		t.Errorf("Expected 1 row deleted, got %d", rows)
	}

	list, _ := postsModel.List(strconv.Itoa(id), false, userId)
	if len(list) != 0 {
		t.Error("Post returned after soft delete")
	}
}

func TestPosts_ById(t *testing.T) {
	db, userId := setupPostsDB(t)
	defer db.Close()
	postsModel := model.NewPosts(db)

	tplId := createDummyTemplate(t, db, userId)
	defer db.Exec("DELETE FROM template WHERE template_id = ?", tplId)
	profileId := getValidProfileId(t, db)

	createInput := model.PostAddData{PostName: "ById Test", TemplateId: tplId, IntProfileId: profileId}
	id, _ := postsModel.Add(createInput, userId)
	defer db.Exec("DELETE FROM post WHERE post_id = ?", id)

	data, err := postsModel.ById(id, userId)
	if err != nil {
		t.Fatalf("ById failed: %v", err)
	}

	if data.PostId != id {
		t.Errorf("ById returned wrong ID. Got %d, want %d", data.PostId, id)
	}
}

func TestPosts_List(t *testing.T) {
	db, userId := setupPostsDB(t)
	defer db.Close()
	postsModel := model.NewPosts(db)

	tplId := createDummyTemplate(t, db, userId)
	defer db.Exec("DELETE FROM template WHERE template_id = ?", tplId)
	profileId := getValidProfileId(t, db)
	credId := getValidCredentialId(t, db)

	createInput := model.PostAddData{
		PostName:     "List Test",
		PostContent:  "Hidden Content",
		TemplateId:   tplId,
		IntProfileId: profileId,
	}
	id, _ := postsModel.Add(createInput, userId)

	defer db.Exec("DELETE FROM post WHERE post_id = ?", id)
	defer db.Exec("DELETE FROM publication WHERE post_id = ?", id)

	listFull, err := postsModel.List(strconv.Itoa(id), true, userId)
	if err != nil {
		t.Fatalf("List (full) failed: %v", err)
	}
	if len(listFull) == 0 {
		t.Fatal("List returned 0 items")
	}
	if listFull[0].PostContent != "Hidden Content" {
		t.Error("Expected content to be present")
	}

	if listFull[0].Status != model.PublicationStatusPublished {
		t.Errorf("Expected default status 'published', got %s", listFull[0].Status)
	}

	listShort, _ := postsModel.List(strconv.Itoa(id), false, userId)
	if len(listShort) > 0 && listShort[0].PostContent != "" {
		t.Error("Expected content to be empty")
	}

	_, err = db.ExecContext(context.Background(),
		"INSERT INTO publication (post_id, int_credential_id, publication_status) VALUES (?, ?, 'failed')",
		id, credId)
	if err != nil {
		t.Fatalf("Failed to inject publication: %v", err)
	}

	listStatus, _ := postsModel.List(strconv.Itoa(id), false, userId)
	if len(listStatus) > 0 {
		if listStatus[0].Status != model.PublicationStatusFailed {
			t.Errorf("Expected status 'failed', got %s", listStatus[0].Status)
		}
	}
}

func TestPostLifecycleWithDependencies(t *testing.T) {
	db, err := app.InitDB(true)
	if err != nil {
		t.Fatalf("db connection failed [%v]", err.Error())
	}
	defer db.Close()

	postsModel := model.NewPosts(db)
	testUserId := 1

	tplId := createDummyTemplate(t, db, testUserId)
	defer db.Exec("DELETE FROM template WHERE template_id = ?", tplId)
	profileId := getValidProfileId(t, db)
	credId := getValidCredentialId(t, db)

	newPost := model.PostAddData{
		PostName:     "Lifecycle Post",
		PostContent:  "Content",
		TemplateId:   tplId,
		IntProfileId: profileId,
	}
	postId, _ := postsModel.Add(newPost, testUserId)

	defer db.Exec("DELETE FROM post WHERE post_id = ?", postId)
	defer db.Exec("DELETE FROM publication WHERE post_id = ?", postId)

	list, _ := postsModel.List(strconv.Itoa(postId), false, testUserId)
	if len(list) > 0 && list[0].Status != model.PublicationStatusPublished {
		t.Errorf("Expected default published")
	}

	db.ExecContext(context.Background(),
		"INSERT INTO publication (post_id, int_credential_id, publication_status) VALUES (?, ?, 'pending')",
		postId, credId)

	listPending, _ := postsModel.List(strconv.Itoa(postId), false, testUserId)
	if len(listPending) > 0 && listPending[0].Status != model.PublicationStatusPending {
		t.Errorf("Expected pending")
	}

	db.ExecContext(context.Background(),
		"INSERT INTO publication (post_id, int_credential_id, publication_status) VALUES (?, ?, 'failed')",
		postId, credId)

	listFailed, _ := postsModel.List(strconv.Itoa(postId), false, testUserId)
	if len(listFailed) > 0 && listFailed[0].Status != model.PublicationStatusFailed {
		t.Errorf("Expected failed")
	}
}
