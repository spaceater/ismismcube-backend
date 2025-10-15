package config

type ServerConfig struct {
	Port             string `json:"port"`
	PageViewFile     string `json:"page_view_file"`
	ExecutedTaskFile string `json:"executed_task_file"`
}

var (
	Port             string
	PageViewFile     string
	ExecutedTaskFile string
)

func InitServerConfig(configData map[string]interface{}) {
	Port = getConfigString(getJSONTag(ServerConfig{}, "Port"), configData, "80")
	PageViewFile = getConfigString(getJSONTag(ServerConfig{}, "PageViewFile"), configData, "./resources/page-view.txt")
	ExecutedTaskFile = getConfigString(getJSONTag(ServerConfig{}, "ExecutedTaskFile"), configData, "./resources/executed-task.txt")
}
