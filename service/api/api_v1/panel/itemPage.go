package panel

import (
	"fmt"
	"math"
	"sun-panel/api/api_v1/common/apiData/commonApiStructs"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/models"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type ItemPage struct {
}

// Edit 创建或编辑页面
func (a *ItemPage) Edit(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	req := models.ItemPage{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	req.UserId = userInfo.ID

	if req.ID != 0 {
		// 修改
		updateField := []string{"Title", "Icon", "Sort"}
		global.Db.Model(&models.ItemPage{}).
			Select(updateField).
			Where("id=? AND user_id=?", req.ID, userInfo.ID).Updates(&req)
	} else {
		// 创建
		global.Db.Create(&req)
	}

	apiReturn.SuccessData(c, req)
}

// GetList 获取用户的页面列表
func (a *ItemPage) GetList(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	pages := []models.ItemPage{}

	err := global.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Order("sort ASC, created_at ASC").Where("user_id=?", userInfo.ID).Find(&pages).Error; err != nil {
			apiReturn.ErrorDatabase(c, err.Error())
			return err
		}

		// 判断页面是否为空，为空将自动创建默认页面
		if len(pages) == 0 {
			defaultPage := models.ItemPage{
				Title:  "首页",
				UserId: userInfo.ID,
				Icon:   "material-symbols:home-outline",
				Sort:   1,
			}
			if err := tx.Create(&defaultPage).Error; err != nil {
				apiReturn.ErrorDatabase(c, err.Error())
				return err
			}

			// 并将当前账号下所有无页面的分组更新到当前页
			if err := tx.Model(&models.ItemIconGroup{}).Where("user_id=? AND page_id IS NULL", userInfo.ID).Update("page_id", defaultPage.ID).Error; err != nil {
				apiReturn.ErrorDatabase(c, err.Error())
				return err
			}

			pages = append(pages, defaultPage)
		}

		// 返回 nil 提交事务
		return nil
	})

	if err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	} else {
		apiReturn.SuccessListData(c, pages, 0)
	}
}

// Deletes 删除页面
func (a *ItemPage) Deletes(c *gin.Context) {
	req := commonApiStructs.RequestDeleteIds[uint]{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	userInfo, _ := base.GetCurrentUserInfo(c)

	var count int64
	if err := global.Db.Model(&models.ItemPage{}).Where("user_id=?", userInfo.ID).Count(&count).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	} else {
		if math.Abs(float64(len(req.Ids))-float64(count)) < 1 {
			apiReturn.ErrorCode(c, 1201, "至少要保留一个页面", nil)
			return
		}
	}

	// 事务：删除页面前先将分组转移到第一个可用页面
	err := global.Db.Transaction(func(tx *gorm.DB) error {
		// 获取第一个可用页面ID
		var firstPage models.ItemPage
		if err := tx.Order("sort ASC, created_at ASC").Where("user_id=? AND id NOT IN ?", userInfo.ID, req.Ids).First(&firstPage).Error; err != nil {
			return fmt.Errorf("无法找到可用的页面来转移分组: %v", err)
		}

		// 将要删除的页面上的分组转移到第一个页面
		if err := tx.Model(&models.ItemIconGroup{}).
			Where("user_id=? AND page_id IN ?", userInfo.ID, req.Ids).
			Update("page_id", firstPage.ID).Error; err != nil {
			return err
		}

		// 删除页面
		if err := tx.Where("id IN ? AND user_id=?", req.Ids, userInfo.ID).Delete(&models.ItemPage{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	apiReturn.Success(c)
}

// SaveSort 保存页面排序
func (a *ItemPage) SaveSort(c *gin.Context) {
	req := commonApiStructs.SortRequest{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	userInfo, _ := base.GetCurrentUserInfo(c)

	for _, item := range req.SortItems {
		global.Db.Model(&models.ItemPage{}).
			Where("id=? AND user_id=?", item.Id, userInfo.ID).
			Update("sort", item.Sort)
	}

	apiReturn.Success(c)
}
