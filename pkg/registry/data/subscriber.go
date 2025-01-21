package data

import (
	"fmt"
	"sync"
)

type SubscriptionManager struct {
	subscribers map[string][]string // 文件名 -> 客户端地址列表
	mutex       sync.Mutex          // 确保线程安全
}

// NewSubscriptionManager 创建新的订阅管理器
func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		subscribers: make(map[string][]string),
	}
}

func getKey(fileName, tag string) string {
	return fmt.Sprintf("%s_%s", fileName, tag)
}

// Subscribe 添加订阅
func (sm *SubscriptionManager) Subscribe(fileName, tag string, client string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 如果文件名未订阅，初始化列表
	key := getKey(fileName, tag)
	if _, exists := sm.subscribers[key]; !exists {
		sm.subscribers[key] = []string{}
	}

	// 添加订阅
	sm.subscribers[key] = append(sm.subscribers[key], client)
}

// GetSubscribers 获取订阅文件的客户端列表
func (sm *SubscriptionManager) GetSubscribers(fileName, tag string) []string {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	key := getKey(fileName, tag)
	return sm.subscribers[key]
}

// Unsubscribe 取消订阅（可选功能）
func (sm *SubscriptionManager) Unsubscribe(fileName, tag string, client string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	key := getKey(fileName, tag)
	subscribers := sm.subscribers[key]
	for i, subscriber := range subscribers {
		if subscriber == client {
			sm.subscribers[key] = append(subscribers[:i], subscribers[i+1:]...)
			break
		}
	}
}

// 判断文件是否订阅
func (sm *SubscriptionManager) IsSubscribed(fileName, tag string) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	key := getKey(fileName, tag)
	_, exists := sm.subscribers[key]
	return exists
}
