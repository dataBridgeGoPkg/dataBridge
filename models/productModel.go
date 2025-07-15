package models

import "database/sql"

// Product represents a product in the system.
type Product struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// CreateProductsTable creates the products table in the database.
func CreateProductsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS products (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE
	);`
	_, err := db.Exec(query)
	return err
}

// CreateProduct adds a new product to the database.
func CreateProduct(db *sql.DB, name string) (int64, error) {
	result, err := db.Exec("INSERT INTO products (name) VALUES (?)", name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetProductByID retrieves a product by its ID.
func GetProductByID(db *sql.DB, id int64) (*Product, error) {
	product := &Product{}
	err := db.QueryRow("SELECT id, name FROM products WHERE id = ?", id).Scan(&product.ID, &product.Name)
	if err != nil {
		return nil, err
	}
	return product, nil
}

// GetAllProducts retrieves all products from the database.
func GetAllProducts(db *sql.DB) ([]*Product, error) {
	rows, err := db.Query("SELECT id, name FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		product := &Product{}
		if err := rows.Scan(&product.ID, &product.Name); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}
