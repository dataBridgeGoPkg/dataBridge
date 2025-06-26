package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"example.com/Product_RoadMap/models"
	"example.com/Product_RoadMap/utils"
	"github.com/gin-gonic/gin"
)

type ReleaseNoteRequest struct {
	ReleaseDate            string          `json:"release_date" binding:"required"`
	ReleasedFeatureDetails json.RawMessage `json:"released_feature_details" binding:"required"`
	ReleasedFeatureIDs     json.RawMessage `json:"released_feature_ids" binding:"required"`
}

type ReleaseNoteResponse struct {
	ID                     int64           `json:"id"`
	ReleaseDate            string          `json:"release_date"`
	ReleasedFeatureDetails json.RawMessage `json:"released_feature_details"`
	ReleasedFeatureIDs     json.RawMessage `json:"released_feature_ids"`
	CreatedAt              int64           `json:"created_at"`
	UpdatedAt              int64           `json:"updated_at"`
}

type ReleasedFeatureDetail struct {
	Type            string  `json:"type" binding:"required"`
	Title           string  `json:"title" binding:"required"`
	Image           *string `json:"image,omitempty"`
	Description     string  `json:"description" binding:"required"`
	HowItWorks      *string `json:"how_it_works,omitempty"`
	JiraID          *string `json:"jira_id,omitempty"`
	ProductBoardURL string  `json:"product_board_url" binding:"required"`
}

func CreateReleaseNote(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	var req ReleaseNoteRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	featureIds := string(req.ReleasedFeatureIDs)

	//Validate if the feature IDs are valid
	if featureIds == "" || featureIds == "null" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "released_feature_ids cannot be empty or null"})
		return
	}

	var featureIdArray []int64
	if err := json.Unmarshal(req.ReleasedFeatureIDs, &featureIdArray); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid released_feature_ids format. Must be a JSON array of integers."})
		return
	}

	for _, id := range featureIdArray {
		getFeature, err := models.GetFeatureByID(models.DB, id)

		if getFeature == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Feature %d not found", id)})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
			return
		}
	}

	// Validate ReleasedFeatureDetails
	var details []ReleasedFeatureDetail
	if err := json.Unmarshal(req.ReleasedFeatureDetails, &details); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid released_feature_details format. Must be a JSON array of objects."})
		return
	}

	//Validate if the feature IDs are valid
	var featureIDs []int64

	if err := json.Unmarshal(req.ReleasedFeatureIDs, &featureIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid released_feature_ids format. Must be a JSON array of integers."})
		return
	}
	for _, id := range featureIDs {
		if id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID in released_feature_ids", "id": id})
			return
		}
	}
	for i, d := range details {
		if d.Type == "" || d.Title == "" || d.Description == "" || d.ProductBoardURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Each released_feature_details object must have type, title, description, and product_board_url", "index": i})
			return
		}
	}
	// Marshal back to json.RawMessage to ensure valid format
	validDetails, _ := json.Marshal(details)

	note := models.ReleaseNote{
		ReleaseDate:            req.ReleaseDate,
		ReleasedFeatureDetails: validDetails,
		ReleasedFeatureIDs:     req.ReleasedFeatureIDs,
	}
	if err := models.CreateReleaseNote(models.DB, &note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	resp := ReleaseNoteResponse{
		ID:                     note.ID,
		ReleaseDate:            note.ReleaseDate,
		ReleasedFeatureDetails: note.ReleasedFeatureDetails,
		ReleasedFeatureIDs:     note.ReleasedFeatureIDs,
		CreatedAt:              note.CreatedAt,
		UpdatedAt:              note.UpdatedAt,
	}
	c.JSON(http.StatusCreated, resp)
}

func GetReleaseNoteByID(c *gin.Context) {
	id := utils.ParseID(c.Param("id"))
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid release note ID"})
		return
	}
	note, err := models.GetReleaseNoteByID(models.DB, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	if note == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Release note not found"})
		return
	}
	resp := ReleaseNoteResponse{
		ID:                     note.ID,
		ReleaseDate:            note.ReleaseDate,
		ReleasedFeatureDetails: note.ReleasedFeatureDetails,
		ReleasedFeatureIDs:     note.ReleasedFeatureIDs,
		CreatedAt:              note.CreatedAt,
		UpdatedAt:              note.UpdatedAt,
	}
	c.JSON(http.StatusOK, resp)
}

func UpdateReleaseNoteByID(c *gin.Context) {
	id := utils.ParseID(c.Param("id"))
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid release note ID"})
		return
	}
	note, err := models.GetReleaseNoteByID(models.DB, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	if note == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Release note not found"})
		return
	}
	var req ReleaseNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}
	// Validate ReleasedFeatureDetails
	var details []ReleasedFeatureDetail
	if err := json.Unmarshal(req.ReleasedFeatureDetails, &details); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid released_feature_details format. Must be a JSON array of objects."})
		return
	}
	for i, d := range details {
		if d.Type == "" || d.Title == "" || d.Description == "" || d.ProductBoardURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Each released_feature_details object must have type, title, description, and product_board_url", "index": i})
			return
		}
	}
	validDetails, _ := json.Marshal(details)

	note.ReleaseDate = req.ReleaseDate
	note.ReleasedFeatureDetails = validDetails
	note.ReleasedFeatureIDs = req.ReleasedFeatureIDs
	if err := models.UpdateReleaseNote(models.DB, note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	resp := ReleaseNoteResponse{
		ID:                     note.ID,
		ReleaseDate:            note.ReleaseDate,
		ReleasedFeatureDetails: note.ReleasedFeatureDetails,
		ReleasedFeatureIDs:     note.ReleasedFeatureIDs,
		CreatedAt:              note.CreatedAt,
		UpdatedAt:              note.UpdatedAt,
	}
	c.JSON(http.StatusOK, resp)
}

func DeleteReleaseNoteByID(c *gin.Context) {
	id := utils.ParseID(c.Param("id"))
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid release note ID"})
		return
	}
	note, err := models.GetReleaseNoteByID(models.DB, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	if note == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Release note not found"})
		return
	}
	if err := models.DeleteReleaseNote(models.DB, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Release note deleted successfully"})
}

func GetAllReleaseNotes(c *gin.Context) {
	notes, err := models.GetAllReleaseNotes(models.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	var resp []ReleaseNoteResponse
	for _, note := range notes {
		resp = append(resp, ReleaseNoteResponse{
			ID:                     note.ID,
			ReleaseDate:            note.ReleaseDate,
			ReleasedFeatureDetails: note.ReleasedFeatureDetails,
			ReleasedFeatureIDs:     note.ReleasedFeatureIDs,
			CreatedAt:              note.CreatedAt,
			UpdatedAt:              note.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, resp)
}

func UploadReleaseNoteImage(c *gin.Context) {
	file, fileHeader, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file"})
		return
	}
	defer file.Close()

	bucket := os.Getenv("AWS_S3_BUCKET")
	url, err := utils.UploadToS3(bucket, file, fileHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}
