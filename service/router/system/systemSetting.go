package system

import (
	"sun-panel/api/api_v1"

	"github.com/gin-gonic/gin"
)

func InitSystemSettingRouter(router *gin.RouterGroup) {
	systemSettingApi := api_v1.ApiGroupApp.ApiSystem.SystemSettingApi
	settingGroup := router.Group("system/setting")
	{
		settingGroup.POST("set", systemSettingApi.Set)
		settingGroup.POST("get", systemSettingApi.Get)
		settingGroup.POST("getSingle", systemSettingApi.GetSingle)
	}
}
