package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type ReleaseNote struct {
	ID                     int64           `json:"id"`
	ReleaseDate            string          `json:"release_date"`             // e.g. "June 20, 2025"
	ReleasedFeatureDetails json.RawMessage `json:"released_feature_details"` // JSON array of objects
	ReleasedFeatureIDs     json.RawMessage `json:"released_feature_ids"`     // JSON array of IDs
	CreatedAt              int64           `json:"created_at"`
	UpdatedAt              int64           `json:"updated_at"`
}

func CreateReleaseNotesTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS release_notes (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		release_date VARCHAR(50),
		released_feature_details JSON,
		released_feature_ids JSON,
		created_at BIGINT,
		updated_at BIGINT
	);`
	_, err := db.Exec(query)
	return err
}

func CreateReleaseNote(db *sql.DB, note *ReleaseNote) error {
	now := time.Now().Unix()
	note.CreatedAt = now
	note.UpdatedAt = now
	query := `
	INSERT INTO release_notes (release_date, released_feature_details, released_feature_ids, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?)`
	result, err := db.Exec(query, note.ReleaseDate, note.ReleasedFeatureDetails, note.ReleasedFeatureIDs, note.CreatedAt, note.UpdatedAt)
	if err != nil {
		return err
	}
	note.ID, err = result.LastInsertId()
	return err
}

func GetReleaseNoteByID(db *sql.DB, id int64) (*ReleaseNote, error) {
	query := `SELECT id, release_date, released_feature_details, released_feature_ids, created_at, updated_at FROM release_notes WHERE id = ?`
	note := &ReleaseNote{}
	err := db.QueryRow(query, id).Scan(&note.ID, &note.ReleaseDate, &note.ReleasedFeatureDetails, &note.ReleasedFeatureIDs, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return note, nil
}

func UpdateReleaseNote(db *sql.DB, note *ReleaseNote) error {
	note.UpdatedAt = time.Now().Unix()
	query := `
	UPDATE release_notes SET release_date = ?, released_feature_details = ?, released_feature_ids = ?, updated_at = ? WHERE id = ?`
	_, err := db.Exec(query, note.ReleaseDate, note.ReleasedFeatureDetails, note.ReleasedFeatureIDs, note.UpdatedAt, note.ID)
	return err
}

func DeleteReleaseNote(db *sql.DB, id int64) error {
	query := `DELETE FROM release_notes WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

func GetAllReleaseNotes(db *sql.DB) ([]*ReleaseNote, error) {
	query := `SELECT id, release_date, released_feature_details, released_feature_ids, created_at, updated_at FROM release_notes`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*ReleaseNote
	for rows.Next() {
		note := &ReleaseNote{}
		err := rows.Scan(&note.ID, &note.ReleaseDate, &note.ReleasedFeatureDetails, &note.ReleasedFeatureIDs, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}
