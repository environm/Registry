// pkg/registry/server/handler/edge_connect_handler.go
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

type EdgeConnectRequest struct {
	CloudIP   string `json:"cloud_ip"`
	CloudPort string `json:"cloud_port"`
	NodeName  string `json:"node_name"`
}

type EdgeConnectResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	LocalIP   string `json:"local_ip"`
	RemoteIP  string `json:"remote_ip"`
	Connected bool   `json:"connected"`
}

type EdgeConnectHandler struct {
	activeConnections map[string]*EdgeClient
	mutex             sync.RWMutex
}

func NewEdgeConnectHandler() *EdgeConnectHandler {
	return &EdgeConnectHandler{
		activeConnections: make(map[string]*EdgeClient),
	}
}

func (ech *EdgeConnectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EdgeConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证参数
	if req.CloudIP == "" || req.CloudPort == "" || req.NodeName == "" {
		http.Error(w, "cloud_ip, cloud_port and node_name are required", http.StatusBadRequest)
		return
	}

	ech.handleEdgeConnect(w, req)
}

func (ech *EdgeConnectHandler) GetHandler() http.HandlerFunc {
	return ech.ServeHTTP
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "unknown"
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0]
}
func (ech *EdgeConnectHandler) handleEdgeConnect(w http.ResponseWriter, req EdgeConnectRequest) {
	cloudAddr := fmt.Sprintf("%s:%s", req.CloudIP, req.CloudPort)

	// 创建新的EdgeClient
	client := NewEdgeClient(cloudAddr, req.NodeName)

	// 尝试连接
	err := client.Connect()
	if err != nil {
		response := EdgeConnectResponse{
			Status:    "error",
			Message:   fmt.Sprintf("Connection failed: %v", err),
			LocalIP:   getLocalIP(),
			Connected: false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 存储连接
	ech.mutex.Lock()
	ech.activeConnections[req.NodeName] = client
	ech.mutex.Unlock()

	// 启动客户端协程
	go client.Start()

	// 返回成功响应
	response := EdgeConnectResponse{
		Status:    "success",
		Message:   "Edge connection established successfully",
		LocalIP:   getLocalIP(),
		RemoteIP:  client.Conn.RemoteAddr().String(),
		Connected: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("Edge connection established to %s as %s", cloudAddr, req.NodeName)
}

type EdgeClient struct {
	Conn        net.Conn
	CloudAddr   string
	NodeID      string
	Reconnect   bool
	MessageChan chan string
	StopChan    chan bool
}

func NewEdgeClient(cloudAddr, nodeID string) *EdgeClient {
	return &EdgeClient{
		CloudAddr:   cloudAddr,
		NodeID:      nodeID,
		Reconnect:   true,
		MessageChan: make(chan string, 100),
		StopChan:    make(chan bool),
	}
}

func (ec *EdgeClient) Connect() error {
	conn, err := net.Dial("tcp", ec.CloudAddr)
	if err != nil {
		return err
	}

	ec.Conn = conn

	// 发送 HTTP 升级请求
	upgradeRequest := fmt.Sprintf("GET /websocket HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nNode-ID: %s\r\n\r\n",
		ec.CloudAddr, ec.NodeID)

	_, err = conn.Write([]byte(upgradeRequest))
	if err != nil {
		conn.Close()
		return err
	}

	log.Printf("Sent WebSocket upgrade request to %s", ec.CloudAddr)
	return nil
}

func (ec *EdgeClient) Start() {
	defer ec.Conn.Close()

	reader := bufio.NewReader(ec.Conn)

	// 读取欢迎消息
	welcome, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Failed to read welcome message: %v", err)
		return
	}
	log.Printf("Cloud server: %s", strings.TrimSpace(welcome))

	// 启动消息发送协程
	go ec.handleSend()

	// 主循环处理接收消息
	for {
		select {
		case <-ec.StopChan:
			log.Println("Stopping edge client")
			return

		default:
			// 设置读取超时
			ec.Conn.SetReadDeadline(time.Now().Add(35 * time.Second))

			message, err := reader.ReadString('\n')
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // 超时继续
				}
				log.Printf("Connection lost: %v", err)
				ec.reconnect()
				return
			}

			message = strings.TrimSpace(message)
			ec.handleMessage(message)
		}
	}
}

func (ec *EdgeClient) handleMessage(message string) {
	switch message {
	case "HEARTBEAT":
		// 响应心跳
		ec.Conn.Write([]byte("PONG\n"))
		log.Printf("Heartbeat responded")

	default:
		if strings.HasPrefix(message, "ACK:") {
			log.Printf("Cloud acknowledged: %s", message[4:])
		} else {
			log.Printf("Received from cloud: %s", message)
		}
	}
}

func (ec *EdgeClient) handleSend() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	counter := 0
	for {
		select {
		case <-ec.StopChan:
			return

		case msg := <-ec.MessageChan:
			// 发送用户输入的消息
			_, err := ec.Conn.Write([]byte(msg + "\n"))
			if err != nil {
				log.Printf("Failed to send message: %v", err)
			}

		case <-ticker.C:
			// 定时发送示例数据
			counter++
			data := fmt.Sprintf("Edge data #%d at %s", counter, time.Now().Format("15:04:05"))
			_, err := ec.Conn.Write([]byte(data + "\n"))
			if err != nil {
				log.Printf("Failed to send data: %v", err)
			} else {
				log.Printf("Sent: %s", data)
			}
		}
	}
}

func (ec *EdgeClient) SendMessage(message string) {
	ec.MessageChan <- message
}

func (ec *EdgeClient) Stop() {
	ec.Reconnect = false
	close(ec.StopChan)
	if ec.Conn != nil {
		ec.Conn.Close()
	}
}

func (ec *EdgeClient) reconnect() {
	if !ec.Reconnect {
		return
	}

	log.Println("Attempting to reconnect...")

	for {
		time.Sleep(5 * time.Second)
		err := ec.Connect()
		if err != nil {
			log.Printf("Reconnect failed: %v, retrying...", err)
			continue
		}

		log.Println("Reconnected successfully")
		go ec.Start()
		break
	}
}
