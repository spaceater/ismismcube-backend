package config

import (
	"time"
)

type WSConfig struct {
	PingIntervalSlowSec int `json:"ping_interval_slow_sec"`
	PongWaitSlowSec     int `json:"pong_wait_slow_sec"`
	PingIntervalFastSec int `json:"ping_interval_fast_sec"`
	PongWaitFastSec     int `json:"pong_wait_fast_sec"`
	WriteWaitSec        int `json:"write_wait_sec"`
}

var (
	WSPingIntervalSlow time.Duration
	WSPongWaitSlow     time.Duration
	WSPingIntervalFast time.Duration
	WSPongWaitFast     time.Duration
	WSWriteWait        time.Duration
)

func InitWSConfig(configData map[string]interface{}) {
	WSPingIntervalSlow = time.Duration(getConfigInt(getJSONTag(WSConfig{}, "PingIntervalSlowSec"), configData, 60)) * time.Second
	WSPongWaitSlow = time.Duration(getConfigInt(getJSONTag(WSConfig{}, "PongWaitSlowSec"), configData, 90)) * time.Second
	WSPingIntervalFast = time.Duration(getConfigInt(getJSONTag(WSConfig{}, "PingIntervalFastSec"), configData, 10)) * time.Second
	WSPongWaitFast = time.Duration(getConfigInt(getJSONTag(WSConfig{}, "PongWaitFastSec"), configData, 15)) * time.Second
	WSWriteWait = time.Duration(getConfigInt(getJSONTag(WSConfig{}, "WriteWaitSec"), configData, 10)) * time.Second
}
