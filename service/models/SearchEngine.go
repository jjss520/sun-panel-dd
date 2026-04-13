package models

import (
	"gorm.io/gorm"
)

type SearchEngine struct {
	BaseModel
	IconSrc   string         `gorm:"type:longtext" json:"iconSrc"`
	Title     string         `gorm:"type:varchar(50)" json:"title"`
	Url       string         `gorm:"type:varchar(1000)" json:"url"`
	Sort      int            `gorm:"type:int(11)" json:"sort"`
	UserId    uint           `json:"userId"`
	User      User           `json:"user"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
