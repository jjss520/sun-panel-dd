package global

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"sun-panel/models"

	"gorm.io/gorm"
)

// RemindClient 表示一个SSE客户端连接
type RemindClient struct {
	UserID uint
	Conn   chan map[string]interface{}
}

// RemindChecker 提醒检查器
type RemindChecker struct {
	mu      sync.RWMutex
	clients map[uint][]*RemindClient // userID -> clients
	db      *gorm.DB
}

var remindChecker *RemindChecker

// InitRemindChecker 初始化提醒检查器
func InitRemindChecker(db *gorm.DB) {
	remindChecker = &RemindChecker{
		clients: make(map[uint][]*RemindClient),
		db:      db,
	}

	// 启动定时检查任务（每分钟检查一次）
	go remindChecker.startChecker()
	log.Println("[提醒检查器] 已启动，每分钟检查一次")
}

// startChecker 启动定时检查
func (rc *RemindChecker) startChecker() {
	// 计算到下一个整分钟的等待时间
	now := time.Now()
	nextMinute := now.Truncate(time.Minute).Add(time.Minute)
	initialDelay := nextMinute.Sub(now)
	
	log.Printf("[提醒检查器] 首次执行将在 %v 后 (%s)", initialDelay, nextMinute.Format("15:04:05"))
	
	// 等待到整分钟
	time.Sleep(initialDelay)
	
	// 立即执行第一次检查
	rc.checkDueReminds()
	
	// 之后每分钟执行一次
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rc.checkDueReminds()
	}
}

// checkDueReminds 检查到期的提醒
func (rc *RemindChecker) checkDueReminds() {
	var notepads []models.Notepad
	
	// 查询所有未提醒的便签（只查询有在线用户的便签）
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	
	if len(rc.clients) == 0 {
		return // 没有在线用户，跳过检查
	}
	
	// 获取所有在线用户ID
	userIDs := make([]uint, 0, len(rc.clients))
	for userID := range rc.clients {
		userIDs = append(userIDs, userID)
	}
	
	// 只查询在线用户的便签
	// 包括：等待触发的(status=0) + 待确认的(status=1，用于强制提醒轮询)
	log.Printf("[提醒检查器] 查询条件: user_id IN %v, remind_status IN (0, 1)", userIDs)
	if err := rc.db.Where("user_id IN ? AND remind_time IS NOT NULL AND remind_time != '' AND remind_status IN (0, 1)", userIDs).Find(&notepads).Error; err != nil {
		log.Printf("[提醒检查器] 查询失败: %v", err)
		return
	}
	log.Printf("[提醒检查器] 查询结果: %d 个便签", len(notepads))

	now := time.Now().Add(2 * time.Second)
	log.Printf("[提醒检查器] 检查 %d 个便签，当前时间: %s", len(notepads), now.Format("2006-01-02 15:04:05"))
	
	// 打印所有查询到的便签信息（调试用）
	for _, note := range notepads {
		// 显示原始时间和实际触发时间
		baseTimeStr := "无"
		actualRemindTime := "无"
		if note.RemindBaseTime != "" {
			baseTimeStr = note.RemindBaseTime
		}
		if note.RemindTime != "" {
			actualRemindTime = note.RemindTime
			if note.RemindRepeat == "" || note.RemindRepeat == "none" {
				actualRemindTime += " (一次性)"
			}
		}
		
		// 提醒方式描述
		remindMode := "普通"
		if note.RemindAdvanceDays > 0 {
			remindMode += fmt.Sprintf(" [提前%d天]", note.RemindAdvanceDays)
		}
		
		repeatText := "不重复"
		if note.RemindRepeat != "" && note.RemindRepeat != "none" {
			repeatMap := map[string]string{
				"daily": "每天",
				"weekly": "每周",
				"monthly": "每月",
				"yearly": "每年",
			}
			if text, ok := repeatMap[note.RemindRepeat]; ok {
				repeatText = text
			}
		}
		
		log.Printf("[提醒检查器]   - ID=%d, Title=%s, 原始时间=%s, 实际触发=%s, Status=%d, UserID=%d, 重复=%s, 方式=%s", 
			note.ID, note.Title, baseTimeStr, actualRemindTime, note.RemindStatus, note.UserID, repeatText, remindMode)
	}

	for _, note := range notepads {
		if note.RemindTime == "" {
			continue
		}

		// 解析提醒时间（使用本地时区）
		var remindTime time.Time
		var err error
		
		// 尝试多种时间格式，都使用本地时区解析
		parseFormats := []string{
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
		}
		
		parsed := false
		for _, format := range parseFormats {
			t, e := time.ParseInLocation(format, note.RemindTime, time.Local)
			if e == nil {
				remindTime = t
				err = nil
				parsed = true
				break
			}
		}
		
		if !parsed {
			log.Printf("[提醒检查器] 时间解析失败: %v, remindTime: %s", err, note.RemindTime)
			continue
		}

		// 状态机处理
		if note.RemindStatus == 0 {
			// === 状态 0：等待触发 ===
			
			// 数据库中的 remindTime 已经是实际触发时间（包含提前天数的计算）
			// 直接使用该时间进行判断，不需要再次处理提前天数
			actualRemindTime := remindTime
			
			// 判断是否到达提醒时间
			// 将双方时间都截断到“分钟”（抹除秒数带来的误差）
			nowMinute := now.Truncate(time.Minute)
			actualMinute := actualRemindTime.Truncate(time.Minute)

			if nowMinute.Sub(actualMinute) >= 0 {
				// 触发提醒，更新状态为 1（待确认）
				log.Printf("[提醒检查器] 触发提醒: UserID=%d, Title=%s, RemindTime=%s, Status=0->1", 
					note.UserID, note.Title, note.RemindTime)
				
				// 构建提醒数据
				remindData := map[string]interface{}{
					"id":                note.ID,
					"title":             note.Title,
					"content":           note.Content,
					"remindTime":        note.RemindTime,
					"remindRepeat":      note.RemindRepeat,
					"remindAdvanceDays": note.RemindAdvanceDays,
				}
				
				// 推送给对应用户的所有客户端
				rc.pushToUser(note.UserID, remindData)
				
				// 更新状态为 1（待确认）
				rc.db.Model(&note).Update("remind_status", 1)
			}
			
		} else if note.RemindStatus == 1 {
			// === 状态 1：待确认 ===
			// 持续推送，直到用户确认（强制提醒）
			log.Printf("[提醒检查器] 强制提醒轮询: UserID=%d, Title=%s", 
				note.UserID, note.Title)
			
			// 构建提醒数据
			remindData := map[string]interface{}{
				"id":                note.ID,
				"title":             note.Title,
				"content":           note.Content,
				"remindTime":        note.RemindTime,
				"remindRepeat":      note.RemindRepeat,
				"remindAdvanceDays": note.RemindAdvanceDays,
			}
			
			// 推送（不改变状态，保持为 1）
			rc.pushToUser(note.UserID, remindData)
		}
	}
}

// CalculateNextRemindTime 计算重复提醒的下一次触发时间（公开方法）
func (rc *RemindChecker) CalculateNextRemindTime(baseTime time.Time, repeatType string, now time.Time) time.Time {
	// 如果基准时间已经是未来的时间，直接返回，不需要计算下一个周期
	if baseTime.After(now) {
		return baseTime
	}
	
	// 基准时间已过，循环计算下一个周期的时间，直到找到未来的时间
	var nextTime time.Time
	switch repeatType {
	case "daily":
		// 每天：从明天开始
		nextTime = time.Date(now.Year(), now.Month(), now.Day(), baseTime.Hour(), baseTime.Minute(), baseTime.Second(), 0, baseTime.Location())
		nextTime = nextTime.AddDate(0, 0, 1) // 明天
		return nextTime

	case "weekly":
		// 每周：找到下一个相同的星期几
		daysDiff := int(baseTime.Weekday() - now.Weekday())
		if daysDiff <= 0 {
			daysDiff += 7
		}
		nextDate := now.AddDate(0, 0, daysDiff)
		nextTime = time.Date(nextDate.Year(), nextDate.Month(), nextDate.Day(), baseTime.Hour(), baseTime.Minute(), baseTime.Second(), 0, baseTime.Location())
		return nextTime

	case "monthly":
		// 每月：将基准时间的月份改为当前月，日期保持不变
		nextTime = time.Date(now.Year(), now.Month(), baseTime.Day(), baseTime.Hour(), baseTime.Minute(), baseTime.Second(), 0, baseTime.Location())
		// 如果本月该日期已过或就是今天但时间已过，则推迟到下个月
		if !nextTime.After(now) {
			nextTime = nextTime.AddDate(0, 1, 0)
		}
		return nextTime

	case "yearly":
		// 每年：将基准时间的年份改为今年
		nextTime = time.Date(now.Year(), baseTime.Month(), baseTime.Day(), baseTime.Hour(), baseTime.Minute(), baseTime.Second(), 0, baseTime.Location())
		// 如果今年的这个日期已过，则推迟到明年
		if !nextTime.After(now) {
			nextTime = nextTime.AddDate(1, 0, 0)
		}
		return nextTime

	default:
		return baseTime
	}
}


// pushToUser 推送提醒给指定用户
func (rc *RemindChecker) pushToUser(userID uint, data map[string]interface{}) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	clients, exists := rc.clients[userID]
	if !exists {
		log.Printf("[提醒检查器] 用户 %d 没有在线客户端", userID)
		return
	}

	successCount := 0
	for _, client := range clients {
		select {
		case client.Conn <- data:
			successCount++
		default:
			log.Printf("[提醒检查器] 客户端通道已满，跳过")
		}
	}

	log.Printf("[提醒检查器] 推送成功: UserID=%d, 成功=%d/%d", userID, successCount, len(clients))
}

// AddClient 添加SSE客户端
func (rc *RemindChecker) AddClient(userID uint) *RemindClient {
	client := &RemindClient{
		UserID: userID,
		Conn:   make(chan map[string]interface{}, 10),
	}

	rc.mu.Lock()
	rc.clients[userID] = append(rc.clients[userID], client)
	rc.mu.Unlock()

	log.Printf("[提醒检查器] 客户端已连接: UserID=%d, 总连接数=%d", userID, len(rc.clients[userID]))
	return client
}

// RemoveClient 移除SSE客户端
func (rc *RemindChecker) RemoveClient(userID uint, client *RemindClient) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	clients, exists := rc.clients[userID]
	if !exists {
		return
	}

	// 从切片中移除
	for i, c := range clients {
		if c == client {
			rc.clients[userID] = append(clients[:i], clients[i+1:]...)
			close(client.Conn)
			log.Printf("[提醒检查器] 客户端已断开: UserID=%d", userID)
			break
		}
	}

	// 如果没有客户端了，删除map entry
	if len(rc.clients[userID]) == 0 {
		delete(rc.clients, userID)
	}
}

// SSEHandler SSE HTTP处理器
func (rc *RemindChecker) SSEHandler(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID（从URL参数或Header）
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		http.Error(w, "Missing userId", http.StatusBadRequest)
		return
	}

	var userID uint
	fmt.Sscanf(userIDStr, "%d", &userID)

	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 添加客户端
	client := rc.AddClient(userID)
	defer rc.RemoveClient(userID, client)

	// 发送心跳
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// 发送初始连接成功消息
	fmt.Fprintf(w, "event: connected\ndata: {\"message\":\"connected\"}\n\n")
	flusher.Flush()

	// 监听客户端断开
	ctx := r.Context()

	// 主循环：等待推送或断开
	for {
		select {
		case <-ctx.Done():
			log.Printf("[提醒检查器] 客户端主动断开: UserID=%d", userID)
			return
		case data, ok := <-client.Conn:
			if !ok {
				return
			}

			// 格式化SSE消息
			fmt.Fprintf(w, "event: remind\ndata: ")
			for k, v := range data {
				fmt.Fprintf(w, "%s:%v,", k, v)
			}
			fmt.Fprintf(w, "\n\n")
			flusher.Flush()
		}
	}
}

// GetRemindChecker 获取全局提醒检查器实例
func GetRemindChecker() *RemindChecker {
	return remindChecker
}
