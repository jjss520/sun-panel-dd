package models

import (
	"sun-panel/models/datatype"

	"gorm.io/gorm"
)

type ItemIcon struct {
	BaseModel
	IconJson        string                    `gorm:"type:varchar(1000)" json:"-"`
	Icon            datatype.ItemIconIconInfo `gorm:"-" json:"icon"`
	Title           string                    `gorm:"type:varchar(50)" json:"title"`
	Url             string                    `gorm:"type:varchar(1000)" json:"url"`
	LanUrl          string                    `gorm:"type:varchar(1000)" json:"lanUrl"`
	Description     string                    `gorm:"type:varchar(1000)" json:"description"`
	OpenMethod      int                       `gorm:"type:tinyint(1)" json:"openMethod"`
	Sort            int                       `gorm:"type:int(11)" json:"sort"`
	ItemIconGroupId int                       `json:"itemIconGroupId"`
	ParentId        int                       `gorm:"type:int(11);default:0;index" json:"parentId"` // 所属文件夹ID，0 表示直接挂在分组下
	IsFolder        int                       `gorm:"type:tinyint(1);default:0" json:"isFolder"`    // 是否为文件夹 0:否 1:是
	LanOnly         int                       `gorm:"type:tinyint(1);default:0" json:"lanOnly"`     // 仅内网展示 0:关闭 1:开启
	UserId          uint                      `json:"userId"`
	User            User                      `json:"user"`
}

func (m *ItemIcon) DeleteByItemIconGroupIds(db *gorm.DB, userId uint, itemIconGroupIds []uint) (err error) {
	err = db.Delete(&ItemIcon{}, "item_icon_group_id in ? AND user_id=?", itemIconGroupIds, userId).Error
	return
}

func (m *ItemIcon) DeleteByUserId(db *gorm.DB, userId uint) (err error) {
	return db.Delete(&ItemIcon{}, "user_id=?", userId).Error
}

// GetAllChildrenIds 递归收集给定图标下所有子孙的ID（用于删除文件夹时级联清理）
func (m *ItemIcon) GetAllChildrenIds(db *gorm.DB, userId uint, ids []uint) (all []uint, err error) {
	if len(ids) == 0 {
		return
	}

	seen := make(map[uint]bool, len(ids))
	queue := make([]uint, 0, len(ids))
	for _, id := range ids {
		seen[id] = true
		queue = append(queue, id)
	}

	for len(queue) > 0 {
		children := []ItemIcon{}
		if err = db.Select("id").Where("user_id=? AND parent_id IN ?", userId, queue).Find(&children).Error; err != nil {
			return
		}

		next := make([]uint, 0, len(children))
		for _, c := range children {
			if seen[c.ID] {
				continue
			}
			seen[c.ID] = true
			all = append(all, c.ID)
			next = append(next, c.ID)
		}
		queue = next
	}

	return
}
