package tests

import (
	"context"
	"database/sql"
	"strconv"
	"synk/gateway/app"
	"synk/gateway/app/model"
	"testing"
)

func setupTemplatesDB(t *testing.T) (*sql.DB, int) {
	db, err := app.InitDB(true)
	if err != nil {
		t.Fatalf("db connection failed: %v", err)
	}
	return db, 1
}

func TestNewTemplates(t *testing.T) {
	db, _ := setupTemplatesDB(t)
	defer db.Close()

	tpl := model.NewTemplates(db)
	if tpl == nil {
		t.Error("NewTemplates returned nil")
	}
}

func TestTemplates_Add(t *testing.T) {
	db, userId := setupTemplatesDB(t)
	defer db.Close()
	tplModel := model.NewTemplates(db)

	input := model.TemplateAddData{
		TemplateName:      "Test Add",
		TemplateContent:   "<p>Content</p>",
		TemplateUrlImport: "http://test.com",
		UserId:            userId,
	}

	id, err := tplModel.Add(input)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if id == 0 {
		t.Fatal("Add returned ID 0")
	}

	defer db.Exec("DELETE FROM template WHERE template_id = ?", id)
}

func TestTemplates_Update(t *testing.T) {
	db, userId := setupTemplatesDB(t)
	defer db.Close()
	tplModel := model.NewTemplates(db)

	createInput := model.TemplateAddData{
		TemplateName:    "Pre-Update",
		TemplateContent: "Old",
		UserId:          userId,
	}
	id, err := tplModel.Add(createInput)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer db.Exec("DELETE FROM template WHERE template_id = ?", id)

	updateInput := model.TemplateUpdateData{
		TemplateId:        id,
		TemplateName:      "Post-Update",
		TemplateContent:   "New",
		TemplateUrlImport: "http://updated.com",
	}

	rows, err := tplModel.Update(updateInput, userId)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if rows != 1 {
		t.Errorf("Expected 1 row updated, got %d", rows)
	}

	list, _ := tplModel.List(strconv.Itoa(id), true, userId)
	if len(list) > 0 {
		if list[0].TemplateName != "Post-Update" {
			t.Errorf("Update name not persisted. Got %s", list[0].TemplateName)
		}
		if list[0].TemplateContent != "New" {
			t.Errorf("Update content not persisted. Got %s", list[0].TemplateContent)
		}
	}
}

func TestTemplates_Delete(t *testing.T) {
	db, userId := setupTemplatesDB(t)
	defer db.Close()
	tplModel := model.NewTemplates(db)

	createInput := model.TemplateAddData{TemplateName: "To Delete", UserId: userId}
	id, _ := tplModel.Add(createInput)

	defer db.Exec("DELETE FROM template WHERE template_id = ?", id)

	rows, err := tplModel.Delete(id, userId)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if rows != 1 {
		t.Errorf("Expected 1 row deleted, got %d", rows)
	}

	list, _ := tplModel.List(strconv.Itoa(id), false, userId)
	if len(list) != 0 {
		t.Error("Item returned after soft delete")
	}
}

func TestTemplates_List(t *testing.T) {
	db, userId := setupTemplatesDB(t)
	defer db.Close()
	tplModel := model.NewTemplates(db)

	createInput := model.TemplateAddData{
		TemplateName:    "List Test",
		TemplateContent: "Secret Content",
		UserId:          userId,
	}
	id, _ := tplModel.Add(createInput)
	defer db.Exec("DELETE FROM template WHERE template_id = ?", id)

	listFull, err := tplModel.List(strconv.Itoa(id), true, userId)
	if err != nil {
		t.Fatalf("List (full) failed: %v", err)
	}
	if len(listFull) == 0 {
		t.Fatal("List returned 0 items")
	}
	if listFull[0].TemplateContent != "Secret Content" {
		t.Error("Expected content to be present")
	}

	listShort, _ := tplModel.List(strconv.Itoa(id), false, userId)
	if len(listShort) > 0 {
		if listShort[0].TemplateContent != "" {
			t.Error("Expected content to be empty string")
		}
	}

	listAll, _ := tplModel.List("", false, userId)
	if len(listAll) == 0 {
		t.Error("List all returned 0 items")
	}
}

func TestTemplates_BasicList(t *testing.T) {
	db, userId := setupTemplatesDB(t)
	defer db.Close()
	tplModel := model.NewTemplates(db)

	createInput := model.TemplateAddData{TemplateName: "Basic List Test", UserId: userId}
	id, _ := tplModel.Add(createInput)
	defer db.Exec("DELETE FROM template WHERE template_id = ?", id)

	list, err := tplModel.BasicList(userId)
	if err != nil {
		t.Fatalf("BasicList failed: %v", err)
	}

	found := false
	for _, tpl := range list {
		if tpl.TemplateId == id {
			found = true
			if tpl.TemplateName != "Basic List Test" {
				t.Error("BasicList returned wrong name")
			}
			break
		}
	}
	if !found {
		t.Error("Created template not found in BasicList")
	}
}

func TestTemplates_ById(t *testing.T) {
	db, userId := setupTemplatesDB(t)
	defer db.Close()
	tplModel := model.NewTemplates(db)

	createInput := model.TemplateAddData{TemplateName: "ById Test", UserId: userId}
	id, _ := tplModel.Add(createInput)
	defer db.Exec("DELETE FROM template WHERE template_id = ?", id)

	data, err := tplModel.ById(id, userId)
	if err != nil {
		t.Fatalf("ById failed: %v", err)
	}

	if data.TemplateId != id {
		t.Errorf("ById returned wrong ID. Got %d, want %d", data.TemplateId, id)
	}
}

func TestTemplateLifecycle(t *testing.T) {
	db, userId := setupTemplatesDB(t)
	defer db.Close()
	tplModel := model.NewTemplates(db)

	newTpl := model.TemplateAddData{
		TemplateName:      "Lifecycle Test",
		TemplateContent:   "<html>...</html>",
		TemplateUrlImport: "http://lifecycle.com",
		UserId:            userId,
	}
	createdId, err := tplModel.Add(newTpl)
	if err != nil {
		t.Fatalf("lifecycle: add failed: %v", err)
	}

	defer db.ExecContext(context.Background(), "DELETE FROM template WHERE template_id = ?", createdId)

	list, _ := tplModel.List(strconv.Itoa(createdId), true, userId)
	if len(list) != 1 {
		t.Error("lifecycle: creation check failed")
	}

	updateData := model.TemplateUpdateData{
		TemplateId:        createdId,
		TemplateName:      "Lifecycle Updated",
		TemplateContent:   "New",
		TemplateUrlImport: "",
	}
	tplModel.Update(updateData, userId)

	tplModel.Delete(createdId, userId)

	finalList, _ := tplModel.List(strconv.Itoa(createdId), false, userId)
	if len(finalList) != 0 {
		t.Error("lifecycle: soft delete check failed")
	}
}
