<template>
  <Teleport to="body">
    <transition name="fade">
      <div 
        v-if="visible" 
        class="notepad-overlay"
        @click.self="handleClose"
      >
        <div 
          ref="notepadRef"
          class="notepad-container"
          :style="{ left: x + 'px', top: y + 'px' }"
        >
          <!-- 左侧列表 -->
          <div class="notepad-sidebar">
            <!-- 顶部标题栏 -->
            <div class="sidebar-header">
              <h2 class="sidebar-title">备忘录</h2>
              <div class="header-actions">
                <SvgIcon 
                  class="action-icon" 
                  icon="mdi:chevron-down" 
                  @click="showSortMenu = !showSortMenu"
                />
              </div>
            </div>

            <!-- 搜索框 -->
            <div class="search-box">
              <SvgIcon class="search-icon" icon="mdi:magnify" />
              <input 
                v-model="searchKeyword"
                type="text" 
                placeholder="search"
                class="search-input"
              />
            </div>

            <!-- 便签列表 -->
            <div class="note-list">
              <div 
                v-for="note in filteredNotes" 
                :key="note.id"
                class="note-item"
                :class="{ active: currentNote.id === note.id }"
                @click="selectNote(note)"
              >
                <div class="note-item-header">
                  <span class="note-item-title">{{ note.title || '无标题' }}</span>
                  <!-- 提醒铃铛图标 -->
                  <SvgIconOnline 
                    v-if="note.remindTime && note.remindStatus !== 2" 
                    class="remind-icon active"
                    icon="mdi:bell-outline"
                  />
                  <SvgIconOnline 
                    v-else
                    class="remind-icon"
                    icon="mdi:bell-off-outline"
                  />
                  <!-- 删除按钮 -->
                  <SvgIcon 
                    class="delete-icon" 
                    icon="material-symbols--delete-outline"
                    @click.stop="deleteNote(note)"
                  />
                </div>
                <div class="note-item-time">{{ formatDate(note.updatedAt) }}</div>
              </div>
            </div>

            <!-- 底部新建按钮 -->
            <div class="sidebar-footer">
              <button class="new-note-btn" @click="createNew">
                <SvgIcon icon="pajamas--doc-new" />
              </button>
            </div>
          </div>

          <!-- 右侧编辑区 -->
          <div class="notepad-editor">
            <!-- 编辑器顶部 -->
            <div class="editor-header">
              <h1 class="editor-title">{{ currentNote.title || '无标题' }}</h1>
              <div class="editor-actions">
                <!-- 关闭按钮 -->
                <SvgIcon class="action-icon" icon="material-symbols--close" @click="handleClose" />
              </div>
            </div>

            <!-- 编辑内容区 -->
            <div 
              ref="editorRef"
              class="editor-content"
              contenteditable="true"
              @input="handleInput"
              @paste="handlePaste"
              placeholder="请输入笔记内容"
            ></div>

            <!-- 底部信息 -->
            <div class="editor-footer">
              <span class="footer-text">
                最后编辑：{{ formatFullDate(currentNote.updatedAt) }}，创建：{{ formatFullDate(currentNote.createdAt) }}
              </span>
            </div>

            <!-- 提醒时间选择器 -->
            <transition name="slide-up">
              <div v-if="showRemindPicker" class="remind-picker">
                <div class="remind-picker-header">
                  <span>选择提醒时间：</span>
                  <button class="close-picker-btn" @click="handleCloseRemindPicker" title="关闭">
                    <SvgIcon icon="material-symbols--close" />
                  </button>
                </div>
                
                <!-- 日期时间选择器 -->
                <NDatePicker
                  v-model:value="remindTimestamp"
                  type="datetime"
                  format="yyyy-MM-dd HH:mm"
                  :time-picker-props="{ format: 'HH:mm' }"
                  placeholder="请选择提醒日期和时间"
                  clearable
                  :actions="['confirm']"
                  placement="bottom-start"
                  class="w-full mb-4"
                />
                
                <!-- 重复选项 -->
                <div class="remind-repeat-section">
                  <div class="repeat-label">重复：</div>
                  <select 
                    v-model="currentRepeatType" 
                    class="repeat-select"
                    @change="handleRepeatChange"
                  >
                    <option value="none">不重复</option>
                    <option value="daily">每天</option>
                    <option value="weekly">每周</option>
                    <option value="monthly">每月</option>
                    <option value="yearly">每年</option>
                  </select>
                </div>
                
                <!-- 提前提醒选项 -->
                <transition name="slide-down">
                  <div v-if="showAdvanceDays" class="remind-advance-section">
                    <div class="advance-label">提前提醒：</div>
                    <select 
                      v-model="currentAdvanceDays" 
                      class="advance-select"
                      @change="handleAdvanceChange"
                    >
                      <option v-for="opt in advanceDaysOptions" :key="opt.value" :value="opt.value">
                        {{ opt.label }}
                      </option>
                    </select>
                  </div>
                </transition>
                
                <!-- 实际提醒时间提示（始终显示） -->
                <div v-if="currentNote.remindTime" class="actual-remind-hint">
                  💡 原始选择：{{ formatRemindTime(currentNote.remindBaseTime || currentNote.remindTime) }} → 下次实际提醒：{{ nextActualRemindTime }}
                </div>
                
                <div class="remind-picker-info">
                  <div v-if="currentNote.remindStatus === 0 && currentNote.remindTime">
                    已设置：{{ formatRemindTime(currentNote.remindBaseTime || currentNote.remindTime) }}
                    <span v-if="currentNote.remindRepeat && currentNote.remindRepeat !== 'none'" class="repeat-badge">
                      {{ getRepeatText(currentNote.remindRepeat) }}
                    </span>
                    <span v-if="currentNote.remindAdvanceDays && currentNote.remindAdvanceDays > 0" class="advance-badge">
                      提前{{ currentNote.remindAdvanceDays }}天
                    </span>
                  </div>
                  <div v-else>未设置提醒</div>
                </div>
                
                <!-- 确认按钮 -->
                <div class="remind-picker-actions">
                  <button class="confirm-btn" @click="handleConfirmRemind">
                    确认
                  </button>
                </div>
              </div>
            </transition>
          </div>

          <!-- 悬浮提醒按钮 -->
          <div 
            class="remind-float-btn"
            :class="{ active: currentNote.remindTime && currentNote.remindStatus !== 2 }"
            :title="currentNote.remindTime ? '已设置提醒: ' + formatRemindTime(currentNote.remindBaseTime || currentNote.remindTime) : '设置提醒'"
            @click="handleOpenRemindPicker"
          >
            <SvgIconOnline :icon="currentNote.remindTime && currentNote.remindStatus !== 2 ? 'mdi:bell-outline' : 'mdi:bell-off-outline'" />
            <span v-if="currentNote.remindTime && currentNote.remindStatus !== 2" class="remind-dot"></span>
          </div>
        </div>
      </div>
    </transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch, h } from 'vue'
import { SvgIcon, SvgIconOnline } from '@/components/common'
import { useMessage, useDialog, NDatePicker } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useDraggable, useDebounceFn, useStorage } from '@vueuse/core'
import { 
    getNotepadList, 
    saveNotepadContent,
    deleteNotepad,
    type NotepadInfo 
} from '@/api/panel/notepad'
import { useAuthStore } from '@/store/modules/auth'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', visible: boolean): void
  (e: 'remindStatusChanged', noteId: number): void  // 提醒状态变化事件
}>()

const { t } = useI18n()
const message = useMessage()
const dialog = useDialog()
const authStore = useAuthStore()
const editorRef = ref<HTMLDivElement | null>(null)
const notepadRef = ref<HTMLElement | null>(null)

// 状态
const currentNote = useStorage<Partial<NotepadInfo>>('sun-panel-notepad-current', { id: 0, title: '', content: '' })
const noteList = useStorage<NotepadInfo[]>('sun-panel-notepad-list', [])
const searchKeyword = ref('')
const showSortMenu = ref(false)
const showRemindPicker = ref(false)
const currentRepeatType = ref<string>('none') // 当前选择的重复类型
const currentAdvanceDays = ref<number>(0) // 当前提前提醒天数

// 窗口拖拽
const { x, y } = useDraggable(notepadRef, {
  initialValue: { x: (window.innerWidth - 1000) / 2, y: (window.innerHeight - 600) / 2 }
})

// 初始化
onMounted(async () => {
    if (noteList.value.length === 0) {
        await loadList()
    }
})

// 加载列表
const loadList = async () => {
    if (!authStore.token) return
    try {
        const res = await getNotepadList()
        if (res.code === 0) {
            noteList.value = res.data || []
        }
    } catch (e) {
        console.error('Load list error', e)
    }
}

// 生成标题
const generateTitle = (textContent?: string) => {
    if (editorRef.value) {
        const h1 = editorRef.value.querySelector('h1')
        if (h1 && h1.innerText.trim()) {
            return h1.innerText.trim()
        }
    }
    const text = textContent !== undefined ? textContent : (editorRef.value?.innerText.trim() || '')
    if (text) {
        return text.substring(0, 10)
    }
    if (currentNote.value.id) {
        return `便签${currentNote.value.id}` 
    }
    return `便签${noteList.value.length + 1}`
}

// 过滤列表
const filteredNotes = computed(() => {
    if (!searchKeyword.value) return noteList.value
    const keyword = searchKeyword.value.toLowerCase()
    return noteList.value.filter(note => 
        note.title?.toLowerCase().includes(keyword) || 
        note.content?.toLowerCase().includes(keyword)
    )
})

// 格式化日期
const formatDate = (dateStr?: string) => {
    if (!dateStr) return ''
    const date = new Date(dateStr)
    return `${date.getFullYear()}/${String(date.getMonth() + 1).padStart(2, '0')}/${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

const formatFullDate = (dateStr?: string) => {
    if (!dateStr) return ''
    const date = new Date(dateStr)
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}:${String(date.getSeconds()).padStart(2, '0')}`
}

// 格式化提醒时间（不显示秒）
const formatRemindTime = (dateStr?: string) => {
    if (!dateStr) return ''
    const date = new Date(dateStr)
    return `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

// 格式化本地时间字符串（避免toISOString导致时区偏差）
const formatLocalDateTime = (date: Date): string => {
    const y = date.getFullYear()
    const m = String(date.getMonth() + 1).padStart(2, '0')
    const d = String(date.getDate()).padStart(2, '0')
    const h = String(date.getHours()).padStart(2, '0')
    const mi = String(date.getMinutes()).padStart(2, '0')
    const s = String(date.getSeconds()).padStart(2, '0')
    return `${y}-${m}-${d}T${h}:${mi}:${s}`
}

// 输入处理
const handleInput = () => {
    if (!editorRef.value) return
    const text = editorRef.value.innerText.trim()
    currentNote.value.title = generateTitle(text)
    saveContent()
}

// 粘贴处理
const handlePaste = (e: ClipboardEvent) => {
    e.preventDefault()
    const text = e.clipboardData?.getData('text/plain') || ''
    document.execCommand('insertText', false, text)
}

// 核心保存逻辑
const handleSave = async () => {
    if (editorRef.value) {
        try {
            const content = editorRef.value.innerHTML
            const text = editorRef.value.innerText.trim()
            const title = generateTitle(text)
            const saveId = currentNote.value.id || 0
            
            const res = await saveNotepadContent({ 
                id: saveId,
                title: title,
                content: content,
                remindTime: currentNote.value.remindTime || null,
                remindStatus: currentNote.value.remindStatus || 0,
                remindRepeat: currentNote.value.remindRepeat || 'none',
                remindAdvanceDays: currentNote.value.remindAdvanceDays || 0
            })
            
            if (res.code === 0) {
                if (currentNote.value.id === saveId) {
                    currentNote.value = res.data
                }
                await loadList()
            }
        } catch (error) {
            console.error('Save notepad error:', error)
        }
    }
}

// 保存内容（防抖）
const saveContent = useDebounceFn(handleSave, 1000)

// 切换便签
const selectNote = (note: NotepadInfo) => {
    console.log('[NotePad] selectNote:', note.title)
    currentNote.value = { ...note }
    currentRepeatType.value = note.remindRepeat || 'none'
    currentAdvanceDays.value = note.remindAdvanceDays || 0
    if (editorRef.value) {
        editorRef.value.innerHTML = note.content || ''
        nextTick(() => {
            bindFileDownloadEvents()
        })
    }
}

// 新建便签
const createNew = () => {
    currentNote.value = { id: 0, title: `便签${noteList.value.length + 1}`, content: '' }
    currentRepeatType.value = 'none'
    currentAdvanceDays.value = 0
    if (editorRef.value) {
        editorRef.value.innerHTML = ''
        editorRef.value.focus()
    }
}

// 关闭
const handleClose = () => {
    handleSave()
    emit('update:visible', false)
}

// 绑定文件下载事件
const bindFileDownloadEvents = () => {
    if (!editorRef.value) return
    const fileLinks = editorRef.value.querySelectorAll('.file-attachment')
    fileLinks.forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault()
            const url = link.getAttribute('href')
            const filename = link.getAttribute('data-filename')
            if (url && filename) {
                downloadFile(url, filename)
            }
        })
    })
}

// 下载文件
const downloadFile = async (url: string, filename: string) => {
    try {
        const response = await fetch(url)
        if (!response.ok) throw new Error('Network response was not ok')
        const blob = await response.blob()
        const urlCreator = window.URL || window.webkitURL
        const objectUrl = urlCreator.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = objectUrl
        link.download = filename
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        setTimeout(() => urlCreator.revokeObjectURL(objectUrl), 100)
    } catch (error) {
        message.error(t('notepad.saveFailed'))
    }
}

// 统一的时间戳变量
const remindTimestamp = ref<number | null>(null)

// 关闭提醒选择器（不保存）
const handleCloseRemindPicker = () => {
    showRemindPicker.value = false
}

// 点击"完成"按钮时保存所有设置
const handleConfirmRemind = () => {
    console.log('[NotePad] ========== handleConfirmRemind ==========')
    console.log('[NotePad] remindTimestamp:', remindTimestamp.value)
    console.log('[NotePad] currentRepeatType:', currentRepeatType.value)
    console.log('[NotePad] currentAdvanceDays:', currentAdvanceDays.value)
    
    if (remindTimestamp.value) {
        // 有设置时间，保存提醒
        setRemind(remindTimestamp.value, true) // autoClose=true
    } else {
        // 没有设置时间，清除提醒
        setRemind(null, true)
    }
    console.log('[NotePad] ==========================================')
}

const setRemind = async (timestamp: number | null, autoClose: boolean = false) => {
    console.log('[NotePad] ========== setRemind ==========')
    console.log('[NotePad] timestamp:', timestamp)
    console.log('[NotePad] autoClose:', autoClose)
    console.log('[NotePad] currentNote.value.id:', currentNote.value.id)
    
    if (!currentNote.value.id) {
        console.log('[NotePad] 未找到便签 ID，提示用户')
        message.warning('请先保存便签后再设置提醒')
        return
    }
    try {
        const remindTime = timestamp ? formatLocalDateTime(new Date(timestamp)) : null
        console.log('[NotePad] formatLocalDateTime 结果:', remindTime)
        
        console.log('[NotePad] 准备保存的数据:', {
            id: currentNote.value.id,
            title: currentNote.value.title,
            remindTime: remindTime,
            remindStatus: remindTime ? 0 : 1,
            remindRepeat: currentRepeatType.value,
            remindAdvanceDays: currentAdvanceDays.value
        })
        
        await saveNotepadContent({
            id: currentNote.value.id,
            title: currentNote.value.title || '',
            content: currentNote.value.content || '',
            remindTime: remindTime,
            remindStatus: remindTime ? 0 : 1,
            remindRepeat: currentRepeatType.value,
            remindAdvanceDays: currentAdvanceDays.value
        })
        
        console.log('[NotePad] 保存成功，更新本地状态')
        currentNote.value.remindTime = remindTime || undefined
        currentNote.value.remindStatus = remindTime ? 0 : 1
        currentNote.value.remindRepeat = currentRepeatType.value
        currentNote.value.remindAdvanceDays = currentAdvanceDays.value
        await loadList()
        
        // 通知父组件清除已提醒记录（无论是设置新提醒还是取消提醒）
        emit('remindStatusChanged', currentNote.value.id)
        
        if (remindTime) {
            const repeatText = getRepeatText(currentRepeatType.value)
            const advanceText = currentAdvanceDays.value > 0 ? ` [提前${currentAdvanceDays.value}天]` : ''
            console.log('[NotePad] 显示成功消息:', `已设置提醒：${formatRemindTime(remindTime)} ${repeatText !== '不重复' ? '(' + repeatText + ')' : ''}${advanceText}`)
            message.success(`已设置提醒：${formatRemindTime(remindTime)} ${repeatText !== '不重复' ? '(' + repeatText + ')' : ''}${advanceText}`)
        } else {
            console.log('[NotePad] 显示取消提醒消息')
            message.success('已取消提醒')
        }
        // 只在明确需要时才关闭选择器
        if (autoClose) {
            showRemindPicker.value = false
        }
    } catch (e) {
        console.error('[NotePad] 设置提醒失败:', e)
        message.error('设置提醒失败')
    }
    console.log('[NotePad] ==========================================')
}


// 打开提醒选择器（确保状态同步）
const handleOpenRemindPicker = () => {
    console.log('[NotePad] ========== 打开提醒选择器 ==========')
    console.log('[NotePad] currentNote.value:', JSON.stringify(currentNote.value, null, 2))
    console.log('[NotePad] currentNote.remindAdvanceDays:', currentNote.value.remindAdvanceDays)
    console.log('[NotePad] currentNote.remindTime:', currentNote.value.remindTime)
    console.log('[NotePad] currentAdvanceDays before:', currentAdvanceDays.value)
    
    // 从 currentNote 重新读取 remindAdvanceDays 状态
    if (currentNote.value.remindAdvanceDays !== undefined) {
        currentAdvanceDays.value = currentNote.value.remindAdvanceDays
        console.log('[NotePad] 从 currentNote 同步 remindAdvanceDays:', currentAdvanceDays.value)
    }
    
    // 从 currentNote.remindTime 解析日期时间
    if (currentNote.value.remindTime) {
        const date = new Date(currentNote.value.remindTime)
        if (!isNaN(date.getTime())) {
            remindTimestamp.value = date.getTime()
        }
    } else {
        // 清除日期时间选择器
        remindTimestamp.value = null
    }
    
    showRemindPicker.value = !showRemindPicker.value
    console.log('[NotePad] showRemindPicker:', showRemindPicker.value)
    console.log('[NotePad] currentAdvanceDays after:', currentAdvanceDays.value)
    console.log('[NotePad] ==========================================')
}

// 处理重复类型变化
const handleRepeatChange = () => {
    // 只更新本地状态，不立即保存，等待用户点击"完成"
    currentNote.value.remindRepeat = currentRepeatType.value
}

// 处理提前天数变化
const handleAdvanceChange = () => {
    // 只更新本地状态，不立即保存，等待用户点击“完成”
    currentNote.value.remindAdvanceDays = currentAdvanceDays.value
}

// 获取重复类型的文本
const getRepeatText = (repeatType?: string) => {
    const map: Record<string, string> = {
        'none': '不重复',
        'daily': '每天',
        'weekly': '每周',
        'monthly': '每月',
        'yearly': '每年'
    }
    return map[repeatType || 'none'] || '不重复'
}

// 是否显示提前提醒选项
const showAdvanceDays = computed(() => {
    return currentRepeatType.value !== 'none' && currentRepeatType.value !== 'daily'
})

// 提前天数选项（根据重复类型限制最大值）
const advanceDaysOptions = computed(() => {
    const repeatType = currentRepeatType.value
    const maxDaysMap: Record<string, number> = {
        'none': 0,
        'daily': 0,
        'weekly': 6,
        'monthly': 29,
        'yearly': 364
    }
    
    const maxDays = maxDaysMap[repeatType] || 0
    const options = [{ value: 0, label: '不提前' }]
    
    for (let i = 1; i <= maxDays; i++) {
        options.push({ value: i, label: `提前${i}天` })
    }
    
    return options
})

// 下次实际提醒时间显示
const nextActualRemindTime = computed(() => {
    if (!currentNote.value.remindTime) {
        return ''
    }
    
    // 计算下一次基准提醒时间（考虑重复类型）
    const baseTime = new Date(currentNote.value.remindTime)
    const now = new Date()
    let nextBaseTime = new Date(baseTime)
    
    // 如果设置了重复类型，计算下一个周期的基准时间
    // 对于重复提醒，总是计算到下一个周期（即使原始日期是未来）
    if (currentNote.value.remindRepeat && currentNote.value.remindRepeat !== 'none') {
        // 先加一个周期，然后再判断是否需要继续往后推
        switch (currentNote.value.remindRepeat) {
            case 'daily':
                nextBaseTime.setDate(nextBaseTime.getDate() + 1)
                break
            case 'weekly':
                nextBaseTime.setDate(nextBaseTime.getDate() + 7)
                break
            case 'monthly':
                nextBaseTime.setMonth(nextBaseTime.getMonth() + 1)
                break
            case 'yearly':
                nextBaseTime.setFullYear(nextBaseTime.getFullYear() + 1)
                break
        }
        
        // 如果加了周期后还是小于等于现在，继续循环直到找到未来的时间
        while (nextBaseTime <= now) {
            switch (currentNote.value.remindRepeat) {
                case 'daily':
                    nextBaseTime.setDate(nextBaseTime.getDate() + 1)
                    break
                case 'weekly':
                    nextBaseTime.setDate(nextBaseTime.getDate() + 7)
                    break
                case 'monthly':
                    nextBaseTime.setMonth(nextBaseTime.getMonth() + 1)
                    break
                case 'yearly':
                    nextBaseTime.setFullYear(nextBaseTime.getFullYear() + 1)
                    break
            }
        }
    }
    
    // 实际提醒时间 = 下一次基准时间 - 提前天数
    const actualTime = new Date(nextBaseTime)
    if (currentAdvanceDays.value > 0) {
        actualTime.setDate(actualTime.getDate() - currentAdvanceDays.value)
    }
    
    console.log('[NotePad] 计算实际提醒时间:', {
        baseTime: formatLocalDateTime(baseTime),
        nextBaseTime: formatLocalDateTime(nextBaseTime),
        advanceDays: currentAdvanceDays.value,
        actualTime: formatLocalDate(actualTime)
    })
    
    return formatLocalDate(actualTime)
})

// 格式化本地日期（不含时间）
const formatLocalDate = (date: Date): string => {
    const y = date.getFullYear()
    const m = String(date.getMonth() + 1).padStart(2, '0')
    const d = String(date.getDate()).padStart(2, '0')
    return `${y}-${m}-${d}`
}

// 提醒检查
const checkReminds = () => {
    const now = new Date()
    noteList.value.forEach(note => {
        if (note.remindTime && note.remindStatus === 0) {
            const remindTime = new Date(note.remindTime)
            if (now >= remindTime && (now.getTime() - remindTime.getTime()) < 60000) {
                showRemindNotification(note)
                markAsReminded(note.id)
            }
        }
    })
}

const showRemindNotification = (note: NotepadInfo) => {
    dialog.info({
        title: '⏰ 提醒',
        content: () => h('div', [
            h('p', { style: 'font-weight: bold; margin-bottom: 8px;' }, note.title),
            h('p', { style: 'color: #666;' }, '设置的提醒时间已到！')
        ]),
        positiveText: '查看',
        negativeText: '关闭',
        onPositiveClick: () => {
            selectNote(note)
            emit('update:visible', true)
        }
    })
    if ('Notification' in window && Notification.permission === 'granted') {
        new Notification('⏰ 提醒', {
            body: note.title,
            icon: '/logo.png'
        })
    }
}

const markAsReminded = async (id: number) => {
    try {
        const note = noteList.value.find(n => n.id === id)
        if (note) {
            // 对于重复提醒，不更新 remindTime，保持原始基准时间
            // 只重置 remindStatus 为 0（未提醒），下次检查时会自动计算下一个周期
            let remindTime = note.remindTime  // 保持原值不变
            let remindStatus = 1  // 默认标记为已提醒
            
            // 如果是重复提醒，保持 remindStatus=0 以便下次继续检查
            if (note.remindRepeat && note.remindRepeat !== 'none') {
                remindStatus = 0
                console.log(`[NotePad] 重复提醒，保持 remindStatus=0，基准时间不变: ${remindTime}`)
            }
            
            await saveNotepadContent({
                id: note.id,
                title: note.title,
                content: note.content,
                remindTime: remindTime,  // 保持原始基准时间
                remindStatus: remindStatus,
                remindRepeat: note.remindRepeat || 'none',
                remindAdvanceDays: note.remindAdvanceDays || 0  // 保留提前天数
            })
            await loadList()
            
            // 通知父组件提醒状态已变化
            emit('remindStatusChanged', id)
        }
    } catch (e) {
        console.error('Mark reminded error', e)
    }
}

// 暴露方法
const refreshData = async () => {
    await loadList()
}

defineExpose({ 
    refreshData,
    checkReminds,
    markAsReminded,
    selectNote
})

// 删除便签
const deleteNote = (note: NotepadInfo) => {
    dialog.warning({
        title: '确认删除',
        content: `确定要删除便签「${note.title || '无标题'}」吗？`,
        positiveText: '删除',
        negativeText: '取消',
        onPositiveClick: async () => {
            try {
                const res = await deleteNotepad({ id: note.id })
                if (res.code === 0) {
                    message.success('删除成功')
                    await loadList()
                    // 如果删除的是当前选中的便签，清空编辑器
                    if (currentNote.value.id === note.id) {
                        if (noteList.value.length > 0) {
                            selectNote(noteList.value[0])
                        } else {
                            createNew()
                        }
                    }
                }
            } catch (e) {
                message.error('删除失败')
            }
        }
    })
}

// 监听显示状态
watch(() => props.visible, (val) => {
    if (val) {
        initData()
    } else {
        handleSave()
    }
})

const initData = async () => {
    console.log('[NotePad] initData 开始')
    await loadList()
    console.log('[NotePad] 列表加载完成，共', noteList.value.length, '个便签')
    
    // 如果有当前便签ID，从列表中查找并恢复
    if (currentNote.value.id && noteList.value.length > 0) {
        const savedNote = noteList.value.find(n => n.id === currentNote.value.id)
        console.log('[NotePad] 查找便签 ID:', currentNote.value.id, '找到:', !!savedNote)
        if (savedNote) {
            // 恢复便签内容和状态
            selectNote(savedNote)
        } else {
            // 如果找不到，选中第一个便签
            console.log('[NotePad] 未找到，选中第一个')
            selectNote(noteList.value[0])
        }
    } else if (noteList.value.length > 0) {
        // 如果没有当前便签ID，选中第一个
        console.log('[NotePad] 无当前ID，选中第一个')
        selectNote(noteList.value[0])
    }
}
</script>

<style scoped lang="less">
// 确保 Naive UI Dialog 关闭按钮显示
:global(.n-dialog__close) {
  display: flex !important;
}

:global(.n-dialog .n-base-icon) {
  color: #666 !important;
}

.notepad-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.3);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.notepad-container {
  position: absolute;
  width: 1000px;
  height: 600px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  display: flex;
  overflow: hidden;
}

// 左侧列表
.notepad-sidebar {
  width: 280px;
  background: #f5f5f7;
  border-right: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #e0e0e0;
}

.sidebar-title {
  font-size: 20px;
  font-weight: 600;
  color: #1d1d1f;
  margin: 0;
}

.action-icon {
  cursor: pointer;
  font-size: 20px;
  color: #86868b;
  transition: color 0.2s;
  
  &:hover {
    color: #007aff;
  }
}

.search-box {
  padding: 12px 16px;
  position: relative;
}

.search-icon {
  position: absolute;
  left: 28px;
  top: 50%;
  transform: translateY(-50%);
  color: #86868b;
  font-size: 18px;
}

.search-input {
  width: 100%;
  padding: 8px 12px 8px 36px;
  border: none;
  border-radius: 8px;
  background: #e8e8ed;
  font-size: 14px;
  outline: none;
  
  &::placeholder {
    color: #86868b;
  }
}

.note-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.note-item {
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
  margin-bottom: 4px;
  
  &:hover {
    background: #e8e8ed;
  }
  
  &.active {
    background: #d1d1d6;
  }
}

.note-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
  gap: 8px;
}

.note-item-title {
  font-size: 15px;
  font-weight: 600;
  color: #1d1d1f;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.remind-icon {
  color: #86868b;
  font-size: 16px;
  flex-shrink: 0;
  transition: color 0.2s;
  
  &.active {
    color: #007aff;
  }
}

.delete-icon {
  color: #86868b;
  font-size: 16px;
  cursor: pointer;
  transition: color 0.2s;
  flex-shrink: 0;
  
  &:hover {
    color: #ff3b30;
  }
}

.note-item-time {
  font-size: 12px;
  color: #86868b;
}

.sidebar-footer {
  padding: 16px;
  border-top: 1px solid #e0e0e0;
  display: flex;
  justify-content: center;
}

.new-note-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #007aff;
  border: none;
  color: white;
  font-size: 24px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.2s, box-shadow 0.2s;
  
  &:hover {
    transform: scale(1.1);
    box-shadow: 0 4px 12px rgba(0, 122, 255, 0.3);
  }
}

// 右侧编辑区
.notepad-editor {
  flex: 1;
  display: flex;
  flex-direction: column;
  position: relative;
}

.editor-header {
  padding: 16px;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.editor-title {
  font-size: 22px;
  font-weight: 600;
  color: #1d1d1f;
  margin: 0;
}

.editor-actions {
  display: flex;
  gap: 12px;
}

.editor-content {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  outline: none;
  font-size: 16px;
  line-height: 1.6;
  color: #1d1d1f;
  
  &:empty:before {
    content: attr(placeholder);
    color: #86868b;
  }
}

.editor-footer {
  padding: 12px 16px;
  border-top: 1px solid #e0e0e0;
  background: #f5f5f7;
}

.footer-text {
  font-size: 12px;
  color: #86868b;
}

// 提醒相关
.remind-float-btn {
  position: absolute;
  bottom: 60px;
  right: 16px;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    transform: scale(1.1);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  }
  
  &.active {
    background: #007aff;
    color: white;
  }
}

.remind-dot {
  position: absolute;
  top: -2px;
  right: -2px;
  width: 8px;
  height: 8px;
  background: #ff3b30;
  border-radius: 50%;
  border: 2px solid white;
}

.remind-picker {
  position: absolute;
  bottom: 60px;
  left: 16px;
  right: 16px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  padding: 16px;
  z-index: 10;
}

.remind-picker-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  
  span {
    font-size: 14px;
    font-weight: 600;
    color: #1d1d1f;
  }
}

.close-picker-btn {
  padding: 4px;
  background: transparent;
  border: none;
  cursor: pointer;
  color: #86868b;
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s;
  
  &:hover {
    background: rgba(0, 0, 0, 0.05);
    color: #1d1d1f;
  }
}

.remind-picker-info {
  margin-top: 12px;
  font-size: 12px;
  color: #86868b;
}

.remind-repeat-section {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.repeat-label {
  font-size: 14px;
  color: #1d1d1f;
  font-weight: 500;
}

.repeat-select {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  background: #f5f5f7;
  font-size: 14px;
  color: #1d1d1f;
  outline: none;
  cursor: pointer;
  transition: border-color 0.2s;
  
  &:hover {
    border-color: #007aff;
  }
  
  &:focus {
    border-color: #007aff;
    box-shadow: 0 0 0 2px rgba(0, 122, 255, 0.1);
  }
}

.repeat-badge {
  display: inline-block;
  margin-left: 8px;
  padding: 2px 8px;
  background: #007aff;
  color: white;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
}

.advance-badge {
  display: inline-block;
  margin-left: 6px;
  padding: 2px 8px;
  background: #ff9500;
  color: white;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
}

// 提前提醒样式
.remind-advance-section {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px 12px;
  background: rgba(255, 149, 0, 0.05);
  border-radius: 8px;
  border: 1px solid rgba(255, 149, 0, 0.2);
  transition: all 0.3s;
}

.advance-label {
  font-size: 14px;
  color: #1d1d1f;
  font-weight: 500;
}

.advance-select {
  padding: 8px 12px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  background: #f5f5f7;
  font-size: 14px;
  color: #1d1d1f;
  outline: none;
  cursor: pointer;
  transition: border-color 0.2s;
  
  &:hover {
    border-color: #ff9500;
  }
  
  &:focus {
    border-color: #ff9500;
    box-shadow: 0 0 0 2px rgba(255, 149, 0, 0.1);
  }
}

.advance-hint {
  margin-top: 4px;
  font-size: 12px;
  color: #ff9500;
  font-weight: 500;
  padding: 6px 8px;
  background: rgba(255, 149, 0, 0.1);
  border-radius: 6px;
}

// 强制提醒样式
.remind-force-section {
  margin-top: 12px;
  padding: 10px 12px;
  background: rgba(255, 59, 48, 0.05);
  border-radius: 8px;
  border: 1px solid rgba(255, 59, 48, 0.2);
  transition: all 0.3s;
  
  &:has(.force-checkbox:checked) {
    background: rgba(255, 59, 48, 0.1);
    border-color: rgba(255, 59, 48, 0.5);
  }
}

.force-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
  color: #ff3b30;
  font-weight: 500;
  user-select: none;
  
  &:hover {
    opacity: 0.8;
  }
}

.force-checkbox {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #ff3b30;
}

.remind-picker-actions {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}

.confirm-btn {
  min-width: 80px;
  padding: 8px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.3);
}

.confirm-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.confirm-btn:active {
  transform: translateY(0);
}

// 动画
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(10px);
}
</style>
