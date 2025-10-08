package config

type LLMConfig struct {
	ApiUrl             string   `json:"api_url"`
	ApiKey             string   `json:"api_key"`
	MaxConcurrentTasks int      `json:"max_concurrent_tasks"`
	Timeout            int      `json:"timeout"`
	AvailableModels    []string `json:"available_models"`
}

type ChatParams struct {
  Prompt           string  `json:"prompt"`
  ContentSize      int     `json:"content_size"`
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float64 `json:"temperature"`
	TopP             float64 `json:"top_p"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty  float64 `json:"presence_penalty"`
	RepeatPenalty    float64 `json:"repeat_penalty"`
}

var (
	LLMConfigure   LLMConfig
	ChatParameters ChatParams
)

func InitLLMConfig() {
	LLMConfigure = LLMConfig{
		ApiUrl:             getEnv("LLM_API_URL", "http://ai.api.maybered.com/AI-VMZ-8B/v1/chat/completions"),
		ApiKey:             getEnv("LLM_API_KEY", ""),
		MaxConcurrentTasks: getEnvInt("LLM_MAX_CONCURRENT_TASKS", 6),
		Timeout:            getEnvInt("LLM_TIMEOUT", 20),
		AvailableModels:    []string{"AI-VMZ-8B"},
	}
	ChatParameters = ChatParams{
		Prompt:           getEnv("LLM_PROMPT", "你是一个AI助手，你的知识广泛，尤其精通哲学、历史、文学等人文学科，你需要以严谨而逻辑清晰的方式响应用户提问。"),
		ContentSize:      getEnvInt("LLM_CONTENT_SIZE", 8192),
		MaxTokens:        getEnvInt("LLM_MAX_TOKENS", 2048),
		Temperature:      getEnvFloat("LLM_TEMPERATURE", 0.3),
		TopP:             getEnvFloat("LLM_TOP_P", 0.7),
		FrequencyPenalty: getEnvFloat("LLM_FREQUENCY_PENALTY", 0.0),
		PresencePenalty:  getEnvFloat("LLM_PRESENCE_PENALTY", 0.0),
		RepeatPenalty:    getEnvFloat("LLM_REPEAT_PENALTY", 1.05),
	}
}
