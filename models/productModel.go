package models

import (
	"database/sql"
)

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

// GetFeaturesWithAssigneesByProductID returns all features and their assignees for a given product ID.
func GetFeaturesWithAssigneesByProductID(db *sql.DB, productID int64) ([]*FeatureWithAssignedUsers, error) {
	query := `
	SELECT
		f.id AS feature_id,
		f.title,
		f.description,
		f.status,
		f.health,
		f.tier,
		f.start_time,
		f.end_time,
		f.notes,
		f.feature_doc_url,
		f.figma_url,
		f.insights,
		f.jira_sync,
		f.product_board_sync,
		f.jira_id,
		f.jira_url,
		f.product_board_id,
		f.business_case,
		f.product_id,
		f.created_at AS feature_created_at,
		f.updated_at AS feature_updated_at,
		u.id AS user_id,
		u.first_name,
		u.last_name
	FROM
		features f
	LEFT JOIN
		feature_assignees fa ON f.id = fa.feature_id
	LEFT JOIN
		users u ON fa.user_id = u.id
	WHERE
		f.product_id = ?
	ORDER BY
		f.id ASC;
	`

	rows, err := db.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	featuresMap := make(map[int64]*FeatureWithAssignedUsers)
	var orderedFeatureIDs []int64

	for rows.Next() {
		var featureID int64
		var title string
		var description string
		var status string
		var health string
		var tier string
		var startTime sql.NullInt64
		var endTime sql.NullInt64
		var notes sql.NullString
		var featureDocUrl sql.NullString
		var figmaUrl sql.NullString
		var insights sql.NullString
		var jiraSync sql.NullBool
		var productBoardSync sql.NullBool
		var jiraID sql.NullString
		var jiraUrl sql.NullString
		var productBoardID sql.NullString
		var businessCase sql.NullString
		var prodID sql.NullInt64
		var featureCreatedAt int64
		var featureUpdatedAt int64
		var userID sql.NullInt64
		var userFirstName sql.NullString
		var userLastName sql.NullString

		err := rows.Scan(
			&featureID,
			&title,
			&description,
			&status,
			&health,
			&tier,
			&startTime,
			&endTime,
			&notes,
			&featureDocUrl,
			&figmaUrl,
			&insights,
			&jiraSync,
			&productBoardSync,
			&jiraID,
			&jiraUrl,
			&productBoardID,
			&businessCase,
			&prodID,
			&featureCreatedAt,
			&featureUpdatedAt,
			&userID,
			&userFirstName,
			&userLastName,
		)
		if err != nil {
			return nil, err
		}

		feature, exists := featuresMap[featureID]
		if !exists {
			feature = &FeatureWithAssignedUsers{
				ID:               featureID,
				Title:            title,
				Description:      description,
				Status:           status,
				Health:           health,
				Tier:             tier,
				StartTime:        startTime,
				EndTime:          endTime,
				Notes:            notes,
				FeatureDocUrl:    featureDocUrl,
				FigmaUrl:         figmaUrl,
				Insights:         insights,
				JiraSync:         jiraSync,
				ProductBoardSync: productBoardSync,
				JiraID:           jiraID,
				JiraUrl:          jiraUrl,
				ProductBoardID:   productBoardID,
				BusinessCase:     businessCase,
				ProductID:        toPtrInt64(prodID),
				CreatedAt:        featureCreatedAt,
				UpdatedAt:        featureUpdatedAt,
				AssignedUsers:    []UserInfoForFeature{},
			}
			featuresMap[featureID] = feature
			orderedFeatureIDs = append(orderedFeatureIDs, featureID)
		}

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

	result := make([]*FeatureWithAssignedUsers, len(orderedFeatureIDs))
	for i, id := range orderedFeatureIDs {
		result[i] = featuresMap[id]
	}

	return result, nil
}
