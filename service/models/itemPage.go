package models

import (
	"gorm.io/gorm"
)

type ItemPage struct {
	BaseModel
	Title    string `gorm:"column:title;type:varchar(100);default:'页面'" json:"title"`
	Icon     string `gorm:"column:icon;type:varchar(50);default:'material-symbols:home-outline'" json:"icon"`
	Sort     int    `gorm:"column:sort;default:9999" json:"sort"`
	UserId   uint   `gorm:"column:user_id;index" json:"userId"`
}

func (ItemPage) TableName() string {
	return "item_pages"
}

// GetListByUserId 获取用户的页面列表
func (p *ItemPage) GetListByUserId(db *gorm.DB, userId uint) ([]ItemPage, error) {
	var pages []ItemPage
	err := db.Order("sort ASC, created_at ASC").Where("user_id = ?", userId).Find(&pages).Error
	return pages, err
}

// Create 创建页面
func (p *ItemPage) Create(db *gorm.DB) error {
	return db.Create(p).Error
}

// Update 更新页面
func (p *ItemPage) Update(db *gorm.DB) error {
	return db.Model(p).Select("title", "icon", "sort").Updates(p).Error
}

// Delete 删除页面
func (p *ItemPage) Delete(db *gorm.DB) error {
	return db.Delete(p).Error
}
