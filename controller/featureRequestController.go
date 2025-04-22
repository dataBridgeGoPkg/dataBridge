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
	Accepted    bool   `json:"accepted" binding:"required"`
	CreatedAt   int64  `json:"created_at,omitempty"` // Unix timestamp
	UpdatedAt   int64  `json:"updated_at,omitempty"` // Unix timestamp

}

type FeatureRequestResponse struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Accepted    bool   `json:"accepted"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

func CreateFeatureRequest(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read request body"})
		return
	}

	// Unmarshal the body into the FeatureRequest struct
	var featureRequest FeatureRequest
	err = json.Unmarshal(body, &featureRequest)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Validate the request
	if featureRequest.Title == "" || featureRequest.Description == "" {
		c.JSON(400, gin.H{"error": "Title and Description are required"})
		return
	}
	// Create a new FeatureRequest object
	featureRequestModel := models.FeatureRequestModel{
		Title:       featureRequest.Title,
		Description: featureRequest.Description,
		Accepted:    featureRequest.Accepted,
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
		CreatedAt:   featureRequestModel.CreatedAt,
		UpdatedAt:   featureRequestModel.UpdatedAt,
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
		CreatedAt:   getFeatureRequest.CreatedAt,
		UpdatedAt:   getFeatureRequest.UpdatedAt,
	}
	c.JSON(http.StatusOK, response)

}

func UpdateFeatureRequestByID(c *gin.Context) {
	iD := c.Param("id")
	featureID := utils.ParseID(iD)
	if featureID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature request ID"})
		return
	}
	// Check if the FeatureRequest exists
	checkFeatureRequest, err := models.FetchFeatureRequestByID(models.DB, featureID)
	if checkFeatureRequest == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	var featureRequest FeatureRequest
	err = c.BindJSON(&featureRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}
	// Validate the request
	if featureRequest.Title == "" || featureRequest.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title and Description are required"})
		return
	}
	// Update the FeatureRequest object with the new values
	featureRequestModel := models.FeatureRequestModel{
		ID:          featureID,
		Title:       featureRequest.Title,
		Description: featureRequest.Description,
		Accepted:    featureRequest.Accepted,
	}
	err = models.UpdateFeatureRequestByID(models.DB, &featureRequestModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	response := FeatureRequestResponse{
		ID:          featureRequestModel.ID,
		Title:       featureRequestModel.Title,
		Description: featureRequestModel.Description,
		Accepted:    featureRequestModel.Accepted,
		CreatedAt:   featureRequestModel.CreatedAt,
		UpdatedAt:   featureRequestModel.UpdatedAt,
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
	// Check if the FeatureRequest exists
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
	// Fetch all features from the database
	features, err := models.FetchAllFeatureRequests(models.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	// Prepare the response
	var responses []FeatureRequestResponse
	for _, f := range features {
		responses = append(responses, FeatureRequestResponse{
			ID:          f.ID,
			Title:       f.Title,
			Description: f.Description,
			Accepted:    f.Accepted,
			CreatedAt:   f.CreatedAt,
			UpdatedAt:   f.UpdatedAt,
		})
	}

	// Return the response
	c.JSON(http.StatusOK, responses)
}
