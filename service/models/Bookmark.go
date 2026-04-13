package models

type Bookmark struct {
	BaseModelNoUpdated        // 使用不包含UpdatedAt的基础模型
	IconJson           string `gorm:"type:text" json:"iconJson"`
	Title              string `gorm:"type:varchar(50)" json:"title"`
	Url                string `gorm:"type:varchar(1000)" json:"url"`
	LanUrl             string `gorm:"type:varchar(1000)" json:"lanUrl"`
	Sort               int    `gorm:"type:int(11)" json:"sort"`
	IsFolder           int    `gorm:"type:int(1)" json:"isFolder"`
	ParentUrl          string `gorm:"type:varchar(1000)" json:"parentUrl"`
	ParentId           uint   `gorm:"type:int(11)" json:"parentId"`
	UserId             uint   `json:"userId"`
	User               User   `json:"user"`
}
