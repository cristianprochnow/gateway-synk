package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"synk/gateway/app/util"
)

type Posts struct {
	db *sql.DB
}

type PostsList struct {
	PostId         int               `json:"post_id"`
	PostName       string            `json:"post_name"`
	TemplateName   string            `json:"template_name"`
	IntProfileName string            `json:"int_profile_name"`
	CreatedAt      string            `json:"created_at"`
	Status         PublicationStatus `json:"status"`
}

type PostAddData struct {
	PostName     string `json:"post_name"`
	PostContent  string `json:"post_content"`
	TemplateId   int    `json:"template_id"`
	IntProfileId int    `json:"int_profile_id"`
	CreatedAt    string `json:"created_at"`
}

type PostUpdateData struct {
	PostId       int    `json:"post_id"`
	PostName     string `json:"post_name"`
	PostContent  string `json:"post_content"`
	TemplateId   int    `json:"template_id"`
	IntProfileId int    `json:"int_profile_id"`
	UpdatedAt    string `json:"updated_at"`
}

type PostByIdData struct {
	PostId int `json:"post_id"`
}

func NewPosts(db *sql.DB) *Posts {
	posts := Posts{db: db}

	return &posts
}

func (p *Posts) List(id string) ([]PostsList, error) {
	var posts []PostsList

	whereList := []string{}
	whereValues := []any{}

	if id != "" {
		whereList = append(whereList, "post_id = ?")
		whereValues = append(whereValues, id)
	}

	where := ""

	if len(whereList) > 0 {
		where = " AND " + strings.Join(whereList, " AND ")
	}

	rows, rowsErr := p.db.Query(
		`SELECT post.post_id, post.post_name, template.template_name,
                int_profile.int_profile_name, post.created_at, "" status
        FROM post
        LEFT JOIN template ON template.template_id = post.template_id
        LEFT JOIN integration_profile int_profile ON int_profile.int_profile_id = post.int_profile_id
        WHERE post.deleted_at IS NULL `+where+`
        ORDER BY post.created_at DESC`, whereValues...,
	)

	if rowsErr != nil {
		return nil, fmt.Errorf("models.posts.list: %s", rowsErr.Error())
	}

	defer rows.Close()

	rowsErr = rows.Err()

	if rowsErr != nil {
		return nil, fmt.Errorf("models.posts.list: %s", rowsErr.Error())
	}

	publicationModel := NewPublication(p.db)

	for rows.Next() {
		var post PostsList

		exception := rows.Scan(
			&post.PostId,
			&post.PostName,
			&post.TemplateName,
			&post.IntProfileName,
			&post.CreatedAt,
			&post.Status,
		)

		post.CreatedAt = util.ToTimeBR(post.CreatedAt)

		if exception != nil {
			return nil, fmt.Errorf("models.posts.list: %s", exception.Error())
		}

		statusCount, statusCountErr := publicationModel.CountByPost(post.PostId)

		if statusCountErr != nil {
			return nil, fmt.Errorf("models.posts.list: %s", statusCountErr.Error())
		}

		if statusCount[PublicationStatusFailed] > 0 {
			post.Status = PublicationStatusFailed
		} else if statusCount[PublicationStatusPending] > 0 {
			post.Status = PublicationStatusPending
		} else {
			post.Status = PublicationStatusPublished
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func (p *Posts) Add(post PostAddData) (int, error) {
	var postId int

	insertRes, insertErr := p.db.ExecContext(
		context.Background(),
		`INSERT INTO synk.post (post_name, post_content, template_id, int_profile_id)
        VALUES (?, ?, ?, ?)`,
		post.PostName, post.PostContent, post.TemplateId, post.IntProfileId,
	)

	if insertErr != nil {
		return postId, fmt.Errorf("models.posts.add: %s", insertErr.Error())
	}

	id, exception := insertRes.LastInsertId()

	if exception != nil {
		return postId, fmt.Errorf("models.posts.add: %s", exception.Error())
	}

	postId = int(id)

	return postId, nil
}

func (p *Posts) Update(post PostUpdateData) (int, error) {
	var rowsAffected int64

	insertRes, insertErr := p.db.ExecContext(
		context.Background(),
		`UPDATE post
        SET post_name = ?,
            post_content = ?,
            template_id = ?,
            int_profile_id = ?,
            updated_at = CURRENT_TIMESTAMP
        WHERE post_id = ? AND deleted_at IS NULL`,
		post.PostName, post.PostContent, post.TemplateId, post.IntProfileId, post.PostId,
	)

	if insertErr != nil {
		return int(rowsAffected), fmt.Errorf("models.posts.update: %s", insertErr.Error())
	}

	rowsAffectedVal, exception := insertRes.RowsAffected()

	if exception != nil {
		return int(rowsAffected), fmt.Errorf("models.posts.update: %s", exception.Error())
	}

	rowsAffected = rowsAffectedVal

	return int(rowsAffected), nil
}

func (p *Posts) Delete(postId int) (int, error) {
	var rowsAffected int64

	insertRes, insertErr := p.db.ExecContext(
		context.Background(),
		`UPDATE post
        SET deleted_at = CURRENT_TIMESTAMP
        WHERE post_id = ?`, postId,
	)

	if insertErr != nil {
		return int(rowsAffected), fmt.Errorf("models.posts.delete: %s", insertErr.Error())
	}

	rowsAffectedVal, exception := insertRes.RowsAffected()

	if exception != nil {
		return int(rowsAffected), fmt.Errorf("models.posts.delete: %s", exception.Error())
	}

	rowsAffected = rowsAffectedVal

	return int(rowsAffected), nil
}

func (p *Posts) ById(postId int) (PostByIdData, error) {
	var post PostByIdData

	rows, rowsErr := p.db.Query(
		`SELECT post_id
        FROM post
        WHERE post_id = ? AND deleted_at IS NULL`, postId,
	)

	if rowsErr != nil {
		return post, fmt.Errorf("models.posts.by_id: %s", rowsErr.Error())
	}

	defer rows.Close()

	rowsErr = rows.Err()

	if rowsErr != nil {
		return post, fmt.Errorf("models.posts.by_id: %s", rowsErr.Error())
	}

	for rows.Next() {
		exception := rows.Scan(
			&post.PostId,
		)

		if exception != nil {
			return post, fmt.Errorf("models.posts.by_id: %s", exception.Error())
		}
	}

	return post, nil
}
