package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"ismismcube-backend/internal/config"
	"ismismcube-backend/internal/model"
	"ismismcube-backend/internal/toolkit"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ChatTask struct {
	CreatedAt     time.Time       `json:"created_at"`
	Content       []byte          `json:"-"`
	WebSocketID   string          `json:"websocket_id"`
	WebSocketConn *websocket.Conn `json:"-"`
	WriteMutex    sync.Mutex      `json:"-"`
}

type QueuePositionData struct {
	QueuePosition int `json:"queue_position"`
}

type QueueBroadcaster interface {
	BroadcastQueueStats(waiting, executing int, broadcastFlag int64)
}

type TaskManager struct {
	pendingTasks   map[string]*ChatTask
	waitingTasks   []*ChatTask
	executingTasks map[string]*ChatTask
	taskMutex      sync.RWMutex
	broadcaster    QueueBroadcaster
	broadcastFlag  int64
	broadcastMutex sync.Mutex
}

var (
	taskManager *TaskManager
)

func GetTaskManager() *TaskManager {
	return taskManager
}

func InitTaskManager(broadcaster QueueBroadcaster) {
	taskManager = &TaskManager{
		pendingTasks:   make(map[string]*ChatTask),
		waitingTasks:   make([]*ChatTask, 0),
		executingTasks: make(map[string]*ChatTask),
	}
	taskManager.broadcaster = broadcaster
}

func (tm *TaskManager) CreateChatTask(content []byte, websocketID string) {
	task := &ChatTask{
		CreatedAt:   time.Now(),
		Content:     content,
		WebSocketID: websocketID,
	}
	tm.taskMutex.Lock()
	tm.pendingTasks[websocketID] = task
	tm.taskMutex.Unlock()
	go func() {
		timer := time.NewTimer(10 * time.Second)
		defer timer.Stop()
		<-timer.C
		tm.taskMutex.Lock()
		delete(tm.pendingTasks, websocketID)
		tm.taskMutex.Unlock()
	}()
}

func (tm *TaskManager) RegisterTaskConnection(websocketID string, conn *websocket.Conn) {
	tm.taskMutex.Lock()
	defer tm.taskMutex.Unlock()
	if task, exists := tm.pendingTasks[websocketID]; exists {
		task.WebSocketConn = conn
		tm.waitingTasks = append(tm.waitingTasks, task)
		delete(tm.pendingTasks, websocketID)
		go tm.broadcaster.BroadcastQueueStats(len(tm.waitingTasks), len(tm.executingTasks), tm.GetBroadcastFlag())
		go tm.sendTaskPosition(task, len(tm.waitingTasks)-1)
		// 触发任务调度
		go tm.checkTasks()
		return
	}
	// 执行中的任务允许重连
	if task, exists := tm.executingTasks[websocketID]; exists {
		task.WebSocketConn = conn
		return
	}
}

func (tm *TaskManager) UnregisterTaskConnection(websocketID string) {
	tm.taskMutex.Lock()
	defer tm.taskMutex.Unlock()
	// 如果是排队中的任务，直接移除并更新广播
	for i, task := range tm.waitingTasks {
		if task.WebSocketID == websocketID {
			tm.waitingTasks = append(tm.waitingTasks[:i], tm.waitingTasks[i+1:]...)
			go tm.broadcaster.BroadcastQueueStats(len(tm.waitingTasks), len(tm.executingTasks), tm.GetBroadcastFlag())
			go tm.broadcastTasksPositions()
			return
		}
	}
	// 如果是执行中的任务，断开后保留在executingTasks中，留给callLLM处理
	if task, exists := tm.executingTasks[websocketID]; exists {
		task.WebSocketConn = nil
		return
	}
}

func (tm *TaskManager) checkTasks() {
	tm.taskMutex.Lock()
	defer tm.taskMutex.Unlock()
	if len(tm.executingTasks) >= config.LLMConfigure.MaxConcurrentTasks {
		return
	}
	tasksScheduled := false
	for i := 0; i < len(tm.waitingTasks); i++ {
		if len(tm.executingTasks) >= config.LLMConfigure.MaxConcurrentTasks {
			break
		}
		tasksScheduled = true
		task := tm.waitingTasks[i]
		tm.waitingTasks = append(tm.waitingTasks[:i], tm.waitingTasks[i+1:]...)
		i--
		tm.executingTasks[task.WebSocketID] = task
		go tm.executeTask(task)
		go tm.sendTaskPosition(task, -1)
	}
	if tasksScheduled {
		go tm.broadcastTasksPositions()
		go tm.broadcaster.BroadcastQueueStats(len(tm.waitingTasks), len(tm.executingTasks), tm.GetBroadcastFlag())
	}
}

func (tm *TaskManager) executeTask(task *ChatTask) {
	model.ExecutedTask.GetAndIncrement()
	defer func() {
		tm.taskMutex.Lock()
		conn := task.WebSocketConn
		delete(tm.executingTasks, task.WebSocketID)
		tm.taskMutex.Unlock()
		if conn != nil {
			task.WriteMutex.Lock()
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			conn.Close()
			task.WriteMutex.Unlock()
		}
		waiting, executing := tm.GetQueueCount()
		go tm.broadcaster.BroadcastQueueStats(waiting, executing, tm.GetBroadcastFlag())
		go tm.checkTasks()
	}()
	timeoutMinutes := config.LLMConfigure.Timeout
	timer := time.NewTimer(time.Duration(timeoutMinutes) * time.Minute)
	defer timer.Stop()
	go func() {
		<-timer.C
		tm.taskMutex.RLock()
		conn := task.WebSocketConn
		tm.taskMutex.RUnlock()
		if conn != nil {
			data := &toolkit.MessageData{
				Type: "error",
				Data: toolkit.ErrorData{
					Error: fmt.Sprintf("Task timed out after %d minutes", timeoutMinutes),
				},
			}
			msg, err := data.ToBytes()
			if err != nil {
				return
			}
			task.WriteMutex.Lock()
			conn.WriteMessage(websocket.TextMessage, msg)
			task.WriteMutex.Unlock()
		}
	}()
	tm.callLLM(task)
}

func (tm *TaskManager) callLLM(task *ChatTask) {
	tm.taskMutex.RLock()
	conn := task.WebSocketConn
	tm.taskMutex.RUnlock()
	if conn == nil {
		return
	}
	var requestData map[string]interface{}
	if err := json.Unmarshal(task.Content, &requestData); err != nil {
		data := &toolkit.MessageData{
			Type: "error",
			Data: toolkit.ErrorData{Error: "Failed to parse request content"},
		}
		msg, _ := data.ToBytes()
		task.WriteMutex.Lock()
		conn.WriteMessage(websocket.TextMessage, msg)
		task.WriteMutex.Unlock()
		return
	}
	var modelName string
	if model, ok := requestData["model"].(string); ok && model != "" {
		modelName = model
	} else {
		if len(config.LLMConfigure.AvailableModels) == 0 {
			data := &toolkit.MessageData{
				Type: "error",
				Data: toolkit.ErrorData{Error: "No available models"},
			}
			msg, _ := data.ToBytes()
			task.WriteMutex.Lock()
			conn.WriteMessage(websocket.TextMessage, msg)
			task.WriteMutex.Unlock()
			return
		}
		modelName = config.LLMConfigure.AvailableModels[0]
	}
	baseApiUrl := config.LLMConfigure.BaseApiUrl
	var LLMApiUrl string
	domainStart := bytes.Index([]byte(baseApiUrl), []byte("://")) + 3
	pathStart := bytes.IndexByte([]byte(baseApiUrl[domainStart:]), '/')
	if pathStart == -1 {
		LLMApiUrl = baseApiUrl + "/" + modelName
	} else {
		LLMApiUrl = baseApiUrl[:domainStart+pathStart] + "/" + modelName + baseApiUrl[domainStart+pathStart:]
	}
	client := &http.Client{
		Timeout: time.Duration(config.LLMConfigure.Timeout) * time.Minute,
	}
	req, err := http.NewRequest("POST", LLMApiUrl, bytes.NewBuffer(task.Content))
	if err != nil {
		data := &toolkit.MessageData{
			Type: "error",
			Data: toolkit.ErrorData{Error: "Failed to create request"},
		}
		msg, _ := data.ToBytes()
		task.WriteMutex.Lock()
		conn.WriteMessage(websocket.TextMessage, msg)
		task.WriteMutex.Unlock()
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Authorization", "Bearer "+config.LLMConfigure.ApiKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to send request to AI API", err)
		data := &toolkit.MessageData{
			Type: "error",
			Data: toolkit.ErrorData{Error: "Failed to send request to AI API"},
		}
		msg, _ := data.ToBytes()
		task.WriteMutex.Lock()
		conn.WriteMessage(websocket.TextMessage, msg)
		task.WriteMutex.Unlock()
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		data := &toolkit.MessageData{
			Type: "error",
			Data: toolkit.ErrorData{
				Error: fmt.Sprintf("AI API returned status %d: %s", resp.StatusCode, string(errorBody)),
			},
		}
		msg, _ := data.ToBytes()
		task.WriteMutex.Lock()
		conn.WriteMessage(websocket.TextMessage, msg)
		task.WriteMutex.Unlock()
		return
	}
	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		tm.taskMutex.RLock()
		conn := task.WebSocketConn
		tm.taskMutex.RUnlock()
		if conn == nil {
			return
		}
		if n > 0 {
			task.WriteMutex.Lock()
			// 需要设置写入超时，防止客户端异常断开后，tcp缓冲区已满导致websocket被长时间阻塞
			conn.SetWriteDeadline(time.Now().Add(config.WSWriteWait))
			err := conn.WriteMessage(websocket.TextMessage, buffer[:n])
			task.WriteMutex.Unlock()
			if err != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
}

func (tm *TaskManager) GetQueueCount() (waiting, executing int) {
	tm.taskMutex.RLock()
	waiting = len(tm.waitingTasks)
	executing = len(tm.executingTasks)
	tm.taskMutex.RUnlock()
	return waiting, executing
}

func (tm *TaskManager) broadcastTasksPositions() {
	tm.taskMutex.RLock()
	tasks := make([]*ChatTask, len(tm.waitingTasks))
	copy(tasks, tm.waitingTasks)
	tm.taskMutex.RUnlock()
	for i, task := range tasks {
		tm.sendTaskPosition(task, i)
	}
}

func (tm *TaskManager) sendTaskPosition(task *ChatTask, position int) {
	tm.taskMutex.RLock()
	conn := task.WebSocketConn
	tm.taskMutex.RUnlock()
	if conn == nil {
		return
	}
	data := &toolkit.MessageData{
		Type: "broadcast",
		Data: QueuePositionData{
			QueuePosition: position,
		},
	}
	msg, err := data.ToBytes()
	if err != nil {
		return
	}
	task.WriteMutex.Lock()
	conn.WriteMessage(websocket.TextMessage, msg)
	task.WriteMutex.Unlock()
}

func (tm *TaskManager) GetBroadcastFlag() int64 {
	tm.broadcastMutex.Lock()
	broadcastFlag := tm.broadcastFlag
	tm.broadcastFlag++
	tm.broadcastMutex.Unlock()
	return broadcastFlag
}
