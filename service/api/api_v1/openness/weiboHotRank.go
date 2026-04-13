package openness

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/global"

	"github.com/gin-gonic/gin"
)

// WeiboHotRank 微博热榜结构体
type WeiboHotRank struct {
	Title string `json:"title"`
	Rank  int    `json:"rank"`
}

// WeiboResponse 微博API响应结构体
type WeiboResponse struct {
	Data struct {
		Realtime []struct {
			Word     string      `json:"word"`
			Monitors interface{} `json:"monitors,omitempty"`
		} `json:"realtime"`
	} `json:"data"`
	Error string `json:"error,omitempty"`
}

// GetWeiboHotRank 获取微博热榜数据
func (a *Openness) GetWeiboHotRank(c *gin.Context) {
	// 构建请求
	req, err := http.NewRequest("GET", "https://weibo.com/ajax/side/searchBand", nil)
	if err != nil {
		apiReturn.Error(c, "请求构建失败："+err.Error())
		return
	}

	// 添加请求头
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")
	req.Header.Set("client-version", "3.0.0")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://weibo.com/")
	req.Header.Set("sec-ch-ua", "\"Not(A:Brand\";v=\"8\", \"Chromium\";v=\"144\", \"Google Chrome\";v=\"144\"")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("server-version", "v2026.02.03.1")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	req.Header.Set("x-xsrf-token", "vvJqBU_YC6NbywDq3RaQ9XlQ")

	// 添加cookies - 注意：这些cookies可能会过期，需要定期更新
	req.Header.Set("cookie", "SINAGLOBAL=732034485695.9249.1769395090616; ULV=1769395090620:1:1:1:732034485695.9249.1769395090616:; SUB=_2AkMe3dWLf8NxqwFRmvERz2LhZYlwzQ7EieKogSRQJRMxHRl-yT9yqkoHtRB6NV37ZcXxeekVVVVwZX-6l-PhHvYOPaQe; SUBP=0033WrSXqPxfM72-Ws9jqgMF55529P9D9WFgAc1KbgwVhT6SoGerR5CR; WBPSESS=3CRDTjRX3vAz9JRDHxBb5JU-w60ZD_uHsSyCan1ITX0bWYY_gBSiHx5y2oVHytqMF6vRTigZMQ9m5TktqMNh19J3uxXmfwkHc1b9pj_5bkeAsHZKly--MUI2jMNG7fwa3mm72LifKfDjs89j8DvyrTC09otYAHUOi2TPQvuLSXM=; XSRF-TOKEN=vvJqBU_YC6NbywDq3RaQ9XlQ")

	// 添加查询参数
	q := req.URL.Query()
	q.Add("last_tab", "hot")
	q.Add("last_tab_time", fmt.Sprintf("%d", time.Now().Unix()))
	req.URL.RawQuery = q.Encode()

	// 打印请求信息
	global.Logger.Info(fmt.Sprintf("微博热榜请求URL: %s", req.URL.String()))

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		global.Logger.Error(fmt.Sprintf("微博热榜请求失败: %s", err.Error()))
		// 即使请求失败，也返回空数据而不是错误，以保证前端正常运行
		apiReturn.SuccessData(c, gin.H{
			"hotRank":    []WeiboHotRank{},
			"updateTime": time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}
	defer resp.Body.Close()

	// 打印响应状态
	global.Logger.Info(fmt.Sprintf("微博热榜响应状态: %d", resp.StatusCode))

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		global.Logger.Error(fmt.Sprintf("微博热榜读取响应失败: %s", err.Error()))
		// 即使读取失败，也返回空数据而不是错误
		apiReturn.SuccessData(c, gin.H{
			"hotRank":    []WeiboHotRank{},
			"updateTime": time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	// 打印响应体（前500个字符）
	respBodyStr := string(body)
	if len(respBodyStr) > 500 {
		respBodyStr = respBodyStr[:500] + "..."
	}
	global.Logger.Info(fmt.Sprintf("微博热榜响应体: %s", respBodyStr))

	// 解析响应
	var weiboResp WeiboResponse
	if err := json.Unmarshal(body, &weiboResp); err != nil {
		global.Logger.Error(fmt.Sprintf("微博热榜解析响应失败: %s", err.Error()))
		// 即使解析失败，也返回空数据而不是错误
		apiReturn.SuccessData(c, gin.H{
			"hotRank":    []WeiboHotRank{},
			"updateTime": time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	// 检查是否有错误
	if weiboResp.Error != "" {
		global.Logger.Error(fmt.Sprintf("微博API返回错误: %s", weiboResp.Error))
		// 即使API返回错误，也返回空数据而不是错误
		apiReturn.SuccessData(c, gin.H{
			"hotRank":    []WeiboHotRank{},
			"updateTime": time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	// 处理数据
	var hotRank []WeiboHotRank
	if len(weiboResp.Data.Realtime) > 0 {
		for i, item := range weiboResp.Data.Realtime {
			// 跳过有monitors字段的数据
			if item.Monitors != nil {
				continue
			}
			// 确保word字段不为空
			if item.Word != "" {
				hotRank = append(hotRank, WeiboHotRank{
					Title: item.Word,
					Rank:  i + 1,
				})
			}
		}
	} else {
		global.Logger.Warn("微博热榜实时数据为空")
	}

	// 如果没有数据，添加一些默认数据，以保证前端正常显示
	if len(hotRank) == 0 {
		hotRank = []WeiboHotRank{
			{Title: "微博热榜数据加载中...", Rank: 1},
			{Title: "请稍后刷新页面重试", Rank: 2},
			{Title: "或检查网络连接", Rank: 3},
		}
		global.Logger.Warn("微博热榜无数据，使用默认数据")
	}

	// 返回数据
	apiReturn.SuccessData(c, gin.H{
		"hotRank":    hotRank,
		"updateTime": time.Now().Format("2006-01-02 15:04:05"),
	})
}
