package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"synk/gateway/app/util"
)

type SocialPlatform string

const (
	Twitter  SocialPlatform = "twitter"
	LinkedIn SocialPlatform = "linkedin"
)

func (sp SocialPlatform) IsValid() bool {
	switch sp {
	case Twitter, LinkedIn:
		return true
	default:
		return false
	}
}

type IntCredentials struct {
	db *sql.DB
}

type IntCredentialsBasicList struct {
	IntCredentialId   int    `json:"int_credential_id"`
	IntCredentialName string `json:"int_credential_name"`
	IntCredentialType string `json:"int_credential_type"`
}

type IntCredentialList struct {
	IntCredentialId     int    `json:"int_credential_id"`
	IntCredentialName   string `json:"int_credential_name"`
	IntCredentialType   string `json:"int_credential_type"`
	IntCredentialConfig string `json:"int_credential_config"`
	CreatedAt           string `json:"created_at"`
}

type IntCredentialAddData struct {
	IntCredentialName   string         `json:"int_credential_name"`
	IntCredentialType   SocialPlatform `json:"int_credential_type"`
	IntCredentialConfig string         `json:"int_credential_config"`
}

type IntCredentialUpdateData struct {
	IntCredentialId     int            `json:"int_credential_id"`
	IntCredentialName   string         `json:"int_credential_name"`
	IntCredentialType   SocialPlatform `json:"int_credential_type"`
	IntCredentialConfig string         `json:"int_credential_config"`
}

func NewIntCredentials(db *sql.DB) *IntCredentials {
	intCredentials := IntCredentials{db: db}

	return &intCredentials
}

func (ic *IntCredentials) BasicList(userId int) ([]IntCredentialsBasicList, error) {
	var intCredentials []IntCredentialsBasicList

	rows, rowsErr := ic.db.Query(
		`SELECT credential.int_credential_id, credential.int_credential_name,
            credential.int_credential_type
        FROM integration_credential credential
        WHERE credential.deleted_at IS NULL AND credential.user_id = ?
        ORDER BY credential.int_credential_name`, userId,
	)

	if rowsErr != nil {
		return nil, fmt.Errorf("models.int_credentials.basic_list: %s", rowsErr.Error())
	}

	defer rows.Close()

	rowsErr = rows.Err()

	if rowsErr != nil {
		return nil, fmt.Errorf("models.int_credentials.basic_list: %s", rowsErr.Error())
	}

	for rows.Next() {
		var intCredential IntCredentialsBasicList

		exception := rows.Scan(
			&intCredential.IntCredentialId,
			&intCredential.IntCredentialName,
			&intCredential.IntCredentialType,
		)

		if exception != nil {
			return nil, fmt.Errorf("models.int_credentials.basic_list: %s", exception.Error())
		}

		intCredentials = append(intCredentials, intCredential)

	}

	return intCredentials, nil
}

func (ic *IntCredentials) BasicListByProfile(profileId int, userId int) ([]IntCredentialsBasicList, error) {
	var intCredentials []IntCredentialsBasicList

	rows, rowsErr := ic.db.Query(
		`SELECT credential.int_credential_id, credential.int_credential_name, credential.int_credential_type
        FROM integration_group int_group
        LEFT JOIN integration_credential credential ON credential.int_credential_id = int_group.int_credential_id
            AND credential.deleted_at IS NULL
        WHERE credential.deleted_at IS NULL AND
            credential.user_id = ? AND
            int_group.int_profile_id = ?
        ORDER BY credential.int_credential_type, credential.int_credential_name`, userId, profileId,
	)

	if rowsErr != nil {
		return nil, fmt.Errorf("models.int_credentials.basic_list_by_profile: %s", rowsErr.Error())
	}

	defer rows.Close()

	rowsErr = rows.Err()

	if rowsErr != nil {
		return nil, fmt.Errorf("models.int_credentials.basic_list_by_profile: %s", rowsErr.Error())
	}

	for rows.Next() {
		var intCredential IntCredentialsBasicList

		exception := rows.Scan(
			&intCredential.IntCredentialId,
			&intCredential.IntCredentialName,
			&intCredential.IntCredentialType,
		)

		if exception != nil {
			return nil, fmt.Errorf("models.int_credentials.basic_list_by_profile: %s", exception.Error())
		}

		intCredentials = append(intCredentials, intCredential)
	}

	return intCredentials, nil
}

func (ic *IntCredentials) List(id string, includeConfig bool, userId int) ([]IntCredentialList, error) {
	var intCredentials []IntCredentialList

	whereList := []string{}
	whereValues := []any{}
	columnsList := []string{}

	whereList = append(whereList, "user_id = ?")
	whereValues = append(whereValues, userId)

	if id != "" {
		whereList = append(whereList, "int_credential_id = ?")
		whereValues = append(whereValues, id)
	}
	if includeConfig {
		columnsList = append(columnsList, "int_credential_config")
	} else {
		columnsList = append(columnsList, "'' int_credential_config")
	}

	where := ""
	columns := ""

	if len(whereList) > 0 {
		where = " AND " + strings.Join(whereList, " AND ")
	}
	if len(columnsList) > 0 {
		columns = ", " + strings.Join(columnsList, ", ")
	}

	rows, rowsErr := ic.db.Query(
		`SELECT int_credential_id, int_credential_name,
            int_credential_type, created_at `+columns+`
        FROM integration_credential
        WHERE deleted_at IS NULL `+where, whereValues...,
	)

	if rowsErr != nil {
		return nil, fmt.Errorf("models.int_credentials.list: %s", rowsErr.Error())
	}

	defer rows.Close()

	rowsErr = rows.Err()

	if rowsErr != nil {
		return nil, fmt.Errorf("models.int_credentials.list: %s", rowsErr.Error())
	}

	for rows.Next() {
		var intCredential IntCredentialList
		var intCredentialConfig sql.NullString

		exception := rows.Scan(
			&intCredential.IntCredentialId,
			&intCredential.IntCredentialName,
			&intCredential.IntCredentialType,
			&intCredential.CreatedAt,
			&intCredentialConfig,
		)

		intCredential.IntCredentialConfig = intCredentialConfig.String
		intCredential.CreatedAt = util.ToTimeBR(intCredential.CreatedAt)

		if exception != nil {
			return nil, fmt.Errorf("models.int_credentials.list: %s", exception.Error())
		}

		intCredentials = append(intCredentials, intCredential)
	}

	return intCredentials, nil
}

func (ic *IntCredentials) Add(intCredential IntCredentialAddData, userId int) (int, error) {
	var intCredentialId int

	insertRes, insertErr := ic.db.ExecContext(
		context.Background(),
		`INSERT INTO synk.integration_credential (
            int_credential_name,
            int_credential_type,
            int_credential_config,
            user_id
        )
        VALUES (?, ?, ?, ?)`,
		intCredential.IntCredentialName,
		intCredential.IntCredentialType,
		intCredential.IntCredentialConfig,
		userId,
	)

	if insertErr != nil {
		return intCredentialId, fmt.Errorf("models.int_credentials.add: %s", insertErr.Error())
	}

	id, exception := insertRes.LastInsertId()

	if exception != nil {
		return intCredentialId, fmt.Errorf("models.int_credentials.add: %s", exception.Error())
	}

	intCredentialId = int(id)

	return intCredentialId, nil
}

func (ic *IntCredentials) Update(intCredential IntCredentialUpdateData, userId int) (int, error) {
	var rowsAffected int64

	updateRes, updateErr := ic.db.ExecContext(
		context.Background(),
		`UPDATE integration_credential
        SET int_credential_id = ?,
            int_credential_name = ?,
            int_credential_type = ?,
            int_credential_config = ?,
            updated_at = CURRENT_TIMESTAMP
        WHERE deleted_at IS NULL AND user_id = ? AND int_credential_id = ?`,
		intCredential.IntCredentialId,
		intCredential.IntCredentialName,
		intCredential.IntCredentialType,
		intCredential.IntCredentialConfig,
		userId,
		intCredential.IntCredentialId,
	)

	if updateErr != nil {
		return int(rowsAffected), fmt.Errorf("models.int_credentials.update: %s", updateErr.Error())
	}

	rowsAffectedVal, exception := updateRes.RowsAffected()

	if exception != nil {
		return int(rowsAffected), fmt.Errorf("models.int_credentials.update: %s", exception.Error())
	}

	rowsAffected = rowsAffectedVal

	return int(rowsAffected), nil
}

func (ic *IntCredentials) Delete(intCredentialId int, userId int) (int, error) {
	var rowsAffected int64

	insertRes, insertErr := ic.db.ExecContext(
		context.Background(),
		`UPDATE integration_credential
        SET deleted_at = CURRENT_TIMESTAMP
        WHERE deleted_at IS NULL AND user_id = ? AND int_credential_id = ?`,
		userId, intCredentialId,
	)

	if insertErr != nil {
		return int(rowsAffected), fmt.Errorf("models.int_credential.delete: %s", insertErr.Error())
	}

	rowsAffectedVal, exception := insertRes.RowsAffected()

	if exception != nil {
		return int(rowsAffected), fmt.Errorf("models.int_credential.delete: %s", exception.Error())
	}

	rowsAffected = rowsAffectedVal

	return int(rowsAffected), nil
}
