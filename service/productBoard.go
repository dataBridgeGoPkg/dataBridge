package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}
}

type ProductFeature struct {
	Type        string `json:"type"`
	StatusID    string `json:"status_id"`
	ProductID   string `json:"product_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	StartDate   string `json:"StartDate"`
	EndDate     string `json:"EndDate"`
}

func ProductBoardAPI(feature ProductFeature) (string, error) {
	productBoardAccessToken := os.Getenv("PRODUCT_BOARD_ACCESS_TOKEN")
	if productBoardAccessToken == "" {
		return "", fmt.Errorf("PRODUCT_BOARD_ACCESS_TOKEN is not set in the environment")
	}

	// Construct JSON payload
	payload, err := json.Marshal(ConstructPayload(feature))
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Set the API URL from environment variable
	// Ensure the environment variable is set
	if os.Getenv("PRODUCTBOARD_API_URL") == "" {
		return "", fmt.Errorf("PRODUCTBOARD_API_URL is not set in the environment")
	}
	// Use the environment variable for the API URL
	url := os.Getenv("PRODUCTBOARD_API_URL")

	fmt.Println(url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("X-Version", "1")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", "Bearer "+productBoardAccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make API request: %w", err)
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if res.StatusCode >= 400 {
		return "", fmt.Errorf("ProductBoard API error: %s", responseBody)
	}

	return string(responseBody), nil
}

func ConstructPayload(feature ProductFeature) map[string]interface{} {
	dataPayload := map[string]interface{}{
		"type": feature.Type,
		"status": map[string]interface{}{
			"id": feature.StatusID,
		},
		"parent": map[string]interface{}{
			"product": map[string]interface{}{
				"id": feature.ProductID,
			},
		},
		"timeframe": map[string]interface{}{
			"startDate": feature.StartDate,
			"endDate":   feature.EndDate,
		},
		"name":        feature.Name,
		"description": feature.Description,
	}

	finalPayload := map[string]interface{}{
		"data": dataPayload,
	}

	// Pretty print the final payload as JSON
	// finalJSON, err := json.MarshalIndent(finalPayload, "", "  ")
	// if err != nil {
	// 	fmt.Println("Failed to marshal payload for preview:", err)
	// } else {
	// 	fmt.Println("Constructed Payload to be sent:")
	// 	fmt.Println(string(finalJSON))
	// }

	return finalPayload
}
