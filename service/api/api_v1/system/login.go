package system

import (
	"strconv"
	"strings"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/lib/cmn"
	"sun-panel/lib/cmn/systemSetting"
	"sun-panel/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoginApi struct {
}

// 登录输入验证
type LoginLoginVerify struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,max=50"`
	VCode    string `json:"vcode" validate:"max=6"`
	Email    string `json:"email"`
}

// @Summary 登录账号
// @Accept application/json
// @Produce application/json
// @Param LoginLoginVerify body LoginLoginVerify true "登陆验证信息"
// @Tags user
// @Router /login [post]
func (l LoginApi) Login(c *gin.Context) {
	param := LoginLoginVerify{}
	if err := c.ShouldBindJSON(&param); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	if errMsg, err := base.ValidateInputStruct(param); err != nil {
		apiReturn.ErrorParamFomat(c, errMsg)
		return
	}

	settings := systemSetting.ApplicationSetting{}
	global.SystemSetting.GetValueByInterface("system_application", &settings)

	mUser := models.User{}
	var (
		err  error
		info models.User
	)
	bToken := ""
	param.Username = strings.TrimSpace(param.Username)
	global.Logger.Infof("DEBUG LOGIN: Received Username=[%s], Password=[%s], PasswordLength=%d, TargetHash=[%s]", param.Username, param.Password, len(param.Password), cmn.PasswordEncryption(param.Password))

	// Check if user exists at all
	var userCheck models.User
	if err := global.Db.Where("username = ?", param.Username).First(&userCheck).Error; err != nil {
		global.Logger.Errorf("DEBUG LOGIN: Username [%s] not found in DB. Error: %v", param.Username, err)
	} else {
		global.Logger.Infof("DEBUG LOGIN: User [%s] found in DB. StoredHash=[%s]", param.Username, userCheck.Password)
	}

	if info, err = mUser.GetUserInfoByUsernameAndPassword(param.Username, cmn.PasswordEncryption(param.Password)); err != nil {
		// 未找到记录 账号或密码错误
		if err == gorm.ErrRecordNotFound {
			global.Logger.Warnf("DEBUG LOGIN: Login failed for [%s] - RECORD NOT FOUND with password index match.", param.Username)
			apiReturn.ErrorByCode(c, 1003)
			return
		} else {
			// 未知错误
			global.Logger.Errorf("DEBUG LOGIN: Database Error for [%s]: %v", param.Username, err)
			apiReturn.Error(c, err.Error())
			return
		}

	}

	global.Logger.Infof("DEBUG LOGIN: GetUserInfo SUCCESS. Found ID=[%d], Status=[%d]", info.ID, info.Status)

	// 停用或未激活
	if info.Status != 1 {
		global.Logger.Warnf("DEBUG LOGIN: User [%s] is DISABLED or INACTIVE. Status=[%d]. Returning 1004.", param.Username, info.Status)
		apiReturn.ErrorByCode(c, 1004)
		return
	}

	bToken = info.Token
	if info.Token == "" {
		// 生成token
		buildTokenOver := false
		for !buildTokenOver {
			bToken = cmn.BuildRandCode(32, cmn.RAND_CODE_MODE2)
			if _, err := mUser.GetUserInfoByToken(bToken); err != nil {
				// 保存token
				mUser.UpdateUserInfoByUserId(info.ID, map[string]interface{}{
					"token": bToken,
				})
				buildTokenOver = true
			}
		}
		info.Token = bToken
	}
	info.Password = ""
	info.ReferralCode = ""

	// global.UserToken.SetDefault(bToken, info)
	cToken := uuid.NewString() + "-" + cmn.Md5(cmn.Md5("userId"+strconv.Itoa(int(info.ID))))
	global.CUserToken.SetDefault(cToken, bToken)
	global.Logger.Debug("token:", cToken, "|", bToken)
	global.Logger.Debug(global.CUserToken.Get(cToken))

	// 设置当前用户信息
	c.Set("userInfo", info)
	info.Token = cToken // 重要 采用cToken,隐藏真实token
	global.Logger.Infof("DEBUG LOGIN: ALL STEPS COMPLETE. Returning SuccessData for [%s]. Token=[%s...]", param.Username, cToken[:10])
	apiReturn.SuccessData(c, info)
}

// 安全退出
func (l *LoginApi) Logout(c *gin.Context) {
	// userInfo, _ := base.GetCurrentUserInfo(c)
	cToken := c.GetHeader("token")
	global.CUserToken.Delete(cToken)
	apiReturn.Success(c)
}
