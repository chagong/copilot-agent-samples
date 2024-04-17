package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

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
