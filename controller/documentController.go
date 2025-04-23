package controller

import (
	"encoding/json"
	"io"
	"net/http"

	"example.com/Product_RoadMap/models"
	"example.com/Product_RoadMap/utils"
	"github.com/gin-gonic/gin"
)

type documentResponse struct {
	ID          int64   `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	URL         string  `json:"url"`
	CreatedAt   int64   `json:"created_at,omitempty"`
	UpdatedAt   int64   `json:"updated_at,omitempty"`
}

func CreateDocument(c *gin.Context) {
	// Parse the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	// Unmarshal the JSON data into a Document struct
	var doc documentResponse
	err = json.Unmarshal(body, &doc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Map documentResponse to models.Document
	document := models.Document{
		Name:        doc.Name,
		Description: doc.Description,
		URL:         doc.URL,
		CreatedAt:   doc.CreatedAt,
		UpdatedAt:   doc.UpdatedAt,
	}

	// Create the document in your database or storage system
	createDoc := models.CreateDocument(models.DB, &document)
	if createDoc != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create document"})
		return
	}
	// Return the created document as a response
	c.JSON(http.StatusCreated, gin.H{
		"message":  "Document created successfully",
		"document": document,
	})
}

func GetDocumentByID(c *gin.Context) {
	iD := c.Param("id")
	documentID := utils.ParseID(iD)
	if iD == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	document, err := models.GetDocumentByID(models.DB, documentID)
	if document == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return

	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve document"})
		return
	}

	if document == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"document": document,
	})
}

func UpdateDocumentById(c *gin.Context) {
	id := c.Param("id")
	documentID := utils.ParseID(id)
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	//Check if the document exists
	document, err := models.GetDocumentByID(models.DB, documentID)
	if document == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	// Parse the request body
	var updatedDocument models.Document
	err = c.BindJSON(&updatedDocument)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	// Update the document in your database or storage system
	err = models.UpdateDocument(models.DB, &updatedDocument)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update document"})
		return
	}
	// Return the updated document as a response
	c.JSON(http.StatusOK, gin.H{
		"message":  "Document updated successfully",
		"document": updatedDocument,
	})

}

func DeleteDocumentById(c *gin.Context) {
	documentID := c.Param("id")
	if len(documentID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}
	// Check if the document exists in your database or storage system
	document, err := models.GetDocumentByID(models.DB, utils.ParseID(documentID))
	if document == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve document"})
		return
	}
	// Delete the document from your database or storage system
	err = models.DeleteDocument(models.DB, utils.ParseID(documentID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}

func GetAllDocuments(c *gin.Context) {
	documents, err := models.GetAllDocuments(models.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve documents"})
		return
	}

	if documents == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No documents found"})
		return
	}

	// Return the list of documents as a response
	c.JSON(http.StatusOK, gin.H{
		"documents": documents,
	})
}
