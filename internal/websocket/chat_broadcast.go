package ws

import (
	"ismismcube-backend/internal/config"
	"ismismcube-backend/internal/server"
	"ismismcube-backend/internal/toolkit"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ClientInfo struct {
	WriteMutex sync.Mutex
}

type QueueStatsData struct {
	WaitingCount   int   `json:"waiting_count"`
	ExecutingCount int   `json:"executing_count"`
	BroadcastFlag  int64 `json:"broadcast_flag"`
}

type LLMConfigData struct {
	MaxConcurrentTasks int      `json:"max_concurrent_tasks"`
	AvailableModels    []string `json:"available_models"`
}

var (
	chatClients    = make(map[*websocket.Conn]*ClientInfo)
	chatClientsMux sync.RWMutex
)

type WebSocketBroadcaster struct{}

func (w *WebSocketBroadcaster) BroadcastQueueStats(waiting, executing int, broadcastFlag int64) {
	chatClientsMux.RLock()
	clients := make([]*websocket.Conn, 0, len(chatClients))
	for conn := range chatClients {
		clients = append(clients, conn)
	}
	chatClientsMux.RUnlock()
	data := &toolkit.MessageData{
		Type: "broadcast",
		Data: QueueStatsData{
			WaitingCount:   waiting,
			ExecutingCount: executing,
			BroadcastFlag:  broadcastFlag,
		},
	}
	msg, err := data.ToBytes()
	if err != nil {
		return
	}
	for _, conn := range clients {
		sendQueueStats(conn, msg)
	}
}

func RegisterChatClient(conn *websocket.Conn, waiting, executing int, broadcastFlag int64) {
	chatClientsMux.Lock()
	chatClients[conn] = &ClientInfo{}
	chatClientsMux.Unlock()
	// 发送队列统计信息
	statsData := &toolkit.MessageData{
		Type: "broadcast",
		Data: QueueStatsData{
			WaitingCount:   waiting,
			ExecutingCount: executing,
			BroadcastFlag:  broadcastFlag,
		},
	}
	statsMsg, err := statsData.ToBytes()
	if err != nil {
		return
	}
	go sendQueueStats(conn, statsMsg)
	// 发送LLM配置信息
	llmConfigData := &toolkit.MessageData{
		Type: "server-config",
		Data: LLMConfigData{
			MaxConcurrentTasks: config.LLMConfigure.MaxConcurrentTasks,
			AvailableModels:    config.LLMConfigure.AvailableModels,
		},
	}
	llmConfigMsg, err := llmConfigData.ToBytes()
	if err != nil {
		return
	}
	go sendQueueStats(conn, llmConfigMsg)
	// 发送聊天参数配置
	chatConfigData := &toolkit.MessageData{
		Type: "chat-config",
		Data: config.ChatParameters,
	}
	chatConfigMsg, err := chatConfigData.ToBytes()
	if err != nil {
		return
	}
	go sendQueueStats(conn, chatConfigMsg)
}

func sendQueueStats(conn *websocket.Conn, data []byte) {
	chatClientsMux.RLock()
	clientInfo, exists := chatClients[conn]
	chatClientsMux.RUnlock()
	if !exists {
		return
	}
	clientInfo.WriteMutex.Lock()
	conn.WriteMessage(websocket.TextMessage, data)
	clientInfo.WriteMutex.Unlock()
}

func UnregisterChatClient(conn *websocket.Conn) {
	chatClientsMux.Lock()
	defer chatClientsMux.Unlock()
	delete(chatClients, conn)
}

func HandleChatBroadcast(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	waiting, executing := server.GetTaskManager().GetQueueCount()
	broadcastFlag := server.GetTaskManager().GetBroadcastFlag()
	RegisterChatClient(conn, waiting, executing, broadcastFlag)
	conn.SetReadDeadline(time.Now().Add(config.WSPongWaitSlow))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(config.WSPongWaitSlow))
		return nil
	})
	ticker := time.NewTicker(config.WSPingIntervalSlow)
	go func() {
		var isNormalClose bool
		defer func() {
			ticker.Stop()
			if !isNormalClose {
				if tcpConn, ok := conn.UnderlyingConn().(*net.TCPConn); ok {
					tcpConn.SetLinger(0)
				}
			}
			conn.Close()
			UnregisterChatClient(conn)
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					isNormalClose = true
				}
				break
			}
		}
	}()
	go func() {
		defer ticker.Stop()
		for {
			<-ticker.C
			conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(config.WSWriteWait))
		}
	}()
}
