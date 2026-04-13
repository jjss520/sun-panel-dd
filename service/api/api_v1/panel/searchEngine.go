package panel

import (
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/models"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// SearchEngine 搜索引擎管理API
type SearchEngine struct {
}

// GetList 获取搜索引擎列表
func (a *SearchEngine) GetList(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	searchEngines := make([]models.SearchEngine, 0)

	// 按sort字段升序排序
	if err := global.Db.Where("user_id = ?", userInfo.ID).Order("sort ASC").Find(&searchEngines).Error; err != nil {
		apiReturn.Error(c, "获取搜索引擎列表失败")
		return
	}

	apiReturn.ListData(c, searchEngines, int64(len(searchEngines)))
}

// Add 添加搜索引擎
func (a *SearchEngine) Add(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	var req models.SearchEngine

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	// 设置用户ID
	req.UserId = userInfo.ID

	// 获取最大sort值并+1作为新的sort值
	var maxSort int
	query := global.Db.Model(&models.SearchEngine{}).Where("user_id = ?", userInfo.ID)
	query.Select("COALESCE(MAX(sort), 0) as max_sort").Scan(&maxSort)

	// 设置新的sort值
	req.Sort = maxSort + 1

	// 插入数据库
	if err := global.Db.Create(&req).Error; err != nil {
		apiReturn.Error(c, "添加搜索引擎失败")
		return
	}

	apiReturn.SuccessData(c, req)
}

// Update 更新搜索引擎
func (a *SearchEngine) Update(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	type UpdateReq struct {
		ID      uint   `json:"id" binding:"required"`
		IconSrc string `json:"iconSrc"`
		Title   string `json:"title" binding:"required"`
		Url     string `json:"url" binding:"required"`
		Sort    int    `json:"sort"`
	}
	var req UpdateReq

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	// 检查搜索引擎是否存在且属于当前用户
	var searchEngine models.SearchEngine
	if err := global.Db.Where("id = ? AND user_id = ?", req.ID, userInfo.ID).First(&searchEngine).Error; err != nil {
		apiReturn.Error(c, "搜索引擎不存在或无权修改")
		return
	}

	// 准备更新数据
	updateData := map[string]interface{}{
		"IconSrc": req.IconSrc,
		"Title":   req.Title,
		"Url":     req.Url,
		"Sort":    req.Sort,
	}

	// 保存更新
	if err := global.Db.Model(&models.SearchEngine{}).Where("id = ? AND user_id = ?", searchEngine.ID, userInfo.ID).Updates(updateData).Error; err != nil {
		apiReturn.Error(c, "修改搜索引擎失败")
		return
	}

	// 查询更新后的搜索引擎信息
	updatedSearchEngine := models.SearchEngine{}
	if err := global.Db.Where("id = ? AND user_id = ?", req.ID, userInfo.ID).First(&updatedSearchEngine).Error; err != nil {
		apiReturn.Error(c, "查询更新后的搜索引擎失败")
		return
	}

	apiReturn.SuccessData(c, updatedSearchEngine)
}

// Delete 删除搜索引擎
func (a *SearchEngine) Delete(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	type DeleteReq struct {
		ID uint `json:"id" binding:"required"`
	}
	var req DeleteReq

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	// 检查搜索引擎是否存在且属于当前用户
	var searchEngine models.SearchEngine
	if err := global.Db.Where("id = ? AND user_id = ?", req.ID, userInfo.ID).First(&searchEngine).Error; err != nil {
		apiReturn.Error(c, "搜索引擎不存在或无权删除")
		return
	}

	// 删除搜索引擎
	if err := global.Db.Where("id = ? AND user_id = ?", req.ID, userInfo.ID).Delete(&models.SearchEngine{}).Error; err != nil {
		apiReturn.Error(c, "删除搜索引擎失败")
		return
	}

	apiReturn.Success(c)
}

// UpdateSort 批量更新搜索引擎排序
func (a *SearchEngine) UpdateSort(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	type SortItem struct {
		ID   uint `json:"id" binding:"required"`
		Sort int  `json:"sort" binding:"required"`
	}
	type UpdateSortReq struct {
		Items []SortItem `json:"items" binding:"required"`
	}
	var req UpdateSortReq

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	// 批量更新排序
	for _, item := range req.Items {
		// 检查搜索引擎是否属于当前用户
		var searchEngine models.SearchEngine
		if err := global.Db.Where("id = ? AND user_id = ?", item.ID, userInfo.ID).First(&searchEngine).Error; err != nil {
			continue // 跳过不属于当前用户的搜索引擎
		}

		// 更新排序
		global.Db.Model(&models.SearchEngine{}).Where("id = ? AND user_id = ?", item.ID, userInfo.ID).Update("sort", item.Sort)
	}

	apiReturn.Success(c)
}
