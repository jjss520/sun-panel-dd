package models

type Notepad struct {
	BaseModel
	UserID  uint   `json:"userId" gorm:"index;comment:用户ID"`
	Title   string `json:"title" gorm:"type:varchar(255);comment:标题"`
	Content string `json:"content" gorm:"type:text;comment:便签内容"`
}
