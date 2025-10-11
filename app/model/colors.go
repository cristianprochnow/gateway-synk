package model

import (
	"database/sql"
	"fmt"
	"strings"
)

type Colors struct {
	db *sql.DB
}

type ColorsList struct {
	ColorId   int    `json:"color_id"`
	ColorName string `json:"color_name"`
	ColorHex  string `json:"color_hex"`
}

func NewColors(db *sql.DB) *Colors {
	colors := Colors{db: db}

	return &colors
}

func (c *Colors) List(id int) ([]ColorsList, error) {
	var colors []ColorsList

	whereList := []string{}
	whereValues := []any{}

	if id != 0 {
		whereList = append(whereList, "color_id = ?")
		whereValues = append(whereValues, id)
	}

	where := ""

	if len(whereList) > 0 {
		where = " WHERE " + strings.Join(whereList, " AND ")
	}

	rows, rowsErr := c.db.Query(
		`SELECT color_id, color_name, color_hex
        FROM color `+where, whereValues...,
	)

	if rowsErr != nil {
		return nil, fmt.Errorf("models.colors.list: %s", rowsErr.Error())
	}

	defer rows.Close()

	rowsErr = rows.Err()

	if rowsErr != nil {
		return nil, fmt.Errorf("models.colors.list: %s", rowsErr.Error())
	}

	for rows.Next() {
		var color ColorsList

		exception := rows.Scan(
			&color.ColorId,
			&color.ColorName,
			&color.ColorHex,
		)

		if exception != nil {
			return nil, fmt.Errorf("models.colors.list: %s", exception.Error())
		}

		colors = append(colors, color)
	}

	return colors, nil
}
