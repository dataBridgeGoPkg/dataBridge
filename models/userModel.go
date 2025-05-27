package models

import (
	"database/sql"
	"time"
)

// Role represents the possible roles a user can have
type Role string

const (
	Admin     Role = "ADMIN"
	Developer Role = "DEVELOPER"
	Viewer    Role = "VIEWER"
)

func (r Role) IsValid() bool {
	switch r {
	case Admin, Developer, Viewer:
		return true
	default:
		return false
	}
}

// User represents a user in the system
type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	EmailId   string `json:"email_id"`
	Role      Role   `json:"role,omitempty"`
	Password  string `json:"-"`
	JiraID    string `json:"jira_id,omitempty"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// CreateUsersTable creates the users table with BIGINT timestamps and a role column
func CreateUsersTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id BIGINT AUTO_INCREMENT PRIMARY KEY,
        first_name VARCHAR(255) NOT NULL,
        last_name VARCHAR(255) NOT NULL,
        email_id VARCHAR(255) NOT NULL UNIQUE,
        password VARCHAR(255) NOT NULL,
		jira_id VARCHAR(255) DEFAULT NULL,
        role VARCHAR(50) NOT NULL,
        created_at BIGINT NOT NULL,
        updated_at BIGINT NOT NULL
    );`

	_, err := db.Exec(query)
	return err
}

// CreateUser inserts a new user into the database
func CreateUser(db *sql.DB, user *User) error {
	now := time.Now().Unix()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `
	INSERT INTO users (first_name, last_name, email_id, password, jira_id, role, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.Exec(query, user.FirstName, user.LastName, user.EmailId, user.Password, user.JiraID, user.Role, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return err
	}

	user.ID, err = result.LastInsertId()
	return err
}

// GetUserByID retrieves a user by their ID
func GetUserByID(db *sql.DB, id int64) (*User, error) {
	query := `
	SELECT id, first_name, last_name, email_id, jira_id, role, created_at, updated_at
	FROM users
	WHERE id = ?`

	user := &User{}
	err := db.QueryRow(query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.EmailId,
		&user.JiraID,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByEmail retrieves a user by their email address
func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	query := `
	SELECT id, first_name, last_name, email_id, password, jira_id, role, created_at, updated_at
	FROM users
	WHERE email_id = ?`

	user := &User{}
	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.EmailId,
		&user.Password,
		&user.JiraID,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUser updates an existing user
func UpdateUser(db *sql.DB, user *User) error {
	user.UpdatedAt = time.Now().Unix()

	query := `
	UPDATE users
	SET first_name = ?, last_name = ?, email_id = ?, role = ?, jira_id = ?, updated_at = ?
	WHERE id = ?`

	_, err := db.Exec(query, user.FirstName, user.LastName, user.EmailId, user.JiraID, user.Role, user.UpdatedAt, user.ID)
	return err
}

// DeleteUser deletes a user by their ID
func DeleteUser(db *sql.DB, id int64) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

// GetAllUsers retrieves all users from the database
func GetAllUsers(db *sql.DB) ([]*User, error) {
	query := `
	SELECT id, first_name, last_name, email_id, jira_id, role, created_at, updated_at
	FROM users`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.EmailId,
			&user.JiraID,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
