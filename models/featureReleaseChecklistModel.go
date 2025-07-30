package models

import (
	"database/sql"
	"time"
)

// FeatureReleaseChecklistModel represents a checklist item for a feature release
// Each item is a row in the feature_release_checklists table

type FeatureReleaseChecklist struct {
	ID        int64  `json:"id,omitempty"`
	FeatureID int64  `json:"feature_id"` // Foreign key to features table
	Item      string `json:"item"`       // Checklist item name
	Validated bool   `json:"validated"`  // true if validated, false otherwise
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

// CreateFeatureReleaseChecklistTable creates the feature_release_checklists table if it doesn't exist
func CreateFeatureReleaseChecklistTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS feature_release_checklists (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		feature_id BIGINT NOT NULL,
		item VARCHAR(255) NOT NULL,
		validated BOOLEAN DEFAULT FALSE,
		created_at BIGINT NOT NULL,
		updated_at BIGINT NOT NULL,
		FOREIGN KEY (feature_id) REFERENCES features(id) ON DELETE CASCADE
	)`
	_, err := db.Exec(query)
	return err
}

// CreateFeatureReleaseChecklist inserts a new checklist item for a feature
func CreateFeatureReleaseChecklist(db *sql.DB, checklist *FeatureReleaseChecklist) error {
	now := time.Now().Unix()
	checklist.CreatedAt = now
	checklist.UpdatedAt = now

	query := `
	INSERT INTO feature_release_checklists (feature_id, item, validated, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?)`
	result, err := db.Exec(query,
		checklist.FeatureID,
		checklist.Item,
		checklist.Validated,
		checklist.CreatedAt,
		checklist.UpdatedAt,
	)
	if err != nil {
		return err
	}
	checklist.ID, err = result.LastInsertId()
	return err
}

// UpdateFeatureReleaseChecklist updates an existing checklist item
func UpdateFeatureReleaseChecklist(db *sql.DB, checklist *FeatureReleaseChecklist) error {
	checklist.UpdatedAt = time.Now().Unix()
	query := `
	UPDATE feature_release_checklists
	SET item = ?, validated = ?, updated_at = ?
	WHERE id = ?`
	_, err := db.Exec(query,
		checklist.Item,
		checklist.Validated,
		checklist.UpdatedAt,
		checklist.ID,
	)
	return err
}

// GetAllFeatureReleaseChecklists retrieves all checklist items for a feature
func GetAllFeatureReleaseChecklists(db *sql.DB, featureID int64) ([]*FeatureReleaseChecklist, error) {
	query := `
	SELECT id, feature_id, item, validated, created_at, updated_at
	FROM feature_release_checklists
	WHERE feature_id = ?
	ORDER BY created_at ASC`
	rows, err := db.Query(query, featureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checklists []*FeatureReleaseChecklist
	for rows.Next() {
		cl := &FeatureReleaseChecklist{}
		err := rows.Scan(
			&cl.ID,
			&cl.FeatureID,
			&cl.Item,
			&cl.Validated,
			&cl.CreatedAt,
			&cl.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		checklists = append(checklists, cl)
	}
	return checklists, nil
}

// GetFeatureReleaseChecklistByID retrieves a checklist item by its ID
func GetFeatureReleaseChecklistByID(db *sql.DB, id int64) (*FeatureReleaseChecklist, error) {
	query := `
	SELECT id, feature_id, item, validated, created_at, updated_at
	FROM feature_release_checklists
	WHERE id = ?`
	cl := &FeatureReleaseChecklist{}
	err := db.QueryRow(query, id).Scan(
		&cl.ID,
		&cl.FeatureID,
		&cl.Item,
		&cl.Validated,
		&cl.CreatedAt,
		&cl.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return cl, nil
}

// DeleteFeatureReleaseChecklist deletes a checklist item by its ID
func DeleteFeatureReleaseChecklist(db *sql.DB, id int64) error {
	query := `DELETE FROM feature_release_checklists WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

func GetFeatureReleaseChecklistByFeatureID(db *sql.DB, featureID int64) ([]*FeatureReleaseChecklist, error) {
	query := `
	SELECT id, feature_id, item, validated, created_at, updated_at
	FROM feature_release_checklists
	WHERE feature_id = ?
	ORDER BY created_at ASC`
	rows, err := db.Query(query, featureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checklists []*FeatureReleaseChecklist
	for rows.Next() {
		cl := &FeatureReleaseChecklist{}
		err := rows.Scan(
			&cl.ID,
			&cl.FeatureID,
			&cl.Item,
			&cl.Validated,
			&cl.CreatedAt,
			&cl.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		checklists = append(checklists, cl)
	}
	return checklists, nil
}
