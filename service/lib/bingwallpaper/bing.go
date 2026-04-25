package bingwallpaper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sun-panel/global"
	"sun-panel/lib/cmn"
	"time"
)

type BingWallpaperResponse struct {
	Images []struct {
		URL       string `json:"url"`
		URLBase   string `json:"urlbase"`
		Copyright string `json:"copyright"`
	} `json:"images"`
}

// DownloadBingWallpaper 下载 BING 每日壁纸
func DownloadBingWallpaper() error {
	apiURL := "https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1&mkt=zh-CN"
	
	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("请求 BING API 失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var wallpaperResp BingWallpaperResponse
	if err := json.Unmarshal(body, &wallpaperResp); err != nil {
		return fmt.Errorf("解析 JSON 失败: %v", err)
	}

	if len(wallpaperResp.Images) == 0 {
		return fmt.Errorf("未获取到壁纸数据")
	}

	imageURL := "https://www.bing.com" + wallpaperResp.Images[0].URL
	
	// 统一使用 source_path 配置
	configUpload := global.Config.GetValueStringOrDefault("base", "source_path")
	if configUpload == "" {
		configUpload = "./files"
	}

	// 创建壁纸目录
	wallpaperDir := filepath.Join(configUpload, "wallpapers", "bing")
	if exists, _ := cmn.PathExists(wallpaperDir); !exists {
		os.MkdirAll(wallpaperDir, os.ModePerm)
	}

	// 固定文件名
	fileName := "bing_wallpaper.jpg"
	filePath := filepath.Join(wallpaperDir, fileName)

	fmt.Printf("开始下载 BING 壁纸到: %s\n", filePath)

	// 下载图片
	imgResp, err := http.Get(imageURL)
	if err != nil {
		return fmt.Errorf("下载图片失败: %v", err)
	}
	defer imgResp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, imgResp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	fmt.Printf("BING 壁纸下载成功: %s\n", filePath)
	return nil
}

// StartDailyTask 启动每日定时任务
func StartDailyTask() {
	fmt.Println("启动 BING 壁纸每日下载任务...")

	// 立即执行一次
	if err := DownloadBingWallpaper(); err != nil {
		fmt.Printf("首次下载 BING 壁纸失败: %v\n", err)
	} else {
		fmt.Println("首次下载 BING 壁纸成功")
	}

	// 设置定时器，每天凌晨 0:05 执行（北京时间）
	go func() {
		// 使用北京时区 (UTC+8)
		beijingZone := time.FixedZone("CST", 8*3600)
		
		for {
			now := time.Now().In(beijingZone)
			// 计算下一个凌晨 0:05 的时间（北京时间）
			next := time.Date(now.Year(), now.Month(), now.Day(), 0, 5, 0, 0, beijingZone)
			if now.After(next) {
				// 如果已经过了今天的 0:05，则设置为明天的 0:05
				next = next.Add(24 * time.Hour)
			}
			
			duration := next.Sub(now)
			fmt.Printf("下次 BING 壁纸下载时间（北京时间）: %s (等待 %v)\n", next.Format("2006-01-02 15:04:05"), duration)
			
			time.Sleep(duration)
			
			// 执行下载
			if err := DownloadBingWallpaper(); err != nil {
				fmt.Printf("下载 BING 壁纸失败: %v\n", err)
			} else {
				fmt.Println("BING 壁纸下载成功")
			}
		}
	}()

	fmt.Println("BING 壁纸每日下载任务已启动（每天北京时间 00:05 执行）")
}
