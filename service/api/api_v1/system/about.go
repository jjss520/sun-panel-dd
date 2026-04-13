package system

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/lib/cmn"

	"github.com/gin-gonic/gin"
)

type About struct {
}

func (a *About) Get(c *gin.Context) {
	version := cmn.GetSysVersionInfo()
	apiReturn.SuccessData(c, gin.H{
		"versionName": version.Version,
		"versionCode": version.Version_code,
	})
}

// CheckUpdate 检查更新
func (a *About) CheckUpdate(c *gin.Context) {
	// 获取当前版本
	currentVersion := cmn.GetSysVersionInfo()

	// 请求GitHub API获取最新版本信息
	versionUrl := "https://api.github.com/repos/jjss520/sun-panel-dd/releases/latest"

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", versionUrl, nil)
	if err != nil {
		apiReturn.Error(c, "创建请求失败: "+err.Error())
		return
	}
	req.Header.Set("User-Agent", "Sun-Panel-v2")

	resp, err := client.Do(req)
	if err != nil {
		apiReturn.Error(c, "检查更新失败: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		apiReturn.Error(c, fmt.Sprintf("检查更新失败: 返回状态码 %d", resp.StatusCode))
		return
	}

	// 解析JSON响应
	var releaseInfo struct {
		TagName string `json:"tag_name"`
		Body    string `json:"body"`
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		apiReturn.Error(c, "读取响应失败: "+err.Error())
		return
	}

	if err := json.Unmarshal(bodyBytes, &releaseInfo); err != nil {
		apiReturn.Error(c, "解析版本信息失败: "+err.Error())
		return
	}

	// 提取版本号
	latestVersionName := strings.TrimPrefix(releaseInfo.TagName, "v")

	// 使用API返回的更新日志
	updateLog := releaseInfo.Body
	if len(updateLog) > 500 {
		updateLog = updateLog[:500] + "..."
	}

	// 比较版本号
	hasNewVersion := false
	if latestVersionName > currentVersion.Version {
		hasNewVersion = true
	}

	apiReturn.SuccessData(c, gin.H{
		"hasNewVersion": hasNewVersion,
		"latestVersion": latestVersionName,
		"updateContent": updateLog,
	})
}
