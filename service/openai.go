package service

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func init() {
	// Load the .env file once during initialization
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func GenerateOpenAIResponse(prompt string) string {
	// Retrieve the API key from the environment
	openAi_API := os.Getenv("OPEN_AI_API")
	if openAi_API == "" {
		log.Fatalf("OPENAI_API_KEY is not set in the environment")
	}

	fmt.Println("Using OpenAI API Key:", openAi_API)

	// Create a new OpenAI client
	client := openai.NewClient(
		option.WithAPIKey(openAi_API),
	)

	// Generate a chat completion
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModelGPT4o,
	})
	if err != nil {
		log.Fatalf("Error generating OpenAI response: %v", err)
	}

	// Extract and return the response content
	chatResponse := chatCompletion.Choices[0].Message.Content
	fmt.Println("Chat Response:", chatResponse)

	return chatResponse
}
