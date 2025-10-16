package config

type LLMConfig struct {
	BaseApiUrl         string   `json:"base_api_url"`
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

func InitLLMConfig(configData map[string]interface{}) {
	LLMConfigure.BaseApiUrl = getConfigString(getJSONTag(LLMConfig{}, "BaseApiUrl"), configData, "")
	LLMConfigure.ApiKey = getConfigString(getJSONTag(LLMConfig{}, "ApiKey"), configData, "")
	LLMConfigure.MaxConcurrentTasks = getConfigInt(getJSONTag(LLMConfig{}, "MaxConcurrentTasks"), configData, 6)
	LLMConfigure.Timeout = getConfigInt(getJSONTag(LLMConfig{}, "Timeout"), configData, 20)
	LLMConfigure.AvailableModels = getConfigStringSlice(getJSONTag(LLMConfig{}, "AvailableModels"), configData, []string{"AI-VMZ-8B"})

	ChatParameters.Prompt = getConfigString(getJSONTag(ChatParams{}, "Prompt"), configData, "你是一个AI助手，你的知识广泛，尤其精通哲学、历史、文学等人文学科，你需要以严谨而逻辑清晰的方式响应用户提问。")
	ChatParameters.ContentSize = getConfigInt(getJSONTag(ChatParams{}, "ContentSize"), configData, 8192)
	ChatParameters.MaxTokens = getConfigInt(getJSONTag(ChatParams{}, "MaxTokens"), configData, 2048)
	ChatParameters.Temperature = getConfigFloat(getJSONTag(ChatParams{}, "Temperature"), configData, 0.3)
	ChatParameters.TopP = getConfigFloat(getJSONTag(ChatParams{}, "TopP"), configData, 0.7)
	ChatParameters.FrequencyPenalty = getConfigFloat(getJSONTag(ChatParams{}, "FrequencyPenalty"), configData, 0.0)
	ChatParameters.PresencePenalty = getConfigFloat(getJSONTag(ChatParams{}, "PresencePenalty"), configData, 0.0)
	ChatParameters.RepeatPenalty = getConfigFloat(getJSONTag(ChatParams{}, "RepeatPenalty"), configData, 1.05)
}
