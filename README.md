
<div align=center>

<img src="./doc/images/logo.png" width="100" height="100" />

# Sun-Panel-DD

一个基于[Sun-Panel](https://github.com/hslr-s/sun-panel)   修改的版本,增加了浏览器的导入书签的功能,使其主页和书签功能分开

在[sun-panel-v2](https://github.com/75412701/sun-panel-v2)二改的基础上再改，个人使用。哈哈

Sun-Panel-dd 一个服务器、NAS导航面板、Homepage、浏览器首页、书签。

个人自用版本,后续会持续更新完善,如果你也喜欢建议点亮右上角星星避免后续迷路
</div>


## ✨ 功能特性

- **区分内外网链接**：自动区分内外网链接
- **个性导航页**：支持导航页自定义设置样式。
- **数据缓存**：避免多次请求数据
- **自适应**： pc端和移动端样式自适应,统一图标
- **移动端**：增加长按图标显示菜单  3.31
- **模式网络切换**：新增三种网络模式：内外网自动-外网-内网，更灵活的网络管理
- **编辑模式**：编辑模式单独显示，打开设置-风格管理-显示编辑模式切换按钮，右下角会显示编辑图标，点击，输入登陆密码进入编辑模式
- **Logo点击隐藏**：点击 Logo 文本快速显示/隐藏所有图标（老板键）


更新内容:
1.区分内外网链接：自动判断打开内外网链接
以前需要通过右下角切换访问模式,特别不方便,现在优化为自动的了

<img  src="https://img.meituan.net/portalweb/ba18a85e1401b1f6a9577f0ee064bc9b2836604.png"/>
<img src="https://nos.netease.com/ysf/d50bb118b2723964b2b837b112bf2c5e.png">
2.编辑模式: 打开设置-风格管理-显示编辑模式切换按钮，右下角会显示编辑图标，点击，输入登陆密码进入编辑模式
<img  src="https://img.meituan.net/portalweb/d42ba6f468f6453c7ffd3eb23f7257483057468.png"/>

3.个性导航页：支持导航页自定义设置样式。
加了个自动获取网络壁纸的功能,避免审美疲劳
<img  src="https://img.meituan.net/portalweb/7dbca78be911a9d4e7c872639c69a148774045.png"/>

4.移动端优化
功能都和pc端图标一样,增加长按显示菜单

5.三种模式网络切换系统（2026.04）
重构了网络模式切换逻辑，提供更清晰的网络管理模式：

**三种网络模式：**
- 🌐 **自动内外网 (auto)**：自动检测内外网环境，智能选择最优链接
- 🌍 **外网模式 (wan)**：强制使用外网链接，不进行内网探测
- 💻 **内网模式 (lan)**：强制使用内网链接，适合纯内网环境

**功能特点：**
- 网络模式循环按钮：在三种公开模式间快速切换（auto → wan → lan）
- 独立编辑模式按钮：可单独开启/隐藏，点击需密码验证
- 自定义图标：每种模式都有独特的视觉标识
- 本地保存：模式状态仅保存在本地浏览器，保护隐私

<img src="https://img.meituan.net/portalweb/ba18a85e1401b1f6a9577f0ee064bc9b2836604.png"/>

**使用场景：**
- 在家使用内网模式访问 NAS 服务
- 在外使用外网模式访问公网服务
- 自动模式让系统智能判断最佳连接方式
- 编辑模式保护敏感内容不被他人查看

8. Logo 点击隐藏功能（2026.04）
新增实用的"老板键"功能，快速隐藏所有内容：

**功能说明：**
- 点击页面顶部的 Logo 文本（如 "Sun-Panel"）
- 第一次点击：隐藏所有图标和应用盒子
- 再次点击：恢复显示所有内容
- 鼠标悬停 Logo 时显示提示文字

**使用场景：**
- 临时需要隐藏屏幕内容，保护隐私
- 快速清理视觉干扰，专注其他工作
- 类似很多导航页的"老板键"功能


添加数据首页不显示?无需重新登录,点击右下角刷新按钮即可.因为数据是缓存在本地,不是随时请求数据,所以这样设计,主要是为了响应快
<img  src="https://nos.netease.com/ysf/d01ace9eb18cde000c0c22861079db84.png" />


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
