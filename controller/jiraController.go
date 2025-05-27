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
	c.JSON(200, gin.H{"message": "Jira issue created successfully", "response": response})

}
