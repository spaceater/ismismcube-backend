package config

import (
	"time"
)

var (
	WSPingIntervalSlow time.Duration
	WSPongWaitSlow     time.Duration
	WSPingIntervalFast time.Duration
	WSPongWaitFast     time.Duration
	WSWriteWait        time.Duration
)

func InitWSConfig() {
	WSPingIntervalSlow = time.Duration(getEnvInt("WS_PING_INTERVAL_SLOW_SEC", 60)) * time.Second
	WSPongWaitSlow = time.Duration(getEnvInt("WS_PONG_WAIT_SLOW_SEC", 90)) * time.Second
	WSPingIntervalFast = time.Duration(getEnvInt("WS_PING_INTERVAL_FAST_SEC", 10)) * time.Second
	WSPongWaitFast = time.Duration(getEnvInt("WS_PONG_WAIT_FAST_SEC", 15)) * time.Second
	WSWriteWait = time.Duration(getEnvInt("WS_WRITE_WAIT_SEC", 10)) * time.Second
}
