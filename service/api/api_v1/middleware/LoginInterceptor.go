package middleware

import (
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/global"
	"sun-panel/models"

	"github.com/gin-gonic/gin"
)

func LoginInterceptor(c *gin.Context) {

	// 继续执行后续的操作，再回来
	// c.Next()

	// 获得token
	cToken := c.GetHeader("token")

	// 没有token信息视为未登录
	if cToken == "" {
		apiReturn.ErrorByCode(c, 1000)
		c.Abort() // 终止执行后续的操作，一般配合return使用
		return
	}

	token := ""
	{
		var ok bool
		token, ok = global.CUserToken.Get(cToken)
		// 可能已经安全退出或者很久没有使用已过期
		if !ok || token == "" {
			// 尝试把客户端提供的 token 当作真实存库的 token(bToken)去数据库查找
			mUser := models.User{}
			if info, err := mUser.GetUserInfoByToken(cToken); err == nil && info.ID != 0 && info.Token != "" {
				// 找到了，说明客户端直接带的是 bToken，建立缓存映射并继续
				token = info.Token
				global.UserToken.SetDefault(token, info)
				global.CUserToken.SetDefault(cToken, token)
				c.Set("userInfo", info)
				return
			}

			// 否则视为登录过期
			apiReturn.ErrorByCode(c, 1001)
			c.Abort() // 终止执行后续的操作，一般配合return使用
			return
		}
	}

	// 直接返回缓存的用户信息
	if userInfo, success := global.UserToken.Get(token); success {
		c.Set("userInfo", userInfo)
		return
	}

	global.Logger.Debug("准备查询数据库的用户资料", token)

	mUser := models.User{}
	// 去库中查询是否存在该用户；否则返回错误
	if info, err := mUser.GetUserInfoByToken(token); err != nil || info.Token == "" || info.ID == 0 {
		apiReturn.ErrorCode(c, 1001, global.Lang.Get("login.err_token_expire"), nil)
		c.Abort()
		return
	} else {
		// 通过 设置当前用户信息
		global.UserToken.SetDefault(info.Token, info)
		global.CUserToken.SetDefault(cToken, token)
		c.Set("userInfo", info)
	}

}

// 不验证缓存直接验证库省去没有缓存每次都要手动登录的问题
func LoginInterceptorDev(c *gin.Context) {

	// 获得token
	token := c.GetHeader("token")
	mUser := models.User{}

	// 去库中查询是否存在该用户；否则返回错误
	if info, err := mUser.GetUserInfoByToken(token); err != nil || info.ID == 0 {
		apiReturn.ErrorCode(c, 1001, global.Lang.Get("login.err_token_expire"), nil)
		c.Abort()
		return
	} else {
		// 通过
		// 设置当前用户信息
		c.Set("userInfo", info)
	}
}
