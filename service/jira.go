package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type JiraInput struct {
	Assignee    string   `json:"assignee"`
	Components  []string `json:"components"`
	IssueType   string   `json:"issuetype"`
	Description string   `json:"description"`
	Labels      []string `json:"labels"`
	Project     string   `json:"project"`
	Summary     string   `json:"summary"`
}

type JiraBody struct {
	Fields JiraFields `json:"fields"`
}

type JiraFields struct {
	Assignee    Assignee    `json:"assignee"`
	Components  []Component `json:"components"`
	IssueType   IssueType   `json:"issuetype"`
	Description Description `json:"description"`
	Labels      []string    `json:"labels"`
	Project     Project     `json:"project"`
	Summary     string      `json:"summary"`
}

type Assignee struct {
	ID string `json:"id"`
}

type Component struct {
	ID string `json:"id"`
}

type IssueType struct {
	ID string `json:"id"`
}

type Project struct {
	ID string `json:"id"`
}

type Description struct {
	Type    string     `json:"type"`
	Version int        `json:"version"`
	Content []DocBlock `json:"content"`
}

type DocBlock struct {
	Type    string      `json:"type"`
	Content []TextBlock `json:"content"`
}

type TextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}
}

func CreateJiraIssue(jiraBody JiraBody) (string, error) {
	// Get environment variables
	jiraAPIURL := os.Getenv("JIRA_API_URL")
	jiraUsername := os.Getenv("JIRA_USER_EMAIL")
	jiraAPIToken := os.Getenv("JIRA_API_TOKEN")

	if jiraAPIURL == "" || jiraUsername == "" || jiraAPIToken == "" {
		return "", fmt.Errorf("missing required Jira environment variables")
	}

	// Encode username:token to Base64 for Basic Auth
	authString := jiraUsername + ":" + jiraAPIToken
	authEncoded := base64.StdEncoding.EncodeToString([]byte(authString))
	authHeader := "Basic " + authEncoded

	// Convert JiraBody struct to JSON
	jsonData, err := json.Marshal(jiraBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Jira body: %v", err)
	}

	// Create HTTP POST request
	req, err := http.NewRequest("POST", jiraAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	// Send the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer res.Body.Close()

	// Read response
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if res.StatusCode >= 400 {
		return "", fmt.Errorf("jira API error (%d): %s", res.StatusCode, string(body))
	}

	return string(body), nil
}

func StructureJiraBody(input JiraInput) JiraBody {
	// Map components
	var components []Component
	for _, c := range input.Components {
		components = append(components, Component{ID: c})
	}

	// Structure description
	description := Description{
		Type:    "doc",
		Version: 1,
		Content: []DocBlock{
			{
				Type: "paragraph",
				Content: []TextBlock{
					{
						Type: "text",
						Text: input.Description,
					},
				},
			},
		},
	}

	// Final structured body (using dynamic input values)
	return JiraBody{
		Fields: JiraFields{
			Assignee:    Assignee{ID: input.Assignee},
			Components:  components,
			IssueType:   IssueType{ID: input.IssueType},
			Description: description,
			Labels:      input.Labels,
			Project:     Project{ID: input.Project},
			Summary:     input.Summary,
		},
	}
}
