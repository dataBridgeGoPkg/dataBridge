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
	Blocked    StatusType = "BLOCKED"
)

func (s StatusType) IsValid() bool {
	switch s {
	case NotStarted, InProgress, Ready, OnProd, Blocked:
		return true
	default:
		return false
	}
}

type FeatureHealth string

const (
	IdeaValidation FeatureHealth = "IDEA_VALIDATION"
	Design         FeatureHealth = "DESIGN"
	Development    FeatureHealth = "DEVELOPMENT"
	Deployed       FeatureHealth = "DEPLOYED"
)

func (h FeatureHealth) IsValid() bool {
	switch h {
	case IdeaValidation, Design, Development, Deployed:
		return true
	default:
		return false
	}
}

// Feature represents a product feature assigned to a user
type Feature struct {
	ID            int64         `json:"id,omitempty"`
	Title         string        `json:"title,omitempty"`
	Description   string        `json:"description,omitempty"`
	Status        StatusType    `json:"status,omitempty"`
	Health        FeatureHealth `json:"health,omitempty"`
	StartTime     *int64        `json:"start_time,omitempty"`
	EndTime       *int64        `json:"end_time,omitempty"`
	Notes         *string       `json:"notes,omitempty"`
	FeatureDocUrl *string       `json:"feature_doc_url,omitempty"`
	FigmaUrl      *string       `json:"figma_url,omitempty"`
	Insights      *string       `json:"insights,omitempty"`
	CreatedAt     int64         `json:"created_at,omitempty"`
	UpdatedAt     int64         `json:"updated_at,omitempty"`
}

func CreateFeaturesTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS features (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		status VARCHAR(50) NOT NULL,
		health VARCHAR(50) NOT NULL,
		start_time VARCHAR(255),
		end_time VARCHAR(255),
		notes TEXT,
		feature_doc_url TEXT,
		figma_url TEXT,
		insights TEXT,
		created_at BIGINT NOT NULL,
		updated_at BIGINT NOT NULL,
	)`
	_, err := db.Exec(query)
	return err
}

type FeatureModelWithUserName struct {
	ID            int64
	Title         string
	Description   string
	Status        string
	Health        string
	StartTime     sql.NullInt64
	EndTime       sql.NullInt64
	Notes         string
	FeatureDocUrl sql.NullString
	FigmaUrl      sql.NullString
	Insights      sql.NullString
	UserFirstName sql.NullString // <-- changed
	UserLastName  sql.NullString // <-- changed
	CreatedAt     int64
	UpdatedAt     int64
}

// ___________//
// UserInfoForFeature holds the subset of user details to be nested within a feature.
type UserInfoForFeature struct {
	UserID    int64          `json:"user_id"`
	FirstName sql.NullString `json:"first_name,omitempty"` // Use omitempty if you want to hide null names
	LastName  sql.NullString `json:"last_name,omitempty"`
}

type FeatureWithAssignedUsers struct {
	ID            int64                `json:"id"`
	Title         string               `json:"title"`
	Description   string               `json:"description"`
	Status        string               `json:"status"` // Consider using your StatusType here for consistency
	Health        string               `json:"health"` // Consider using your FeatureHealth here for consistency
	StartTime     sql.NullInt64        `json:"start_time,omitempty"`
	EndTime       sql.NullInt64        `json:"end_time,omitempty"`
	Notes         sql.NullString       `json:"notes,omitempty"` // Assuming notes can be NULL
	FeatureDocUrl sql.NullString       `json:"feature_doc_url,omitempty"`
	FigmaUrl      sql.NullString       `json:"figma_url,omitempty"`
	Insights      sql.NullString       `json:"insights,omitempty"`
	AssignedUsers []UserInfoForFeature `json:"assigned_users"`
	CreatedAt     int64                `json:"created_at"`
	UpdatedAt     int64                `json:"updated_at"`
}

// GetAllFeaturesWithAssignedUsers retrieves all features along with their assigned users.
func GetAllFeaturesWithUserName(db *sql.DB) ([]*FeatureWithAssignedUsers, error) {
	query := `
    SELECT
        f.id AS feature_id,
        f.title,
        f.description,
        f.status,
		f.health,
        f.start_time,
        f.end_time,
        f.notes,
        f.feature_doc_url,
        f.figma_url,
        f.insights,
        f.created_at AS feature_created_at,
        f.updated_at AS feature_updated_at,
        u.id AS user_id,         -- User's ID
        u.first_name,
        u.last_name
    FROM
        features f
    LEFT JOIN
        feature_assignees fa ON f.id = fa.feature_id
    LEFT JOIN
        users u ON fa.user_id = u.id
    ORDER BY
        f.id ASC; -- IMPORTANT: Order by feature ID for easier grouping in Go
    `

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Use a map to group users by feature ID
	featuresMap := make(map[int64]*FeatureWithAssignedUsers)
	// Maintain order of features as encountered (map iteration is not ordered)
	var orderedFeatureIDs []int64

	for rows.Next() {
		var featureID int64
		var title string
		var description string
		var status string // Or your models.StatusType if you prefer
		var health string // Or your models.FeatureHealth if you prefer
		var startTime sql.NullInt64
		var endTime sql.NullInt64
		var notes sql.NullString
		var featureDocUrl sql.NullString
		var figmaUrl sql.NullString
		var insights sql.NullString
		var featureCreatedAt int64
		var featureUpdatedAt int64

		// These user fields can be NULL if a feature has no assignees (due to LEFT JOIN)
		var userID sql.NullInt64
		var userFirstName sql.NullString
		var userLastName sql.NullString

		err := rows.Scan(
			&featureID,
			&title,
			&description,
			&status,
			&health,
			&startTime,
			&endTime,
			&notes,
			&featureDocUrl,
			&figmaUrl,
			&insights,
			&featureCreatedAt,
			&featureUpdatedAt,
			&userID, // Scan the user's ID
			&userFirstName,
			&userLastName,
		)
		if err != nil {
			return nil, err
		}

		// Check if we've already started processing this feature
		feature, exists := featuresMap[featureID]
		if !exists {
			// First time seeing this feature
			feature = &FeatureWithAssignedUsers{
				ID:            featureID,
				Title:         title,
				Description:   description,
				Status:        status,
				Health:        health,
				StartTime:     startTime,
				EndTime:       endTime,
				Notes:         notes,
				FeatureDocUrl: featureDocUrl,
				FigmaUrl:      figmaUrl,
				Insights:      insights,
				CreatedAt:     featureCreatedAt,
				UpdatedAt:     featureUpdatedAt,
				AssignedUsers: []UserInfoForFeature{}, // Initialize empty slice
			}
			featuresMap[featureID] = feature
			orderedFeatureIDs = append(orderedFeatureIDs, featureID)
		}

		// If userID is valid (i.e., there was a user joined for this row), add them
		if userID.Valid {
			assignedUser := UserInfoForFeature{
				UserID:    userID.Int64,
				FirstName: userFirstName,
				LastName:  userLastName,
			}
			feature.AssignedUsers = append(feature.AssignedUsers, assignedUser)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Convert map to slice based on the order encountered
	result := make([]*FeatureWithAssignedUsers, len(orderedFeatureIDs))
	for i, id := range orderedFeatureIDs {
		result[i] = featuresMap[id]
	}

	return result, nil
}

// func GetAllFeaturesWithUserName(db *sql.DB) ([]*FeatureModelWithUserName, error) {
// 	query := `
// 	SELECT
// 		f.id, f.title, f.description, f.status, f.start_time, f.end_time,
// 		f.notes, f.feature_doc_url, f.figma_url, f.insights,
// 		IFNULL(u.first_name, ''), IFNULL(u.last_name, ''),
// 		f.created_at, f.updated_at
// 	FROM features f
// 	LEFT JOIN feature_assignees fa ON f.id = fa.feature_id
// 	LEFT JOIN users u ON fa.user_id = u.id
// 	`

// 	rows, err := db.Query(query)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var features []*FeatureModelWithUserName
// 	for rows.Next() {
// 		feature := &FeatureModelWithUserName{}
// 		err := rows.Scan(
// 			&feature.ID,
// 			&feature.Title,
// 			&feature.Description,
// 			&feature.Status,
// 			&feature.StartTime,
// 			&feature.EndTime,
// 			&feature.Notes,
// 			&feature.FeatureDocUrl,
// 			&feature.FigmaUrl,
// 			&feature.Insights,
// 			&feature.UserFirstName,
// 			&feature.UserLastName,
// 			&feature.CreatedAt,
// 			&feature.UpdatedAt,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		features = append(features, feature)
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return features, nil
// }

func CreateFeature(db *sql.DB, feature *Feature) error {
	if !feature.Status.IsValid() {
		return errors.New("invalid status value")
	}

	now := time.Now().Unix()
	feature.CreatedAt = now
	feature.UpdatedAt = now

	query := `
	INSERT INTO features (
		title, description, status, health, start_time, end_time, notes, feature_doc_url, figma_url, insights, created_at, updated_at
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.Exec(query,
		feature.Title,
		feature.Description,
		string(feature.Status),
		string(feature.Health),
		feature.StartTime,
		feature.EndTime,
		feature.Notes,
		feature.FeatureDocUrl,
		feature.FigmaUrl,
		feature.Insights,
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
	SELECT 
		id, title, description, status, health, start_time, end_time, notes, feature_doc_url, figma_url, insights, created_at, updated_at
	FROM features
	WHERE id = ?`

	feature := &Feature{}

	err := db.QueryRow(query, id).Scan(
		&feature.ID,
		&feature.Title,
		&feature.Description,
		&feature.Status,
		&feature.Health,
		&feature.StartTime,
		&feature.EndTime,
		&feature.Notes,
		&feature.FeatureDocUrl,
		&feature.FigmaUrl,
		&feature.Insights,
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
	SET title = ?, description = ?, status = ?, health = ?, start_time = ?, end_time = ?, 
		notes = ?, feature_doc_url = ?, figma_url = ?, insights = ?, updated_at = ?
	WHERE id = ?`

	_, err := db.Exec(query,
		feature.Title,
		feature.Description,
		string(feature.Status),
		string(feature.Health),
		feature.StartTime,
		feature.EndTime,
		feature.Notes,
		feature.FeatureDocUrl,
		feature.FigmaUrl,
		feature.Insights,
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
	SELECT 
		id, title, description, status, health, start_time, end_time, notes, feature_doc_url, figma_url, insights, created_at, updated_at
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
			&feature.Health,
			&feature.StartTime,
			&feature.EndTime,
			&feature.Notes,
			&feature.FeatureDocUrl,
			&feature.FigmaUrl,
			&feature.Insights,
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
