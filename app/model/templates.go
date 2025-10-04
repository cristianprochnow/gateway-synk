package model

import (
	"database/sql"
	"fmt"
)

type Templates struct {
	db *sql.DB
}

type TemplatesBasicList struct {
	TemplateId   int    `json:"template_id"`
	TemplateName string `json:"template_name"`
}

type TemplatesByIdData struct {
	TemplateId int `json:"template_id"`
}

func NewTemplates(db *sql.DB) *Templates {
	templates := Templates{db: db}

	return &templates
}

func (t *Templates) BasicList() ([]TemplatesBasicList, error) {
	var templates []TemplatesBasicList

	rows, rowsErr := t.db.Query(
		`SELECT template_id, template_name
        FROM template
        ORDER BY template_name`,
	)

	if rowsErr != nil {
		return nil, fmt.Errorf("models.templates.basic_list: %s", rowsErr.Error())
	}

	defer rows.Close()

	rowsErr = rows.Err()

	if rowsErr != nil {
		return nil, fmt.Errorf("models.templates.basic_list: %s", rowsErr.Error())
	}

	for rows.Next() {
		var template TemplatesBasicList

		exception := rows.Scan(
			&template.TemplateId,
			&template.TemplateName,
		)

		if exception != nil {
			return nil, fmt.Errorf("models.templates.basic_list: %s", exception.Error())
		}

		templates = append(templates, template)

	}

	return templates, nil
}

func (t *Templates) ById(templateId int) (TemplatesByIdData, error) {
	var template TemplatesByIdData

	rows, rowsErr := t.db.Query(
		`SELECT template_id
        FROM template
        WHERE template_id = ?`, templateId,
	)

	if rowsErr != nil {
		return template, fmt.Errorf("models.templates.by_id: %s", rowsErr.Error())
	}

	defer rows.Close()

	rowsErr = rows.Err()

	if rowsErr != nil {
		return template, fmt.Errorf("models.templates.by_id: %s", rowsErr.Error())
	}

	for rows.Next() {
		exception := rows.Scan(
			&template.TemplateId,
		)

		if exception != nil {
			return template, fmt.Errorf("models.templates.by_id: %s", exception.Error())
		}
	}

	return template, nil
}
