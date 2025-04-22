package models

import (
	"database/sql"
	"errors"
	"time"
)

// StatusType represents the possible states of a feature
type StatusType string

const (
	NotStarted StatusType = "NOT_STARTED"
	InProgress StatusType = "IN_PROGRESS"
	Ready      StatusType = "READY"
	OnProd     StatusType = "ON_PROD"
)

func (s StatusType) IsValid() bool {
	switch s {
	case NotStarted, InProgress, Ready, OnProd:
		return true
	default:
		return false
	}
}

// Feature represents a product feature assigned to a user
type Feature struct {
	ID           int64      `json:"id,omitempty"`
	Title        string     `json:"title,omitempty"`
	Description  string     `json:"description,omitempty"`
	Status       StatusType `json:"status,omitempty"`
	StartTime    *string    `json:"start_time,omitempty"` // consider using int64 or time.Time if doing time calculations
	EndTime      *string    `json:"end_time,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
	AssignedUser *int64     `json:"assigned_user,omitempty"`
	CreatedAt    int64      `json:"created_at,omitempty"`
	UpdatedAt    int64      `json:"updated_at,omitempty"`
}

func CreateFeaturesTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS features (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		status VARCHAR(50) NOT NULL,
		start_time VARCHAR(255),
		end_time VARCHAR(255),
		notes TEXT,
		assigned_user_id BIGINT,
		created_at BIGINT NOT NULL,
		updated_at BIGINT NOT NULL,
		FOREIGN KEY (assigned_user_id) REFERENCES users(id)
	)`
	_, err := db.Exec(query)
	return err
}

func CreateFeature(db *sql.DB, feature *Feature) error {
	if !feature.Status.IsValid() {
		return errors.New("invalid status value")
	}

	now := time.Now().Unix()
	feature.CreatedAt = now
	feature.UpdatedAt = now

	query := `
	INSERT INTO features (title, description, status, start_time, end_time, notes, assigned_user_id, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.Exec(query,
		feature.Title,
		feature.Description,
		string(feature.Status),
		feature.StartTime,
		feature.EndTime,
		feature.Notes,
		feature.AssignedUser,
		feature.CreatedAt,
		feature.UpdatedAt,
	)
	if err != nil {
		return err
	}

	feature.ID, err = result.LastInsertId()
	return err
}

func GetFeatureByID(db *sql.DB, id int64) (*Feature, error) {
	query := `
	SELECT id, title, description, status, start_time, end_time, notes, assigned_user_id, created_at, updated_at
	FROM features
	WHERE id = ?`

	feature := &Feature{}

	err := db.QueryRow(query, id).Scan(
		&feature.ID,
		&feature.Title,
		&feature.Description,
		&feature.Status,
		&feature.StartTime,
		&feature.EndTime,
		&feature.Notes,
		&feature.AssignedUser,
		&feature.CreatedAt,
		&feature.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return feature, nil
}

func UpdateFeature(db *sql.DB, feature *Feature) error {
	if !feature.Status.IsValid() {
		return errors.New("invalid status value")
	}

	feature.UpdatedAt = time.Now().Unix()

	query := `
	UPDATE features
	SET title = ?, description = ?, status = ?, start_time = ?, end_time = ?, notes = ?, assigned_user_id = ?, updated_at = ?
	WHERE id = ?`

	_, err := db.Exec(query,
		feature.Title,
		feature.Description,
		string(feature.Status),
		feature.StartTime,
		feature.EndTime,
		feature.Notes,
		feature.AssignedUser,
		feature.UpdatedAt,
		feature.ID,
	)
	return err
}

func DeleteFeature(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM features WHERE id = ?`, id)
	return err
}

func GetAllFeatures(db *sql.DB) ([]*Feature, error) {
	query := `
	SELECT id, title, description, status, start_time, end_time, notes, assigned_user_id, created_at, updated_at
	FROM features`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []*Feature
	for rows.Next() {
		feature := &Feature{}
		err := rows.Scan(
			&feature.ID,
			&feature.Title,
			&feature.Description,
			&feature.Status,
			&feature.StartTime,
			&feature.EndTime,
			&feature.Notes,
			&feature.AssignedUser,
			&feature.CreatedAt,
			&feature.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		features = append(features, feature)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return features, nil
}
