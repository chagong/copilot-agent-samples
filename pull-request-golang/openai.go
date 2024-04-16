package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	openai "github.com/sashabaranov/go-openai"
)

var OpenaAIClient *openai.Client

func invokeOpenAI(c *gin.Context, prompt string) (*string, error) {
	resp, err := OpenaAIClient.CreateChatCompletion(
		c,
		openai.ChatCompletionRequest{
			Model: os.Getenv("OPENAI_MODEL_NAME"),
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return nil, err
	}
	return &resp.Choices[0].Message.Content, nil
}

func extractPullRequestID(c *gin.Context, msg string) (*string, error) {
	prompt := fmt.Sprintf("Try your best to extract the pull request ID from user's message: %s. The only thing you should return is the ID. Nothing else.", msg)
	return invokeOpenAI(c, prompt)
}

func detectIntension(c *gin.Context, msg string) (*string, error) {
	prompt := fmt.Sprintf("Try your best to detect the user's intension from user's message and choos from one of these options: %s. Here is the user's message: %s. The only thing you should return is the choice.\n", strings.Join(Intensions, ","), msg)

	return invokeOpenAI(c, prompt)
}

func summarizePullRequest(c *gin.Context, content string) (*string, error) {
	prompt := fmt.Sprintf("Help me summarize this pull request. %s", content)
	return invokeOpenAI(c, prompt)
}

func summarizePullRequestList(c *gin.Context, content string) (*string, error) {
	prompt := fmt.Sprintf("Help me summarize these pull requests and keep their links: %s", content)
	return invokeOpenAI(c, prompt)
}
