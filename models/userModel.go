package models

import (
	"database/sql"
	"time"
)

// User represents a user in the system
type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	EmailId   string `json:"email_id"`
	Password  string `json:"-"` // Exclude from JSON responses
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// CreateUsersTable creates the users table with BIGINT timestamps
func CreateUsersTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id BIGINT AUTO_INCREMENT PRIMARY KEY,
        first_name VARCHAR(255) NOT NULL,
        last_name VARCHAR(255) NOT NULL,
        email_id VARCHAR(255) NOT NULL UNIQUE,
        password VARCHAR(255) NOT NULL,
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
	INSERT INTO users (first_name, last_name, email_id, password, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	result, err := db.Exec(query, user.FirstName, user.LastName, user.EmailId, user.Password, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return err
	}

	user.ID, err = result.LastInsertId()
	return err
}

// GetUserByID retrieves a user by their ID
func GetUserByID(db *sql.DB, id int64) (*User, error) {
	query := `
	SELECT id, first_name, last_name, email_id, created_at, updated_at
	FROM users
	WHERE id = ?`

	user := &User{}
	err := db.QueryRow(query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.EmailId,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Get user by email
func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	query := `
		SELECT id, first_name, last_name, email_id, password, created_at, updated_at
		FROM users
		WHERE email_id = ?`

	user := &User{}
	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.EmailId,
		&user.Password,
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
	SET first_name = ?, last_name = ?, email_id = ?, updated_at = ?
	WHERE id = ?`

	_, err := db.Exec(query, user.FirstName, user.LastName, user.EmailId, user.UpdatedAt, user.ID)
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
	SELECT id, first_name, last_name, email_id, created_at, updated_at
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
