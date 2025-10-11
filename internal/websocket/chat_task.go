package ws

import (
	"ismismcube-backend/internal/config"
	"ismismcube-backend/internal/server"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func HandleChatTask(w http.ResponseWriter, r *http.Request) {
	websocketID := r.URL.Query().Get("id")
	if websocketID == "" {
		http.Error(w, "Missing websocket ID", http.StatusBadRequest)
		return
	}
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
	taskManager := server.GetTaskManager()
	taskManager.RegisterTaskConnection(websocketID, conn)
	conn.SetReadDeadline(time.Now().Add(config.WSPongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(config.WSPongWait))
		return nil
	})
	ticker := time.NewTicker(config.WSPingInterval)
	go func() {
		defer func() {
			ticker.Stop()
			conn.Close()
			taskManager.UnregisterTaskConnection(websocketID)
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(config.WSWriteWait))
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
