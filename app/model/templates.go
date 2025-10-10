package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"synk/gateway/app/util"
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

type TemplatesList struct {
	TemplateId        int    `json:"template_id"`
	TemplateName      string `json:"template_name"`
	TemplateContent   string `json:"template_content"`
	TemplateUrlImport string `json:"template_url_import"`
	CreatedAt         string `json:"created_at"`
}

type TemplateAddData struct {
	TemplateName      string `json:"template_name"`
	TemplateContent   string `json:"template_content"`
	TemplateUrlImport string `json:"template_url_import"`
	UserId            int    `json:"user_id"`
	CreatedAt         string `json:"created_at"`
}

type TemplateByIdData struct {
	TemplateId int `json:"template_id"`
}

type TemplateUpdateData struct {
	TemplateId        int    `json:"template_id"`
	TemplateName      string `json:"template_name"`
	TemplateContent   string `json:"template_content"`
	TemplateUrlImport string `json:"template_url_import"`
	UpdatedAt         string `json:"updated_at"`
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

func (t *Templates) List(id string, includeContent bool) ([]TemplatesList, error) {
	var templates []TemplatesList

	whereList := []string{}
	whereValues := []any{}
	columnsList := []string{}

	if id != "" {
		whereList = append(whereList, "template_id = ?")
		whereValues = append(whereValues, id)
	}
	if includeContent {
		columnsList = append(columnsList, "template_content")
	} else {
		columnsList = append(columnsList, "'' template_content")
	}

	where := ""
	columns := ""

	if len(whereList) > 0 {
		where = " AND " + strings.Join(whereList, " AND ")
	}
	if len(columnsList) > 0 {
		columns = ", " + strings.Join(columnsList, ", ")
	}

	rows, rowsErr := t.db.Query(
		`SELECT template_id, template_name,
            template_url_import, created_at `+columns+`
        FROM template
        WHERE deleted_at IS NULL `+where, whereValues...,
	)

	if rowsErr != nil {
		return nil, fmt.Errorf("models.templates.list: %s", rowsErr.Error())
	}

	defer rows.Close()

	rowsErr = rows.Err()

	if rowsErr != nil {
		return nil, fmt.Errorf("models.templates.list: %s", rowsErr.Error())
	}

	for rows.Next() {
		var template TemplatesList
		var templateUrlImport sql.NullString

		exception := rows.Scan(
			&template.TemplateId,
			&template.TemplateName,
			&templateUrlImport,
			&template.CreatedAt,
			&template.TemplateContent,
		)

		template.TemplateUrlImport = templateUrlImport.String
		template.CreatedAt = util.ToTimeBR(template.CreatedAt)

		if exception != nil {
			return nil, fmt.Errorf("models.templates.list: %s", exception.Error())
		}

		templates = append(templates, template)
	}

	return templates, nil
}

func (t *Templates) Add(template TemplateAddData) (int, error) {
	var templateId int

	insertRes, insertErr := t.db.ExecContext(
		context.Background(),
		`INSERT INTO synk.template (template_name, template_content, template_url_import, user_id)
        VALUES (?, ?, ?, ?)`,
		template.TemplateName, template.TemplateContent, template.TemplateUrlImport, template.UserId,
	)

	if insertErr != nil {
		return templateId, fmt.Errorf("models.templates.add: %s", insertErr.Error())
	}

	id, exception := insertRes.LastInsertId()

	if exception != nil {
		return templateId, fmt.Errorf("models.templates.add: %s", exception.Error())
	}

	templateId = int(id)

	return templateId, nil
}

func (t *Templates) Update(template TemplateUpdateData) (int, error) {
	var rowsAffected int64

	updateRes, updateErr := t.db.ExecContext(
		context.Background(),
		`UPDATE template
        SET template_name = ?,
            template_content = ?,
            template_url_import = ?,
            updated_at = CURRENT_TIMESTAMP
        WHERE template_id = ? AND
              deleted_at IS NULL`,
		template.TemplateName, template.TemplateContent, template.TemplateUrlImport, template.TemplateId,
	)

	if updateErr != nil {
		return int(rowsAffected), fmt.Errorf("models.templates.update: %s", updateErr.Error())
	}

	rowsAffectedVal, exception := updateRes.RowsAffected()

	if exception != nil {
		return int(rowsAffected), fmt.Errorf("models.templates.update: %s", exception.Error())
	}

	rowsAffected = rowsAffectedVal

	return int(rowsAffected), nil
}
