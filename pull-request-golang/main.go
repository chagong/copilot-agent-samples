package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v61/github"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

type Delta struct {
	Content string `json:"content"`
}

type Choice struct {
	Index int   `json:"index"`
	Delta Delta `json:"delta"`
}

type Chunk struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

func stringToChunk(s string) string {
	c := Chunk{
		ID:      "chunk",
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   os.Getenv("OPENAI_MODEL_NAME"),
		Choices: []Choice{
			{
				Index: 0,
				Delta: Delta{
					Content: s,
				},
			},
		},
	}

	d, _ := json.Marshal(c)
	return fmt.Sprintf("data: %s\n\n", d)
}

func main() {

	// dotenv
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// gin
	router := gin.Default()
	router.GET("/oauth/callback", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": "true"})
	})
	router.POST("/handle-turn", headersMiddleware(), func(c *gin.Context) {
		// Bind the JSON body to the RequestBody struct
		var requestBody RequestBody
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
		}

		token := c.GetHeader("X-Github-Token")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Empty github token"})
			c.Abort()
		}

		currentMsg := requestBody.Messages[len(requestBody.Messages)-1]
		project := currentMsg.Content[strings.Index(currentMsg.Content, " ")+1:]
		params := strings.Split(project, "/")
		if len(params) != 2 {
			c.Writer.WriteString(stringToChunk("Please tell me the repository, like gin-gonic/gin"))
			c.Writer.Flush()
			return
		}

		pulls := listPullRequests(c, token, params[0], params[1])
		if len(pulls) == 0 {
			c.Writer.WriteString(stringToChunk("Cannot find any pull requests"))
			c.Writer.Flush()
			return
		}

		resp, err := summarizePullRequests(c, pulls)

		if err != nil {
			c.Writer.WriteString(stringToChunk("Invoke OpenAPI failed"))
			c.Writer.Flush()
			return
		}
		c.Writer.WriteString(stringToChunk(*resp))
		c.Writer.Flush()
	})

	if err := router.Run(":8085"); err != nil {
		log.Panicf("error: %s", err)
	}
}

func headersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}

type RequestBody struct {
	CopilotThreadID  string      `json:"copilot_thread_id"`
	Messages         []Message   `json:"messages"`
	Stop             interface{} `json:"stop"`
	TopP             float64     `json:"top_p"`
	Temperature      float64     `json:"temperature"`
	MaxTokens        int         `json:"max_tokens"`
	PresencePenalty  float64     `json:"presence_penalty"`
	FrequencyPenalty float64     `json:"frequency_penalty"`
	CopilotSkills    interface{} `json:"copilot_skills"`
	Agent            string      `json:"agent"`
}

type Message struct {
	Role                 string        `json:"role"`
	Content              string        `json:"content"`
	CopilotReferences    []interface{} `json:"copilot_references"`
	CopilotConfirmations interface{}   `json:"copilot_confirmations"`
}

func listPullRequests(c *gin.Context, token string, owner string, repo string) []*github.PullRequest {
	client := github.NewClient(nil).WithAuthToken(token)

	// List pull requests where you are a requested reviewer
	opt := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 5},
		// You can adjust this to include closed PRs if needed
		State: "open",
	}
	pulls, _, err := client.PullRequests.List(c, owner, repo, opt)
	if err != nil {
		fmt.Printf("Error listing pull requests: %v\n", err)
		return nil
	}
	return pulls
}

func summarizePullRequests(c *gin.Context, pulls []*github.PullRequest) (*string, error) {

	config := openai.DefaultAzureConfig(os.Getenv("OPENAI_API_KEY"), os.Getenv("OPENAI_API_BASE"))
	if version := os.Getenv("OPENAI_API_VERSION"); version != "" {
		config.APIVersion = version
	}
	client := openai.NewClientWithConfig(config)

	content, _ := json.Marshal(pulls)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: os.Getenv("OPENAI_MODEL_NAME"),
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("Help me summarize these pull requests and keep their links: %s", content),
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
