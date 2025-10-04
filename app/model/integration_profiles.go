package model

import (
	"database/sql"
	"fmt"
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

func NewIntProfiles(db *sql.DB) *IntProfiles {
	intProfiles := IntProfiles{db: db}

	return &intProfiles
}

func (ip *IntProfiles) BasicList() ([]IntProfilesBasicList, error) {
	var intProfiles []IntProfilesBasicList

	rows, rowsErr := ip.db.Query(
		`SELECT profile.int_profile_id, profile.int_profile_name,
            color.color_hex, color.color_name
        FROM integration_profile profile
        LEFT JOIN color ON color.color_id = profile.color_id
        ORDER BY profile.int_profile_name`,
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

func (ip *IntProfiles) ById(intProfileId int) (IntProfilesByIdData, error) {
	var intProfile IntProfilesByIdData

	rows, rowsErr := ip.db.Query(
		`SELECT int_profile_id
        FROM integration_profile
        WHERE int_profile_id = ?`, intProfileId,
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
