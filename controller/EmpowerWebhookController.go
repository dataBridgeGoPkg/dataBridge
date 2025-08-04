package controller

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
)

// Define only the structures you care about

type UserInfo struct {
	Email  string `json:"email"`
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	TeamID int64  `json:"team_id"`
}

type CallDetails struct {
	CallUUID   string `json:"call_uuid"`
	RingoverID string `json:"ringover_id"`
}

type Speech struct {
	Text string `json:"text"`
}

type Transcription struct {
	Speeches []Speech `json:"speeches"`
	Text     string   `json:"text"`
}

type Payload struct {
	UserInfo      UserInfo      `json:"user_info"`
	CallDetails   CallDetails   `json:"call_details"`
	Transcription Transcription `json:"transcription"`
}

type FinalTranscript struct {
	CallUUID   string   `json:"call_uuid"`
	RingoverID string   `json:"ringover_id"`
	UserEmail  string   `json:"user_email"`
	UserID     int64    `json:"user_id"`
	UserName   string   `json:"user_name"`
	Speeches   []Speech `json:"speeches"`
	FullText   string   `json:"full_text"`
}

var filteredTranscripts []FinalTranscript

func HandleWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read request body"})
		return
	}

	// fmt.Println(string(body))

	var payload Payload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	//Pretty print just the transcription part
	// jsonBytes, _ := json.MarshalIndent(payload.Transcription, "", "  ")
	// fmt.Println(string(jsonBytes))

	// response := map[string]any{
	// 	"User_Info":     payload.UserInfo,
	// 	"Call_Details":  payload.CallDetails,
	// 	"Transcription": payload.Transcription,
	// }

	// jsonBytes, err := json.MarshalIndent(response, "", "  ")
	// if err != nil {
	// 	fmt.Println("Error marshaling response:", err)
	// } else {
	// 	fmt.Println(string(jsonBytes))
	// }

	// Check User ID
	if payload.UserInfo.ID == 56381 {
		final := FinalTranscript{
			CallUUID:   payload.CallDetails.CallUUID,
			RingoverID: payload.CallDetails.RingoverID,
			UserEmail:  payload.UserInfo.Email,
			UserID:     payload.UserInfo.ID,
			UserName:   payload.UserInfo.Name,
			Speeches:   payload.Transcription.Speeches,
			FullText:   payload.Transcription.Text,
		}
		filteredTranscripts = append(filteredTranscripts, final)
		fmt.Println("✅ Stored transcript for user ID 61838")
	}

	fmt.Println(filteredTranscripts)

}
