package models

type Notepad struct {
	BaseModel
	UserID       uint   `json:"userId" gorm:"index;comment:用户ID"`
	Title        string `json:"title" gorm:"type:varchar(255);comment:标题"`
	Content      string `json:"content" gorm:"type:text;comment:便签内容"`
	RemindTime   string `json:"remindTime,omitempty" gorm:"type:varchar(50);comment:提醒时间(本地时间格式)"`
	RemindStatus int    `json:"remindStatus" gorm:"default:0;comment:提醒状态 0=未提醒 1=已提醒"`
	RemindRepeat string `json:"remindRepeat,omitempty" gorm:"type:varchar(20);default:none;comment:重复类型 none=不重复 daily=每天 weekly=每周 monthly=每月 yearly=每年"`
	RemindForce  int    `json:"remindForce" gorm:"default:0;comment:强制提醒 0=关闭 1=开启"`
	RemindAdvanceDays int `json:"remindAdvanceDays" gorm:"default:0;comment:提前提醒天数（0=不提前）"`
}
