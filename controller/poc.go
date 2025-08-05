package controller

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"example.com/ringover_kb/prompt" // Adjust the import path as necessary
	"example.com/ringover_kb/service"
	"github.com/gin-gonic/gin"
	// Adjust the import path as necessary
)

type transcriptDump struct {
	RingoTranscriptions []struct {
		FullConversation string `json:"full_conversation"`
		UserID           string `json:"user_id"`
	} `json:"Ringo_transcriptions"`
}

type structuredData struct {
	Question        string   `json:"question"`
	Answer          string   `json:"answer"`
	Category        string   `json:"category"`
	Tags            []string `json:"tags"`
	ConfidenceScore float64  `json:"confidence_score"`
}

func GetCleanDataFromTranscriptDump(c *gin.Context) {
	// Step 1 - Load JSON file
	rawDataTranscript, err := ioutil.ReadFile("/Users/atmadeep.das/Desktop/Go_Product_RoadMap/transcript.json")
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	var rawDataDump transcriptDump
	if err := json.Unmarshal(rawDataTranscript, &rawDataDump); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Step 2 - Collect prompts
	var knowledgeBaseArray []string

	for _, transcriptEntry := range rawDataDump.RingoTranscriptions {
		transcript := transcriptEntry.FullConversation
		if transcript == "" {
			log.Println("Empty transcript, skipping...")
			continue
		}

		promptText := prompt.OpenAIPrompt(transcript)

		//Step 3 - Convert to structured data with OpenAI API
		var structuredEntry structuredData

		knowledgeBase := service.GenerateOpenAIResponse(promptText)

		if err := json.Unmarshal([]byte(knowledgeBase), &structuredEntry); err != nil {
			log.Printf("Failed to unmarshal OpenAI response: %v", err)
			continue
		}

		knowledgeBaseArray = append(knowledgeBaseArray, knowledgeBase)
	}

	c.JSON(200, knowledgeBaseArray)
}
