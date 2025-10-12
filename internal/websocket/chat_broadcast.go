package ws

import (
	"encoding/json"
	"fmt"
	"ismismcube-backend/internal/config"
	"ismismcube-backend/internal/server"
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
	data := fmt.Appendf(nil, `broadcast:{"waiting_count":%d,"executing_count":%d,"broadcast_flag":%d}`, waiting, executing, broadcastFlag)
	for _, conn := range clients {
		sendQueueStats(conn, data)
	}
}

func RegisterChatClient(conn *websocket.Conn, waiting, executing int, broadcastFlag int64) {
	chatClientsMux.Lock()
	defer chatClientsMux.Unlock()
	chatClients[conn] = &ClientInfo{}
	go sendQueueStats(conn, fmt.Appendf(nil, `broadcast:{"waiting_count":%d,"executing_count":%d,"broadcast_flag":%d}`, waiting, executing, broadcastFlag))
	llmConfigData, err := json.Marshal(map[string]interface{}{
		"max_concurrent_tasks": config.LLMConfigure.MaxConcurrentTasks,
		"available_models":     config.LLMConfigure.AvailableModels,
	})
	if err != nil {
		return
	}
	go sendQueueStats(conn, fmt.Appendf(nil, "server-config:%s", string(llmConfigData)))
	chatParamsData, err := json.Marshal(config.ChatParameters)
	if err != nil {
		return
	}
	go sendQueueStats(conn, fmt.Appendf(nil, "chat-config:%s", string(chatParamsData)))
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
