package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

// GetJiraIssue is your desired target structure
type GetJiraIssue struct {
	Assignee    string   `json:"assignee"`
	Components  []string `json:"components"`
	IssueType   string   `json:"issuetype"`
	Description string   `json:"description"`
	Labels      []string `json:"labels"`
	Project     string   `json:"project"`
	Summary     string   `json:"summary"`
}

// --- Intermediate structs to capture the full Jira API response ---
// These structs are designed to parse the parts of the Jira API response
// that are needed to populate the GetJiraIssue struct.

type ApiJiraResponse struct { // Top-level structure of the Jira API response
	Fields ApiJiraFields `json:"fields"`
	// Other top-level fields like "id", "key", "self" are ignored for this transformation
}

type ApiJiraFields struct {
	Assignee    *ApiJiraUser        `json:"assignee"` // Pointer to handle null
	Components  []ApiJiraComponent  `json:"components"`
	IssueType   *ApiJiraIssueType   `json:"issuetype"`   // Pointer
	Description *ApiJiraDescription `json:"description"` // Pointer for complex object
	Labels      []string            `json:"labels"`
	Project     *ApiJiraProject     `json:"project"` // Pointer
	Summary     string              `json:"summary"`
	// Many other customfield_xxxxx and other fields from Jira API are ignored
}

type ApiJiraUser struct {
	AccountID string `json:"accountId"`
	// Other fields (self, emailAddress, etc.) are ignored
}

type ApiJiraComponent struct {
	ID string `json:"id"`
	// Other fields (self, name, etc.) are ignored
}

type ApiJiraIssueType struct {
	ID string `json:"id"`
	// Other fields (self, name, etc.) are ignored
}

// ApiJiraDescription represents the structure of the 'description' field (Atlassian Document Format)
type ApiJiraDescription struct {
	Content []ApiAdfContentNode `json:"content"`
}

type ApiAdfContentNode struct { // ADF = Atlassian Document Format
	Type    string           `json:"type"`
	Content []ApiAdfTextNode `json:"content,omitempty"`
}

type ApiAdfTextNode struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ApiJiraProject struct {
	ID string `json:"id"`
	// Other fields (self, key, name, etc.) are ignored
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

func UpdateJiraIssue(issueID string, jiraBody map[string]interface{}) (string, error) {
	// Get environment variables
	jiraBaseURL := os.Getenv("JIRA_API_URL")
	jiraUsername := os.Getenv("JIRA_USER_EMAIL")
	jiraAPIToken := os.Getenv("JIRA_API_TOKEN")

	if jiraBaseURL == "" || jiraUsername == "" || jiraAPIToken == "" {
		return "", fmt.Errorf("missing required Jira environment variables")
	}

	// Build the full URL with issue ID
	updateURL := fmt.Sprintf("%s/%s", jiraBaseURL, issueID)

	// Encode username:token to Base64 for Basic Auth
	authString := jiraUsername + ":" + jiraAPIToken
	authEncoded := base64.StdEncoding.EncodeToString([]byte(authString))
	authHeader := "Basic " + authEncoded

	// Convert map to JSON
	jsonData, err := json.Marshal(jiraBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Jira body: %v", err)
	}

	// Create HTTP PUT request
	req, err := http.NewRequest("PUT", updateURL, bytes.NewBuffer(jsonData))
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

func GetJiraIssueWithID(id string) (GetJiraIssue, error) {
	// Get environment variables
	jiraAPIURL := os.Getenv("JIRA_API_URL")
	jiraUsername := os.Getenv("JIRA_USER_EMAIL")
	jiraAPIToken := os.Getenv("JIRA_API_TOKEN")

	if jiraAPIURL == "" || jiraUsername == "" || jiraAPIToken == "" {
		return GetJiraIssue{}, fmt.Errorf("missing required Jira environment variables (JIRA_API_URL, JIRA_USER_EMAIL, JIRA_API_TOKEN)")
	}

	// Build the full URL with issue ID
	getURL := fmt.Sprintf("%s/%s", jiraAPIURL, id)

	fmt.Println("Fetching Jira issue from URL:", getURL)

	// Encode username:token to Base64 for Basic Auth
	authString := jiraUsername + ":" + jiraAPIToken
	authEncoded := base64.StdEncoding.EncodeToString([]byte(authString))
	authHeader := "Basic " + authEncoded

	// Create HTTP GET request
	req, err := http.NewRequest("GET", getURL, nil)
	if err != nil {
		return GetJiraIssue{}, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	// req.Header.Set("Content-Type", "application/json") // Not strictly necessary for GET but good practice
	req.Header.Set("Authorization", authHeader)

	// Send the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return GetJiraIssue{}, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer res.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(res.Body) // For Go 1.16+ consider io.ReadAll(res.Body)
	if err != nil {
		return GetJiraIssue{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode >= 400 {
		return GetJiraIssue{}, fmt.Errorf("jira API error (%d): %s", res.StatusCode, string(body))
	}

	// Unmarshal the full response body into our intermediate ApiJiraResponse struct
	var apiResponse ApiJiraResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return GetJiraIssue{}, fmt.Errorf("failed to unmarshal full Jira API response: %w. Body: %s", err, string(body))
	}

	// Now, transform data from apiResponse to the target GetJiraIssue struct
	var targetIssue GetJiraIssue

	// Assignee
	if apiResponse.Fields.Assignee != nil {
		targetIssue.Assignee = apiResponse.Fields.Assignee.AccountID
	}

	// Components
	if apiResponse.Fields.Components != nil {
		targetIssue.Components = make([]string, 0, len(apiResponse.Fields.Components)) // Initialize with capacity
		for _, comp := range apiResponse.Fields.Components {
			targetIssue.Components = append(targetIssue.Components, comp.ID)
		}
	}

	// IssueType
	if apiResponse.Fields.IssueType != nil {
		targetIssue.IssueType = apiResponse.Fields.IssueType.ID
	}

	// Description (extracting the first text found in the document structure)
	if apiResponse.Fields.Description != nil && len(apiResponse.Fields.Description.Content) > 0 {
	descriptionLoop:
		for _, L1Content := range apiResponse.Fields.Description.Content { // Iterate through first level content nodes
			if L1Content.Type == "paragraph" && len(L1Content.Content) > 0 {
				for _, L2Content := range L1Content.Content { // Iterate through second level (text nodes within paragraph)
					if L2Content.Type == "text" && L2Content.Text != "" {
						targetIssue.Description = L2Content.Text
						break descriptionLoop // Found the first text, break out of all description loops
					}
				}
			}
			// You might want to handle other ADF node types here if necessary
		}
	}

	// Labels
	targetIssue.Labels = apiResponse.Fields.Labels // Direct copy

	// Project
	if apiResponse.Fields.Project != nil {
		targetIssue.Project = apiResponse.Fields.Project.ID
	}

	// Summary
	targetIssue.Summary = apiResponse.Fields.Summary // Direct copy

	return targetIssue, nil
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

func UpdateStructureJiraBody(input JiraInput) map[string]interface{} {
	// Map components
	var components []map[string]string
	for _, c := range input.Components {
		components = append(components, map[string]string{"id": c})
	}

	// Structure description
	description := map[string]interface{}{
		"type":    "doc",
		"version": 1,
		"content": []map[string]interface{}{
			{
				"type": "paragraph",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": input.Description,
					},
				},
			},
		},
	}

	// Final structured body (dynamic input)
	return map[string]interface{}{
		"fields": map[string]interface{}{
			"assignee": map[string]string{
				"accountId": input.Assignee,
			},
			"components": components,
			"issuetype": map[string]string{
				"id": input.IssueType,
			},
			"description": description,
			"labels":      input.Labels,
			"project": map[string]string{
				"id": input.Project,
			},
			"summary": input.Summary,
		},
	}
}
