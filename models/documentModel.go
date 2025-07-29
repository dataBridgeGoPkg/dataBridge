package models

import (
	"database/sql"
	"time"
)

// Document represents a standalone document in the system
type Document struct {
	ID          int64   `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	URL         string  `json:"url"`
	ProductID   int64   `json:"product_id"`
	CreatedAt   int64   `json:"created_at,omitempty"`
	UpdatedAt   int64   `json:"updated_at,omitempty"`
}

func CreateDocumentsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS documents (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		url TEXT,
		product_id BIGINT NOT NULL,
		created_at BIGINT NOT NULL,
		updated_at BIGINT NOT NULL
	)`
	_, err := db.Exec(query)
	return err
}

func CreateDocument(db *sql.DB, doc *Document) error {
	now := time.Now().Unix()
	doc.CreatedAt = now
	doc.UpdatedAt = now

	query := `
	INSERT INTO documents (name, description, url, product_id, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	result, err := db.Exec(query,
		doc.Name,
		doc.Description,
		doc.URL,
		doc.ProductID,
		doc.CreatedAt,
		doc.UpdatedAt,
	)
	if err != nil {
		return err
	}

	doc.ID, err = result.LastInsertId()
	return err
}

func GetDocumentByID(db *sql.DB, id int64) (*Document, error) {
	query := `
	SELECT id, name, description, url, product_id, created_at, updated_at
	FROM documents
	WHERE id = ?`

	doc := &Document{}

	err := db.QueryRow(query, id).Scan(
		&doc.ID,
		&doc.Name,
		&doc.Description,
		&doc.URL,
		&doc.ProductID,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func UpdateDocument(db *sql.DB, doc *Document) error {
	doc.UpdatedAt = time.Now().Unix()

	query := `
	UPDATE documents
	SET name = ?, description = ?, url = ?, product_id = ?, updated_at = ?
	WHERE id = ?`

	_, err := db.Exec(query,
		doc.Name,
		doc.Description,
		doc.URL,
		doc.ProductID,
		doc.UpdatedAt,
		doc.ID,
	)
	return err
}

func DeleteDocument(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM documents WHERE id = ?`, id)
	return err
}

func GetAllDocuments(db *sql.DB) ([]*Document, error) {
	query := `
	SELECT id, name, description, url, product_id, created_at, updated_at
	FROM documents`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []*Document
	for rows.Next() {
		doc := &Document{}
		err := rows.Scan(
			&doc.ID,
			&doc.Name,
			&doc.Description,
			&doc.URL,
			&doc.ProductID,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return documents, nil
}
