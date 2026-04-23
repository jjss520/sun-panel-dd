package models

type Notepad struct {
	BaseModel
	UserID       uint   `json:"userId" gorm:"index;comment:用户ID"`
	Title        string `json:"title" gorm:"type:varchar(255);comment:标题"`
	Content      string `json:"content" gorm:"type:text;comment:便签内容"`
	RemindBaseTime string `json:"remindBaseTime,omitempty" gorm:"type:varchar(50);comment:提醒基准时间(用户原始选择的时间)"`
	RemindTime   string `json:"remindTime,omitempty" gorm:"type:varchar(50);comment:提醒时间(实际触发时间，已计算周期和提前天数)"`
	RemindStatus int    `json:"remindStatus" gorm:"default:0;comment:提醒状态 0=等待触发 1=待确认 2=已结束"`
	RemindRepeat string `json:"remindRepeat,omitempty" gorm:"type:varchar(20);default:none;comment:重复类型 none=不重复 daily=每天 weekly=每周 monthly=每月 yearly=每年"`
	RemindAdvanceDays int `json:"remindAdvanceDays" gorm:"default:0;comment:提前提醒天数（0=不提前）"`
}
