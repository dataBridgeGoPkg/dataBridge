package controller

import (
	"encoding/json"
	"io"
	"net/http"

	"example.com/Product_RoadMap/models"
	"example.com/Product_RoadMap/utils"
	"github.com/gin-gonic/gin"
)

type FeatureRequest struct {
	ID          int64  `json:"id,omitempty"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	Accepted    *bool  `json:"accepted,omitempty"`
	RequestedBy string `json:"requested_by,omitempty"`
	ProductID   int64  `json:"product_id"`
	CreatedAt   int64  `json:"created_at,omitempty"` // Unix timestamp
	UpdatedAt   int64  `json:"updated_at,omitempty"` // Unix timestamp
}

type FeatureRequestResponse struct {
	ID          int64   `json:"id,omitempty"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ProductID   int64   `json:"product_id"`
	Accepted    *bool   `json:"accepted,omitempty"`
	RequestedBy *string `json:"requested_by,omitempty"`
	CreatedAt   int64   `json:"created_at,omitempty"`
	UpdatedAt   int64   `json:"updated_at,omitempty"`
}

type UpdateFeatureRequestInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Accepted    *bool   `json:"accepted"`
	RequestedBy *string `json:"requested_by"`
}

func CreateFeatureRequest(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read request body"})
		return
	}

	var featureRequest FeatureRequest
	err = json.Unmarshal(body, &featureRequest)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	if featureRequest.Title == "" || featureRequest.Description == "" {
		c.JSON(400, gin.H{"error": "Title and Description are required"})
		return
	}

	productID := featureRequest.ProductID

	//Check product ID is valid
	product, err := models.GetProductByID(models.DB, productID)
	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	featureRequestModel := models.FeatureRequestModel{
		Title:       featureRequest.Title,
		Description: featureRequest.Description,
		Accepted:    featureRequest.Accepted,
		RequestedBy: &featureRequest.RequestedBy,
		ProductID:   productID,
	}

	if featureRequest.Accepted != nil {
		featureRequestModel.Accepted = featureRequest.Accepted
	}

	if err := models.CreateFeatureRequest(models.DB, &featureRequestModel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	response := FeatureRequestResponse{
		ID:          featureRequestModel.ID,
		Title:       featureRequestModel.Title,
		Description: featureRequestModel.Description,
		Accepted:    featureRequestModel.Accepted,
		RequestedBy: featureRequestModel.RequestedBy,
		CreatedAt:   featureRequestModel.CreatedAt,
		UpdatedAt:   featureRequestModel.UpdatedAt,
		ProductID:   productID,
	}

	c.JSON(http.StatusCreated, response)
}

func GetFeatureRequestByID(c *gin.Context) {
	iD := c.Param("id")
	featureID := utils.ParseID(iD)
	if featureID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature request ID"})
		return
	}
	getFeatureRequest, err := models.FetchFeatureRequestByID(models.DB, featureID)
	if getFeatureRequest == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	response := FeatureRequestResponse{
		ID:          getFeatureRequest.ID,
		Title:       getFeatureRequest.Title,
		Description: getFeatureRequest.Description,
		Accepted:    getFeatureRequest.Accepted,
		RequestedBy: getFeatureRequest.RequestedBy,
		CreatedAt:   getFeatureRequest.CreatedAt,
		UpdatedAt:   getFeatureRequest.UpdatedAt,
		ProductID:   getFeatureRequest.ProductID,
	}
	c.JSON(http.StatusOK, response)
}

func UpdateFeatureRequestByID(c *gin.Context) {
	// Parse feature request ID
	id := c.Param("id")
	featureID := utils.ParseID(id)
	if featureID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature request ID"})
		return
	}

	// Fetch existing feature request
	existingFeatureRequest, err := models.FetchFeatureRequestByID(models.DB, featureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	if existingFeatureRequest == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature request not found"})
		return
	}

	// Bind input
	var input UpdateFeatureRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Merge fields
	if input.Title != nil {
		existingFeatureRequest.Title = *input.Title
	}
	if input.Description != nil {
		existingFeatureRequest.Description = *input.Description
	}
	existingFeatureRequest.Accepted = input.Accepted
	existingFeatureRequest.RequestedBy = input.RequestedBy

	// Validate required fields
	if existingFeatureRequest.Title == "" || existingFeatureRequest.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title and Description are required"})
		return
	}

	// Update in DB
	if err := models.UpdateFeatureRequestByID(models.DB, existingFeatureRequest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	// Return updated response
	response := FeatureRequestResponse{
		ID:          existingFeatureRequest.ID,
		Title:       existingFeatureRequest.Title,
		Description: existingFeatureRequest.Description,
		Accepted:    existingFeatureRequest.Accepted,
		RequestedBy: existingFeatureRequest.RequestedBy,
		CreatedAt:   existingFeatureRequest.CreatedAt,
		UpdatedAt:   existingFeatureRequest.UpdatedAt,
		ProductID:   existingFeatureRequest.ProductID,
	}
	c.JSON(http.StatusOK, response)
}

func DeleteFeatureRequestByID(c *gin.Context) {
	iD := c.Param("id")
	featureID := utils.ParseID(iD)
	if featureID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature request ID"})
		return
	}
	checkFeatureRequest, err := models.FetchFeatureRequestByID(models.DB, featureID)
	if checkFeatureRequest == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	err = models.DeleteFeatureRequestByID(models.DB, featureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Feature deleted successfully"})
}

func GetAllFeatureRequests(c *gin.Context) {
	features, err := models.FetchAllFeatureRequests(models.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	productID := utils.ParseID(c.Param("product_id"))
	var responses []FeatureRequestResponse
	for _, f := range features {
		if productID == 0 || f.ProductID == productID {
			responses = append(responses, FeatureRequestResponse{
				ID:          f.ID,
				Title:       f.Title,
				Description: f.Description,
				Accepted:    f.Accepted,
				RequestedBy: f.RequestedBy,
				CreatedAt:   f.CreatedAt,
				UpdatedAt:   f.UpdatedAt,
				ProductID:   f.ProductID,
			})
		}
	}

	c.JSON(http.StatusOK, responses)
}
