package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	openai "github.com/sashabaranov/go-openai"
)

func main() {

	// dotenv
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	//openai
	config := openai.DefaultAzureConfig(os.Getenv("OPENAI_API_KEY"), os.Getenv("OPENAI_API_BASE"))
	if version := os.Getenv("OPENAI_API_VERSION"); version != "" {
		config.APIVersion = version
	}
	OpenaAIClient = openai.NewClientWithConfig(config)

	// gin
	router := gin.Default()
	router.GET("/oauth/callback", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": "true"})
	})
	router.POST("/handle-turn", headersMiddleware(), func(c *gin.Context) {

		// Bind the JSON body to the RequestBody struct
		var requestBody RequestBody
		if err := c.BindJSON(&requestBody); err != nil {
			c.Writer.WriteString(stringToChunk(err.Error()))
			c.Writer.Flush()
			c.Abort()
		}
		fmt.Println(requestBody)

		token := c.GetHeader("X-Github-Token")
		if token == "" {
			c.Writer.WriteString(stringToChunk("Empty github token"))
			c.Writer.Flush()
			c.Abort()
		}

		currentMsg := requestBody.Messages[len(requestBody.Messages)-1]

		// detect user intension from the message content
		it, err := detectIntension(c, currentMsg.Content)
		if err != nil {
			c.Writer.WriteString(stringToChunk(err.Error()))
			c.Writer.Flush()
			c.Abort()
		}
		fmt.Println(*it)

		owner, name := currentMsg.CopilotReferences[0].Data.OwnerLogin, currentMsg.CopilotReferences[0].Data.Name
		switch *it {
		case Intensions[0]:
			pulls := listPullRequests(c, token, owner, name)
			content, _ := json.Marshal(pulls)
			resp, err := summarizePullRequestList(c, string(content))
			fmt.Println(resp)
			if err != nil {
				fmt.Println(err)
				c.Writer.WriteString(stringToChunk(err.Error()))
				c.Writer.Flush()
				c.Abort()
			}
			c.Writer.WriteString(stringToChunk(*resp))
			c.Writer.Flush()
			break
		case Intensions[1]:
			idStr, err := extractPullRequestID(c, currentMsg.Content)
			fmt.Println(*idStr)
			if err != nil {
				c.Writer.WriteString(stringToChunk(err.Error()))
				c.Writer.Flush()
				c.Abort()
			}
			id, err := strconv.Atoi(*idStr)
			if err != nil {
				c.Writer.WriteString(stringToChunk(err.Error()))
				c.Writer.Flush()
				c.Abort()
			}
			pull, err := getPullRequest(c, token, owner, name, id)
			if err != nil {
				c.Writer.WriteString(stringToChunk(err.Error()))
				c.Writer.Flush()
				c.Abort()
			}
			resp, err := summarizePullRequest(c, *pull)
			if err != nil {
				c.Writer.WriteString(stringToChunk(err.Error()))
				c.Writer.Flush()
				c.Abort()
			}
			c.Writer.WriteString(stringToChunk(*resp))
			c.Writer.Flush()
			break
		case Intensions[2]:
			idStr, err := extractPullRequestID(c, currentMsg.Content)
			fmt.Println(*idStr)
			if err != nil {
				c.Writer.WriteString(stringToChunk(err.Error()))
				c.Writer.Flush()
				c.Abort()
			}
			id, err := strconv.Atoi(*idStr)
			if err != nil {
				c.Writer.WriteString(stringToChunk(err.Error()))
				c.Writer.Flush()
				c.Abort()
			}
			err = approvePullRequest(c, token, owner, name, id)
			if err != nil {
				c.Writer.WriteString(stringToChunk(err.Error()))
				c.Writer.Flush()
				c.Abort()
			}
			c.Writer.WriteString(stringToChunk(fmt.Sprintf("Pull Request %d approved", id)))
			c.Writer.Flush()
			break
		default:
			fmt.Println()
		}

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
