package system

import (
	"sun-panel/api/api_v1/common/apiReturn"

	"github.com/gin-gonic/gin"
)

type Ping struct {
}

// Ping接口 - 用于网络连通性检测
// 只返回成功状态,用于判断服务是否可达
func (p *Ping) Get(c *gin.Context) {
	apiReturn.Success(c)
}
