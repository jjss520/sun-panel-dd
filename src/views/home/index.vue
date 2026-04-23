<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import { NBackTop, NButton, NButtonGroup, NDropdown, NModal, NSkeleton, NSpin, useDialog, useMessage } from 'naive-ui'
import { nextTick, onMounted, onActivated, onUnmounted, ref, h } from 'vue'
import { AppIcon, AppStarter, EditItem, NotePad } from './components'
import { Clock, SystemMonitor } from '@/components/deskModule'
import SearchBoxWithSuggestions from '@/components/deskModule/SearchBoxWithSuggestions/index.vue'
import { SvgIcon, SvgIconOnline } from '@/components/common'
import { deletes, getListByGroupId, saveSort } from '@/api/panel/itemIcon'
import { acknowledgeReminder, getNotepadList } from '@/api/panel/notepad'  // 确认提醒 API + 获取便签列表

import { setTitle, updateLocalUserInfo, openUrlWithoutReferer } from '@/utils/cmn'
import { useAuthStore, usePanelState } from '@/store'
import { PanelPanelConfigStyleEnum, PanelStateNetworkModeEnum } from '@/enums'
import { VisitMode } from '@/enums/auth'
import { router } from '@/router'
import { onBeforeRouteUpdate } from 'vue-router'
import { t } from '@/locales'
import { useWindowSize, useStorage } from "@vueuse/core"
interface ItemGroup extends Panel.ItemIconGroup {
  sortStatus?: boolean
  hoverStatus: boolean
  items?: Panel.ItemInfo[]
}

const ms = useMessage()
const dialog = useDialog()
const panelState = usePanelState()
const authStore = useAuthStore()


const scrollContainerRef = ref<HTMLElement | undefined>(undefined)

const editItemInfoShow = ref<boolean>(false)
const editItemInfoData = ref<Panel.ItemInfo | null>(null)
const windowShow = ref<boolean>(false)
const windowSrc = ref<string>('')
const windowTitle = ref<string>('')

const windowIframeRef = ref(null)
const windowIframeIsLoad = ref<boolean>(false)

const dropdownMenuX = ref(0)
const dropdownMenuY = ref(0)
const dropdownShow = ref(false)
const currentRightSelectItem = ref<Panel.ItemInfo | null>(null)
const currentAddItenIconGroupId = ref<number | undefined>()
const notepadVisible = ref(false)
const notepadInstance = ref(null) // 便签实例
let remindEventSource: EventSource | null = null // SSE 连接
const isMobile = ref(false)
const showIcons = ref(true) // 控制图标显示/隐藏
const showGroupNav = ref(false) // 控制分组导航显示/隐藏（默认隐藏）
const currentGroupIndex = ref(0) // 当前滚动到的分组索引
const groupNavTimer = ref<number | null>(null) // 自动隐藏定时器

// 提醒通知列表
interface RemindNotification {
  id: number
  noteId: number
  title: string
  time: string
  visible: boolean
  remindForce?: number // 强制提醒状态
  remindRepeat?: string // 重复类型
}
const remindNotifications = ref<RemindNotification[]>([])
// 切换图标显示/隐藏
function handleToggleIcons() {
  showIcons.value = !showIcons.value
}

// 显示分组导航
function showGroupNavMenu() {
  showGroupNav.value = true
  // 清除之前的隐藏定时器
  if (groupNavTimer.value) {
    clearTimeout(groupNavTimer.value)
    groupNavTimer.value = null
  }
}

// 隐藏分组导航（延迟）
function hideGroupNavMenu() {
  if (groupNavTimer.value) {
    clearTimeout(groupNavTimer.value)
  }
  groupNavTimer.value = window.setTimeout(() => {
    showGroupNav.value = false
  }, 1500) // 1.5秒后自动隐藏
}

// 滚动到指定分组
function scrollToGroup(index: number) {
  const groupElements = document.querySelectorAll('.item-list')
  if (groupElements[index]) {
    groupElements[index].scrollIntoView({ behavior: 'smooth', block: 'start' })
    currentGroupIndex.value = index
  }
}

// 监听滚动，更新当前分组索引
function handleScroll() {
  if (!scrollContainerRef.value) return
  
  const container = scrollContainerRef.value
  const scrollTop = container.scrollTop
  const containerHeight = container.clientHeight
  const scrollPosition = scrollTop + containerHeight / 3
  
  const groupElements = container.querySelectorAll('.item-list')
  
  groupElements.forEach((element, index) => {
    const elementTop = (element as HTMLElement).offsetTop
    const elementBottom = elementTop + (element as HTMLElement).offsetHeight
    
    if (scrollPosition >= elementTop && scrollPosition < elementBottom) {
      currentGroupIndex.value = index
    }
  })
}

// 组件挂载时添加滚动监听和鼠标监听
onMounted(async () => {
  // 监听滚动容器的滚动事件
  if (scrollContainerRef.value) {
    scrollContainerRef.value.addEventListener('scroll', handleScroll)
  }
  
  // 使用事件委托监听鼠标移动
  document.addEventListener('mousemove', handleMouseMove)
  
  // ✅ 启动 SSE 提醒推送并执行离线补偿
  if (authStore.visitMode === VisitMode.VISIT_MODE_LOGIN) {
    // 1. 先把数据库里“挂起”的提醒弹出来（离线补偿）
    await checkInitialReminders()
    // 2. 再开启实时推送
    startRemindSSE()
  }
})

// 组件卸载时清除定时器和SSE连接
onUnmounted(() => {
  stopRemindSSE()
})

// 鼠标移动事件处理
function handleMouseMove(e: MouseEvent) {
  // 鼠标在左侧 60px 区域内
  if (e.clientX <= 60) {
    showGroupNavMenu()
  } 
  // 鼠标离开导航条区域（超过 120px）
  else if (e.clientX > 120 && showGroupNav.value) {
    hideGroupNavMenu()
  }
}

// ========== SSE 提醒功能 ==========

// 启动 SSE 提醒推送
const startRemindSSE = () => {
  if (!authStore.userInfo?.id) {
    console.log('[SSE提醒] 未登录，跳过连接')
    return
  }

  // 如果已有连接，先关闭
  if (remindEventSource) {
    remindEventSource.close()
  }

  const userId = authStore.userInfo.id
  const url = `/api/panel/notepad/remindStream?userId=${userId}`
  
  console.log('[SSE提醒] 正在连接:', url)
  remindEventSource = new EventSource(url)

  // 连接成功
  remindEventSource.addEventListener('connected', (event) => {
    console.log('[SSE提醒] 连接成功')
  })

  // 接收提醒
  remindEventSource.addEventListener('remind', (event) => {
    console.log('[SSE提醒] 收到提醒:', event.data)
    try {
      // 解析数据（格式为 key:value,key:value）
      const dataStr = event.data as string
      const pairs = dataStr.split(',')
      const data: any = {}
      pairs.forEach(pair => {
        const [key, ...valueParts] = pair.split(':')
        if (key && valueParts.length > 0) {
          data[key.trim()] = valueParts.join(':').trim()
        }
      })

      // 转换为数字类型
      if (data.id) data.id = parseInt(data.id)
      if (data.remindForce) data.remindForce = parseInt(data.remindForce)
      if (data.remindAdvanceDays) data.remindAdvanceDays = parseInt(data.remindAdvanceDays)

      console.log('[SSE提醒] 解析后的数据:', data)
      showRemindNotification(data)
    } catch (error) {
      console.error('[SSE提醒] 解析失败:', error)
    }
  })

  // 错误处理
  remindEventSource.onerror = (error) => {
    console.error('[SSE提醒] 连接错误:', error)
    if (remindEventSource?.readyState === EventSource.CLOSED) {
      console.log('[SSE提醒] 连接已关闭')
    }
  }
}

// 停止 SSE 提醒推送
const stopRemindSSE = () => {
  if (remindEventSource) {
    remindEventSource.close()
    remindEventSource = null
    console.log('[SSE提醒] 连接已关闭')
  }
}

// 显示提醒通知（卡片式）
const showRemindNotification = (note: any) => {
  // ✅ 去重机制：如果该便签已经在显示了，不再弹出
  if (remindNotifications.value.some(n => n.noteId === note.id)) {
    return
  }

  const id = Date.now() // 生成一个前端展示用的唯一 ID
  
  // 计算实际提醒时间用于显示
  let displayTime = note.remindTime
  if (note.remindAdvanceDays && note.remindAdvanceDays > 0 && note.remindTime) {
    const baseTime = new Date(note.remindTime)
    const actualTime = new Date(baseTime)
    actualTime.setDate(actualTime.getDate() - note.remindAdvanceDays)
    displayTime = actualTime.toISOString()
  }
  
  const notification: RemindNotification = {
    id,
    noteId: note.id,
    title: note.title || '无标题',
    time: new Date(displayTime).toLocaleString('zh-CN'),
    visible: true,
    remindForce: note.remindForce || 0,
    remindRepeat: note.remindRepeat || 'none'
  }
  
  remindNotifications.value.push(notification)
  
  // ✅ 移除定时自动消失逻辑：不再设置 window.setTimeout

  // 浏览器原生通知保持不变
  if ('Notification' in window && Notification.permission === 'granted') {
    new Notification('⏰ 提醒', {
      body: note.title,
      icon: '/logo.png'
    })
  }
}

// 关闭通知（用户点击“知道了”）
const closeNotification = async (id: number) => {
  const index = remindNotifications.value.findIndex(n => n.id === id)
  if (index > -1) {
    const notification = remindNotifications.value[index]
    
    // 移除定时器清理代码（因为已经没有定时器了）
    
    try {
      // 1. 调用后端确认 API
      await acknowledgeReminder({ id: notification.noteId })
      console.log('[提醒确认] 成功:', notification.noteId)
      
      // 2. ✅ 刷新列表：确保桌面图标、便签列表的状态实时同步为“已处理”
      handleRefreshData()
      
    } catch (error) {
      console.error('[提醒确认] 失败:', error)
    }
    
    // 3. UI 隐藏动画
    notification.visible = false
    setTimeout(() => {
      remindNotifications.value = remindNotifications.value.filter(n => n.id !== id)
    }, 300)
  }
}


// 查看便签
const viewNotepad = (noteId: number, notificationId: number) => {
  notepadVisible.value = true
  closeNotification(notificationId)
  // 等待组件加载后选中该便签
  setTimeout(() => {
    if (notepadInstance.value) {
      // @ts-ignore
      notepadInstance.value.selectNote?.({ id: noteId })
    }
  }, 100)
}

// ✅ 离线补偿检查：页面加载时检查是否有未确认的提醒（remindStatus === 1）
const checkInitialReminders = async () => {
  if (authStore.visitMode !== VisitMode.VISIT_MODE_LOGIN) return
  
  try {
    const res = await getNotepadList()
    if (res.code === 0 && res.data) {
      // 找出所有状态为 1 (待确认) 的便签
      const pendingReminds = res.data.filter((note: any) => note.remindStatus === 1)
      console.log('[SSE提醒] 发现离线期间未确认提醒:', pendingReminds.length)
      
      // 逐个弹出（showRemindNotification 内部已有去重逻辑，不会重复弹出）
      pendingReminds.forEach((note: any) => {
        showRemindNotification(note)
      })
    }
  } catch (error) {
    console.error('[SSE提醒] 获取离线提醒失败:', error)
  }
}

// ✅ 离线补偿检查：页面加载时检查是否有未确认的提醒（remindStatus === 1）
function checkOfflineReminders() {
  filterItems.value.forEach(group => {
    group.items?.forEach(item => {
      // item 本身是 NotepadInfo 的子集，如果它的状态是 1
      if ((item as any).remindStatus === 1) {
        showRemindNotification(item)
      }
    })
  })
}

async function handleRefreshData() {
  try {
    // 删除除用户登录信息外的所有缓存
    ss.remove(BOOKMARKS_CACHE_KEY)
    ss.remove(GROUP_LIST_CACHE_KEY)
    ss.remove('searchEngineListCache')

    // 直接清除所有localStorage中的图标列表缓存
    // 由于ss没有getAllKeys方法，我们直接使用原生localStorage API
    Object.keys(localStorage).forEach(key => {
      if (key.startsWith(ITEM_ICON_LIST_CACHE_KEY_PREFIX)) {
        ss.remove(key)
      }
    })

    // 重新加载数据
    getList()

    // 刷新便签数据（仅登录状态下）
    if (authStore.visitMode === VisitMode.VISIT_MODE_LOGIN && notepadInstance.value) {
        // @ts-ignore
        notepadInstance.value.refreshData?.()
    }

    // 检查是否需要重新获取网络壁纸
    if (panelState.panelConfig.autoNetworkWallpaper) {
      try {
        // 添加时间戳参数，确保每次请求的URL不同，避免浏览器缓存
        const timestamp = new Date().getTime()
        const baseUrl = panelState.panelConfig.autoNetworkWallpaperApi || 'https://img.xjh.me/random_img.php?return=302&type=bg&ctype=nature'
        const apiUrl = baseUrl.includes('?') ? `${baseUrl}&t=${timestamp}` : `${baseUrl}?t=${timestamp}`
        panelState.panelConfig.backgroundImageSrc = apiUrl
        panelState.recordState()
      } catch (error) {
        console.error('重新获取网络壁纸失败', error)
      }
    }

    ms.success(t('common.refreshSuccess'))
  } catch (error) {
    console.error('刷新数据失败:', error)
    ms.error(t('common.refreshFailed'))
  }
}
// 1. 定义树形节点的接口（包含 children）
interface TreeItem {
	key: string | number;
	label: string;
	isLeaf: boolean;
	bookmark?: {
		id: number;
		title: string;
		url: string;
		folderId: string | null;
	};
	children?: TreeItem[]; // 补充 children 属性（可选，叶子节点没有）
}

const settingModalShow = ref(false)

const items = ref<ItemGroup[]>([])
const filterItems = ref<ItemGroup[]>([])



useWindowSize()

// 从API导入获取书签列表的函数
import { getList as getBookmarksList } from '@/api/panel/bookmark'
import { getList as getGroupList } from '@/api/panel/itemIconGroup'
import { ss } from '@/utils/storage/local'
import { getSystemSettings } from '@/api/system/systemSetting'


// 书签数据树
const treeData = ref<any[]>([])
// 缓存键名
const BOOKMARKS_CACHE_KEY = 'bookmarksTreeCache'
const GROUP_LIST_CACHE_KEY = 'groupListCache'
// 图标列表缓存键前缀
const ITEM_ICON_LIST_CACHE_KEY_PREFIX = 'itemIconList_'

const systemPingUrl = useStorage('systemPingUrl', '')

// 检测内网连接
async function checkIntranetConnection(): Promise<boolean> {
  if (!systemPingUrl.value) return false

  let url = systemPingUrl.value.trim()
  if (!url.startsWith('http://') && !url.startsWith('https://')) {
    url = 'http://' + url
  }

  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 150)

  try {
    const response = await fetch(url, {
      method: 'GET',
      signal: controller.signal
    })
    clearTimeout(timeoutId)
    return response.status === 200
  } catch (e) {
    return false
  }
}


// 获取书签数据并转换为前端需要的格式
async function loadBookmarkTree(forceRefresh = false) {
  try {
    // 如果不是强制刷新且缓存存在，则使用缓存
    if (!forceRefresh) {
      const cachedData = ss.get(BOOKMARKS_CACHE_KEY)
      if (cachedData) {
        // 处理缓存的原始fullData格式数据
        let treeDataResult = [];

        // 检查是否已经是树形结构（直接包含children字段）
        if (Array.isArray(cachedData) && cachedData.length > 0 && 'children' in cachedData[0]) {
          treeDataResult = convertServerTreeToFrontendTree(cachedData)
        } else if (cachedData.list && Array.isArray(cachedData.list)) {
          // 后端返回的是带list字段的结构
          const serverBookmarks = cachedData.list
          if (serverBookmarks.length > 0 && 'children' in serverBookmarks[0]) {
            treeDataResult = convertServerTreeToFrontendTree(serverBookmarks)
          } else {
            treeDataResult = buildBookmarkTree(serverBookmarks)
          }
        } else {
          treeDataResult = buildBookmarkTree(Array.isArray(cachedData) ? cachedData : [])
        }

        treeData.value = treeDataResult
        return
      }
    } else {
      // 强制刷新时清除缓存
      ss.remove(BOOKMARKS_CACHE_KEY)
    }
    const response = await getBookmarksList()
    if (response.code === 0) {
      // 检查数据结构
      const data: any = response.data || []
      let treeDataResult = []

      // 检查是否已经是树形结构（直接包含children字段）
      if (Array.isArray(data) && data.length > 0 && 'children' in data[0]) {
        treeDataResult = convertServerTreeToFrontendTree(data)
      } else if (data.list && Array.isArray(data.list)) {
        // 后端返回的是带list字段的结构
        const serverBookmarks = data.list
        if (serverBookmarks.length > 0 && 'children' in serverBookmarks[0]) {
          treeDataResult = convertServerTreeToFrontendTree(serverBookmarks)
        } else {
          treeDataResult = buildBookmarkTree(serverBookmarks)
        }
      } else {
        treeDataResult = buildBookmarkTree(Array.isArray(data) ? data : [])
      }

      // 更新treeData
      treeData.value = treeDataResult
      ss.set(BOOKMARKS_CACHE_KEY, data)
    }
  } catch (error) {
    console.error('获取书签数据失败:', error)
    // 出错时尝试使用缓存
    const cachedData = ss.get(BOOKMARKS_CACHE_KEY)
    if (cachedData) {
      // 处理缓存的原始fullData格式数据
      let treeDataResult = [];

      // 检查是否已经是树形结构（直接包含children字段）
      if (Array.isArray(cachedData) && cachedData.length > 0 && 'children' in cachedData[0]) {
        treeDataResult = convertServerTreeToFrontendTree(cachedData)
      } else if (cachedData.list && Array.isArray(cachedData.list)) {
        // 后端返回的是带list字段的结构
        const serverBookmarks = cachedData.list
        if (serverBookmarks.length > 0 && 'children' in serverBookmarks[0]) {
          treeDataResult = convertServerTreeToFrontendTree(serverBookmarks)
        } else {
          treeDataResult = buildBookmarkTree(serverBookmarks)
        }
      } else {
        treeDataResult = buildBookmarkTree(Array.isArray(cachedData) ? cachedData : [])
      }
      treeData.value = treeDataResult
    }
  }
}

// 将服务器返回的树形结构转换为前端组件需要的格式
function convertServerTreeToFrontendTree(serverTree: any[]): any[] {
  // 先对顶层节点按sort字段排序
  const sortedServerTree = [...serverTree].sort((a, b) => (a.sort || 0) - (b.sort || 0));
  const result = sortedServerTree.map(node => {
    // 处理两种可能的节点结构：
    // 1. 服务器原始数据格式 (id, title, isFolder, url, iconJson)
    // 2. 前端节点格式 (key, label, isFolder, bookmark)
    const isFrontendFormat = node.hasOwnProperty('key') && node.hasOwnProperty('label');

    // 提取基本属性
    const nodeId = isFrontendFormat ? node.key : node.id;
    const title = isFrontendFormat ? node.label : node.title;
    const isFolder = isFrontendFormat ? (node.isFolder ? 1 : 0) : node.isFolder;
    const url = isFrontendFormat ? (node.bookmark?.url || '') : node.url;
    const iconJson = isFrontendFormat ? (node.bookmark?.iconJson || '') : node.iconJson;
    const parentId = isFrontendFormat ? (node.rawNode?.parentId || node.ParentId || '0') : (node.parentId || node.ParentId || '0');

    // 提取排序字段
    const sortOrder = node.sort || 0;

    // 处理bookmark对象
    let bookmarkObj = undefined;
    if (isFolder !== 1 && url) {
      // 确保folderId是字符串类型
      const folderId = parentId !== undefined ? String(parentId) : null;
      bookmarkObj = {
        id: nodeId,
        title: title,
        url: url,
        folderId: folderId,
        iconJson: iconJson, // 保存base64图标数据
        sort: sortOrder // 保存排序字段到书签对象
      };
    }

    const frontendNode = {
        key: nodeId,
        label: title || '未命名',
        isLeaf: isFolder !== 1,
        isFolder: isFolder === 1, // 添加isFolder属性
        sort: sortOrder, // 保存排序字段到前端节点
        bookmark: bookmarkObj
    };

    // 递归处理子节点
    if (node.children && node.children.length > 0) {
      // 对子节点先按sort字段排序再递归转换
      const sortedChildren = [...node.children].sort((a, b) => (a.sort || 0) - (b.sort || 0));
      (frontendNode as TreeItem).children = convertServerTreeToFrontendTree(sortedChildren);
    }

    return frontendNode;
  });

  return result;
}

// 构建书签树
function buildBookmarkTree(bookmarks: any[]): any[] {
  // 首先分离文件夹和书签
  const folders = bookmarks.filter(b => {
    return (b.isFolder === 1 || (b.isFolder && typeof b.isFolder === 'boolean'));
  });
  const items = bookmarks.filter(b => {
    return (b.isFolder === 0 || (!b.isFolder && typeof b.isFolder === 'boolean'));
  });

  // 构建文件夹树
  const rootFolders: any[] = []
  const folderMap = new Map<string, any>() // 使用字符串键

  // 先创建所有文件夹节点
  folders.forEach(folder => {
    // 处理两种可能的文件夹结构
    const isFrontendFormat = folder.hasOwnProperty('key') && folder.hasOwnProperty('label');
    const folderId = isFrontendFormat ? folder.key : folder.id;
    const folderTitle = isFrontendFormat ? folder.label : folder.title;
    const folderSort = folder.sort || 0;
    const folderNode = {
      key: folderId,
      label: folderTitle,
      children: [],
      isFolder: true,
      sort: folderSort // 保存排序字段
    };
    // 使用id作为map的键
    folderMap.set(folderId.toString(), folderNode);
    // 同时也将文件夹名称作为键，以便处理嵌套关系
    folderMap.set(folderTitle, folderNode);
  });

  // 将文件夹添加到其父文件夹中
  folders.forEach(folder => {
    const folderNode = folderMap.get(folder.id.toString())
    // 检查是否有ParentUrl并且不是根节点(0)
    if (folder.ParentUrl && folder.ParentUrl !== '0' && folder.ParentUrl !== 0) {
      // 尝试用不同的方式查找父文件夹
      let parentFolder = folderMap.get(folder.ParentUrl.toString())

      if (!parentFolder) {
        // 如果找不到，尝试用文件夹标题匹配
        parentFolder = folderMap.get(folder.ParentUrl)
      }

      if (parentFolder) {
        parentFolder.children.push(folderNode)
        return
      }
    }
    // 如果没有父文件夹或父文件夹不存在，则作为根文件夹
    rootFolders.push(folderNode)
  })

  // 将书签项添加到对应的文件夹中
  items.forEach(item => {
    // 处理两种可能的书签结构
    const isFrontendFormat = item.hasOwnProperty('key') && item.hasOwnProperty('label');
    // 提取书签基本信息
    const bookmarkId = isFrontendFormat ? item.key : item.id;
    const bookmarkTitle = isFrontendFormat ? item.label : (item.title || '未命名');
    const bookmarkUrl = isFrontendFormat ? (item.bookmark?.url || '') : (item.url || '');
    const bookmarkIconJson = isFrontendFormat ? (item.bookmark?.iconJson || '') : (item.iconJson || '');
    // 确保folderId是字符串类型
    const folderId = isFrontendFormat ? (item.rawNode?.parentId || item.ParentId || '0') : (item.parentId || item.ParentId || '0');
    const stringFolderId = String(folderId);
    // 获取排序字段
    const sortOrder = isFrontendFormat ? (item.rawNode?.sort || 0) : (item.sort || 0);

    let targetFolder;

    if (stringFolderId === '0' || stringFolderId === 'null' || stringFolderId === 'undefined') {
      // 根目录的书签，创建一个"未分类"文件夹
      targetFolder = folderMap.get('未分类');
      if (!targetFolder) {
        targetFolder = {
          key: '未分类',
          label: '未分类',
          children: [],
          isFolder: true,
          sort: 0 // 设置默认排序
        };
        folderMap.set('未分类', targetFolder);
        rootFolders.push(targetFolder);
      }
    } else {
      // 查找对应的文件夹
      targetFolder = folderMap.get(stringFolderId);
    }

    if (targetFolder) {
      // 创建书签节点
      const bookmarkNode = {
        key: bookmarkId,
        label: bookmarkTitle,
        isLeaf: true,
        sort: sortOrder, // 保存排序字段
        bookmark: {
          id: bookmarkId,
          title: bookmarkTitle,
          url: bookmarkUrl,
          folderId: stringFolderId,
          iconJson: bookmarkIconJson
        }
      };
      targetFolder.children.push(bookmarkNode);
    }
  })

  // 递归排序所有节点的子节点
  function sortTreeNodes(nodes: any[]) {
    nodes.sort((a, b) => (a.sort || 0) - (b.sort || 0));
    nodes.forEach(node => {
      if (node.isFolder && node.children) {
        sortTreeNodes(node.children);
      }
    });
  }

  sortTreeNodes(rootFolders);

  return rootFolders
}



function openPage(openMethod: number, url: string, title?: string) {
  switch (openMethod) {
    case 1:
      window.location.replace(url)
      break
    case 2:
      openUrlWithoutReferer(url, '_blank')
      break
    case 3:
      windowShow.value = true
      windowSrc.value = url
      windowTitle.value = title || url
      windowIframeIsLoad.value = true
      break

    default:
      break
  }
}

async function handleItemClick(itemGroupIndex: number, item: Panel.ItemInfo) {
  // 如果是移动端且刚刚是长按，则不触发点击事件
  if (isMobile.value && isLongPressing) {
    isLongPressing = false
    return
  }
  
  if (items.value[itemGroupIndex] && items.value[itemGroupIndex].sortStatus) {
    handleEditItem(item)
    return
  }

  // 辅助函数：标准化URL（自动添加http://）
  const normalizeUrl = (url: string | undefined | null) => {
    if (!url) return ''
    let trimmed = url.trim()
    if (!trimmed) return ''

    // 如果是 javascript: 等特殊协议或已经是 http/https 开头，或者是相对路径，则不处理
    if (/^[a-z]+:/i.test(trimmed) || trimmed.startsWith('/') || trimmed.startsWith('./') || trimmed.startsWith('../')) {
      return trimmed
    }

    // 默认为 http
    return 'http://' + trimmed
  }

  // 辅助函数：检查URL是否有效
  const isValidUrl = (url: string | undefined | null) => {
    if (!url) return false
    const trimmed = url.trim()
    return trimmed !== '' && trimmed !== 'null' && trimmed !== 'undefined'
  }

  // 预处理 URL
  const publicUrl = normalizeUrl(item.url)
  const lanUrl = normalizeUrl(item.lanUrl)

  // wan 模式：直接显示外网，不进行内网探测
  if (panelState.networkMode === PanelStateNetworkModeEnum.wan) {
    openPage(item.openMethod, publicUrl, item.title)
    return
  }

  // lan 模式：直接显示内网，不进行内网探测（复制 wan 逻辑）
  if (panelState.networkMode === PanelStateNetworkModeEnum.lan) {
    openPage(item.openMethod, lanUrl, item.title)
    return
  }

  // 默认使用公网地址
  let jumpUrl = publicUrl

  // 检查是否需要进行内网探测（auto 和 edit 模式）
  // 条件：有内网地址 AND 内网地址有效 AND 系统配置了PingUrl
  // 注意：这里我们检查原始的 item.lanUrl 是否有效，但使用标准化的 lanUrl 进行跳转
  const shouldCheckIntranet = isValidUrl(item.lanUrl) && systemPingUrl.value

  if (shouldCheckIntranet) {
    // 情况1：新窗口打开 (openMethod === 2)
    // 需要先打开空白窗口避开拦截，然后异步探测
    if (item.openMethod === 2) {
      const newWindow = window.open('about:blank', '_blank')
      if (newWindow) {
        // 探测
        const isIntranet = await checkIntranetConnection()

        // 确定最终URL
        let finalUrl = publicUrl
        if (isIntranet && isValidUrl(item.lanUrl)) {
             finalUrl = lanUrl
        }

        newWindow.location.href = finalUrl
        return // 结束，不执行后面的 openPage
      }
    }

    // 情况2：当前窗口或弹窗 (openMethod === 1, 3等)
    const isIntranet = await checkIntranetConnection()
    if (isIntranet && isValidUrl(item.lanUrl)) {
      jumpUrl = lanUrl
    }
  }

  // 执行打开页面 (如果是新窗口且上面处理过了，这里就不会执行)
  openPage(item.openMethod, jumpUrl, item.title)
}

function handWindowIframeIdLoad(payload: Event) {
  windowIframeIsLoad.value = false
}

// 根据网络模式过滤项目
function filterItemsByNetworkMode() {
  // WAN、LAN 和 auto 模式需要过滤
  if (panelState.networkMode === PanelStateNetworkModeEnum.wan || 
      panelState.networkMode === PanelStateNetworkModeEnum.auto ||
      panelState.networkMode === PanelStateNetworkModeEnum.lan) {
    const filteredGroups = items.value.map(group => {
      if (group.items) {
        // 过滤掉lanOnly为1的项目
        const filteredItems = group.items.filter(item => item.lanOnly !== 1)
        return { ...group, items: filteredItems }
      }
      return group
    })
    // 过滤掉没有项目的组
    filterItems.value = filteredGroups.filter(group => !group.items || group.items.length > 0)
  } else {
    // 只有编辑模式下显示所有项目
    filterItems.value = items.value
  }
}

async function getList() {
  try {
    // 1. 首先尝试从缓存读取数据
    const cachedData = ss.get(GROUP_LIST_CACHE_KEY)
    if (cachedData) {
      items.value = cachedData
      // 为每个分组加载图标数据
      for (let i = 0; i < cachedData.length; i++) {
        const element = cachedData[i]
        if (element.id)
          updateItemIconGroupByNet(i, element.id)
      }
      // 应用网络模式过滤
      filterItemsByNetworkMode()
      // ✅ 离线补偿检查：检查是否有未确认的提醒
      checkOfflineReminders()
      return
    }

    // 2. 缓存中没有数据，请求接口获取数据
    const response = await getGroupList<Common.ListResponse<ItemGroup[]>>()
    if (response.code === 0) {
      items.value = response.data.list
      // 3. 将数据永久保存到缓存中
      ss.set(GROUP_LIST_CACHE_KEY, response.data.list)

      // 为每个分组加载图标数据
      for (let i = 0; i < response.data.list.length; i++) {
        const element = response.data.list[i]
        if (element.id)
          updateItemIconGroupByNet(i, element.id)
      }
      // 应用网络模式过滤
      filterItemsByNetworkMode()
      // ✅ 离线补偿检查：检查是否有未确认的提醒
      checkOfflineReminders()
    }
  } catch (error) {
    // 出错时尝试从缓存获取
    const cachedData = ss.get(GROUP_LIST_CACHE_KEY)
    if (cachedData) {
      items.value = cachedData
      // 为每个分组加载图标数据
      for (let i = 0; i < cachedData.length; i++) {
        const element = cachedData[i]
        if (element.id)
          updateItemIconGroupByNet(i, element.id)
      }
      // 应用网络模式过滤
      filterItemsByNetworkMode()
      // ✅ 离线补偿检查：检查是否有未确认的提醒
      checkOfflineReminders()
    }
  }
}

// 从后端获取组下面的图标
async function updateItemIconGroupByNet(itemIconGroupIndex: number, itemIconGroupId: number) {
  try {
    // 1. 定义缓存键
    const cacheKey = `${ITEM_ICON_LIST_CACHE_KEY_PREFIX}${itemIconGroupId}`

    // 2. 首先尝试从缓存读取数据
    const cachedData = ss.get(cacheKey)
    if (cachedData) {
      items.value[itemIconGroupIndex].items = cachedData

      // 当所有组的数据都加载完成后，应用网络模式过滤
      const allGroupsLoaded = items.value.every(group => group.items !== undefined)
      if (allGroupsLoaded) {
        filterItemsByNetworkMode()
      }
      return
    }
    const res = await getListByGroupId<Common.ListResponse<Panel.ItemInfo[]>>(itemIconGroupId)

    if (res.code === 0) {
      items.value[itemIconGroupIndex].items = res.data.list
      // 4. 将数据永久保存到缓存中
      ss.set(cacheKey, res.data.list)

      // 当所有组的数据都加载完成后，应用网络模式过滤
      const allGroupsLoaded = items.value.every(group => group.items !== undefined)
      if (allGroupsLoaded) {
        filterItemsByNetworkMode()
      }
    }
  } catch (error) {
    // 出错时尝试从缓存获取
    const cacheKey = `${ITEM_ICON_LIST_CACHE_KEY_PREFIX}${itemIconGroupId}`
    const cachedData = ss.get(cacheKey)
    if (cachedData) {
      items.value[itemIconGroupIndex].items = cachedData

      // 当所有组的数据都加载完成后，应用网络模式过滤
      const allGroupsLoaded = items.value.every(group => group.items !== undefined)
      if (allGroupsLoaded) {
        filterItemsByNetworkMode()
      }
    }
  }
}

// 组件激活时刷新书签数据确保显示最新顺序
onActivated(() => {
  // 延迟执行，优先保证页面切换流畅
  setTimeout(() => {
    loadBookmarkTree(false);
  }, 20);
});

function handleRightMenuSelect(key: string | number) {
  dropdownShow.value = false
  // LAN 模式和编辑模式使用 LAN URL，WAN 和 auto 模式使用 WAN URL
  const isLanMode = panelState.networkMode === PanelStateNetworkModeEnum.lan || 
                    panelState.networkMode === PanelStateNetworkModeEnum.edit
  let jumpUrl = isLanMode ? currentRightSelectItem.value?.lanUrl : currentRightSelectItem.value?.url
  if (currentRightSelectItem.value?.lanUrl === '')
    jumpUrl = currentRightSelectItem.value.url
  switch (key) {
    case 'newWindows':
      if (jumpUrl) {
        openUrlWithoutReferer(jumpUrl, '_blank')
      }
      break
    case 'openWanUrl':
      if (currentRightSelectItem.value)
        openPage(currentRightSelectItem.value?.openMethod, currentRightSelectItem.value?.url, currentRightSelectItem.value?.title)
      break
    case 'openLanUrl':
      if (currentRightSelectItem.value && currentRightSelectItem.value.lanUrl)
        openPage(currentRightSelectItem.value?.openMethod, currentRightSelectItem.value.lanUrl, currentRightSelectItem.value?.title)
      break
    case 'edit':
      // 这里有个奇怪的问题，如果不使用{...}的方式 父组件的值会同步修改 标记一下
      handleEditItem({ ...currentRightSelectItem.value } as Panel.ItemInfo)
      break
    case 'delete':
      dialog.warning({
        title: t('common.warning'),
        content: t('common.deleteConfirmByName', { name: currentRightSelectItem.value?.title }),
        positiveText: t('common.confirm'),
        negativeText: t('common.cancel'),
        onPositiveClick: () => {
          if (currentRightSelectItem.value) {
            const itemIconGroupId = currentRightSelectItem.value.itemIconGroupId
            deletes([currentRightSelectItem.value.id as number]).then(({ code, msg }) => {
              if (code === 0) {
                ms.success(t('common.deleteSuccess'))
                // 清除该分组的图标缓存
                ss.remove(`${ITEM_ICON_LIST_CACHE_KEY_PREFIX}${itemIconGroupId}`)
                getList()
              }
              else {
                ms.error(`${t('common.deleteFail')}:${msg}`)
              }
            })
          }
        },
      })

      break
    default:
      break
  }
}

function handleContextMenu(e: MouseEvent, itemGroupIndex: number, item: Panel.ItemInfo) {
  if (items.value[itemGroupIndex] && items.value[itemGroupIndex].sortStatus)
    return

  e.preventDefault()
  currentRightSelectItem.value = item
  dropdownShow.value = false
  nextTick().then(() => {
    dropdownShow.value = true
    dropdownMenuX.value = e.clientX
    dropdownMenuY.value = e.clientY
  })
}

// 处理触摸开始事件
let longPressTimerId: number | null = null
let touchStartX = 0
let touchStartY = 0
let isLongPressing = false // 标记是否正在长按

function handleTouchStart(e: TouchEvent, itemGroupIndex: number, item: Panel.ItemInfo) {
  const touch = e.touches[0]
  touchStartX = touch.clientX
  touchStartY = touch.clientY
  isLongPressing = false
  
  // 设置长按定时器（500ms）
  longPressTimerId = window.setTimeout(() => {
    isLongPressing = true // 标记为长按状态
    handleLongPress(itemGroupIndex, item)
  }, 500)
}

function handleTouchEnd() {
  // 清除定时器
  if (longPressTimerId !== null) {
    clearTimeout(longPressTimerId)
    longPressTimerId = null
  }
}

function handleLongPress(itemGroupIndex: number, item: Panel.ItemInfo) {
  if (items.value[itemGroupIndex] && items.value[itemGroupIndex].sortStatus)
    return

  // 菜单显示后立即清除定时器，防止手指移动时菜单消失
  if (longPressTimerId !== null) {
    clearTimeout(longPressTimerId)
    longPressTimerId = null
  }

  currentRightSelectItem.value = item
  dropdownShow.value = false
  nextTick().then(() => {
    dropdownShow.value = true
    // 使用触摸起始位置作为菜单位置
    dropdownMenuX.value = touchStartX
    dropdownMenuY.value = touchStartY
  })
}

// 检测是否为移动设备
function checkMobile() {
  isMobile.value = window.innerWidth < 768
}

function onClickoutside() {
  // message.info('clickoutside')
  dropdownShow.value = false
}

function handleEditSuccess(item: Panel.ItemInfo) {
  // 查找编辑的图标所属的分组
  for (let i = 0; i < items.value.length; i++) {
    const group = items.value[i]
    if (group.id === item.itemIconGroupId) {
      // 清除该分组的图标缓存
      ss.remove(`${ITEM_ICON_LIST_CACHE_KEY_PREFIX}${item.itemIconGroupId}`)
      break
    }
  }
  getList()
}

// function handleChangeNetwork(mode: PanelStateNetworkModeEnum) {
//   panelState.setNetworkMode(mode)
//   if (mode === PanelStateNetworkModeEnum.lan)
//     ms.success(t('panelHome.changeToLanModelSuccess'))
//
//   else
//     ms.success(t('panelHome.changeToWanModelSuccess'))
//
//   // 切换网络模式后，重新应用过滤
//   filterItemsByNetworkMode()
// }


function handleSaveSort(itemGroup: ItemGroup) {
  const saveItems: Common.SortItemRequest[] = []
  if (itemGroup.items) {
    for (let i = 0; i < itemGroup.items.length; i++) {
      const element = itemGroup.items[i]
      saveItems.push({
        id: element.id as number,
        sort: i + 1,
      })
    }

    saveSort({ itemIconGroupId: itemGroup.id as number, sortItems: saveItems }).then(({ code, msg }) => {
      if (code === 0) {
        // 清除该分组的图标缓存
        ss.remove(`${ITEM_ICON_LIST_CACHE_KEY_PREFIX}${itemGroup.id}`)
        // itemGroup.sortStatus = false // 不要自动关闭排序状态，允许用户继续操作
      }
      else {
        console.error(`${t('common.saveFail')}:${msg}`)
      }
    })
  }
}

function getDropdownMenuOptions() {
  const dropdownMenuOptions = [
    {
      label: t('iconItem.newWindowOpen'),
      key: 'newWindows',
    },

  ]

  // 当图标有公网地址时，显示打开公网地址选项
  if (currentRightSelectItem.value?.url) {
    dropdownMenuOptions.push({
      label: t('panelHome.openWanUrl'),
      key: 'openWanUrl',
    })
  }

  // 当图标有内网地址时，显示打开内网地址选项
  if (currentRightSelectItem.value?.lanUrl) {
    dropdownMenuOptions.push({
      label: t('panelHome.openLanUrl'),
      key: 'openLanUrl',
    })
  }

  if (authStore.visitMode === VisitMode.VISIT_MODE_LOGIN) {
    dropdownMenuOptions.push({
      label: t('common.edit'),
      key: 'edit',
    }, {
      label: t('common.delete'),
      key: 'delete',
    })
  }

  return dropdownMenuOptions
}

onMounted(async () => {
  // 初始化移动端检测
  checkMobile()
  window.addEventListener('resize', checkMobile)
  
  // 更新用户信息
  updateLocalUserInfo()
  getList()

  // 加载Ping Url设置
  try {
    if (!systemPingUrl.value) {
      const res = await getSystemSettings<{pingUrl: string}>(['pingUrl'])
      if (res.code === 0 && res.data && res.data.pingUrl) {
        systemPingUrl.value = res.data.pingUrl
      }
    }
  } catch (error) {
    console.error('获取Ping Url设置失败', error)
  }

  // 更新同步云端配置
  panelState.updatePanelConfigByCloud()

  // 检查是否需要自动获取网络壁纸
  if (panelState.panelConfig.autoNetworkWallpaper) {
    try {
      const apiUrl = panelState.panelConfig.autoNetworkWallpaperApi || 'https://img.xjh.me/random_img.php?return=302&type=bg&ctype=nature'
      panelState.panelConfig.backgroundImageSrc = apiUrl
      panelState.recordState()
    } catch (error) {
      console.error('自动获取网络壁纸失败', error)
    }
  }

  // 设置标题
  if (panelState.panelConfig.logoText)
    setTitle(panelState.panelConfig.logoText)

  // 确保公开模式下始终使用 auto 模式
  if (authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC) {
    panelState.setNetworkMode(PanelStateNetworkModeEnum.auto)
  }

  // 加载书签数据，使用forceRefresh=true确保获取最新排序
  await loadBookmarkTree(false)

  // 启动 SSE 提醒推送（登录状态下）
  if (authStore.visitMode === VisitMode.VISIT_MODE_LOGIN) {
    startRemindSSE()
  }
})

onActivated(() => {
  // Reload bookmark tree when returning from manager to reflect cache changes
  loadBookmarkTree(true)
})
onBeforeRouteUpdate(() => {
  // Reload bookmark tree when route is updated
  loadBookmarkTree(true)
})

// 递归搜索书签树函数已被移除，搜索功能已迁移到SearchBoxWithSuggestions组件

// 前端搜索过滤
function itemFrontEndSearch(keyword?: string) {
  const trimmedKeyword = keyword?.trim() || ''
  if (trimmedKeyword !== '' && panelState.panelConfig.searchBoxSearchIcon) {
    const filteredData = ref<ItemGroup[]>([])
    const lowerCaseKeyword = trimmedKeyword.toLowerCase()

    // 只搜索原有图标（首页书签），不再搜索左侧书签
    for (let i = 0; i < items.value.length; i++) {
      const element = items.value[i].items?.filter((item: Panel.ItemInfo) => {
        // 首先应用网络模式过滤 - wan、auto 和 lan 模式过滤掉 lanOnly 项目
        const shouldFilterLanOnly = panelState.networkMode === PanelStateNetworkModeEnum.wan || 
                                   panelState.networkMode === PanelStateNetworkModeEnum.auto ||
                                   panelState.networkMode === PanelStateNetworkModeEnum.lan
        const networkModeMatch = !shouldFilterLanOnly || item.lanOnly !== 1
        if (!networkModeMatch) return false

        // 然后应用搜索关键词过滤
        return (
          item.title.toLowerCase().includes(lowerCaseKeyword)
          || item.url.toLowerCase().includes(lowerCaseKeyword)
          || item.description?.toLowerCase().includes(lowerCaseKeyword)
        )
      })
      if (element && element.length > 0) {
        filteredData.value.push({ items: element, hoverStatus: false })
      }
    }

    filterItems.value = filteredData.value
  }
  else {
    // 没有搜索关键词时，应用网络模式过滤
    filterItemsByNetworkMode()
  }
}

function handleSetHoverStatus(groupIndex: number, hoverStatus: boolean) {
  if (items.value[groupIndex])
    items.value[groupIndex].hoverStatus = hoverStatus
}

function handleSetSortStatus(itemGroup: ItemGroup, sortStatus: boolean) {
  itemGroup.sortStatus = sortStatus

  // 并未保存排序重新更新数据
  if (!sortStatus) {
    if (itemGroup.id) {
       // Find the index in the original items array to ensure data consistency
       const idx = items.value.findIndex(x => x.id === itemGroup.id)
       if (idx !== -1) {
           updateItemIconGroupByNet(idx, itemGroup.id as number)
       }
    }
  }
}

function handleEditItem(item: Panel.ItemInfo) {
  editItemInfoData.value = item
  editItemInfoShow.value = true
  currentAddItenIconGroupId.value = undefined
}

function handleAddItem(itemIconGroupId?: number) {
  editItemInfoData.value = null
  editItemInfoShow.value = true
  if (itemIconGroupId)
    currentAddItenIconGroupId.value = itemIconGroupId
}

// 网络模式切换处理
function handleChangeNetwork(targetMode: PanelStateNetworkModeEnum) {
  // 只有切换到编辑模式才需要验证密码
  if (targetMode === PanelStateNetworkModeEnum.edit) {
    // 显示密码输入对话框
    const passwordInput = ref('')

    dialog.create({
      title: t('panelHome.verifyPassword'),
      content: () => h('div', { class: 'mt-4' }, [
        h('div', { class: 'mb-2 text-sm text-gray-600 dark:text-gray-400' }, t('panelHome.enterPasswordToSwitchLan')),
        h('input', {
          type: 'password',
          class: 'w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white',
          placeholder: t('common.password'),
          value: passwordInput.value,
          onInput: (e: Event) => {
            passwordInput.value = (e.target as HTMLInputElement).value
          },
          onKeydown: (e: KeyboardEvent) => {
            if (e.key === 'Enter') {
              e.preventDefault()
              // 触发确定按钮
              const positiveButton = document.querySelector('.n-dialog__action button:last-child') as HTMLButtonElement
              if (positiveButton) positiveButton.click()
            }
          }
        })
      ]),
      positiveText: t('common.confirm'),
      negativeText: t('common.cancel'),
      onPositiveClick: async () => {
        if (!passwordInput.value) {
          ms.warning(t('panelHome.passwordRequired'))
          return false // 阻止对话框关闭
        }

        try {
          // 验证密码 - 调用登录接口验证
          const response = await fetch('/api/login', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify({
              username: authStore.userInfo?.username,
              password: passwordInput.value,
            }),
          })

          const result = await response.json()

          if (result.code === 0) {
            // 密码正确,切换模式
            // 切换前自动保存排序状态
            items.value.forEach(group => {
              if (group.sortStatus) {
                handleSaveSort(group)
                group.sortStatus = false // 立即关闭排序状态
              }
            })

            panelState.setNetworkMode(targetMode)
            ms.success('已切换到编辑模式（模式状态仅保存在本地）')
            filterItemsByNetworkMode() // 确保视图更新
            return true
          } else {
            // 密码错误
            ms.error(t('panelHome.passwordIncorrect'))
            return false // 阻止对话框关闭
          }
        } catch (error) {
          console.error('验证密码失败:', error)
          ms.error(t('common.networkError'))
          return false
        }
      },
    })
  } else {
    // 切换前自动保存排序状态
    items.value.forEach(group => {
      if (group.sortStatus) {
        handleSaveSort(group)
        group.sortStatus = false // 立即关闭排序状态
      }
    })

    // 其他模式切换,直接切换
    panelState.setNetworkMode(targetMode)
    filterItemsByNetworkMode() // 确保视图更新
    
    // 显示成功提示
    const modeNames = {
      [PanelStateNetworkModeEnum.auto]: '已切换到自动内外网模式（模式状态仅保存在本地）',
      [PanelStateNetworkModeEnum.wan]: '已切换到外网模式（模式状态仅保存在本地）',
      [PanelStateNetworkModeEnum.lan]: '已切换到内网模式（模式状态仅保存在本地）',
    }
    if (modeNames[targetMode]) {
      ms.success(modeNames[targetMode])
    }
  }
}

// 循环切换网络模式（只切换三种公开模式）
function handleCycleNetworkMode() {
  const modes = [
    PanelStateNetworkModeEnum.auto,   // 自动内外网
    PanelStateNetworkModeEnum.wan,   // 外网模式
    PanelStateNetworkModeEnum.lan,   // 内网模式
  ]
  
  const currentIndex = modes.indexOf(panelState.networkMode)
  // 如果当前是编辑模式，从第一个开始
  const startIndex = currentIndex === -1 ? 0 : currentIndex
  const nextIndex = (startIndex + 1) % modes.length
  const nextMode = modes[nextIndex]
  
  handleChangeNetwork(nextMode)
}

// 获取网络模式按钮文本
function getNetworkModeButtonText() {
  const modeTexts: Record<number, string> = {
    [PanelStateNetworkModeEnum.auto]: '自动内外网',
    [PanelStateNetworkModeEnum.wan]: '外网模式',
    [PanelStateNetworkModeEnum.lan]: '内网模式',
    [PanelStateNetworkModeEnum.edit]: '编辑模式',
  }
  const currentMode = modeTexts[panelState.networkMode] || '未知模式'
  return `当前：${currentMode}`
}

// 获取网络模式按钮图标
function getNetworkModeButtonIcon() {
  const modeIcons: Record<number, string> = {
    [PanelStateNetworkModeEnum.auto]: 'carbon--network-3',
    [PanelStateNetworkModeEnum.wan]: 'mdi:wan',
    [PanelStateNetworkModeEnum.lan]: 'material-symbols:lan-outline-rounded',
    // 编辑模式下不切换图标，保持显示上一个公开模式的图标
    [PanelStateNetworkModeEnum.edit]: 'mdi:wan',
  }
  return modeIcons[panelState.networkMode] || 'mdi:wan'
}
</script>

<template>
  <div class="w-full h-full sun-main">
    <div
      class="cover wallpaper" :style="{
        filter: `blur(${panelState.panelConfig.backgroundBlur}px)`,
        background: `url(${panelState.panelConfig.backgroundImageSrc}) no-repeat`,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }"
    />
    <div class="mask" :style="{ backgroundColor: `rgba(0,0,0,${panelState.panelConfig.backgroundMaskNumber})` }" />
    <div ref="scrollContainerRef" class="absolute w-full h-full overflow-auto">
      <div
        class="p-2.5 mx-auto"
        :style="{
          marginTop: `${panelState.panelConfig.marginTop}%`,
          marginBottom: `${panelState.panelConfig.marginBottom}%`,
          maxWidth: (panelState.panelConfig.maxWidth ?? '1200') + panelState.panelConfig.maxWidthUnit,
        }"
      >
        <!-- 头 -->
        <div class="mx-[auto] w-[80%]">
          <!-- 右上角便签按钮 -->
          <div v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" 
               class="fixed top-4 right-4 z-50 cursor-pointer shadow-[0_0_10px_2px_rgba(0,0,0,0.2)]" 
               style="background-color: #2a2a2a6b; border-radius: 4px; width: 40px; height: 40px; display: flex; align-items: center; justify-content: center;"
               title="便签"
               @click="notepadVisible = true">
            <SvgIcon class="text-white" style="width: 25px; height: 25px;" icon="glyphs--note" />
          </div>
        
          <div class="flex mx-[auto] items-center justify-center text-white">
            <div class="logo cursor-pointer" @click="handleToggleIcons" title="点击显示/隐藏图标">
              <span class="text-2xl md:text-6xl font-bold text-shadow">
                {{ panelState.panelConfig.logoText }}
              </span>
            </div>
            <div class="divider text-base lg:text-2xl mx-[10px]">
              |
            </div>
            <div class="text-shadow">
              <Clock :hide-second="!panelState.panelConfig.clockShowSecond" />
            </div>
          </div>
          <div v-if="panelState.panelConfig.searchBoxShow" class="flex mt-[20px] mx-auto sm:w-full lg:w-[80%]">
            <SearchBoxWithSuggestions @itemSearch="itemFrontEndSearch" />
          </div>
        </div>

        <!-- 应用盒子 -->
        <div v-if="showIcons" :style="{ marginLeft: `${panelState.panelConfig.marginX}px`, marginRight: `${panelState.panelConfig.marginX}px` }">
          <!-- 系统监控状态 -->
          <div
            v-if="panelState.panelConfig.systemMonitorShow
              && ((panelState.panelConfig.systemMonitorPublicVisitModeShow && authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC)
                || authStore.visitMode === VisitMode.VISIT_MODE_LOGIN)"
            class="flex mx-auto"
          >
            <SystemMonitor
              :allow-edit="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN"
              :show-title="panelState.panelConfig.systemMonitorShowTitle"
            />
          </div>

          <!-- 组纵向排列 -->
          <div
            v-for="(itemGroup, itemGroupIndex) in filterItems" :key="itemGroupIndex"
            class="item-list mt-[50px]"
            :class="itemGroup.sortStatus ? 'shadow-2xl border shadow-[0_0_30px_10px_rgba(0,0,0,0.3)]  p-[10px] rounded-2xl' : ''"
            @mouseenter="handleSetHoverStatus(itemGroupIndex, true)"
            @mouseleave="handleSetHoverStatus(itemGroupIndex, false)"
          >
            <!-- 分组标题 -->
            <div class="text-white text-xl font-extrabold mb-[20px] ml-[10px] flex items-center">
              <span class="group-title text-shadow">
                {{ itemGroup.title }}
              </span>
              <div
                v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN && panelState.networkMode === PanelStateNetworkModeEnum.edit"
                class="group-buttons ml-2 delay-100 transition-opacity flex"
              >
                <span class="mr-2 cursor-pointer" :title="t('common.add')" @click="handleAddItem(itemGroup.id)">
                  <SvgIcon class="text-white font-xl" icon="typcn:plus" />
                </span>
                <span class="mr-2 cursor-pointer " :title="t('common.sort')" @click="handleSetSortStatus(itemGroup, !itemGroup.sortStatus)">
                  <SvgIcon class="text-white font-xl" icon="ri:drag-drop-line" />
                </span>
              </div>
            </div>

            <!-- 详情图标 -->
            <div v-if="panelState.panelConfig.iconStyle === PanelPanelConfigStyleEnum.info">
              <div v-if="itemGroup.items">
                <VueDraggable
                  v-model="itemGroup.items" item-key="sort" :animation="300"
                  class="icon-info-box"
                  filter=".not-drag"
                  :disabled="!itemGroup.sortStatus"
                  @end="handleSaveSort(itemGroup)"
                >
                  <div v-for="item, index in itemGroup.items" :key="index" :title="item.description" 
                    @contextmenu="(e) => !isMobile && handleContextMenu(e, itemGroupIndex, item)"
                    @touchstart="(e) => handleTouchStart(e, itemGroupIndex, item)"
                    @touchend="handleTouchEnd()"
                    @touchmove="handleTouchEnd()"
                  >
                    <AppIcon
                      :class="itemGroup.sortStatus ? 'cursor-move' : 'cursor-pointer'"
                      :item-info="item"
                      :icon-text-color="panelState.panelConfig.iconTextColor"
                      :icon-text-info-hide-description="panelState.panelConfig.iconTextInfoHideDescription || false"
                      :icon-text-icon-hide-title="panelState.panelConfig.iconTextIconHideTitle || false"
                      :style="0"
                      @click="handleItemClick(itemGroupIndex, item)"
                    />
                  </div>

                  <div v-if="itemGroup.items.length === 0" class="not-drag">
                    <AppIcon
                      :class="itemGroup.sortStatus ? 'cursor-move' : 'cursor-pointer'"
                      :item-info="{ icon: { itemType: 3, text: 'subway:add' }, title: t('common.add'), url: '', openMethod: 0 }"
                      :icon-text-color="panelState.panelConfig.iconTextColor"
                      :icon-text-info-hide-description="panelState.panelConfig.iconTextInfoHideDescription || false"
                      :icon-text-icon-hide-title="panelState.panelConfig.iconTextIconHideTitle || false"
                      :style="0"
                      @click="handleAddItem(itemGroup.id)"
                    />
                  </div>
                </VueDraggable>
              </div>
            </div>

            <!-- APP图标宫型盒子 -->
            <div v-if="panelState.panelConfig.iconStyle === PanelPanelConfigStyleEnum.icon">
              <div v-if="itemGroup.items">
                <VueDraggable
                  v-model="itemGroup.items" item-key="sort" :animation="300"
                  class="icon-small-box"

                  filter=".not-drag"
                  :disabled="!itemGroup.sortStatus"
                  @end="handleSaveSort(itemGroup)"
                >
                  <div v-for="item, index in itemGroup.items" :key="index" :title="item.description" 
                    @contextmenu="(e) => !isMobile && handleContextMenu(e, itemGroupIndex, item)"
                    @touchstart="(e) => handleTouchStart(e, itemGroupIndex, item)"
                    @touchend="handleTouchEnd()"
                    @touchmove="handleTouchEnd()"
                  >
                    <AppIcon
                      :class="itemGroup.sortStatus ? 'cursor-move' : 'cursor-pointer'"
                      :item-info="item"
                      :icon-text-color="panelState.panelConfig.iconTextColor"
                      :icon-text-info-hide-description="!panelState.panelConfig.iconTextInfoHideDescription"
                      :icon-text-icon-hide-title="panelState.panelConfig.iconTextIconHideTitle || false"
                      :style="1"
                      @click="handleItemClick(itemGroupIndex, item)"
                    />
                  </div>

                  <div v-if="itemGroup.items.length === 0" class="not-drag">
                    <AppIcon
                      class="cursor-pointer"
                      :item-info="{ icon: { itemType: 3, text: 'subway:add' }, title: $t('common.add'), url: '', openMethod: 0 }"
                      :icon-text-color="panelState.panelConfig.iconTextColor"
                      :icon-text-info-hide-description="!panelState.panelConfig.iconTextInfoHideDescription"
                      :icon-text-icon-hide-title="panelState.panelConfig.iconTextIconHideTitle || false"
                      :style="1"
                      @click="handleAddItem(itemGroup.id)"
                    />
                  </div>
                </vuedraggable>
              </div>
            </div>

            <!-- 编辑栏 -->

          </div>
        </div>
        <div class="mt-5 footer" v-html="panelState.panelConfig.footerHtml" />
      </div>
    </div>

    <!-- 左侧分组导航条 -->
    <Transition name="fade">
      <div 
        v-if="showGroupNav && filterItems.length > 0" 
        class="group-nav-sidebar"
      >
        <div class="group-nav-line">
          <div
            v-for="(group, index) in filterItems"
            :key="index"
            class="group-nav-dot"
            :class="{ 'group-nav-dot-active': currentGroupIndex === index }"
            @click="scrollToGroup(index)"
          >
            <span class="group-nav-tooltip">{{ group.title }}</span>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 右键菜单 -->
    <NDropdown
      placement="bottom-start" trigger="manual" :x="dropdownMenuX" :y="dropdownMenuY"
      :options="getDropdownMenuOptions()" :show="dropdownShow" :on-clickoutside="onClickoutside" @select="handleRightMenuSelect"
    />

    <!-- 悬浮按钮 -->
    <div class="fixed-element shadow-[0_0_10px_2px_rgba(0,0,0,0.2)]">
      <NButtonGroup vertical>
        <!-- 网络模式切换按钮 - 循环切换三种公开模式 -->
        <NButton
          v-if="panelState.panelConfig.netModeChangeButtonShow && authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" 
          color="#2a2a2a6b"
          :title="getNetworkModeButtonText()" 
          @click="handleCycleNetworkMode()"
        >
          <template #icon>
            <SvgIcon class="text-white font-xl" :icon="getNetworkModeButtonIcon()" />
          </template>
        </NButton>

        <!-- 编辑模式切换按钮 - 独立显示 -->
        <NButton
          v-if="panelState.panelConfig.secretModeButtonShow && authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" 
          color="#2a2a2a6b"
          title="编辑模式"
          @click="handleChangeNetwork(PanelStateNetworkModeEnum.edit)"
        >
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="boxicons--edit" />
          </template>
        </NButton>

        <NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" color="#2a2a2a6b" @click="settingModalShow = !settingModalShow">
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="majesticons-applications" />
          </template>
        </NButton>

        <NButton color="#2a2a2a6b" :title="t('panelHome.refreshData')" @click="handleRefreshData">
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="shuaxin" />
          </template>
        </NButton>

        <NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC" color="#2a2a2a6b" :title="$t('panelHome.goToLogin')" @click="router.push('/login')">
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="material-symbols:account-circle" />
          </template>
        </NButton>
      </NButtonGroup>

      <AppStarter v-model:visible="settingModalShow" />
      <NotePad 
        ref="notepadInstance" 
        v-model:visible="notepadVisible" 
      />
      <!-- <Setting v-model:visible="settingModalShow" /> -->
    </div>

    <NBackTop
      :listen-to="() => scrollContainerRef"
      :right="10"
      :bottom="10"
      style="background-color:transparent;border: none;box-shadow: none;"
    >
      <div class="shadow-[0_0_10px_2px_rgba(0,0,0,0.2)]">
        <NButton color="#2a2a2a6b">
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="icon-park-outline:to-top" />
          </template>
        </NButton>
      </div>
    </NBackTop>

    <EditItem v-model:visible="editItemInfoShow" :item-info="editItemInfoData" :item-group-id="currentAddItenIconGroupId" @done="handleEditSuccess" />

    <!-- 弹窗 -->
    <NModal
      v-model:show="windowShow" :mask-closable="false" preset="card"
      style="max-width: 1000px;height: 600px;border-radius: 1rem;" :bordered="true" size="small" role="dialog"
      aria-modal="true"
    >
      <template #header>
        <div class="flex items-center">
          <span class="mr-[20px]">
            {{ windowTitle }}
          </span>

          <NSpin v-if="windowIframeIsLoad" size="small" />
        </div>
      </template>
      <div class="w-full h-full rounded-2xl overflow-hidden border dark:border-zinc-700">
        <div v-if="windowIframeIsLoad" class="flex flex-col p-5">
          <NSkeleton height="50px" width="100%" class="rounded-lg" />
          <NSkeleton height="180px" width="100%" class="mt-[20px] rounded-lg" />
          <NSkeleton height="180px" width="100%" class="mt-[20px] rounded-lg" />
        </div>
        <iframe
          v-show="!windowIframeIsLoad" id="windowIframeId" ref="windowIframeRef" :src="windowSrc"
          class="w-full h-full" frameborder="0" @load="handWindowIframeIdLoad"
        />
      </div>
    </NModal>

    <!-- 现代化提醒通知卡片 -->
    <Teleport to="body">
      <TransitionGroup name="notification-slide" tag="div" class="remind-notifications-container">
        <div
          v-for="notification in remindNotifications"
          :key="notification.id"
          v-show="notification.visible"
          class="remind-notification-card"
        >
          <div class="notification-header">
            <div class="notification-icon">
              <SvgIcon icon="boxicons--bell-ring" />
            </div>
            <div class="notification-title">提醒</div>
            <button class="notification-close" @click="closeNotification(notification.id)">
              <SvgIcon icon="material-symbols--close" />
            </button>
          </div>
          
          <div class="notification-content">
            <div class="notification-note-title">
              <SvgIcon icon="note" class="note-icon" />
              {{ notification.title }}
            </div>
            <div class="notification-time">
              <SvgIconOnline icon="mdi:clock-outline" class="time-icon" />
              {{ notification.time }}
            </div>
          </div>
          
          <div class="notification-actions">
            <button class="action-btn view-btn" @click="viewNotepad(notification.noteId, notification.id)">
              <SvgIconOnline icon="mdi:eye-outline" />
              查看
            </button>
            <button class="action-btn close-btn" @click="closeNotification(notification.id)">
              <SvgIconOnline icon="mdi:check" />
              知道了
            </button>
          </div>
        </div>
      </TransitionGroup>
    </Teleport>
  </div>
</template>

<style>
body,
html {
  overflow: hidden;
  background-color: rgb(54, 54, 54);
}
</style>

<style scoped>
.mask {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.sun-main {
  user-select: none;
}

.cover {
  position: absolute;
  width: 100%;
  height: 100%;
  overflow: hidden;
  /* background: url(@/assets/start_sky.jpg) no-repeat; */

  transform: scale(1.05);
}

.text-shadow {
  text-shadow: 2px 2px 50px rgb(0, 0, 0);
}

.app-icon-text-shadow {
  text-shadow: 2px 2px 5px rgb(0, 0, 0);
}

.fixed-element {
  position: fixed;
  /* 将元素固定在屏幕上 */
  right: 10px;
  /* 距离屏幕顶部的距离 */
  bottom: 50px;
  /* 距离屏幕左侧的距离 */
}

.icon-info-box {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(200px, 100%), 1fr));
  gap: 2px;

}

.icon-small-box {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100px, 100%), 1fr));
  gap: 2px;

}

/* 响应式图标块布局 */
@media (max-width: 1024px) {
  .icon-info-box {
    grid-template-columns: repeat(auto-fill, minmax(min(160px, 100%), 1fr));
    gap: 14px;
  }

  .icon-small-box {
    grid-template-columns: repeat(auto-fill, minmax(min(85px, 100%), 1fr));
    gap: 14px;
  }
}

/* 响应式图标块布局 - 继续使用grid布局，但减小最小宽度 */
@media (max-width: 768px) {
  .icon-info-box {
    grid-template-columns: repeat(auto-fill, minmax(min(100px, 100%), 1fr));
    gap: 12px;
  }

  .icon-small-box {
    grid-template-columns: repeat(auto-fill, minmax(min(60px, 100%), 1fr));
    gap: 12px;
  }
}

@media (max-width: 480px) {
  .icon-info-box {
    grid-template-columns: repeat(auto-fill, minmax(min(100px, 100%), 1fr));
    gap: 10px;
  }

  .icon-small-box {
    grid-template-columns: repeat(auto-fill, minmax(min(60px, 100%), 1fr));
    gap: 10px;
  }
}

@media (max-width: 360px) {
  .icon-info-box {
    grid-template-columns: repeat(auto-fill, minmax(min(100px, 100%), 1fr));
    gap: 8px;
  }

  .icon-small-box {
    grid-template-columns: repeat(auto-fill, minmax(min(60px, 100%), 1fr));
    gap: 8px;
  }
}



/* 优化条状按钮阴影 */
/* 优化条状按钮阴影 - 已移除，避免污染全局 .fixed 类 */


:deep(.no-focus-outline:focus) {
  box-shadow: none !important;
}

.no-tap-highlight {
  -webkit-tap-highlight-color: transparent !important;
  outline: none !important;
}

/* 防止 iOS 长按图片时显示系统菜单 */
.app-icon img,
.app-icon-info-icon img,
.app-icon-small-icon img {
  -webkit-touch-callout: none !important;
  -webkit-user-select: none !important;
  user-select: none !important;
  pointer-events: none;
}

/* 移动端长按样式优化 */
@media (max-width: 768px) {
  /* 禁用默认的长按行为 */
  .app-icon,
  .app-icon-info,
  .app-icon-small {
    -webkit-touch-callout: none !important;
    -webkit-user-select: none !important;
    user-select: none !important;
  }
  
  /* 确保图标容器在长按时不响应默认行为 */
  .icon-info-box > div,
  .icon-small-box > div {
    touch-action: manipulation;
  }
}

/* 左侧分组导航条样式 */
.group-nav-sidebar {
  position: fixed;
  left: 15px;
  top: 50%;
  transform: translateY(-50%);
  z-index: 1000;
}

.group-nav-line {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 32px;
  padding: 16px 0;
}

/* 竖线 */
.group-nav-line::before {
  content: '';
  position: absolute;
  left: 50%;
  top: 0;
  bottom: 0;
  width: 2px;
  background: rgba(255, 255, 255, 0.15);
  transform: translateX(-50%);
  border-radius: 1px;
}

.group-nav-dot {
  position: relative;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.3);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  margin: 0 auto;
  z-index: 1;
}

.group-nav-dot:hover {
  background: rgba(255, 255, 255, 0.6);
  transform: scale(1.15);
}

.group-nav-dot-active {
  background: #fff;
  box-shadow: 0 0 10px rgba(255, 255, 255, 0.8), 0 0 20px rgba(255, 255, 255, 0.4);
  transform: scale(1.3);
}

/* 悬停提示 */
.group-nav-tooltip {
  position: absolute;
  left: 20px;
  top: 50%;
  transform: translateY(-50%);
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(8px);
  color: rgba(255, 255, 255, 0.95);
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
  opacity: 0;
  visibility: hidden;
  transition: all 0.2s ease;
  pointer-events: none;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.1);
  letter-spacing: 0.3px;
}

.group-nav-dot:hover .group-nav-tooltip {
  opacity: 1;
  visibility: visible;
  left: 24px;
}

/* 淡入淡出动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 移动端隐藏导航条 */
@media (max-width: 768px) {
  .group-nav-sidebar {
    display: none;
  }
}

/* 隐藏左侧悬停检测区域 */
.left-hover-area {
  pointer-events: none !important;
}

/* ========== 现代化提醒通知卡片样式 ========== */
.remind-notifications-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-width: 400px;
  pointer-events: none; /* 容器不阻挡点击 */
}

.remind-notification-card {
  pointer-events: auto; /* 卡片可以交互 */
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95) 0%, rgba(248, 249, 250, 0.95) 100%);
  backdrop-filter: blur(20px);
  border-radius: 16px;
  box-shadow: 
    0 8px 32px rgba(0, 0, 0, 0.12),
    0 2px 8px rgba(0, 0, 0, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.3);
  padding: 16px;
  min-width: 320px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  animation: slideInRight 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.remind-notification-card:hover {
  transform: translateY(-2px);
  box-shadow: 
    0 12px 40px rgba(0, 0, 0, 0.15),
    0 4px 12px rgba(0, 0, 0, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.notification-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

.notification-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #667eea;
  animation: bellShake 2s ease-in-out infinite;
  flex-shrink: 0;
}

.notification-icon :deep(svg),
.notification-icon :deep(.svg-icon),
.notification-icon :deep(iconify-icon),
.notification-icon :deep(span) {
  width: 28px !important;
  height: 28px !important;
  color: #667eea !important;
  display: block;
  font-size: 28px !important;
}

@keyframes bellShake {
  0%, 100% { transform: rotate(0deg); }
  10%, 30%, 50% { transform: rotate(-10deg); }
  20%, 40%, 60% { transform: rotate(10deg); }
}

.notification-title {
  flex: 1;
  font-size: 16px;
  font-weight: 600;
  color: #1a1a1a;
  letter-spacing: 0.3px;
}

.notification-close {
  width: 32px;
  height: 32px;
  border: none;
  background: rgba(0, 0, 0, 0.05);
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #666;
  transition: all 0.2s;
  padding: 0;
  flex-shrink: 0;
}

.notification-close :deep(svg),
.notification-close :deep(.svg-icon),
.notification-close :deep(iconify-icon),
.notification-close :deep(span) {
  width: 18px !important;
  height: 18px !important;
  color: #666 !important;
  display: block;
  fill: currentColor !important;
  font-size: 18px !important;
}

.notification-close:hover {
  background: rgba(0, 0, 0, 0.1);
  color: #333;
  transform: scale(1.1);
}

.notification-close:hover :deep(svg),
.notification-close:hover :deep(.svg-icon) {
  color: #333 !important;
}

.notification-content {
  margin-bottom: 14px;
}

.notification-note-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 500;
  color: #2c3e50;
  margin-bottom: 8px;
  line-height: 1.5;
}

.note-icon {
  color: #667eea;
  font-size: 18px;
  flex-shrink: 0;
}

.notification-time {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #7f8c8d;
  padding-left: 26px;
}

.time-icon {
  font-size: 14px;
  opacity: 0.7;
}

.notification-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}

.action-btn {
  flex: 1;
  padding: 10px 16px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.view-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.view-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.4);
}

.view-btn:active {
  transform: translateY(0);
}

.close-btn {
  background: rgba(0, 0, 0, 0.06);
  color: #555;
}

.close-btn:hover {
  background: rgba(0, 0, 0, 0.1);
  color: #333;
  transform: translateY(-1px);
}

.close-btn:active {
  transform: translateY(0);
}

/* 滑入动画 */
@keyframes slideInRight {
  from {
    opacity: 0;
    transform: translateX(100px) scale(0.9);
  }
  to {
    opacity: 1;
    transform: translateX(0) scale(1);
  }
}

.notification-slide-enter-active,
.notification-slide-leave-active {
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.notification-slide-enter-from {
  opacity: 0;
  transform: translateX(100px) scale(0.9);
}

.notification-slide-leave-to {
  opacity: 0;
  transform: translateX(100px) scale(0.8);
}

/* 移动端适配 */
@media (max-width: 768px) {
  .remind-notifications-container {
    top: 10px;
    right: 10px;
    left: 10px;
    max-width: none;
  }
  
  .remind-notification-card {
    min-width: auto;
    width: 100%;
  }
}
</style>

