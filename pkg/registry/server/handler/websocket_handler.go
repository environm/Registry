// pkg/registry/server/handler/websocket_handler.go
package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type EdgeNode struct {
	Conn    net.Conn
	ID      string
	Address string
}

type EdgeMessage struct {
	From    string
	To      string
	Content string
	Type    string
	Time    time.Time
}

type WebSocketHandler struct {
	EdgeNodes   sync.Map
	MessageChan chan *EdgeMessage
}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		MessageChan: make(chan *EdgeMessage, 100),
	}
}

func (wsh *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("WebSocket connection request from %s", r.RemoteAddr)
	wsh.handleWebSocketConnection(w, r)
}

func (wsh *WebSocketHandler) GetHandler() http.HandlerFunc {
	return wsh.ServeHTTP
}

func (wsh *WebSocketHandler) handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}

	conn, bufrw, err := hj.Hijack()
	if err != nil {
		log.Printf("Hijack failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = conn.Write([]byte("HTTP/1.1 101 Switching Protocols\r\nConnection: upgrade\r\nUpgrade: tcp\r\n\r\n"))
	if err != nil {
		log.Printf("Failed to send switching protocols: %v", err)
		conn.Close()
		return
	}

	// 读取节点ID
	nodeID, err := bufrw.ReadString('\n')
	if err != nil {
		log.Printf("Failed to read node ID: %v", err)
		conn.Close()
		return
	}
	nodeID = strings.TrimSpace(nodeID)

	// 创建边节点记录
	edgeNode := &EdgeNode{
		Conn:    conn,
		ID:      nodeID,
		Address: r.RemoteAddr,
	}

	wsh.EdgeNodes.Store(nodeID, edgeNode)
	log.Printf("Edge node %s connected via WebSocket from %s", nodeID, edgeNode.Address)
	welcomeMsg := fmt.Sprintf("Welcome edge node %s! WebSocket connection established at %s\n",
		nodeID, time.Now().Format("2006-01-02 15:04:05"))
	conn.Write([]byte(welcomeMsg))

	go wsh.handleEdgeMessageLoop(conn, bufrw.Reader, nodeID)
}

func (wsh *WebSocketHandler) handleEdgeMessageLoop(conn net.Conn, reader *bufio.Reader, nodeID string) {
	defer func() {
		conn.Close()
		wsh.EdgeNodes.Delete(nodeID)
		log.Printf("Edge node %s disconnected from WebSocket", nodeID)
	}()

	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-heartbeatTicker.C:
			// 发送心跳
			_, err := conn.Write([]byte("HEARTBEAT\n"))
			if err != nil {
				log.Printf("Failed to send heartbeat to %s: %v", nodeID, err)
				return
			}

		default:
			// 设置读超时
			conn.SetReadDeadline(time.Now().Add(35 * time.Second))

			message, err := reader.ReadString('\n')
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Printf("Read error from %s: %v", nodeID, err)
				return
			}

			message = strings.TrimSpace(message)
			wsh.processMessage(nodeID, message, conn)
		}
	}
}

func (wsh *WebSocketHandler) processMessage(nodeID, message string, conn net.Conn) {
	switch message {
	case "PONG":
		log.Printf("Received PONG from %s", nodeID)

	default:
		log.Printf("Received from %s: %s", nodeID, message)

		wsh.MessageChan <- &EdgeMessage{
			From:    nodeID,
			Content: message,
			Type:    "data",
			Time:    time.Now(),
		}

		// 回复确认
		ackMsg := fmt.Sprintf("ACK: %s\n", message)
		conn.Write([]byte(ackMsg))
	}
}

type EdgeNodeInfo struct {
	NodeID       string    `json:"node_id"`
	IPAddress    string    `json:"ip_address"`
	ConnectedAt  time.Time `json:"connected_at"`
	LastActivity time.Time `json:"last_activity"`
	Status       string    `json:"status"`
}

// 获取所有连接的边节点信息
func (wsh *WebSocketHandler) GetAllEdgeNodes() []EdgeNodeInfo {
	var nodes []EdgeNodeInfo

	wsh.EdgeNodes.Range(func(key, value interface{}) bool {
		edgeNode := value.(*EdgeNode)

		nodeInfo := EdgeNodeInfo{
			NodeID:       edgeNode.ID,
			IPAddress:    wsh.extractIPFromAddress(edgeNode.Address),
			ConnectedAt:  time.Now(),
			LastActivity: time.Now(),
			Status:       "connected",
		}

		nodes = append(nodes, nodeInfo)
		return true
	})

	return nodes
}

// 获取连接统计信息
func (wsh *WebSocketHandler) GetConnectionStats() map[string]interface{} {
	stats := make(map[string]interface{})
	count := 0
	var nodeList []string

	wsh.EdgeNodes.Range(func(key, value interface{}) bool {
		count++
		nodeList = append(nodeList, key.(string))
		return true
	})

	stats["total_connections"] = count
	stats["connected_nodes"] = nodeList
	stats["timestamp"] = time.Now().Format(time.RFC3339)

	return stats
}

// 从地址中提取IP（处理各种格式的地址）
func (wsh *WebSocketHandler) extractIPFromAddress(address interface{}) string {
	switch addr := address.(type) {
	case string:
		if strings.Contains(addr, ":") {
			return strings.Split(addr, ":")[0]
		}
		return addr
	case net.Addr:
		return strings.Split(addr.String(), ":")[0]
	default:
		return "unknown"
	}
}

type EdgeNodesHandler struct {
	WebSocketHandler *WebSocketHandler
}

func NewEdgeNodesHandler(wsHandler *WebSocketHandler) *EdgeNodesHandler {
	return &EdgeNodesHandler{
		WebSocketHandler: wsHandler,
	}
}

func (enh *EdgeNodesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch r.URL.Path {
	case "/edge/nodes":
		enh.handleGetAllNodes(w, r)
	case "/edge/connections":
		enh.handleGetConnections(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (enh *EdgeNodesHandler) GetHandler() http.HandlerFunc {
	return enh.ServeHTTP
}

// 获取所有节点详细信息
func (enh *EdgeNodesHandler) handleGetAllNodes(w http.ResponseWriter, r *http.Request) {
	nodes := enh.WebSocketHandler.GetAllEdgeNodes()

	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"nodes": nodes,
			"count": len(nodes),
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 获取连接统计信息
func (enh *EdgeNodesHandler) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats := enh.WebSocketHandler.GetConnectionStats()

	response := map[string]interface{}{
		"status": "success",
		"data":   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 获取连接列表（简化版）
func (enh *EdgeNodesHandler) handleGetConnections(w http.ResponseWriter, r *http.Request) {
	var connections []map[string]string

	enh.WebSocketHandler.EdgeNodes.Range(func(key, value interface{}) bool {
		edgeNode := value.(*EdgeNode)

		connInfo := map[string]string{
			"node_id":         edgeNode.ID,
			"ip_address":      enh.WebSocketHandler.extractIPFromAddress(edgeNode.Address),
			"connection_time": time.Now().Format("2006-01-02 15:04:05"),
		}

		connections = append(connections, connInfo)
		return true
	})

	response := map[string]interface{}{
		"status":      "success",
		"connections": connections,
		"total":       len(connections),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
