package models

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// InitDB initializes the SQLite database connection
var DB *sql.DB

func Init() {
	dsn := "root:Bahvyg-dadzaj-bocti1@tcp(127.0.0.1:3306)/product_road_map"
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Error connecting to DB: %v", err)
	}

	fmt.Println("Connected to MySQL!")
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
