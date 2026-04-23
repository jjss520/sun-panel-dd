package panel

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/lib/cmn"
	"sun-panel/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Notepad struct{}

// Get 获取单个便签
func (a *Notepad) Get(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	var req struct {
		Id uint `form:"id"`
	}
	c.ShouldBindQuery(&req)

	var notepad models.Notepad
	db := global.Db.Where("user_id = ?", userInfo.ID)

	if req.Id > 0 {
		db = db.Where("id = ?", req.Id)
	} else {
		db = db.Order("updated_at desc") // 默认最近的
	}

	if err := db.First(&notepad).Error; err != nil {
		// 没找到，返回nil，前端视为新建
		apiReturn.SuccessData(c, nil)
		return
	}
	apiReturn.SuccessData(c, notepad)
}

// GetList 获取便签列表
func (a *Notepad) GetList(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	var list []models.Notepad
	if err := global.Db.Where("user_id = ?", userInfo.ID).Order("updated_at desc").Find(&list).Error; err != nil {
		apiReturn.Error(c, "Get List Failed")
		return
	}
	apiReturn.SuccessData(c, list)
}

// Save 保存（新增或更新）
func (a *Notepad) Save(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	type Req struct {
		Id                uint   `json:"id"`
		Title             string `json:"title"`
		Content           string `json:"content"`
		RemindTime        string `json:"remindTime,omitempty"`
		RemindStatus      int    `json:"remindStatus"`
		RemindRepeat      string `json:"remindRepeat,omitempty"`
		RemindForce       int    `json:"remindForce"`
		RemindAdvanceDays int    `json:"remindAdvanceDays"`
	}
	var req Req
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	var notepad models.Notepad
	if req.Id > 0 {
		// Update
		if err := global.Db.Where("user_id = ? AND id = ?", userInfo.ID, req.Id).First(&notepad).Error; err != nil {
			apiReturn.Error(c, "Not Found")
			return
		}
		notepad.Title = req.Title
		notepad.Content = req.Content
		notepad.RemindTime = req.RemindTime
		notepad.RemindStatus = req.RemindStatus
		notepad.RemindRepeat = req.RemindRepeat
		notepad.RemindForce = req.RemindForce
		notepad.RemindAdvanceDays = req.RemindAdvanceDays
		if err := global.Db.Save(&notepad).Error; err != nil {
			apiReturn.Error(c, "Update Failed")
			return
		}
	} else {
		// Create
		notepad = models.Notepad{
			UserID:            userInfo.ID,
			Title:             req.Title,
			Content:           req.Content,
			RemindTime:        req.RemindTime,
			RemindStatus:      req.RemindStatus,
			RemindRepeat:      req.RemindRepeat,
			RemindForce:       req.RemindForce,
			RemindAdvanceDays: req.RemindAdvanceDays,
		}
		if err := global.Db.Create(&notepad).Error; err != nil {
			apiReturn.Error(c, "Create Failed")
			return
		}
	}

	apiReturn.SuccessData(c, notepad)
}

// Delete 删除便签
func (a *Notepad) Delete(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	type Req struct {
		Id uint `json:"id"`
	}
	var req Req
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	// 先获取便签内容，用于解析文件链接
	var notepad models.Notepad
	if err := global.Db.Where("user_id = ? AND id = ?", userInfo.ID, req.Id).First(&notepad).Error; err != nil {
		apiReturn.Error(c, "Not Found")
		return
	}

	// 解析并删除文件
	deleteNotepadFiles(notepad.Content, userInfo.ID)

	// 删除便签记录
	if err := global.Db.Where("user_id = ? AND id = ?", userInfo.ID, req.Id).Delete(&models.Notepad{}).Error; err != nil {
		apiReturn.Error(c, "Delete Failed")
		return
	}
	apiReturn.Success(c)
}

// deleteNotepadFiles 解析便签内容中的文件链接并删除对应文件
func deleteNotepadFiles(content string, userId uint) {
	// 解析图片链接
	imgRegex := regexp.MustCompile(`<img[^>]+src="([^"]+)"`)
	imgMatches := imgRegex.FindAllStringSubmatch(content, -1)

	// 解析文件链接
	fileRegex := regexp.MustCompile(`<a[^>]+href="([^"]+)"`)
	fileMatches := fileRegex.FindAllStringSubmatch(content, -1)

	// 收集所有文件路径
	var filePaths []string
	for _, match := range imgMatches {
		if len(match) > 1 {
			filePaths = append(filePaths, match[1])
		}
	}
	for _, match := range fileMatches {
		if len(match) > 1 {
			filePaths = append(filePaths, match[1])
		}
	}

	// 使用系统配置的上传路径
	configUpload := global.Config.GetValueString("base", "source_path")

	// 获取所有文件记录，用于匹配删除
	var allFiles []models.File
	global.Db.Find(&allFiles, "user_id = ?", userId)

	// 删除文件和数据库记录
	for _, path := range filePaths {
		// 处理相对路径
		if strings.HasPrefix(path, "/") {
			path = path[1:]
		}

		// 构建完整文件路径
		fullPath := configUpload + "/" + path
		if strings.HasPrefix(path, "notepad/") {
			fullPath = configUpload + "/" + path
		}

		// 检查文件是否存在并删除
		if _, err := os.Stat(fullPath); err == nil {
			os.Remove(fullPath)
		}

		// 删除数据库记录
		for _, file := range allFiles {
			// 匹配文件路径
			if strings.Contains(file.Src, path) || strings.Contains(path, file.Src) {
				global.Db.Delete(&file)
				break
			}
		}
	}
}

// Upload 上传文件
func (a *Notepad) Upload(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	// 使用系统配置的上传路径
	configUpload := global.Config.GetValueString("base", "source_path")

	f, err := c.FormFile("file") // 前端这类请求通常用 file 字段
	if err != nil {
		apiReturn.ErrorByCode(c, 1300)
		return
	}

	fileExt := strings.ToLower(path.Ext(f.Filename))
	// 允许的扩展名，可以和 System File 配置一致，或者稍微放宽
	agreeExts := []string{".png", ".jpg", ".gif", ".jpeg", ".webp", ".svg", ".ico", ".txt", ".md", ".json", ".pdf", ".doc", ".docx", ".xls", ".xlsx"}

	if !cmn.InArray(agreeExts, fileExt) {
		// 暂时不严格限制，或者复用 system/file.go 的逻辑
		// 这里简化允许上传，但前端要显示为链接
	}

	fileName := cmn.Md5(fmt.Sprintf("%s%s", f.Filename, time.Now().String()))
	// 存放到 uploads/notepad/{year}/{month}/{day}/
	fildDir := fmt.Sprintf("%s/notepad/%d/%d/%d/", configUpload, time.Now().Year(), time.Now().Month(), time.Now().Day())

	isExist, _ := cmn.PathExists(fildDir)
	if !isExist {
		os.MkdirAll(fildDir, os.ModePerm)
	}
	filepath := fmt.Sprintf("%s%s%s", fildDir, fileName, fileExt)

	if err := c.SaveUploadedFile(f, filepath); err != nil {
		apiReturn.Error(c, "Upload Write Failed")
		return
	}

	// 记录到 models.File 表吗？
	// 最好记录，为了日后管理文件。
	mFile := models.File{}
	mFile.AddFile(userInfo.ID, f.Filename, fileExt, filepath)

	// 返回相对路径，前端补全 URL
	// 注意 filepath 是 ./uploads/... 前端需要 /uploads/...
	// 系统其他地方返回的是 filepath[1:] 即去掉开头的 .

	downloadUrl := filepath
	if strings.HasPrefix(filepath, ".") {
		downloadUrl = filepath[1:]
	}

	apiReturn.SuccessData(c, gin.H{
		"url":  downloadUrl,
		"name": f.Filename,
		"type": f.Header.Get("Content-Type"),
	})
}

// RemindStream SSE 实时推送提醒
func (a *Notepad) RemindStream(c *gin.Context) {
	userIdStr := c.Query("userId")
	if userIdStr == "" {
		c.JSON(400, gin.H{"error": "Missing userId"})
		return
	}

	var userId uint
	fmt.Sscanf(userIdStr, "%d", &userId)

	// 获取提醒检查器
	checker := global.GetRemindChecker()
	if checker == nil {
		c.JSON(500, gin.H{"error": "Remind checker not initialized"})
		return
	}

	// 设置 SSE 响应头（必须在写入任何内容之前）
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.WriteHeaderNow()

	// 调用 SSE 处理器
	checker.SSEHandler(c.Writer, c.Request)
}

// Acknowledge 确认提醒（用户点击“我知道”）
func (a *Notepad) Acknowledge(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	type Req struct {
		Id uint `json:"id"`
	}
	var req Req
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	// 查询便签
	var notepad models.Notepad
	if err := global.Db.Where("user_id = ? AND id = ?", userInfo.ID, req.Id).First(&notepad).Error; err != nil {
		apiReturn.Error(c, "Not Found")
		return
	}

	// 检查状态是否为待确认
	if notepad.RemindStatus != 1 {
		apiReturn.Error(c, "Reminder is not in pending state")
		return
	}

	// 根据重复类型处理
	if notepad.RemindRepeat == "" || notepad.RemindRepeat == "none" {
		// 一次性提醒：标记为已结束
		global.Db.Model(&notepad).Update("remind_status", 2)
		log.Printf("[提醒确认] 一次性提醒已确认: ID=%d, Title=%s", notepad.ID, notepad.Title)
	} else {
		// 重复提醒：计算下一个周期并重置状态
		parseFormats := []string{
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
		}
		
		var remindTime time.Time
		parsed := false
		for _, format := range parseFormats {
			t, err := time.ParseInLocation(format, notepad.RemindTime, time.Local)
			if err == nil {
				remindTime = t
				parsed = true
				break
			}
		}
		
		if !parsed {
			apiReturn.Error(c, "Invalid remind time format")
			return
		}
		
		// 计算下一个周期时间（基于当前时间，避免历史遗留欠账导致连环弹窗）
		checker := global.GetRemindChecker()
		if checker == nil {
			apiReturn.Error(c, "Remind checker not initialized")
			return
		}
		
		now := time.Now()
		nextTime := checker.CalculateNextRemindTime(remindTime, notepad.RemindRepeat, now)
		
		// 如果有提前天数，需要减去提前天数得到实际触发时间
		if notepad.RemindAdvanceDays > 0 {
			nextTime = nextTime.AddDate(0, 0, -notepad.RemindAdvanceDays)
		}
		
		// 更新数据库
		global.Db.Model(&notepad).Updates(map[string]interface{}{
			"remind_time": nextTime.Format("2006-01-02T15:04:05"),
			"remind_status": 0, // 重置为等待触发
		})
		
		log.Printf("[提醒确认] 重复提醒已确认，下次提醒: ID=%d, NextTime=%s", notepad.ID, nextTime.Format("2006-01-02T15:04:05"))
	}
	
	apiReturn.Success(c)
}
