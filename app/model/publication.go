package model

import (
	"database/sql"
	"fmt"
)

type PublicationStatus string

const (
	PublicationStatusPending   PublicationStatus = "pending"
	PublicationStatusFailed    PublicationStatus = "failed"
	PublicationStatusPublished PublicationStatus = "published"
)

type PublicationStatusCount struct {
	Total  int
	Status PublicationStatus
}

type Publication struct {
	db *sql.DB
}

func NewPublication(db *sql.DB) *Publication {
	publication := Publication{db: db}

	return &publication
}

func (p *Publication) CountByPost(postId int) (map[PublicationStatus]int, error) {
	posts := map[PublicationStatus]int{}

	rows, rowsErr := p.db.Query(
		`SELECT COUNT(*) total, publication.publication_status status
        FROM publication
        WHERE publication.post_id = ?
        GROUP BY publication.publication_status`, postId,
	)

	if rowsErr != nil {
		return nil, fmt.Errorf("models.publication.listByPost: %s", rowsErr.Error())
	}

	defer rows.Close()

	rowsErr = rows.Err()

	if rowsErr != nil {
		return nil, fmt.Errorf("models.publication.listByPost: %s", rowsErr.Error())
	}

	for rows.Next() {
		var post PublicationStatusCount

		exception := rows.Scan(
			&post.Total,
			&post.Status,
		)

		if exception != nil {
			return nil, fmt.Errorf("models.publication.listByPost: %s", exception.Error())
		}

		posts[post.Status] = post.Total
	}

	return posts, nil
}
