package panelApiStructs

import (
	"sun-panel/api/api_v1/common/apiData/commonApiStructs"
	"sun-panel/models"
)

type ItemIconEditRequest struct {
	models.ItemIcon
	IconJson string
}

type ItemIconSaveSortRequest struct {
	SortItems       []commonApiStructs.SortRequestItem `json:"sortItems"`
	ItemIconGroupId uint                               `json:"itemIconGroupId"`
}

type ItemIconGetSiteFaviconReq struct {
	Url string `json:"url"`
}

type ItemIconGetSiteFaviconResp struct {
	IconUrl string `json:"iconUrl"`
}

// ItemIconMoveRequest 图标移动到文件夹
type ItemIconMoveRequest struct {
	Id       uint `json:"id"`
	ParentId int  `json:"parentId"` // 目标文件夹ID，0=分组根目录
}
