package panel

import (
	"sun-panel/api/api_v1/middleware"
	"sun-panel/api/api_v1/panel"

	"github.com/gin-gonic/gin"
)

func InitNotepad(router *gin.RouterGroup) {
	notepad := panel.Notepad{}
	// 使用登录拦截器
	r := router.Group("", middleware.LoginInterceptor)
	{
		r.GET("panel/notepad/get", notepad.Get)
		r.GET("panel/notepad/getList", notepad.GetList)
		r.POST("panel/notepad/save", notepad.Save)
		r.POST("panel/notepad/delete", notepad.Delete)
		r.POST("panel/notepad/upload", notepad.Upload)
	}
}
