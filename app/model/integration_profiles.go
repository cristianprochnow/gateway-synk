package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"synk/gateway/app/util"
)

type IntProfiles struct {
	db *sql.DB
}

type IntProfilesBasicList struct {
	IntProfileId   int    `json:"int_profile_id"`
	IntProfileName string `json:"int_profile_name"`
	ColorName      string `json:"color_name"`
	ColorHex       string `json:"color_hex"`
}

type IntProfilesByIdData struct {
	IntProfileId int `json:"int_profile_id"`
}

type IntProfileList struct {
	IntProfileId   int                       `json:"int_profile_id"`
	IntProfileName string                    `json:"int_profile_name"`
	ColorId        int                       `json:"color_id"`
	ColorName      string                    `json:"color_name"`
	ColorHex       string                    `json:"color_hex"`
	CreatedAt      string                    `json:"created_at"`
	Credentials    []IntCredentialsBasicList `json:"credentials"`
}

type IntProfileAddData struct {
	IntProfileName string `json:"int_profile_name"`
	ColorId        int    `json:"color_id"`
}

type IntProfileUpdateData struct {
	IntProfileId   int    `json:"int_profile_id"`
	IntProfileName string `json:"int_profile_name"`
	ColorId        int    `json:"color_id"`
}

func NewIntProfiles(db *sql.DB) *IntProfiles {
	intProfiles := IntProfiles{db: db}

	return &intProfiles
}

func (ip *IntProfiles) BasicList(userId int) ([]IntProfilesBasicList, error) {
	var intProfiles []IntProfilesBasicList

	rows, rowsErr := ip.db.Query(
		`SELECT profile.int_profile_id, profile.int_profile_name,
            color.color_hex, color.color_name
        FROM integration_profile profile
        LEFT JOIN color ON color.color_id = profile.color_id
        WHERE profile.deleted_at IS NULL AND profile.user_id = ?
        ORDER BY profile.int_profile_name`, userId,
	)

	if rowsErr != nil {
		return nil, fmt.Errorf("models.int_profiles.basic_list: %s", rowsErr.Error())
	}

	defer rows.Close()

	rowsErr = rows.Err()

	if rowsErr != nil {
		return nil, fmt.Errorf("models.int_profiles.basic_list: %s", rowsErr.Error())
	}

	for rows.Next() {
		var intProfile IntProfilesBasicList

		exception := rows.Scan(
			&intProfile.IntProfileId,
			&intProfile.IntProfileName,
			&intProfile.ColorHex,
			&intProfile.ColorName,
		)

		if exception != nil {
			return nil, fmt.Errorf("models.int_profiles.basic_list: %s", exception.Error())
		}

		intProfiles = append(intProfiles, intProfile)

	}

	return intProfiles, nil
}

func (ip *IntProfiles) ById(intProfileId int, userId int) (IntProfilesByIdData, error) {
	var intProfile IntProfilesByIdData

	rows, rowsErr := ip.db.Query(
		`SELECT int_profile_id
        FROM integration_profile
        WHERE deleted_at IS NULL AND user_id = ? AND int_profile_id = ?`,
		userId, intProfileId,
	)

	if rowsErr != nil {
		return intProfile, fmt.Errorf("models.int_profile.by_id: %s", rowsErr.Error())
	}

	defer rows.Close()

	rowsErr = rows.Err()

	if rowsErr != nil {
		return intProfile, fmt.Errorf("models.int_profile.by_id: %s", rowsErr.Error())
	}

	for rows.Next() {
		exception := rows.Scan(
			&intProfile.IntProfileId,
		)

		if exception != nil {
			return intProfile, fmt.Errorf("models.int_profile.by_id: %s", exception.Error())
		}
	}

	return intProfile, nil
}

func (ip *IntProfiles) List(id string, userId int) ([]IntProfileList, error) {
	var intProfiles []IntProfileList

	whereList := []string{}
	whereValues := []any{}

	whereList = append(whereList, "profile.user_id = ?")
	whereValues = append(whereValues, userId)

	if id != "" {
		whereList = append(whereList, "profile.int_profile_id = ?")
		whereValues = append(whereValues, id)
	}

	where := ""

	if len(whereList) > 0 {
		where = " AND " + strings.Join(whereList, " AND ")
	}

	rows, rowsErr := ip.db.Query(
		`SELECT profile.int_profile_id, profile.int_profile_name, color.color_id,
               color.color_name, color.color_hex, profile.created_at
        FROM integration_profile profile
        LEFT JOIN color ON color.color_id = profile.color_id
        WHERE profile.deleted_at IS NULL `+where, whereValues...,
	)

	if rowsErr != nil {
		return nil, fmt.Errorf("models.integration_profiles.list: %s", rowsErr.Error())
	}

	defer rows.Close()

	rowsErr = rows.Err()

	if rowsErr != nil {
		return nil, fmt.Errorf("models.integration_profiles.list: %s", rowsErr.Error())
	}

	for rows.Next() {
		var intProfile IntProfileList

		exception := rows.Scan(
			&intProfile.IntProfileId,
			&intProfile.IntProfileName,
			&intProfile.ColorId,
			&intProfile.ColorName,
			&intProfile.ColorHex,
			&intProfile.CreatedAt,
		)

		if exception != nil {
			return nil, fmt.Errorf("models.integration_profiles.list: %s", exception.Error())
		}

		intProfile.CreatedAt = util.ToTimeBR(intProfile.CreatedAt)

		intProfiles = append(intProfiles, intProfile)
	}

	return intProfiles, nil
}

func (ip *IntProfiles) Add(intProfile IntProfileAddData, intCredentials []int, userId int) (int, error) {
	var intProfileId int

	insertRes, insertErr := ip.db.ExecContext(
		context.Background(),
		`INSERT INTO synk.integration_profile (int_profile_name, color_id, user_id)
        VALUES (?, ?, ?)`,
		intProfile.IntProfileName, intProfile.ColorId, userId,
	)

	if insertErr != nil {
		return intProfileId, fmt.Errorf("models.integration_profiles.add: %s", insertErr.Error())
	}

	id, exception := insertRes.LastInsertId()

	if exception != nil {
		return intProfileId, fmt.Errorf("models.integration_profiles.add: %s", exception.Error())
	}

	intProfileId = int(id)

	for _, credentialId := range intCredentials {
		ip.db.ExecContext(
			context.Background(),
			`INSERT INTO synk.integration_group (int_profile_id, int_credential_id)
            VALUES (?, ?)`,
			intProfileId, credentialId,
		)
	}

	return intProfileId, nil
}

func (ip *IntProfiles) Update(intProfile IntProfileUpdateData, intCredentials []int, userId int) (int, error) {
	var rowsAffected int64

	updateRes, updateErr := ip.db.ExecContext(
		context.Background(),
		`UPDATE integration_profile
        SET int_profile_name = ?,
            color_id = ?,
            updated_at = CURRENT_TIMESTAMP
        WHERE deleted_at IS NULL AND user_id = ? AND int_profile_id = ?`,
		intProfile.IntProfileName, intProfile.ColorId, userId, intProfile.IntProfileId,
	)

	if updateErr != nil {
		return int(rowsAffected), fmt.Errorf("models.integration_profiles.update: %s", updateErr.Error())
	}

	rowsAffectedVal, exception := updateRes.RowsAffected()

	if exception != nil {
		return int(rowsAffected), fmt.Errorf("models.integration_profiles.update: %s", exception.Error())
	}

	ip.db.ExecContext(
		context.Background(),
		`DELETE FROM integration_group WHERE int_profile_id = ?`, intProfile.IntProfileId,
	)

	for _, credentialId := range intCredentials {
		ip.db.ExecContext(
			context.Background(),
			`INSERT INTO synk.integration_group (int_profile_id, int_credential_id)
            VALUES (?, ?)`,
			intProfile.IntProfileId, credentialId,
		)
	}

	rowsAffected = rowsAffectedVal

	return int(rowsAffected), nil
}

func (ip *IntProfiles) Delete(intProfileId int, userId int) (int, error) {
	var rowsAffected int64

	insertRes, insertErr := ip.db.ExecContext(
		context.Background(),
		`UPDATE integration_profile
        SET deleted_at = CURRENT_TIMESTAMP
        WHERE deleted_at IS NULL AND user_id = ? AND int_profile_id = ?`,
		userId, intProfileId,
	)

	if insertErr != nil {
		return int(rowsAffected), fmt.Errorf("models.integration_profiles.delete: %s", insertErr.Error())
	}

	rowsAffectedVal, exception := insertRes.RowsAffected()

	if exception != nil {
		return int(rowsAffected), fmt.Errorf("models.integration_profiles.delete: %s", exception.Error())
	}

	rowsAffected = rowsAffectedVal

	return int(rowsAffected), nil
}
