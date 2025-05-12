package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"example.com/Product_RoadMap/models"
	"example.com/Product_RoadMap/utils"
	"github.com/gin-gonic/gin"
)

type Feature struct {
	Title        string            `json:"title" binding:"required"`
	Description  string            `json:"description" binding:"required"`
	Status       models.StatusType `json:"status" binding:"required"`
	StartTime    *int64            `json:"start_time,omitempty"`
	EndTime      *int64            `json:"end_time,omitempty"`
	Notes        string            `json:"notes,omitempty"`
	AssignedUser *int64            `json:"assigned_user,omitempty"` // Accepts int64 or null
}

type response struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       models.StatusType `json:"status"`
	StartTime    *int64            `json:"start_time,omitempty"`
	EndTime      *int64            `json:"end_time,omitempty"`
	Notes        *string           `json:"notes,omitempty"`
	AssignedUser *int64            `json:"assigned_user,omitempty"`
	CreatedAt    int64             `json:"created_at"`
	UpdatedAt    int64             `json:"updated_at"`
}

func CreateFeatures(c *gin.Context) {

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//unmarshal the body into the FeatureRequest struct
	var featureRequest Feature
	if err := json.Unmarshal(body, &featureRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}
	// Create a new Feature object
	var userID int64
	if featureRequest.AssignedUser != nil {
		userID = *featureRequest.AssignedUser // Dereference the pointer to get the actual value
	} else {
		userID = 0 // Default to 0 if AssignedUser is nil
	}

	//Check if the userID exists
	checkUser, err := models.GetUserByID(models.DB, userID)
	if checkUser == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	fmt.Println(checkUser.Role)

	if checkUser.Role != "ADMIN" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Only Admins can create features"})
		return
	}

	// Create a new Feature object
	feature := models.Feature{
		Title:        featureRequest.Title,
		Description:  featureRequest.Description,
		Status:       featureRequest.Status,
		StartTime:    featureRequest.StartTime,
		EndTime:      featureRequest.EndTime,
		Notes:        &featureRequest.Notes,
		AssignedUser: featureRequest.AssignedUser,
	}

	if err := models.CreateFeature(models.DB, &feature); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	response := response{
		ID:           feature.ID,
		Title:        feature.Title,
		Description:  feature.Description,
		Status:       feature.Status,
		StartTime:    feature.StartTime,
		EndTime:      feature.EndTime,
		Notes:        feature.Notes,
		AssignedUser: feature.AssignedUser,
		CreatedAt:    feature.CreatedAt,
		UpdatedAt:    feature.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)

}

func GetFeatureByID(c *gin.Context) {
	id := c.Param("id")
	featureID := utils.ParseID(id)
	if featureID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	//Check if the Feature exists
	getFeature, err := models.GetFeatureByID(models.DB, featureID)
	if getFeature == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	response := response{
		ID:           getFeature.ID,
		Title:        getFeature.Title,
		Description:  getFeature.Description,
		Status:       getFeature.Status,
		StartTime:    getFeature.StartTime,
		EndTime:      getFeature.EndTime,
		Notes:        getFeature.Notes,
		AssignedUser: getFeature.AssignedUser,
		CreatedAt:    getFeature.CreatedAt,
		UpdatedAt:    getFeature.UpdatedAt,
	}
	c.JSON(http.StatusOK, response)
}

func DeletFeatureById(c *gin.Context) {
	id := c.Param("id")
	featureID := utils.ParseID(id)

	if featureID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	//Check if the feature exists
	getFeature, err := models.GetFeatureByID(models.DB, featureID)
	if getFeature == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	if err := models.DeleteFeature(models.DB, featureID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Feature deleted successfully"})
}

func UpdateFeatureById(c *gin.Context) {

	type UpdateFeatureInput struct {
		Title        *string `json:"title"`
		Description  *string `json:"description"`
		Status       *string `json:"status"`
		StartTime    *int64  `json:"start_time"`
		EndTime      *int64  `json:"end_time"`
		Notes        *string `json:"notes"`
		AssignedUser *int64  `json:"assigned_user"`
	}

	// Parse feature ID
	id := c.Param("id")
	featureID := utils.ParseID(id)
	if featureID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	// Check if feature exists
	existingFeature, err := models.GetFeatureByID(models.DB, featureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	if existingFeature == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	// Bind input
	var input UpdateFeatureInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	// Merge updates into existing feature
	if input.Title != nil {
		existingFeature.Title = *input.Title
	}
	if input.Description != nil {
		existingFeature.Description = *input.Description
	}
	if input.Status != nil {
		existingFeature.Status = models.StatusType(*input.Status)
	}
	if input.StartTime != nil {
		startTime := *input.StartTime
		existingFeature.StartTime = &startTime
	}
	if input.EndTime != nil {
		endTime := *input.EndTime
		existingFeature.EndTime = &endTime
	}
	if input.Notes != nil {
		existingFeature.Notes = input.Notes
	}
	if input.AssignedUser != nil {
		existingFeature.AssignedUser = input.AssignedUser
	}

	// Update in DB
	if err := models.UpdateFeature(models.DB, existingFeature); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	// Response
	updateResponse := response{
		ID:           existingFeature.ID,
		Title:        existingFeature.Title,
		Description:  existingFeature.Description,
		Status:       existingFeature.Status,
		StartTime:    existingFeature.StartTime,
		EndTime:      existingFeature.EndTime,
		Notes:        existingFeature.Notes,
		AssignedUser: existingFeature.AssignedUser,
		CreatedAt:    existingFeature.CreatedAt,
		UpdatedAt:    existingFeature.UpdatedAt,
	}

	c.JSON(http.StatusOK, updateResponse)
}

func GetAllFeatures(c *gin.Context) {
	// Fetch all features from the database
	features, err := models.GetAllFeatures(models.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	// Prepare the response
	var responses []response
	for _, f := range features {
		responses = append(responses, response{
			ID:           f.ID,
			Title:        f.Title,
			Description:  f.Description,
			Status:       f.Status,
			StartTime:    f.StartTime, // Use the pointer field directly
			EndTime:      f.EndTime,   // Use the pointer field directly
			Notes:        f.Notes,     // Use the pointer field directly
			AssignedUser: f.AssignedUser,
			CreatedAt:    f.CreatedAt,
			UpdatedAt:    f.UpdatedAt,
		})
	}

	// Return the response
	c.JSON(http.StatusOK, responses)
}

func GetAllFeaturesWithUserName(c *gin.Context) {
	features, err := models.GetAllFeaturesWithUserName(models.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	var responses []models.FeatureModelWithUserName

	for _, f := range features {
		responses = append(responses, models.FeatureModelWithUserName{
			ID:            f.ID,
			Title:         f.Title,
			Description:   f.Description,
			Status:        f.Status,
			StartTime:     f.StartTime,
			EndTime:       f.EndTime,
			Notes:         f.Notes,
			AssignedUser:  f.AssignedUser,
			UserFirstName: f.UserFirstName,
			UserLastName:  f.UserLastName,
			CreatedAt:     f.CreatedAt,
			UpdatedAt:     f.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, responses)
}
