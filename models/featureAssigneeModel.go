package models

import "database/sql"

type FeatureAssignee struct {
	FeatureAssigneeID int64 `json:"feature_assignee_id"` // Primary key, auto-increment
	FeatureID         int64 `json:"feature_id"`          // Foreign key to features table
	UserID            int64 `json:"user_id"`             // Foreign key to users table
}

// CreateFeatureAssignee inserts a new row into feature_assignees
func CreateFeatureAssignee(db *sql.DB, fa *FeatureAssignee) error {
	query := `
    INSERT INTO feature_assignees (feature_id, user_id)
    VALUES (?, ?)
    `

	result, err := db.Exec(query, fa.FeatureID, fa.UserID)
	if err != nil {
		return err
	}

	fa.FeatureAssigneeID, err = result.LastInsertId()
	return err
}

func GetFeatureAssigneeByID(db *sql.DB, id int64) (*FeatureAssignee, error) {
	query := `SELECT feature_assignee_id, feature_id, user_id FROM feature_assignees WHERE feature_assignee_id = ?`
	featureAssignee := &FeatureAssignee{}
	err := db.QueryRow(query, id).Scan(
		&featureAssignee.FeatureAssigneeID,
		&featureAssignee.FeatureID,
		&featureAssignee.UserID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return featureAssignee, nil
}

// DeleteFeatureAssignee deletes a feature_assignee row by its ID
func DeleteFeatureAssignee(db *sql.DB, id int64) error {
	query := `DELETE FROM feature_assignees WHERE feature_assignee_id = ?`
	_, err := db.Exec(query, id)
	return err
}

func DeleteFeatureAssigneeWithUserId(db *sql.DB, id int64) error {
	query := `DELETE FROM feature_assignees WHERE user_id = ?`
	_, err := db.Exec(query, id)
	return err
}

func DeleteFeatureAssigneeWithFeatureId(db *sql.DB, id int64) error {
	query := `DELETE FROM feature_assignees WHERE feature_id = ?`
	_, err := db.Exec(query, id)
	return err
}

// CheckIfUserAlreadyAssigned checks if a user is already assigned to a specific feature
func CheckIfUserAlreadyAssigned(db *sql.DB, featureID, userID int64) (bool, error) {
	query := `
		SELECT 1 FROM feature_assignees
		WHERE feature_id = ? AND user_id = ?
		LIMIT 1
	`
	var exists int
	err := db.QueryRow(query, featureID, userID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
