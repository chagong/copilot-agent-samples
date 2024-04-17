package main

// github copilot request struct
type RepositoryData struct {
	Type        string         `json:"type"`
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	OwnerLogin  string         `json:"ownerLogin"`
	OwnerType   string         `json:"ownerType"`
	ReadmePath  string         `json:"readmePath"`
	Description string         `json:"description"`
	CommitOID   string         `json:"commitOID"`
	Ref         string         `json:"ref"`
	RefInfo     RefInfo        `json:"refInfo"`
	Visibility  string         `json:"visibility"`
	Languages   []LanguageData `json:"languages"`
}

type RefInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type LanguageData struct {
	Name    string  `json:"name"`
	Percent float64 `json:"percent"`
}

type CopilotReference struct {
	Type     string         `json:"type"`
	Data     RepositoryData `json:"data"`
	ID       string         `json:"id"`
	Metadata Metadata       `json:"metadata"`
}

type Metadata struct {
	DisplayName string `json:"display_name"`
	DisplayIcon string `json:"display_icon"`
}

type Message struct {
	Role                 string             `json:"role"`
	Content              string             `json:"content"`
	CopilotReferences    []CopilotReference `json:"copilot_references"`
	CopilotConfirmations interface{}        `json:"copilot_confirmations"`
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

// pre-defined intensions
var Intensions = []string{"ListPullRequests", "SummarizePullRequest", "ApprovePullRequest"}

// chunk defination for github copilot streaming
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
