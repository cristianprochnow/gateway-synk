package model

import (
	"database/sql"
	"fmt"
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

func NewPosts(db *sql.DB) *Posts {
	posts := Posts{db: db}

	return &posts
}

func (p *Posts) List() ([]PostsList, error) {
	var posts []PostsList

	rows, rowsErr := p.db.Query(
		`SELECT post.post_id, post.post_name, template.template_name,
                int_profile.int_profile_name, post.created_at, NULL status
        FROM post
        LEFT JOIN template ON template.template_id = post.template_id
        LEFT JOIN integration_profile int_profile ON int_profile.int_profile_id = post.int_profile_id
        ORDER BY post.created_at DESC`,
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
		)

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
