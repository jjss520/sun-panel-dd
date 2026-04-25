package system

import (
	"sun-panel/api/api_v1"
	"sun-panel/api/api_v1/middleware"

	"github.com/gin-gonic/gin"
)

func InitFileRouter(router *gin.RouterGroup) {
	FileApi := api_v1.ApiGroupApp.ApiSystem.FileApi

	// 验证项目的权限(有访问密码的需要验证访问token)
	private := router.Group("", middleware.LoginInterceptor)
	{
		private.POST("/file/uploadImg", FileApi.UploadImg)           // 原有：图标/通用图片（base64）
		private.POST("/file/uploadWallpaper", FileApi.UploadWallpaper) // 新增：壁纸（文件系统）
		private.POST("/file/uploadFiles", FileApi.UploadFiles)       // 原有：便签文件

		private.POST("/file/getList", FileApi.GetList)
		private.POST("/file/deletes", FileApi.Deletes)

	}

}
