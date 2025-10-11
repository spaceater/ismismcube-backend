package ws

import (
	"fmt"
	"ismismcube-backend/internal/config"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ismismcubeClients    = make(map[*websocket.Conn]*ClientInfo)
	ismismcubeClientsMux sync.RWMutex
)

func RegisterIsmismcubeClient(conn *websocket.Conn) {
	ismismcubeClientsMux.Lock()
	defer ismismcubeClientsMux.Unlock()
	ismismcubeClients[conn] = &ClientInfo{}
	go broadcastOnlineCount()
}

func UnregisterIsmismcubeClient(conn *websocket.Conn) {
	ismismcubeClientsMux.Lock()
	defer ismismcubeClientsMux.Unlock()
	delete(ismismcubeClients, conn)
	go broadcastOnlineCount()
}

func broadcastOnlineCount() {
	ismismcubeClientsMux.RLock()
	data := fmt.Appendf(nil, `broadcast:{"online":%d}`, len(ismismcubeClients))
	clients := make([]*websocket.Conn, 0, len(ismismcubeClients))
	for conn := range ismismcubeClients {
		clients = append(clients, conn)
	}
	ismismcubeClientsMux.RUnlock()
	for _, conn := range clients {
		ismismcubeClientsMux.RLock()
		clientInfo, exists := ismismcubeClients[conn]
		ismismcubeClientsMux.RUnlock()
		if !exists {
			continue
		}
		clientInfo.WriteMutex.Lock()
		conn.WriteMessage(websocket.TextMessage, data)
		clientInfo.WriteMutex.Unlock()
	}
}

func HandleIsmismcubeOnline(w http.ResponseWriter, r *http.Request) {
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
	RegisterIsmismcubeClient(conn)
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
			UnregisterIsmismcubeClient(conn)
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					ismismcubeClientsMux.RLock()
					clientInfo, exists := ismismcubeClients[conn]
					ismismcubeClientsMux.RUnlock()
					if exists {
						clientInfo.WriteMutex.Lock()
						conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(config.WSWriteWait))
						clientInfo.WriteMutex.Unlock()
					}
				}
				break
			}
		}
	}()
	go func() {
		defer ticker.Stop()
		for {
			<-ticker.C
			ismismcubeClientsMux.RLock()
			clientInfo, exists := ismismcubeClients[conn]
			ismismcubeClientsMux.RUnlock()
			if !exists {
				return
			}
			clientInfo.WriteMutex.Lock()
			conn.WriteMessage(websocket.PingMessage, nil)
			clientInfo.WriteMutex.Unlock()
		}
	}()
}
