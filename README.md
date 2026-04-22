
<div align=center>

<img src="./public/favicon.svg" width="100" height="100" />

# Sun-Panel-DD

一个基于[Sun-Panel](https://github.com/hslr-s/sun-panel) 修改的版本，增加了浏览器的导入书签功能，使其主页和书签功能分开。

在[sun-panel-v2](https://github.com/75412701/sun-panel-v2)二改的基础上再改，个人使用。哈哈

Sun-Panel-dd 一个服务器、NAS导航面板、Homepage、浏览器首页、书签管理器。

**个人自用版本，持续更新完善中**。如果你也喜欢，建议点亮右上角星星 ⭐ 避免后续迷路！
</div>


## ✨ 功能特性

### 🎯 核心功能
- **区分内外网链接**：自动区分内外网链接，智能切换
- **个性导航页**：支持导航页自定义设置样式
- **BING每日壁纸**：服务器每天自动下载最新BING壁纸，一键设置为背景
- **数据缓存**：避免多次请求数据，提升加载速度
- **自适应布局**：PC端和移动端样式自适应，统一图标风格

### 📱 移动端优化
- **长按菜单**：移动端长按图标显示操作菜单（v3.31+
- **触摸友好**：优化的触摸交互体验

### 🌐 网络管理
- **三种网络模式**：内外网自动 / 外网 / 内网，更灵活的网络管理
- **编辑模式**：独立编辑模式，右下角快捷进入，密码保护

### 🔖 书签管理
- **浏览器书签导入**：支持从浏览器导入HTML格式书签文件
- **书签导出**：支持导出为HTML格式，方便备份和迁移
- **树形结构**：完整的文件夹层级支持
- **搜索功能**：快速查找书签内容

### 📝 记事本与提醒
- **富文本记事本**：支持文字、图片、文件附件
- **智能提醒系统**：
  - ⏰ 定时提醒：设置具体时间的提醒
  - 🔄 重复提醒：支持每日/每周/每月/每年重复
  - ⚡ 强制提醒：重要事项强制弹窗提醒
  - 📅 提前提醒：可设置提前N天提醒
- **导入导出**：支持记事本和提醒配置的完整导入导出（覆盖模式）

### 💾 数据管理
- **配置导入导出**：图标配置、记事本数据一键备份恢复
- **上传文件管理**：集中管理上传的壁纸和图标
- **分组管理**：灵活的图标分组管理

### 🎨 界面优化
- **Logo点击隐藏**：点击 Logo 文本快速显示/隐藏所有图标（老板键）
- **侧边分组导航**：左侧智能导航条，鼠标悬停显示，高亮跟随滚动位置
- **深色/浅色主题**：支持自动跟随系统或手动切换

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
