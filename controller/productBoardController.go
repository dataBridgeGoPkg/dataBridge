package controller

import (
	"encoding/json"
	"io"
	"net/http"

	"example.com/Product_RoadMap/service" // Import the service package
	"github.com/gin-gonic/gin"
)

type validateProductFeature struct {
	Type        string `json:"type"`
	StatusID    string `json:"status_id"`
	ProductID   string `json:"product_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	StartDate   string `json:"StartDate"`
	EndDate     string `json:"EndDate"`
}

func CreateProductBoardFeature(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var productFeature validateProductFeature

	err = json.Unmarshal(body, &productFeature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Call the ConstructPayload function from the service package
	productFeatureConverted := service.ProductFeature{
		Type:        productFeature.Type,
		StatusID:    productFeature.StatusID,
		ProductID:   productFeature.ProductID,
		Name:        productFeature.Name,
		Description: productFeature.Description,
		StartDate:   productFeature.StartDate,
		EndDate:     productFeature.EndDate,
	}

	//Calling the product board API
	response, err := service.ProductBoardAPI(productFeatureConverted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call ProductBoardAPI", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ProductBoardAPI called successfully", "response": response})
}
