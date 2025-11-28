package tests

import (
	"context"
	"synk/gateway/app"
	"synk/gateway/app/model"
	"testing"
)

func TestPublicationCountByPost(t *testing.T) {
	db, err := app.InitDB(true)
	if err != nil {
		t.Fatalf("publication: db connection failed [%v]", err.Error())
	}
	defer db.Close()

	pubModel := model.NewPublication(db)

	var templateId, intProfileId, intCredentialId int

	err = db.QueryRow("SELECT template_id FROM template WHERE deleted_at IS NULL LIMIT 1").Scan(&templateId)
	if err != nil {
		t.Fatalf("publication: valid 'template_id' required. %v", err)
	}

	err = db.QueryRow("SELECT int_profile_id FROM integration_profile WHERE deleted_at IS NULL LIMIT 1").Scan(&intProfileId)
	if err != nil {
		t.Fatalf("publication: valid 'int_profile_id' required. %v", err)
	}

	err = db.QueryRow("SELECT int_credential_id FROM integration_credential WHERE deleted_at IS NULL LIMIT 1").Scan(&intCredentialId)
	if err != nil {
		t.Fatalf("publication: valid 'int_credential_id' required. %v", err)
	}

	res, err := db.Exec(`INSERT INTO post (post_name, post_content, template_id, int_profile_id, user_id)
						 VALUES ('Pub Test Post', 'Dummy Content', ?, ?, 1)`, templateId, intProfileId)
	if err != nil {
		t.Fatalf("publication: could not create dummy post for testing: %v", err)
	}
	postIdInt64, _ := res.LastInsertId()
	postId := int(postIdInt64)

	defer db.Exec("DELETE FROM post WHERE post_id = ?", postId)

	queries := []string{
		"INSERT INTO publication (post_id, int_credential_id, publication_status) VALUES (?, ?, 'failed')",
		"INSERT INTO publication (post_id, int_credential_id, publication_status) VALUES (?, ?, 'failed')",
		"INSERT INTO publication (post_id, int_credential_id, publication_status) VALUES (?, ?, 'pending')",
		"INSERT INTO publication (post_id, int_credential_id, publication_status) VALUES (?, ?, 'published')",
	}

	for _, q := range queries {
		_, err := db.ExecContext(context.Background(), q, postId, intCredentialId)
		if err != nil {
			t.Fatalf("publication: failed to insert test data: %v", err)
		}
	}
	defer db.Exec("DELETE FROM publication WHERE post_id = ?", postId)

	counts, err := pubModel.CountByPost(postId)
	if err != nil {
		t.Fatalf("publication: CountByPost failed: %v", err)
	}

	if counts[model.PublicationStatusFailed] != 2 {
		t.Errorf("publication: expected 2 failed, got %d", counts[model.PublicationStatusFailed])
	}
	if counts[model.PublicationStatusPending] != 1 {
		t.Errorf("publication: expected 1 pending, got %d", counts[model.PublicationStatusPending])
	}
	if counts[model.PublicationStatusPublished] != 1 {
		t.Errorf("publication: expected 1 published, got %d", counts[model.PublicationStatusPublished])
	}
}
