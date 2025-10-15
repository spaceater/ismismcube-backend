package toolkit

import (
	"encoding/json"
	"fmt"
)

type MessageData struct {
	Type string      `json:"-"`
	Data interface{} `json:"-"`
}

func (md *MessageData) ToBytes() ([]byte, error) {
	jsonData, err := json.Marshal(md.Data)
	if err != nil {
		return nil, err
	}
	return fmt.Appendf(nil, "%s:%s", md.Type, string(jsonData)), nil
}

type ErrorData struct {
	Error string `json:"error"`
}
