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
	Title            string               `json:"title" binding:"required"`
	Description      string               `json:"description" binding:"required"`
	Status           models.StatusType    `json:"status" binding:"required"`
	Health           models.FeatureHealth `json:"health,omitempty"`
	StartTime        *int64               `json:"start_time,omitempty"`
	EndTime          *int64               `json:"end_time,omitempty"`
	Notes            string               `json:"notes,omitempty"`
	AssignedUser     *int64               `json:"assigned_user,omitempty"`
	FeatureDocUrl    *string              `json:"feature_doc_url,omitempty"`
	FigmaUrl         *string              `json:"figma_url,omitempty"`
	Insights         *string              `json:"insights,omitempty"`
	JiraSync         *bool                `json:"jira_sync,omitempty"`
	ProductBoardSync *bool                `json:"product_board_sync,omitempty"`
	JiraID           *string              `json:"jira_id,omitempty"`
	ProductBoardID   *string              `json:"product_board_id,omitempty"`
	BusinessCase     *string              `json:"business_case,omitempty"`
}

type response struct {
	ID               int64                `json:"id"`
	Title            string               `json:"title"`
	Description      string               `json:"description"`
	Status           models.StatusType    `json:"status"`
	Health           models.FeatureHealth `json:"health,omitempty"`
	StartTime        *int64               `json:"start_time,omitempty"`
	EndTime          *int64               `json:"end_time,omitempty"`
	Notes            *string              `json:"notes,omitempty"`
	FeatureDocUrl    *string              `json:"feature_doc_url,omitempty"`
	FigmaUrl         *string              `json:"figma_url,omitempty"`
	Insights         *string              `json:"insights,omitempty"`
	JiraSync         *bool                `json:"jira_sync,omitempty"`
	ProductBoardSync *bool                `json:"product_board_sync,omitempty"`
	JiraID           *string              `json:"jira_id,omitempty"`
	ProductBoardID   *string              `json:"product_board_id,omitempty"`
	BusinessCase     *string              `json:"business_case,omitempty"`
	CreatedAt        int64                `json:"created_at"`
	UpdatedAt        int64                `json:"updated_at"`
}

type FeatureAssignee struct {
	UserIds   []int `json:"user_ids"`
	FeatureId int   `json:"feature_id"`
}

type AssigneeFeature struct {
	FeatureID int64 `json:"feature_id"`
	UserID    int64 `json:"user_id"`
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
	}

	//Check if the assigned userID exists
	checkUser, err := models.GetUserByID(models.DB, userID)
	if checkUser == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	// Create a new Feature object
	feature := models.Feature{
		Title:            featureRequest.Title,
		Description:      featureRequest.Description,
		Status:           featureRequest.Status,
		Health:           featureRequest.Health,
		StartTime:        featureRequest.StartTime,
		EndTime:          featureRequest.EndTime,
		Notes:            &featureRequest.Notes,
		FeatureDocUrl:    featureRequest.FeatureDocUrl,
		FigmaUrl:         featureRequest.FigmaUrl,
		Insights:         featureRequest.Insights,
		JiraSync:         featureRequest.JiraSync,
		ProductBoardSync: featureRequest.ProductBoardSync,
		JiraID:           featureRequest.JiraID,
		ProductBoardID:   featureRequest.ProductBoardID,
		BusinessCase:     featureRequest.BusinessCase,
	}

	if err := models.CreateFeature(models.DB, &feature); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	response := response{
		ID:               feature.ID,
		Title:            feature.Title,
		Description:      feature.Description,
		Status:           feature.Status,
		Health:           feature.Health,
		StartTime:        feature.StartTime,
		EndTime:          feature.EndTime,
		Notes:            feature.Notes,
		FeatureDocUrl:    feature.FeatureDocUrl,
		FigmaUrl:         feature.FigmaUrl,
		Insights:         feature.Insights,
		JiraSync:         feature.JiraSync,
		ProductBoardSync: feature.ProductBoardSync,
		JiraID:           feature.JiraID,
		ProductBoardID:   feature.ProductBoardID,
		BusinessCase:     feature.BusinessCase,
		CreatedAt:        feature.CreatedAt,
		UpdatedAt:        feature.UpdatedAt,
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
	fmt.Println(getFeature)

	if getFeature == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	response := response{
		ID:               getFeature.ID,
		Title:            getFeature.Title,
		Description:      getFeature.Description,
		Status:           getFeature.Status,
		Health:           getFeature.Health,
		StartTime:        getFeature.StartTime,
		EndTime:          getFeature.EndTime,
		Notes:            getFeature.Notes,
		FeatureDocUrl:    getFeature.FeatureDocUrl,
		FigmaUrl:         getFeature.FigmaUrl,
		Insights:         getFeature.Insights,
		JiraSync:         getFeature.JiraSync,
		ProductBoardSync: getFeature.ProductBoardSync,
		JiraID:           getFeature.JiraID,
		ProductBoardID:   getFeature.ProductBoardID,
		BusinessCase:     getFeature.BusinessCase,
		CreatedAt:        getFeature.CreatedAt,
		UpdatedAt:        getFeature.UpdatedAt,
	}
	c.JSON(http.StatusOK, response)
}

// func DeletFeatureById(c *gin.Context) {
// 	id := c.Param("id")
// 	featureID := utils.ParseID(id)

// 	if featureID == 0 {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
// 		return
// 	}

// 	//Check if the feature exists
// 	getFeature, err := models.GetFeatureByID(models.DB, featureID)
// 	if getFeature == nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
// 		return
// 	}
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
// 		return
// 	}

// 	if err := models.DeleteFeature(models.DB, featureID); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{"message": "Feature deleted successfully"})
// }

func DeletFeatureById(c *gin.Context) {
	id := c.Param("id")
	featureID := utils.ParseID(id)

	if featureID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	// 1. Check if the feature exists
	getFeature, err := models.GetFeatureByID(models.DB, featureID)
	if getFeature == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	// 2. Delete associated feature assignments from feature_assignees table
	if err := models.DeleteFeatureAssigneeWithFeatureId(models.DB, featureID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete feature assignments", "details": err.Error()})
		return
	}

	// 3. Delete the feature from the features table
	if err := models.DeleteFeature(models.DB, featureID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete feature", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feature and associated assignments deleted successfully"})
}

func UpdateFeatureById(c *gin.Context) {

	type UpdateFeatureInput struct {
		Title            *string               `json:"title"`
		Description      *string               `json:"description"`
		Status           *models.StatusType    `json:"status" binding:"required"`
		Health           *models.FeatureHealth `json:"health,omitempty"`
		StartTime        *int64                `json:"start_time"`
		EndTime          *int64                `json:"end_time"`
		Notes            *string               `json:"notes"`
		FeatureDocUrl    *string               `json:"feature_doc_url,omitempty"`
		FigmaUrl         *string               `json:"figma_url,omitempty"`
		Insights         *string               `json:"insights,omitempty"`
		JiraSync         *bool                 `json:"jira_sync,omitempty"`
		ProductBoardSync *bool                 `json:"product_board_sync,omitempty"`
		JiraID           *string               `json:"jira_id,omitempty"`
		ProductBoardID   *string               `json:"product_board_id,omitempty"`
		BusinessCase     *string               `json:"business_case,omitempty"`
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
	if input.Health != nil {
		existingFeature.Health = models.FeatureHealth(*input.Health)
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
	if input.FeatureDocUrl != nil {
		existingFeature.FeatureDocUrl = input.FeatureDocUrl
	}
	if input.FigmaUrl != nil {
		existingFeature.FigmaUrl = input.FigmaUrl
	}
	if input.Insights != nil {
		existingFeature.Insights = input.Insights
	}
	if input.JiraSync != nil {
		existingFeature.JiraSync = input.JiraSync
	}
	if input.ProductBoardSync != nil {
		existingFeature.ProductBoardSync = input.ProductBoardSync
	}
	if input.JiraID != nil {
		existingFeature.JiraID = input.JiraID
	}
	if input.ProductBoardID != nil {
		existingFeature.ProductBoardID = input.ProductBoardID
	}
	if input.BusinessCase != nil {
		existingFeature.BusinessCase = input.BusinessCase
	}

	// Update in DB
	if err := models.UpdateFeature(models.DB, existingFeature); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	// Response
	updateResponse := response{
		ID:               existingFeature.ID,
		Title:            existingFeature.Title,
		Description:      existingFeature.Description,
		Status:           existingFeature.Status,
		Health:           existingFeature.Health,
		StartTime:        existingFeature.StartTime,
		EndTime:          existingFeature.EndTime,
		Notes:            existingFeature.Notes,
		FeatureDocUrl:    existingFeature.FeatureDocUrl,
		FigmaUrl:         existingFeature.FigmaUrl,
		Insights:         existingFeature.Insights,
		JiraSync:         existingFeature.JiraSync,
		ProductBoardSync: existingFeature.ProductBoardSync,
		JiraID:           existingFeature.JiraID,
		ProductBoardID:   existingFeature.ProductBoardID,
		BusinessCase:     existingFeature.BusinessCase,
		CreatedAt:        existingFeature.CreatedAt,
		UpdatedAt:        existingFeature.UpdatedAt,
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
			ID:               f.ID,
			Title:            f.Title,
			Description:      f.Description,
			Status:           f.Status,
			Health:           f.Health,
			StartTime:        f.StartTime,
			EndTime:          f.EndTime,
			Notes:            f.Notes,
			FeatureDocUrl:    f.FeatureDocUrl,
			FigmaUrl:         f.FigmaUrl,
			Insights:         f.Insights,
			JiraSync:         f.JiraSync,
			ProductBoardSync: f.ProductBoardSync,
			JiraID:           f.JiraID,
			ProductBoardID:   f.ProductBoardID,
			BusinessCase:     f.BusinessCase,
			CreatedAt:        f.CreatedAt,
			UpdatedAt:        f.UpdatedAt,
		})
	}

	// Return the response
	c.JSON(http.StatusOK, responses)
}

func GetAllFeaturesWithAssginness(c *gin.Context) {

	//fmt.Println(c.Keys["user_id"])

	sourceFeatures, err := models.GetAllFeaturesWithUserName(models.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	var responses []models.FeatureWithAssignedUsers
	for _, f := range sourceFeatures {
		if f == nil {
			continue
		}
		responses = append(responses, models.FeatureWithAssignedUsers{
			ID:               f.ID,
			Title:            f.Title,
			Description:      f.Description,
			Status:           f.Status,
			StartTime:        f.StartTime,
			EndTime:          f.EndTime,
			Notes:            f.Notes,
			FeatureDocUrl:    f.FeatureDocUrl,
			FigmaUrl:         f.FigmaUrl,
			Insights:         f.Insights,
			JiraSync:         f.JiraSync,
			ProductBoardSync: f.ProductBoardSync,
			JiraID:           f.JiraID,
			ProductBoardID:   f.ProductBoardID,
			BusinessCase:     f.BusinessCase,
			AssignedUsers:    f.AssignedUsers,
			CreatedAt:        f.CreatedAt,
			UpdatedAt:        f.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, responses)
}

func CreateFeatureAssignee(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var featureAssigneeRequest FeatureAssignee
	if err := json.Unmarshal(body, &featureAssigneeRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	// Check if the feature exists
	checkFeature, err := models.GetFeatureByID(models.DB, int64(featureAssigneeRequest.FeatureId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	if checkFeature == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Feature with ID %d not found", featureAssigneeRequest.FeatureId)})
		return
	}

	// Loop over user IDs
	for _, userID := range featureAssigneeRequest.UserIds {
		// Check if user exists
		checkUser, err := models.GetUserByID(models.DB, int64(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
			return
		}
		if checkUser == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("User with ID %d not found", userID)})
			return
		}

		// Check for duplicate assignment
		alreadyAssigned, err := models.CheckIfUserAlreadyAssigned(models.DB, int64(featureAssigneeRequest.FeatureId), int64(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error (check assignment)", "details": err.Error()})
			return
		}
		if alreadyAssigned {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("User ID %d is already assigned to Feature ID %d", userID, featureAssigneeRequest.FeatureId)})
			return
		}

		// Create new feature_assignee row
		featureAssignee := models.FeatureAssignee{
			FeatureID: int64(featureAssigneeRequest.FeatureId),
			UserID:    int64(userID),
		}
		if err := models.CreateFeatureAssignee(models.DB, &featureAssignee); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Feature assignees created successfully"})
}

func AddUserToAFeature(c *gin.Context) {
	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Unmarshal into your existing AssigneeFeature struct
	var featureAssigneeRequest AssigneeFeature
	if err := json.Unmarshal(body, &featureAssigneeRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	// Check if the user exists
	user, err := models.GetUserByID(models.DB, featureAssigneeRequest.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error (user lookup)", "details": err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if the feature exists
	feature, err := models.GetFeatureByID(models.DB, featureAssigneeRequest.FeatureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error (feature lookup)", "details": err.Error()})
		return
	}
	if feature == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	// Check for duplicate assignment
	alreadyAssigned, err := models.CheckIfUserAlreadyAssigned(models.DB, featureAssigneeRequest.FeatureID, featureAssigneeRequest.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error (check assignment)", "details": err.Error()})
		return
	}
	if alreadyAssigned {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("User ID %d is already assigned to Feature ID %d", featureAssigneeRequest.UserID, featureAssigneeRequest.FeatureID)})
		return
	}

	// Create the assignment
	newAssignee := models.FeatureAssignee{
		FeatureID: featureAssigneeRequest.FeatureID,
		UserID:    featureAssigneeRequest.UserID,
	}
	if err := models.CreateFeatureAssignee(models.DB, &newAssignee); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feature assignee", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":             "Feature assignee created successfully",
		"feature_assignee_id": newAssignee.FeatureAssigneeID,
	})
}

func DeleteAssigneeFromAFeature(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body", "details": err.Error()})
		return
	}

	// Unmarshal the request body into the UpdateDeleteAssignees struct
	var requestBody AssigneeFeature
	if err := json.Unmarshal(body, &requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	// Check if the feature exists
	feature, err := models.GetFeatureByID(models.DB, requestBody.FeatureID)
	if feature == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error (feature lookup)", "details": err.Error()})
		return
	}

	// Check if the user exists
	user, err := models.GetUserByID(models.DB, requestBody.UserID)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error (user lookup)", "details": err.Error()})
		return
	}

	//Delete the feature assignee
	if err := models.DeleteFeatureAssigneeWithUserId(models.DB, requestBody.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error (delete assignment)", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feature assignee deleted successfully"})
}
