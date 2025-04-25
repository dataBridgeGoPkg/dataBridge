package models

import (
	"database/sql"
	"time"
)

// FeatureRequestModel represents a feature request in the system
type FeatureRequestModel struct {
	ID          int64  `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Accepted    bool   `json:"accepted,omitempty"`
	RequestedBy string `json:"requested_by,omitempty"` // Optional field for the user who requested the feature
	CreatedAt   int64  `json:"created_at,omitempty"`   // Unix timestamp
	UpdatedAt   int64  `json:"updated_at,omitempty"`   // Unix timestamp
}

// CreateFeatureRequestTable creates the feature_requests table if it doesn't exist
func CreateFeatureRequestTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS feature_requests (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		accepted BOOLEAN DEFAULT FALSE,
		requested_by VARCHAR(255),
		created_at BIGINT NOT NULL,
		updated_at BIGINT NOT NULL
	)`
	_, err := db.Exec(query)
	return err
}

// CreateFeatureRequest inserts a new feature request into the database
func CreateFeatureRequest(db *sql.DB, fr *FeatureRequestModel) error {
	now := time.Now().Unix()
	fr.CreatedAt = now
	fr.UpdatedAt = now

	query := `
	INSERT INTO feature_requests (title, description, accepted, requested_by, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	result, err := db.Exec(query,
		fr.Title,
		fr.Description,
		fr.Accepted,
		fr.RequestedBy,
		fr.CreatedAt,
		fr.UpdatedAt,
	)
	if err != nil {
		return err
	}

	fr.ID, err = result.LastInsertId()
	return err
}

// FetchFeatureRequestByID retrieves a feature request by its ID
func FetchFeatureRequestByID(db *sql.DB, id int64) (*FeatureRequestModel, error) {
	query := `
	SELECT id, title, description, accepted, requested_by, created_at, updated_at
	FROM feature_requests
	WHERE id = ?`

	featureRequest := &FeatureRequestModel{}
	err := db.QueryRow(query, id).Scan(
		&featureRequest.ID,
		&featureRequest.Title,
		&featureRequest.Description,
		&featureRequest.Accepted,
		&featureRequest.RequestedBy,
		&featureRequest.CreatedAt,
		&featureRequest.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return featureRequest, nil
}

// FetchAllFeatureRequests retrieves all feature requests
func FetchAllFeatureRequests(db *sql.DB) ([]*FeatureRequestModel, error) {
	query := `
	SELECT id, title, description, accepted, requested_by, created_at, updated_at
	FROM feature_requests
	ORDER BY created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var featureRequests []*FeatureRequestModel
	for rows.Next() {
		fr := &FeatureRequestModel{}
		err := rows.Scan(
			&fr.ID,
			&fr.Title,
			&fr.Description,
			&fr.Accepted,
			&fr.RequestedBy,
			&fr.CreatedAt,
			&fr.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		featureRequests = append(featureRequests, fr)
	}

	return featureRequests, nil
}

// UpdateFeatureRequestByID updates a feature request with new data
func UpdateFeatureRequestByID(db *sql.DB, fr *FeatureRequestModel) error {
	fr.UpdatedAt = time.Now().Unix()

	query := `
	UPDATE feature_requests
	SET title = ?, description = ?, accepted = ?, requested_by = ?, updated_at = ?
	WHERE id = ?`

	_, err := db.Exec(query,
		fr.Title,
		fr.Description,
		fr.Accepted,
		fr.RequestedBy,
		fr.UpdatedAt,
		fr.ID,
	)

	return err
}

// DeleteFeatureRequestByID deletes a feature request from the database
func DeleteFeatureRequestByID(db *sql.DB, id int64) error {
	query := `DELETE FROM feature_requests WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}
