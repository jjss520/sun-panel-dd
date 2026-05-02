package system

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
	"io"
	"os"
	"path"
	"strings"
	"sun-panel/api/api_v1/common/apiData/commonApiStructs"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/lib/cmn"
	"sun-panel/models"
	"time"
)

type FileApi struct{}

func (a *FileApi) UploadImg(c *gin.Context) {
	// 获取上传的文件
	f, err := c.FormFile("imgfile")
	if err != nil {
		apiReturn.ErrorByCode(c, 1300)
		return
	}

	// 检查文件类型
	fileExt := strings.ToLower(path.Ext(f.Filename))
	agreeExts := []string{
		".png",
		".jpg",
		".gif",
		".jpeg",
		".webp",
		".svg",
		".ico",
	}

	if !strings.Contains(strings.Join(agreeExts, "|"), fileExt) {
		apiReturn.ErrorByCode(c, 1301)
		return
	}

	// 打开上传的文件
	file, err := f.Open()
	if err != nil {
		apiReturn.ErrorByCode(c, 1300)
		return
	}
	defer file.Close()

	// 读取文件内容
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		apiReturn.ErrorByCode(c, 1300)
		return
	}

	// 获取文件的MIME类型
	contentType := "image/png" // 默认使用PNG
	switch fileExt {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	case ".svg":
		contentType = "image/svg+xml"
	case ".ico":
		contentType = "image/x-icon"
	}

	// 转换为base64
	base64Str := base64.StdEncoding.EncodeToString(buffer.Bytes())

	// 返回data URL格式
	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, base64Str)

	// 直接返回base64数据
	apiReturn.SuccessData(c, gin.H{
		"imageUrl": dataURL,
	})
}

// UploadWallpaper 上传壁纸（保存到文件系统）
func (a *FileApi) UploadWallpaper(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	configUpload := global.Config.GetValueStringOrDefault("base", "source_path")

	// 获取上传的文件
	f, err := c.FormFile("imgfile")
	if err != nil {
		apiReturn.ErrorByCode(c, 1300)
		return
	}

	// 检查文件类型（只允许图片）
	fileExt := strings.ToLower(path.Ext(f.Filename))
	agreeExts := []string{".png", ".jpg", ".jpeg", ".webp", ".gif"}
	if !strings.Contains(strings.Join(agreeExts, "|"), fileExt) {
		apiReturn.ErrorByCode(c, 1301)
		return
	}

	// 检查文件大小（限制 10MB）
	if f.Size > 10*1024*1024 {
		apiReturn.ErrorByCode(c, 1302) // 文件过大
		return
	}

	// 生成唯一文件名
	fileName := fmt.Sprintf("wallpaper_%d_%s%s",
		userInfo.ID,
		time.Now().Format("20060102_150405"),
		fileExt)

	// 创建壁纸目录
	wallpaperDir := fmt.Sprintf("%s/wallpapers/user_%d/", configUpload, userInfo.ID)
	if exists, _ := cmn.PathExists(wallpaperDir); !exists {
		os.MkdirAll(wallpaperDir, os.ModePerm)
	}

	filepath := fmt.Sprintf("%s%s", wallpaperDir, fileName)

	// 保存文件
	if err := c.SaveUploadedFile(f, filepath); err != nil {
		apiReturn.ErrorByCode(c, 1300)
		return
	}

	// 记录到数据库
	mFile := models.File{}
	mFile.UserId = userInfo.ID
	mFile.FileName = f.Filename
	mFile.Ext = fileExt
	// 保存完整路径，与便签文件保持一致
	mFile.Src = filepath
	mFile.FileType = "wallpaper"

	if err := global.Db.Create(&mFile).Error; err != nil {
		global.Logger.Warn("Failed to save wallpaper record:", err)
		// 即使数据库记录失败，文件已上传，仍然返回成功
	}

	// 返回完整路径（前端直接使用）
	apiReturn.SuccessData(c, gin.H{
		"imageUrl": filepath,
	})
}

func (a *FileApi) UploadFiles(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	configUpload := global.Config.GetValueStringOrDefault("base", "source_path")

	form, err := c.MultipartForm()
	if err != nil {
		apiReturn.ErrorByCode(c, 1300)
		return
	}
	files := form.File["files[]"]
	errFiles := []string{}
	succMap := map[string]string{}
	for _, f := range files {
		fileExt := strings.ToLower(path.Ext(f.Filename))
		fileName := cmn.Md5(fmt.Sprintf("%s%s", f.Filename, time.Now().String()))
		fildDir := fmt.Sprintf("%s/%d/%d/%d/", configUpload, time.Now().Year(), time.Now().Month(), time.Now().Day())
		isExist, _ := cmn.PathExists(fildDir)
		if !isExist {
			os.MkdirAll(fildDir, os.ModePerm)
		}
		filepath := fmt.Sprintf("%s%s%s", fildDir, fileName, fileExt)
		if c.SaveUploadedFile(f, filepath) != nil {
			errFiles = append(errFiles, f.Filename)
		} else {
			// 成功
			// 像数据库添加记录
			mFile := models.File{}
			mFile.AddFile(userInfo.ID, f.Filename, fileExt, filepath)
			succMap[f.Filename] = filepath
		}
	}

	apiReturn.SuccessData(c, gin.H{
		"succMap":  succMap,
		"errFiles": errFiles,
	})
}

func (a *FileApi) GetList(c *gin.Context) {
	list := []models.File{}
	userInfo, _ := base.GetCurrentUserInfo(c)
	var count int64
	if err := global.Db.Order("created_at desc").Find(&list, "user_id=?", userInfo.ID).Count(&count).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	data := []map[string]interface{}{}
	for _, v := range list {
		data = append(data, map[string]interface{}{
			"src":        v.Src,
			"fileName":   v.FileName,
			"id":         v.ID,
			"createTime": v.CreatedAt,
			"updateTime": v.UpdatedAt,
			"path":       v.Src,
			"fileType":   v.FileType, // 添加文件类型
		})
	}
	apiReturn.SuccessListData(c, data, count)
}

func (a *FileApi) Deletes(c *gin.Context) {
	req := commonApiStructs.RequestDeleteIds[uint]{}
	userInfo, _ := base.GetCurrentUserInfo(c)
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	global.Db.Transaction(func(tx *gorm.DB) error {
		files := []models.File{}

		if err := tx.Order("created_at desc").Find(&files, "user_id=? AND id in ?", userInfo.ID, req.Ids).Error; err != nil {
			return err
		}

		for _, v := range files {
			os.Remove(v.Src)
		}

		if err := tx.Order("created_at desc").Delete(&files, "user_id=? AND id in ?", userInfo.ID, req.Ids).Error; err != nil {
			return err
		}

		return nil
	})

	apiReturn.Success(c)

}
