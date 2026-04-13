package panel

import (
	"sun-panel/api/api_v1"
	"sun-panel/api/api_v1/middleware"

	"github.com/gin-gonic/gin"
)

// InitSearchEngine 初始化搜索引擎相关路由
func InitSearchEngine(router *gin.RouterGroup) {
	searchEngineApi := api_v1.ApiGroupApp.ApiPanel.SearchEngine

	// 需要登录的路由
	private := router.Group("", middleware.LoginInterceptor)
	{
		private.POST("/panel/searchEngine/add", searchEngineApi.Add)
		private.POST("/panel/searchEngine/update", searchEngineApi.Update)
		private.POST("/panel/searchEngine/delete", searchEngineApi.Delete)
		private.POST("/panel/searchEngine/updateSort", searchEngineApi.UpdateSort)
	}

	// 公开模式下可以获取搜索引擎列表
	public := router.Group("", middleware.PublicModeInterceptor)
	{
		public.POST("/panel/searchEngine/getList", searchEngineApi.GetList)
	}
}
