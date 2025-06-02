package controller

import (
	"encoding/json"
	"io"

	"example.com/Product_RoadMap/service"
	"github.com/gin-gonic/gin"
)

type JiraBody struct {
	Assignee    string   `json:"assignee"`
	Components  []string `json:"components"`
	IssueType   string   `json:"issuetype"`
	Description string   `json:"description"`
	Labels      []string `json:"labels"`
	Project     string   `json:"project"`
	Summary     string   `json:"summary"`
}

type ResponseBody struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Self string `json:"self"`
}

func CreateJiraIssue(c *gin.Context) {

	//Readd the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to read request body", "details": err.Error()})
		return
	}

	//unmarshal the request body into the JiraIssueBody struct
	var jiraIssue JiraBody

	err = json.Unmarshal(body, &jiraIssue)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	formatJira := service.StructureJiraBody(service.JiraInput{
		Assignee:    jiraIssue.Assignee,
		Components:  jiraIssue.Components,
		IssueType:   jiraIssue.IssueType,
		Description: jiraIssue.Description,
		Labels:      jiraIssue.Labels,
		Project:     jiraIssue.Project,
		Summary:     jiraIssue.Summary,
	})

	//Create Jira Issue
	response, err := service.CreateJiraIssue(formatJira)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create Jira issue", "details": err.Error()})
		return
	}

	//Unmarshal response to check for errors
	var createResponse ResponseBody
	err = json.Unmarshal([]byte(response), &createResponse)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse Jira response", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Jira issue created successfully", "response": createResponse})

}

func UpdateJiraIssue(c *gin.Context) {
	// Get the issue ID from the URL parameters
	issueID := c.Param("id")
	if issueID == "" {
		c.JSON(400, gin.H{"error": "Missing Jira issue ID in URL parameter"})
		return
	}

	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to read request body", "details": err.Error()})
		return
	}

	// Unmarshal the request body into JiraBody struct
	var jiraIssue JiraBody
	err = json.Unmarshal(body, &jiraIssue)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Format the Jira issue body into a map
	formatJira := service.UpdateStructureJiraBody(service.JiraInput{
		Assignee:    jiraIssue.Assignee,
		Components:  jiraIssue.Components,
		IssueType:   jiraIssue.IssueType,
		Description: jiraIssue.Description,
		Labels:      jiraIssue.Labels,
		Project:     jiraIssue.Project,
		Summary:     jiraIssue.Summary,
	})

	//Call service to update Jira Issue (map format supported!)
	response, err := service.UpdateJiraIssue(issueID, formatJira)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update Jira issue", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Jira issue updated successfully", "response": response})
}
