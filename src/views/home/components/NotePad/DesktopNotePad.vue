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
          :class="{ 'dark-mode': isDarkMode }"
          :style="{ left: x + 'px', top: y + 'px' }"
        >
          <!-- 左侧列表 -->
          <div ref="sidebarRef" class="notepad-sidebar">
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
                  <input 
                    v-if="editingNoteId === note.id" 
                    v-model="editingTitle" 
                    class="note-item-title-input"
                    @blur="saveEditingTitle(note)"
                    @keyup.enter="saveEditingTitle(note)"
                    @click.stop
                  />
                  <span v-else class="note-item-title" @dblclick.stop="startEditing(note)">{{ note.title || '无标题' }}</span>
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
                <div class="note-item-time">{{ formatDate(note.updateTime) }}</div>
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
          <div class="notepad-editor" @mousedown="preventEditorDrag">
            <!-- 编辑器顶部 -->
            <div class="editor-header">
              <input 
                v-if="isEditingHeaderTitle" 
                v-model="headerEditingTitle" 
                class="editor-title-input"
                @blur="saveHeaderTitle"
                @keyup.enter="saveHeaderTitle"
                @click.stop
              />
              <h1 v-else class="editor-title" :class="{ 'editable': isEditMode }" @dblclick="isEditMode ? startHeaderEditing() : null">{{ currentNote.title || '无标题' }}</h1>
              <div class="editor-actions">
                <!-- 查看模式：显示编辑按钮 -->
                <SvgIcon 
                  v-if="!isEditMode" 
                  class="action-icon" 
                  icon="basil--edit-outline" 
                  @click="enterEditMode"
                  title="编辑"
                />
                <!-- 编辑模式：显示保存和取消按钮 -->
                <button 
                  v-else
                  class="save-btn" 
                  @click="handleManualSave"
                >
                  保存
                </button>
                <button 
                  v-if="isEditMode"
                  class="cancel-edit-btn" 
                  @click="handleCancelEdit"
                >
                  取消
                </button>
                <!-- 关闭按钮 -->
                <SvgIcon class="action-icon" icon="material-symbols--close" @click="handleClose" />
              </div>
            </div>

            <!-- 编辑内容区 -->
            <div 
              ref="editorRef"
              class="editor-content"
              :class="{ 'readonly': !isEditMode }"
              contenteditable="true"
              @input="handleInput"
              @paste="handlePaste"
              @drop="handleDrop"
              @dragover="handleDragOver"
              @mousedown.stop
              placeholder="请输入笔记内容，或拖拽文件到这里"
            ></div>

            <!-- 底部信息 -->
            <div class="editor-footer">
              <span class="footer-text">
                {{ isEditMode ? '编辑模式' : '查看模式' }} | 最后编辑：{{ formatFullDate(currentNote.updateTime) }}，创建：{{ formatFullDate(currentNote.createTime) }}
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
                  <button 
                    v-if="currentNote.remindTime && currentNote.remindStatus !== 2" 
                    class="cancel-btn" 
                    @click="handleCancelRemind"
                  >
                    取消提醒
                  </button>
                  <button class="confirm-btn" @click="handleConfirmRemind">
                    {{ currentNote.remindTime && currentNote.remindStatus !== 2 ? '修改' : '确认' }}
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
import { ref, computed, onMounted, onUnmounted, nextTick, watch, h } from 'vue'
import { SvgIcon, SvgIconOnline } from '@/components/common'
import { useMessage, useDialog, NDatePicker } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useDraggable, useDebounceFn, useStorage } from '@vueuse/core'
import { 
    getNotepadList, 
    saveNotepadContent,
    deleteNotepad,
    uploadNotepadFile,
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

// 检测深色模式
const isDarkMode = ref(document.documentElement.classList.contains('dark'))
const editorRef = ref<HTMLDivElement | null>(null)
const notepadRef = ref<HTMLElement | null>(null)
const sidebarRef = ref<HTMLElement | null>(null)

// 状态
const currentNote = useStorage<Partial<NotepadInfo>>('sun-panel-notepad-current', { id: 0, title: '', content: '' })
const noteList = useStorage<NotepadInfo[]>('sun-panel-notepad-list', [])
const searchKeyword = ref('')
const showSortMenu = ref(false)
const showRemindPicker = ref(false)
const currentRepeatType = ref<string>('none') // 当前选择的重复类型
const currentAdvanceDays = ref<number>(0) // 当前提前提醒天数
const editingNoteId = ref<number | null>(null) // 正在编辑的便签ID
const editingTitle = ref('') // 编辑中的标题
const isEditingHeaderTitle = ref(false) // 是否正在编辑顶部标题
const headerEditingTitle = ref('') // 顶部标题编辑值
const isEditMode = ref(false) // 是否处于编辑模式

// 监听当前便签变化，同步到编辑器
watch(() => currentNote.value, (newNote) => {
    if (editorRef.value && newNote) {
        // 只在内容不同时才更新，避免光标跳动
        if (editorRef.value.innerHTML !== newNote.content) {
            editorRef.value.innerHTML = newNote.content || ''
            // 查看模式下禁用编辑
            if (!isEditMode.value) {
                editorRef.value.contentEditable = 'false'
            }
            nextTick(() => {
                bindFileDownloadEvents()
            })
        }
    }
}, { deep: true })

// 监听视图变化，从提醒页返回编辑页时恢复内容
watch(() => showRemindPicker.value, (newValue) => {
    if (!newValue && editorRef.value && currentNote.value) {
        // 从提醒页返回时，恢复编辑器内容
        nextTick(() => {
            if (editorRef.value) {
                editorRef.value.innerHTML = currentNote.value.content || ''
                // 根据编辑模式设置可编辑状态
                editorRef.value.contentEditable = isEditMode.value ? 'true' : 'false'
                bindFileDownloadEvents()
            }
        })
    }
})

// 阻止编辑器区域的拖动
const preventEditorDrag = (e: MouseEvent) => {
    e.stopPropagation()
}

// 动态计算窗口初始位置（居中）
const getInitialPosition = () => {
  const width = Math.min(Math.max(window.innerWidth * 0.8, 800), 1200);
  const height = Math.min(Math.max(window.innerHeight * 0.75, 500), 800);
  return {
    x: (window.innerWidth - width) / 2,
    y: (window.innerHeight - height) / 2
  };
};

const { x, y } = useDraggable(notepadRef, {
  initialValue: getInitialPosition(),
  handle: sidebarRef // 只在左侧列表区域可以拖动
})

// 监听 dark class 变化
let observer: MutationObserver | null = null
onMounted(async () => {
    if (noteList.value.length === 0) {
        await loadList()
    }
    
    // 阻止编辑器区域的拖动（使用原生事件监听）
    const editorElement = document.querySelector('.notepad-editor')
    if (editorElement) {
        editorElement.addEventListener('mousedown', (e) => {
            e.stopPropagation()
        }, true) // 使用捕获阶段
    }
    
    // 监听 dark class 变化
    observer = new MutationObserver((mutations) => {
        mutations.forEach((mutation) => {
            if (mutation.attributeName === 'class') {
                isDarkMode.value = document.documentElement.classList.contains('dark')
            }
        })
    })
    observer.observe(document.documentElement, { attributes: true })
})

onUnmounted(() => {
    if (observer) {
        observer.disconnect()
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

// 输入处理 - 只在编辑模式下才触发
const handleInput = () => {
    if (!editorRef.value || !isEditMode.value) return
    // 只保存内容，不自动生成标题
    saveContent()
}

// 粘贴处理
const handlePaste = async (e: ClipboardEvent) => {
    const items = e.clipboardData?.items
    if (!items) return
    
    // 检查是否有图片
    for (let i = 0; i < items.length; i++) {
        if (items[i].type.indexOf('image') !== -1) {
            e.preventDefault()
            const file = items[i].getAsFile()
            if (file) {
                await uploadAndInsertImage(file)
            }
            return
        }
    }
    
    // 纯文本粘贴
    e.preventDefault()
    const text = e.clipboardData?.getData('text/plain') || ''
    document.execCommand('insertText', false, text)
}

// 拖拽覆盖事件
const handleDragOver = (e: DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
}

// 拖拽放下事件
const handleDrop = async (e: DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    
    const files = e.dataTransfer?.files
    if (!files || files.length === 0) return
    
    for (let i = 0; i < files.length; i++) {
        const file = files[i]
        // 只处理图片文件
        if (file.type.startsWith('image/')) {
            await uploadAndInsertImage(file)
        } else {
            // 其他文件插入下载链接
            await insertFileLink(file)
        }
    }
}

// 上传图片并插入到编辑器
const uploadAndInsertImage = async (file: File) => {
    try {
        const formData = new FormData()
        formData.append('file', file)
        
        const res = await uploadNotepadFile(formData)
        if (res.code === 0 && res.data) {
            const imgHtml = `<img src="${res.data.url}" alt="${res.data.name}" style="max-width:100%;height:auto;margin:8px 0;" />`
            if (editorRef.value) {
                editorRef.value.focus()
                document.execCommand('insertHTML', false, imgHtml)
                saveContent()
            }
            message.success('图片上传成功')
        }
    } catch (error) {
        console.error('Upload image error:', error)
        message.error('图片上传失败')
    }
}

// 插入文件链接
const insertFileLink = async (file: File) => {
    try {
        const formData = new FormData()
        formData.append('file', file)
        
        const res = await uploadNotepadFile(formData)
        if (res.code === 0 && res.data) {
            const linkHtml = `<a href="${res.data.url}" class="file-attachment" data-filename="${res.data.name}" style="color:#007aff;text-decoration:none;display:inline-flex;align-items:center;gap:4px;padding:4px 8px;background:#f5f5f7;border-radius:6px;margin:4px 0;"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>${res.data.name}</a>`
            if (editorRef.value) {
                editorRef.value.focus()
                document.execCommand('insertHTML', false, linkHtml)
                saveContent()
            }
            message.success('文件上传成功')
        }
    } catch (error) {
        console.error('Upload file error:', error)
        message.error('文件上传失败')
    }
}

// 核心保存逻辑（只保存便签内容，不保存提醒设置）
const handleSave = async () => {
    if (editorRef.value) {
        try {
            const content = editorRef.value.innerHTML
            // 使用当前标题，不自动生成
            const title = currentNote.value.title || '无标题'
            const saveId = currentNote.value.id || 0
            
            // 只保存标题和内容，不传递提醒相关字段，避免覆盖数据库中的提醒设置
            const res = await saveNotepadContent({ 
                id: saveId,
                title: title,
                content: content
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

// 切换便签 - 进入查看模式
const selectNote = (note: NotepadInfo) => {
    console.log('[NotePad] selectNote:', note.title)
    currentNote.value = { ...note }
    currentRepeatType.value = note.remindRepeat || 'none'
    currentAdvanceDays.value = note.remindAdvanceDays || 0
    
    // 如果重复类型是“不重复”或“每天”，强制重置提前天数为0
    if (currentRepeatType.value === 'none' || currentRepeatType.value === 'daily') {
        currentAdvanceDays.value = 0
        currentNote.value.remindAdvanceDays = 0
    }
    
    // 设置为查看模式
    isEditMode.value = false
    
    if (editorRef.value) {
        editorRef.value.innerHTML = note.content || ''
        // 查看模式下禁用编辑
        editorRef.value.contentEditable = 'false'
        nextTick(() => {
            bindFileDownloadEvents()
        })
    }
}

// 进入编辑模式
const enterEditMode = () => {
    isEditMode.value = true
    nextTick(() => {
        if (editorRef.value) {
            editorRef.value.contentEditable = 'true'
            editorRef.value.focus()
        }
    })
}

// 保存内容(手动保存)
const handleManualSave = async () => {
    await handleSave()
    isEditMode.value = false
    // 保存后禁用编辑
    nextTick(() => {
        if (editorRef.value) {
            editorRef.value.contentEditable = 'false'
        }
    })
    message.success('保存成功')
}

// 取消编辑
const handleCancelEdit = () => {
    dialog.warning({
        title: '提示',
        content: '是否放弃当前修改？',
        positiveText: '放弃',
        negativeText: '继续编辑',
        onPositiveClick: () => {
            isEditMode.value = false
            // 恢复原始内容
            nextTick(() => {
                if (editorRef.value && currentNote.value) {
                    editorRef.value.innerHTML = currentNote.value.content || ''
                    editorRef.value.contentEditable = 'false'
                }
            })
        }
    })
}

// 新建便签 - 直接进入编辑模式
const createNew = () => {
    const finalTitle = `便签${noteList.value.length + 1}`
    currentNote.value = { id: 0, title: finalTitle, content: '' }
    currentRepeatType.value = 'none'
    currentAdvanceDays.value = 0
    isEditMode.value = true // 新建时直接进入编辑模式
    if (editorRef.value) {
        editorRef.value.innerHTML = ''
        editorRef.value.contentEditable = 'true'
        editorRef.value.focus()
    }
}

// 开始编辑标题
const startEditing = (note: NotepadInfo) => {
    editingNoteId.value = note.id
    editingTitle.value = note.title || ''
}

// 保存编辑的标题
const saveEditingTitle = async (note: NotepadInfo) => {
    const newTitle = editingTitle.value.trim() || '无标题'
    
    // 更新列表中的标题
    const index = noteList.value.findIndex(n => n.id === note.id)
    if (index !== -1) {
        noteList.value[index].title = newTitle
    }
    
    // 如果当前选中的是这个便签，也更新
    if (currentNote.value.id === note.id) {
        currentNote.value.title = newTitle
    }
    
    // 保存到后端
    try {
        await saveNotepadContent({
            id: note.id,
            title: newTitle,
            content: note.content || ''
        })
    } catch (e) {
        console.error('Save title error', e)
        message.error('保存标题失败')
    }
    
    editingNoteId.value = null
    editingTitle.value = ''
}

// 开始编辑顶部标题
const startHeaderEditing = () => {
    isEditingHeaderTitle.value = true
    headerEditingTitle.value = currentNote.value.title || ''
}

// 保存顶部标题
const saveHeaderTitle = async () => {
    const newTitle = headerEditingTitle.value.trim() || '无标题'
    
    // 更新当前便签标题
    currentNote.value.title = newTitle
    
    // 更新列表中的标题
    if (currentNote.value.id) {
        const index = noteList.value.findIndex(n => n.id === currentNote.value.id)
        if (index !== -1) {
            noteList.value[index].title = newTitle
        }
        
        // 保存到后端
        try {
            await saveNotepadContent({
                id: currentNote.value.id,
                title: newTitle,
                content: currentNote.value.content || ''
            })
        } catch (e) {
            console.error('Save title error', e)
            message.error('保存标题失败')
        }
    }
    
    isEditingHeaderTitle.value = false
    headerEditingTitle.value = ''
}

// 关闭
const handleClose = () => {
    // 如果在编辑模式，提示保存
    if (isEditMode.value) {
        dialog.warning({
            title: '提示',
            content: '是否保存当前修改？',
            positiveText: '保存',
            negativeText: '不保存',
            onPositiveClick: async () => {
                await handleSave()
                emit('update:visible', false)
                isEditMode.value = false
            },
            onNegativeClick: () => {
                emit('update:visible', false)
                isEditMode.value = false
                // 恢复原始内容
                nextTick(() => {
                    if (editorRef.value && currentNote.value) {
                        editorRef.value.innerHTML = currentNote.value.content || ''
                    }
                })
            }
        })
    } else {
        // 查看模式直接关闭
        emit('update:visible', false)
    }
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

// 取消提醒
const handleCancelRemind = async () => {
    dialog.warning({
        title: '确认取消',
        content: `确定要取消便签「${currentNote.value.title || '无标题'}」的提醒吗？`,
        positiveText: '取消提醒',
        negativeText: '再想想',
        onPositiveClick: async () => {
            await setRemind(null, true)
            message.success('已取消提醒')
        }
    })
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
        
        // 保存提醒设置时，保留当前的编辑器内容
        let contentToSave = currentNote.value.content || ''
        if (editorRef.value) {
            contentToSave = editorRef.value.innerHTML
        }
        
        console.log('[NotePad] 准备保存的数据:', {
            id: currentNote.value.id,
            title: currentNote.value.title,
            remindTime: remindTime,
            remindStatus: remindTime ? 0 : 2,  // 有时间为等待触发(0)，无时间为已结束(2)
            remindRepeat: currentRepeatType.value,
            remindAdvanceDays: currentAdvanceDays.value
        })
        
        await saveNotepadContent({
            id: currentNote.value.id,
            title: currentNote.value.title || '',
            content: contentToSave,
            remindTime: remindTime,
            remindStatus: remindTime ? 0 : 2,  // 有时间为等待触发(0)，无时间为已结束(2)
            remindRepeat: currentRepeatType.value,
            remindAdvanceDays: currentAdvanceDays.value
        })
        
        console.log('[NotePad] 保存成功，更新本地状态')
        currentNote.value.remindTime = remindTime || undefined
        currentNote.value.remindStatus = remindTime ? 0 : 2  // 有时间为等待触发(0)，无时间为已结束(2)
        currentNote.value.remindRepeat = currentRepeatType.value
        currentNote.value.remindAdvanceDays = currentAdvanceDays.value
        
        // 更新本地缓存中的便签内容
        const index = noteList.value.findIndex(n => n.id === currentNote.value.id)
        if (index !== -1) {
            noteList.value[index].remindTime = remindTime || undefined
            noteList.value[index].remindStatus = remindTime ? 0 : 2
            noteList.value[index].remindRepeat = currentRepeatType.value
            noteList.value[index].remindAdvanceDays = currentAdvanceDays.value
        }
        
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
    console.log('[NotePad] currentNote.remindRepeat:', currentNote.value.remindRepeat)
    console.log('[NotePad] currentAdvanceDays before:', currentAdvanceDays.value)
    
    // 从 currentNote 重新读取 remindRepeat 状态
    currentRepeatType.value = currentNote.value.remindRepeat || 'none'
    
    // 从 currentNote 重新读取 remindAdvanceDays 状态
    if (currentNote.value.remindAdvanceDays !== undefined) {
        currentAdvanceDays.value = currentNote.value.remindAdvanceDays
    } else {
        currentAdvanceDays.value = 0
    }
    
    // 如果重复类型是“不重复”或“每天”，强制重置提前天数为0
    if (currentRepeatType.value === 'none' || currentRepeatType.value === 'daily') {
        currentAdvanceDays.value = 0
        console.log('[NotePad] 重复类型为不重复/每天，重置提前天数为0')
    }
    
    console.log('[NotePad] 从 currentNote 同步 remindAdvanceDays:', currentAdvanceDays.value)
    
    // 优先从 remindBaseTime 解析日期时间（用户原始选择的时间），如果没有则使用 remindTime
    const timeStr = currentNote.value.remindBaseTime || currentNote.value.remindTime
    if (timeStr) {
        const date = new Date(timeStr)
        if (!isNaN(date.getTime())) {
            remindTimestamp.value = date.getTime()
            console.log('[NotePad] 从', currentNote.value.remindBaseTime ? 'remindBaseTime' : 'remindTime', '加载时间:', timeStr)
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
    // 只更新本地状态，不立即保存，等待用户点击“完成”
    currentNote.value.remindRepeat = currentRepeatType.value
    
    // 如果切换到“不重复”或“每天”，重置提前天数为0
    if (currentRepeatType.value === 'none' || currentRepeatType.value === 'daily') {
        currentAdvanceDays.value = 0
        currentNote.value.remindAdvanceDays = 0
    }
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
    // 直接使用后端计算的 remindTime（实际触发时间）
    if (currentNote.value.remindTime) {
        const remindTime = new Date(currentNote.value.remindTime)
        return formatLocalDate(remindTime)
    }
    return ''
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
    
    // 强制从 localStorage 重新读取 currentNote，确保获取最新值
    const storedCurrentNote = localStorage.getItem('sun-panel-notepad-current')
    if (storedCurrentNote) {
        try {
            const parsed = JSON.parse(storedCurrentNote)
            console.log('[NotePad] 从 storage 读取到 currentNote:', parsed)
            // 更新 currentNote.value
            currentNote.value = parsed
        } catch (e) {
            console.error('[NotePad] 解析 currentNote 失败:', e)
        }
    }
    
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
  width: 80vw !important;
  max-width: 1200px !important;
  min-width: 800px !important;
  height: 75vh !important;
  max-height: 800px !important;
  min-height: 500px !important;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  display: flex;
  overflow: hidden;
  user-select: none; /* 禁用拖拽时的文本选择 */
  -webkit-user-select: none;
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

.note-item-title-input {
  font-size: 15px;
  font-weight: 600;
  color: #1d1d1f;
  border: none;
  outline: none;
  background: transparent;
  padding: 2px 4px;
  border-radius: 4px;
  width: 100%;
  
  &:focus {
    background: rgba(0, 122, 255, 0.1);
  }
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
  background: #f5f5f7; /* 浅色模式：灰色底色 */
  border: none;
  color: #007aff; /* 图标保持蓝色 */
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
  pointer-events: auto; /* 确保编辑器区域处理自己的鼠标事件，避免被父级拖拽捕获 */
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
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    opacity: 0.7;
  }
  
  // 编辑模式下显示可编辑提示
  &.editable {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    
    &::after {
      content: '双击编辑';
      font-size: 12px;
      font-weight: normal;
      color: #86868b;
      opacity: 0.6;
      white-space: nowrap;
    }
    
    &:hover::after {
      opacity: 1;
    }
  }
}

.editor-title-input {
  font-size: 22px;
  font-weight: 600;
  color: #1d1d1f;
  border: none;
  outline: none;
  background: transparent;
  padding: 2px 4px;
  border-radius: 4px;
  width: 100%;
  max-width: 400px;
  position: relative;
  
  &:focus {
    background: rgba(0, 122, 255, 0.1);
  }
  
  // 编辑模式下显示可编辑提示
  &::placeholder {
    content: '双击编辑';
    color: #86868b;
    opacity: 0.6;
  }
}

.editor-actions {
  display: flex;
  gap: 12px;
}

// 保存按钮
.save-btn {
  padding: 6px 16px;
  background: #007aff;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    opacity: 0.8;
    transform: scale(1.05);
  }
  
  &:active {
    transform: scale(0.95);
  }
}

// 取消编辑按钮
.cancel-edit-btn {
  padding: 6px 16px;
  background: #f5f5f7;
  color: #1d1d1f;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    background: #e8e8ed;
    transform: scale(1.05);
  }
  
  &:active {
    transform: scale(0.95);
  }
}

.editor-content {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  outline: none;
  font-size: 16px;
  line-height: 1.6;
  color: #1d1d1f;
  user-select: text; /* 允许编辑器内选择文本 */
  -webkit-user-select: text;
  cursor: text;
  pointer-events: auto; /* 确保编辑器处理自己的鼠标事件，避免被父级拖拽捕获 */
  
  &:empty:before {
    content: attr(placeholder);
    color: #86868b;
  }
  
  // 只读模式样式
  &.readonly {
    background: #fafafa;
    cursor: default;
    user-select: text;
  }
}

// 深色模式适配
.dark .editor-content {
  color: #f5f5f7;
  
  &:empty:before {
    color: #86868b;
  }
  
  &.readonly {
    background: #2c2c2e;
  }
}

.editor-footer {
  padding: 12px 16px;
  border-top: 1px solid #e0e0e0;
  background: #f5f5f7;
}

// 深色模式适配
.dark .editor-footer {
  border-top-color: #3a3a3c;
  background: #1c1c1e;
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
  background: white; /* 浅色模式：白色底色 */
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #007aff; /* 图标保持蓝色 */
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

// 深色模式适配
.dark .remind-float-btn {
  background: #2c2c32; /* 深色模式：深色底色 */
  color: #0a84ff; /* 图标保持蓝色 */
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  
  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
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

// 深色模式适配
.dark .remind-picker {
  background: #2c2c2e;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
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

// 深色模式适配
.dark .repeat-select {
  border-color: #3a3a3c;
  background: #1c1c1e;
  color: #f5f5f7;
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
  justify-content: space-between;
  gap: 8px;
}

.cancel-btn {
  min-width: 80px;
  padding: 8px 20px;
  background: #f5f5f7;
  color: #ff3b30;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: all 0.2s;
}

.cancel-btn:hover {
  background: #ffebee;
  border-color: #ff3b30;
}

.cancel-btn:active {
  transform: scale(0.98);
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

// 深色模式
.notepad-container.dark-mode {
  background: #18181c;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
}

.notepad-container.dark-mode .notepad-sidebar {
  background: #1c1c20;
  border-right-color: #2c2c32;
}

.notepad-container.dark-mode .sidebar-header,
.notepad-container.dark-mode .editor-header,
.notepad-container.dark-mode .sidebar-footer {
  border-color: #2c2c32;
}

.notepad-container.dark-mode .editor-footer {
  border-top-color: #2c2c32;
  background: #1c1c20;
}

.notepad-container.dark-mode .sidebar-title,
.notepad-container.dark-mode .editor-title,
.notepad-container.dark-mode .note-item-title,
.notepad-container.dark-mode .note-item-title-input,
.notepad-container.dark-mode .editor-title-input,
.notepad-container.dark-mode .repeat-label,
.notepad-container.dark-mode .advance-label {
  color: rgba(255, 255, 255, 0.9);
}

.notepad-container.dark-mode .action-icon,
.notepad-container.dark-mode .delete-icon,
.notepad-container.dark-mode .remind-icon,
.notepad-container.dark-mode .note-item-time,
.notepad-container.dark-mode .footer-text,
.notepad-container.dark-mode .search-icon,
.notepad-container.dark-mode .remind-picker-info,
.notepad-container.dark-mode .close-picker-btn {
  color: rgba(255, 255, 255, 0.52);
}

.notepad-container.dark-mode .action-icon:hover {
  color: #0a84ff;
}

.notepad-container.dark-mode .delete-icon:hover {
  color: #ff453a;
}

.notepad-container.dark-mode .remind-icon.active {
  color: #0a84ff;
}

.notepad-container.dark-mode .search-input,
.notepad-container.dark-mode .repeat-select,
.notepad-container.dark-mode .advance-select {
  background: #2c2c32;
  color: rgba(255, 255, 255, 0.9);
  border-color: #2c2c32;
}

.notepad-container.dark-mode .search-input::placeholder {
  color: rgba(255, 255, 255, 0.52);
}

.notepad-container.dark-mode .note-item:hover {
  background: #2c2c32;
}

.notepad-container.dark-mode .note-item.active {
  background: #323238;
}

.notepad-container.dark-mode .editor-content {
  color: rgba(255, 255, 255, 0.9);
}

.notepad-container.dark-mode .editor-content:empty:before {
  color: rgba(255, 255, 255, 0.52);
}

.notepad-container.dark-mode .remind-float-btn {
  background: #2c2c32; /* 深色模式：深色底色 */
  color: #0a84ff; /* 图标保持蓝色 */
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.notepad-container.dark-mode .remind-float-btn.active {
  background: #0a84ff;
  color: white;
}

.notepad-container.dark-mode .remind-picker {
  background: #1c1c20;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
}

.notepad-container.dark-mode .cancel-btn {
  background: #2c2c32;
  color: #ff453a;
  border-color: #3c3c42;
}

.notepad-container.dark-mode .cancel-btn:hover {
  background: rgba(255, 69, 58, 0.15);
  border-color: #ff453a;
}
.notepad-container.dark-mode .remind-picker-header span {
  color: rgba(255, 255, 255, 0.9);
}

.notepad-container.dark-mode .close-picker-btn {
  color: rgba(255, 255, 255, 0.52);
}

.notepad-container.dark-mode .close-picker-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.9);
}

// 新建按钮：底色随深浅模式变化，但图标保持蓝色
.notepad-container.dark-mode .new-note-btn {
  background: #2c2c32; /* 深色模式：深色底色 */
  color: #0a84ff; /* 图标保持蓝色 */
}
</style>
