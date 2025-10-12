package ws

import (
	"fmt"
	"ismismcube-backend/internal/config"
	"log"
	"net"
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
			UnregisterIsmismcubeClient(conn)
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
