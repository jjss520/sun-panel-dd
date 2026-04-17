
<div align=center>

<img src="./public/favicon.svg" width="100" height="100" />

# Sun-Panel-DD

一个基于[Sun-Panel](https://github.com/hslr-s/sun-panel)   修改的版本,增加了浏览器的导入书签的功能,使其主页和书签功能分开

在[sun-panel-v2](https://github.com/75412701/sun-panel-v2)二改的基础上再改，个人使用。哈哈

Sun-Panel-dd 一个服务器、NAS导航面板、Homepage、浏览器首页、书签。

个人自用版本,后续会持续更新完善,如果你也喜欢建议点亮右上角星星避免后续迷路
</div>


## ✨ 功能特性

- **区分内外网链接**：自动区分内外网链接
- **个性导航页**：支持导航页自定义设置样式。
- **BING每日壁纸**：服务器每天自动下载最新BING壁纸，一键设置为背景
- **数据缓存**：避免多次请求数据
- **自适应**： pc端和移动端样式自适应,统一图标
- **移动端**：增加长按图标显示菜单  3.31
- **模式网络切换**：新增三种网络模式：内外网自动-外网-内网，更灵活的网络管理
- **编辑模式**：编辑模式单独显示，打开设置-风格管理-显示编辑模式切换按钮，右下角会显示编辑图标，点击，输入登陆密码进入编辑模式
- **Logo点击隐藏**：点击 Logo 文本快速显示/隐藏所有图标（老板键）
- **侧边分组导航**：左侧智能导航条，鼠标悬停显示，高亮跟随滚动位置，点击快速跳转

## 部署
本项目支持 Docker 或其他基于 Docker 的平台部署。<br>
1.编写docker-compose.yml文件<br>
2.运行docker-compose up -d<br>
3.打开 域名/ip:3002<br><br><br>
账号:admin<br>
密码:123456
### docker

```yml
docker run -d \
  --name sun-panel \
  -v $(pwd)/sun-panel/conf:/app/conf \
  -v $(pwd)/sun-panel/uploads:/app/uploads \
  -v $(pwd)/sun-panel/database:/app/database \
  -v $(pwd)/sun-panel/runtime:/app/runtime \
  -p 3002:3002 \
  --restart always \
  ghcr.io/jjss520/sun-panel-dd:latest
```
国内加速
```yml
docker run -d \
  --name sun-panel \
  -v $(pwd)/sun-panel/conf:/app/conf \
  -v $(pwd)/sun-panel/uploads:/app/uploads \
  -v $(pwd)/sun-panel/database:/app/database \
  -v $(pwd)/sun-panel/runtime:/app/runtime \
  -p 3002:3002 \
  --restart always \
  ghcr.1ms.run/jjss520/sun-panel-dd:latest
```

## ❤️ Thanks

- [红烧猎人](https://blog.enianteam.com/u/sun/content/11)

---

[![Star History Chart](https://api.star-history.com/svg?repos=75412701/sun-panel-v2&type=Date)](https://star-history.com/#75412701/sun-panel-v2&Date)
