package panel

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sun-panel/api/api_v1/common/apiData/commonApiStructs"
	"sun-panel/api/api_v1/common/apiData/panelApiStructs"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/lib/siteFavicon"
	"sun-panel/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type ItemIcon struct {
}

func (a *ItemIcon) Edit(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	req := models.ItemIcon{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	if req.ItemIconGroupId == 0 {
		// apiReturn.Error(c, "Group is mandatory")
		apiReturn.ErrorParamFomat(c, "Group is mandatory")
		return
	}

	req.UserId = userInfo.ID

	// json转字符串
	if j, err := json.Marshal(req.Icon); err == nil {
		req.IconJson = string(j)
	}

	if req.ID != 0 {
		// 修改
		updateField := []string{"IconJson", "Icon", "Title", "Url", "LanUrl", "Description", "OpenMethod", "GroupId", "UserId", "ItemIconGroupId", "LanOnly"}
		if req.Sort != 0 {
			updateField = append(updateField, "Sort")
		}
		global.Db.Model(&models.ItemIcon{}).
			Select(updateField).
			Where("id=?", req.ID).Updates(&req)
	} else {
		req.Sort = 9999
		// 创建
		global.Db.Create(&req)
	}

	apiReturn.SuccessData(c, req)
}

// 添加多个图标
func (a *ItemIcon) AddMultiple(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	// type Request
	req := []models.ItemIcon{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	for i := 0; i < len(req); i++ {
		if req[i].ItemIconGroupId == 0 {
			apiReturn.ErrorParamFomat(c, "Group is mandatory")
			return
		}
		req[i].UserId = userInfo.ID
		// json转字符串
		if j, err := json.Marshal(req[i].Icon); err == nil {
			req[i].IconJson = string(j)
		}
	}

	global.Db.Create(&req)

	apiReturn.SuccessData(c, req)
}

// // 获取详情
// func (a *ItemIcon) GetInfo(c *gin.Context) {
// 	req := systemApiStructs.AiDrawGetInfoReq{}

// 	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
// 		apiReturn.ErrorParamFomat(c, err.Error())
// 		return
// 	}

// 	userInfo, _ := base.GetCurrentUserInfo(c)

// 	aiDraw := models.AiDraw{}
// 	aiDraw.ID = req.ID
// 	if err := aiDraw.GetInfo(global.Db); err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			apiReturn.Error(c, "不存在记录")
// 			return
// 		}
// 		apiReturn.ErrorDatabase(c, err.Error())
// 		return
// 	}

// 	if userInfo.ID != aiDraw.UserID {
// 		apiReturn.ErrorNoAccess(c)
// 		return
// 	}

// 	apiReturn.SuccessData(c, aiDraw)
// }

func (a *ItemIcon) GetListByGroupId(c *gin.Context) {
	req := models.ItemIcon{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	userInfo, _ := base.GetCurrentUserInfo(c)
	itemIcons := []models.ItemIcon{}

	if err := global.Db.Order("sort ,created_at").Find(&itemIcons, "item_icon_group_id = ? AND user_id=?", req.ItemIconGroupId, userInfo.ID).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	for k, v := range itemIcons {
		json.Unmarshal([]byte(v.IconJson), &itemIcons[k].Icon)
	}

	apiReturn.SuccessListData(c, itemIcons, 0)
}

func (a *ItemIcon) Deletes(c *gin.Context) {
	req := commonApiStructs.RequestDeleteIds[uint]{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	userInfo, _ := base.GetCurrentUserInfo(c)
	if err := global.Db.Delete(&models.ItemIcon{}, "id in ? AND user_id=?", req.Ids, userInfo.ID).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	apiReturn.Success(c)
}

// 保存排序
func (a *ItemIcon) SaveSort(c *gin.Context) {
	req := panelApiStructs.ItemIconSaveSortRequest{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	userInfo, _ := base.GetCurrentUserInfo(c)

	transactionErr := global.Db.Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		for _, v := range req.SortItems {
			if err := tx.Model(&models.ItemIcon{}).Where("user_id=? AND id=? AND item_icon_group_id=?", userInfo.ID, v.Id, req.ItemIconGroupId).Update("sort", v.Sort).Error; err != nil {
				// 返回任何错误都会回滚事务
				return err
			}
		}

		// 返回 nil 提交事务
		return nil
	})

	if transactionErr != nil {
		apiReturn.ErrorDatabase(c, transactionErr.Error())
		return
	}

	apiReturn.Success(c)
}

// 将远程图片转换为base64
func getImageBase64(imageUrl string) (string, error) {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second, // 10秒超时
	}

	// 发送HTTP GET请求
	resp, err := client.Get(imageUrl)
	if err != nil {
		return "", fmt.Errorf("failed to get image: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get image, status code: %d", resp.StatusCode)
	}

	// 读取响应体
	buffer := bytes.Buffer{}
	if _, err := io.Copy(&buffer, resp.Body); err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	// 获取Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png" // 默认使用PNG
	}

	// 转换为base64
	base64Str := base64.StdEncoding.EncodeToString(buffer.Bytes())

	// 返回data URL格式
	return fmt.Sprintf("data:%s;base64,%s", contentType, base64Str), nil
}

// 支持获取并直接返回base64格式的图标
func (a *ItemIcon) GetSiteFavicon(c *gin.Context) {
	req := panelApiStructs.ItemIconGetSiteFaviconReq{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	resp := panelApiStructs.ItemIconGetSiteFaviconResp{}

	// 解析URL获取域名和协议，只解析一次
	parsedURL, parseErr := url.Parse(req.Url)
	if parseErr != nil {
		apiReturn.Error(c, "acquisition failed:"+parseErr.Error())
		return
	}

	var fullUrl string

	// 1. 首先尝试从网站本身获取favicon
	iconUrl, err := siteFavicon.GetOneFaviconURL(req.Url)
	if err != nil {
		// 2. 如果失败，使用谷歌的favicon服务作为回退
		global.Logger.Debug("Failed to get favicon from site, trying Google API:", err)

		// 使用谷歌的favicon服务作为回退
		domain := parsedURL.Host
		fullUrl = fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=64", domain)
		global.Logger.Debug("Using Google favicon service for domain:", domain)
	} else {
		// 处理获取到的favicon URL
		if strings.HasPrefix(iconUrl, "//") {
			// 协议相对URL，只添加当前协议
			fullUrl = parsedURL.Scheme + ":" + iconUrl
			global.Logger.Debug("Protocol relative URL, converted to:", fullUrl)
		} else if strings.HasPrefix(iconUrl, "/") {
			// 绝对路径，需要添加完整域名
			fullUrl = parsedURL.Scheme + "://" + parsedURL.Host + iconUrl
			global.Logger.Debug("Absolute path, converted to:", fullUrl)
		} else if strings.HasPrefix(iconUrl, "http://") || strings.HasPrefix(iconUrl, "https://") {
			// 完整URL，直接使用
			fullUrl = iconUrl
			global.Logger.Debug("Full URL, using directly:", fullUrl)
		} else {
			// 相对路径，需要添加完整域名和路径前缀
			basePath := parsedURL.Path
			if !strings.HasSuffix(basePath, "/") {
				// 获取目录部分
				lastSlash := strings.LastIndex(basePath, "/")
				if lastSlash != -1 {
					basePath = basePath[:lastSlash+1]
				} else {
					basePath = "/"
				}
			}
			fullUrl = parsedURL.Scheme + "://" + parsedURL.Host + basePath + iconUrl
			global.Logger.Debug("Relative path, converted to:", fullUrl)
		}
	}

	// 移除URL中的查询参数，因为有些图标服务器不支持
	if strings.Contains(fullUrl, "?") {
		fullUrl = strings.Split(fullUrl, "?")[0]
		global.Logger.Debug("Removed query params, final URL:", fullUrl)
	}

	// 获取base64格式的图标
	base64Icon, getErr := getImageBase64(fullUrl)
	if getErr != nil {
		// 如果获取失败，使用谷歌的favicon服务作为最终回退
		global.Logger.Debug("Failed to get favicon from URL, trying Google API as final fallback:", getErr)
		domain := parsedURL.Host
		googleUrl := fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=64", domain)
		base64Icon, getErr = getImageBase64(googleUrl)
		if getErr != nil {
			apiReturn.Error(c, "acquisition failed:"+getErr.Error())
			return
		}
	}

	// 返回base64格式的图标
	resp.IconUrl = base64Icon
	apiReturn.SuccessData(c, resp)
}
