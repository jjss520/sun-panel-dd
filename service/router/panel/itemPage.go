package panel

import (
	"sun-panel/api/api_v1"
	"sun-panel/api/api_v1/middleware"

	"github.com/gin-gonic/gin"
)

func InitItemPage(router *gin.RouterGroup) {
	itemPage := api_v1.ApiGroupApp.ApiPanel.ItemPage
	r := router.Group("", middleware.LoginInterceptor)
	{
		r.POST("/panel/itemPage/edit", itemPage.Edit)
		r.POST("/panel/itemPage/deletes", itemPage.Deletes)
		r.POST("/panel/itemPage/saveSort", itemPage.SaveSort)
	}

	// 公开模式
	rPublic := router.Group("", middleware.PublicModeInterceptor)
	{
		rPublic.POST("/panel/itemPage/getList", itemPage.GetList)
	}
}
