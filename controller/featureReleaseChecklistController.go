package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"example.com/Product_RoadMap/models"
	"github.com/gin-gonic/gin"
)

type FeatureReleaseChecklistController struct {
	FeatureId *int64  `json:"feature_id"` // Foreign key to features table
	Item      *string `json:"item"`       // Checklist item name
	Validated *bool   `json:"validated"`  // true if validated, false otherwise
}

type DefaultCheckList []struct {
	Item      string `json:"item"`
	Validated bool   `json:"validated"`
}

func CreateFeatureReleaseChecklist(c *gin.Context) {
	checkListBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	// Unmarshal the request body into the controller struct
	var checkList FeatureReleaseChecklistController
	err = json.Unmarshal(checkListBody, &checkList)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	fmt.Println(checkList)

	//Verify if the feature ID exists

	if checkList.FeatureId == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Feature ID is required"})
		return
	}
	checkFeatureId, err := models.GetFeatureByID(models.DB, *checkList.FeatureId)

	if checkFeatureId == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error (feature lookup)", "details": err.Error()})
		return
	}

	// Create the checklist item
	checklist := models.FeatureReleaseChecklist{
		FeatureID: *checkList.FeatureId,
		Item:      *checkList.Item,
		Validated: *checkList.Validated,
	}

	err = models.CreateFeatureReleaseChecklist(models.DB, &checklist)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error (checklist creation)", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Feature release checklist created successfully", "checklist_id": checklist.ID})

}

// Get Featture Realease Checklist by ID
func GetFeatureReleaseChecklistByCheckListID(c *gin.Context) {
	checklistIDStr := c.Param("id")

	// Convert checklistID from string to int64
	checklistID, err := strconv.ParseInt(checklistIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checklist_id"})
		return
	}

	// Fetch the checklist item by ID
	checklistItem, err := models.GetFeatureReleaseChecklistByID(models.DB, checklistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error (checklist retrieval)", "details": err.Error()})
		return
	}
	if checklistItem == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Checklist item not found"})
		return
	}

	c.JSON(http.StatusOK, checklistItem)
}

func UpdateFeatureReleaseChecklistByID(c *gin.Context) {

	// Parse and validate checklist ID
	checklistIDStr := c.Param("id")
	checklistID, err := strconv.ParseInt(checklistIDStr, 10, 64)
	if err != nil || checklistID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checklist ID"})
		return
	}

	// Fetch existing checklist
	existingChecklist, err := models.GetFeatureReleaseChecklistByID(models.DB, checklistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	if existingChecklist == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Checklist not found"})
		return
	}

	// Bind JSON input
	var input FeatureReleaseChecklistController
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	// Update only provided fields
	if input.FeatureId != nil {
		existingChecklist.FeatureID = *input.FeatureId
	}
	if input.Item != nil {
		existingChecklist.Item = *input.Item
	}
	if input.Validated != nil {
		existingChecklist.Validated = *input.Validated
	}

	// Save the updated checklist
	if err := models.UpdateFeatureReleaseChecklist(models.DB, existingChecklist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update checklist", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Feature release checklist updated successfully",
		"checklist": existingChecklist,
	})
}

func DeleteFeatureReleaseChecklistByID(c *gin.Context) {
	checklistIDStr := c.Param("id")

	// Convert checklistID from string to int64
	checklistID, err := strconv.ParseInt(checklistIDStr, 10, 64)
	if err != nil || checklistID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checklist_id"})
		return
	}

	// Delete the checklist item by ID
	err = models.DeleteFeatureReleaseChecklist(models.DB, checklistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error (checklist deletion)", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feature release checklist deleted successfully"})
}

func GetFeatureReleaseChecklistByFeatureID(c *gin.Context) {
	featureIDStr := c.Param("feature_id")

	// Convert featureID from string to int64
	featureID, err := strconv.ParseInt(featureIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature_id"})
		return
	}

	// Fetch the checklist items for the given feature ID
	checklistItems, err := models.GetFeatureReleaseChecklistByFeatureID(models.DB, featureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error (checklist retrieval)", "details": err.Error()})
		return
	}
	if len(checklistItems) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No checklist items found for this feature"})
		return
	}

	c.JSON(http.StatusOK, checklistItems)
}

func CreateDefaultFeatureCheckList(c *gin.Context) {

	// Verify if the feature ID exists
	featureIDStr := c.Param("feature_id")

	// Convert featureID from string to int64
	featureID, err := strconv.ParseInt(featureIDStr, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature_id"})
		return
	}
	checkFeatureId, err := models.GetFeatureByID(models.DB, featureID)
	if checkFeatureId == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error (feature lookup)", "details": err.Error()})
		return
	}

	// Read the body
	defaultChecklistBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}
	var defaultChecklist DefaultCheckList
	err = json.Unmarshal(defaultChecklistBody, &defaultChecklist)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Create the default checklist items
	for _, item := range defaultChecklist {
		checklist := models.FeatureReleaseChecklist{
			FeatureID: featureID,
			Item:      item.Item,
			Validated: item.Validated,
		}

		err = models.CreateFeatureReleaseChecklist(models.DB, &checklist)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error (default checklist creation)", "details": err.Error()})
			return
		}

	}

	c.JSON(http.StatusCreated, gin.H{"message": "Default feature release checklist created successfully"})

}
