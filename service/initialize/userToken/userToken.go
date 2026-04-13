package userToken

import (
	"sun-panel/global"
	"sun-panel/lib/cache"
	"sun-panel/models"

	"time"
)

func InitUserToken() cache.Cacher[models.User] {
	// 设置为0表示永不过期
	return global.NewCache[models.User](0*time.Second, 1*time.Hour, "UserToken")
}

// func InitVerifyCodeCachePool() {
// 	global.VerifyCodeCachePool = cache.NewGoCache(10*time.Minute, 60*time.Second)
// }
