package config

import (
	"time"
)

var (
	WSPingInterval time.Duration
	WSPongWait     time.Duration
	WSWriteWait    time.Duration
)

func InitWSConfig() {
	WSPingInterval = time.Duration(getEnvInt("WS_PING_INTERVAL_SEC", 60)) * time.Second
	WSPongWait = time.Duration(getEnvInt("WS_PONG_WAIT_SEC", 90)) * time.Second
	WSWriteWait = time.Duration(getEnvInt("WS_WRITE_WAIT_SEC", 10)) * time.Second
}
