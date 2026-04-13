package system

import (
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/models"

	"github.com/gin-gonic/gin"
)

type SystemSettingApi struct {
}

// 批量设置系统配置
func (s *SystemSettingApi) Set(c *gin.Context) {
	type Req struct {
		Settings map[string]interface{} `json:"settings" binding:"required"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	// 遍历设置项并保存
	systemSetting := models.SystemSetting{}
	for configName, configValue := range req.Settings {
		if err := systemSetting.Set(configName, configValue); err != nil {
			apiReturn.Error(c, "保存配置失败: "+err.Error())
			return
		}
	}

	apiReturn.Success(c)
}

// 批量获取系统配置
func (s *SystemSettingApi) Get(c *gin.Context) {
	type Req struct {
		ConfigNames []string `json:"configNames"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	systemSetting := models.SystemSetting{}
	result := make(map[string]string)

	// 如果没有指定配置名,返回所有配置
	if len(req.ConfigNames) == 0 {
		var allSettings []models.SystemSetting
		if err := models.Db.Find(&allSettings).Error; err != nil {
			apiReturn.Error(c, "获取配置失败: "+err.Error())
			return
		}
		for _, setting := range allSettings {
			result[setting.ConfigName] = setting.ConfigValue
		}
	} else {
		// 获取指定的配置
		for _, configName := range req.ConfigNames {
			value, err := systemSetting.Get(configName)
			if err != nil {
				// 如果配置不存在,跳过
				continue
			}
			result[configName] = value
		}
	}

	apiReturn.SuccessData(c, result)
}

// 获取单个系统配置
func (s *SystemSettingApi) GetSingle(c *gin.Context) {
	type Req struct {
		ConfigName string `json:"configName" binding:"required"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	systemSetting := models.SystemSetting{}
	value, err := systemSetting.Get(req.ConfigName)
	if err != nil {
		apiReturn.Error(c, "配置不存在")
		return
	}

	apiReturn.SuccessData(c, gin.H{
		"configName":  req.ConfigName,
		"configValue": value,
	})
}
