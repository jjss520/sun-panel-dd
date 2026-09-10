<template>
  <Teleport to="body">
    <transition name="fade">
      <div 
        v-if="visible" 
        class="mobile-notepad-overlay"
        @click.self="handleClose"
      >
        <div 
          class="mobile-notepad-container"
          :class="{ 'dark-mode': isDarkMode }"
        >
          
          <!-- 视图1：列表页 -->
          <div v-if="currentView === 'list'" class="mobile-view list-view">
            <!-- 顶部标题栏 -->
            <div class="mobile-header">
              <h2 class="mobile-title">备忘录</h2>
              <div class="header-actions">
                <SvgIcon 
                  class="action-icon" 
                  icon="mdi:chevron-down" 
                  @click="showSortMenu = !showSortMenu"
                />
                <SvgIcon 
                  class="action-icon" 
                  icon="material-symbols--close" 
                  @click="handleClose"
                />
              </div>
            </div>

            <!-- 搜索框 -->
            <div class="mobile-search-box">
              <SvgIcon class="search-icon" icon="mdi:magnify" />
              <input 
                v-model="searchKeyword"
                type="text" 
                placeholder="search"
                class="search-input"
              />
            </div>

            <!-- 便签列表 -->
            <div class="mobile-note-list">
              <div 
                v-for="note in filteredNotes" 
                :key="note.id"
                class="mobile-note-item"
                @click="selectNote(note)"
              >
                <div class="mobile-note-item-header">
                  <span class="mobile-note-item-title">{{ note.title || '无标题' }}</span>
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
                  <SvgIcon 
                    class="delete-icon" 
                    icon="material-symbols--delete-outline"
                    @click.stop="deleteNote(note)"
                  />
                </div>
                <div class="mobile-note-item-time">{{ formatDate(note.updateTime) }}</div>
              </div>
            </div>

            <!-- 底部新建按钮 -->
            <div class="mobile-footer">
              <button class="new-note-btn" @click="createNew">
                <SvgIcon icon="pajamas--doc-new" />
              </button>
            </div>
          </div>

          <!-- 视图2：编辑页 -->
          <div v-if="currentView === 'editor'" class="mobile-view editor-view">
            <!-- 顶部导航栏 -->
            <div class="mobile-editor-header">
              <div class="nav-left">
                <input 
                  v-if="isEditingTitle" 
                  v-model="editingTitleValue" 
                  class="mobile-editor-title-input"
                  @blur="saveMobileTitle"
                  @keyup.enter="saveMobileTitle"
                  @click.stop
                />
                <h1 v-else class="mobile-editor-title" :class="{ 'editable': isEditMode }" @click="isEditMode ? startMobileEditing() : null">{{ currentNote.title || '无标题' }}</h1>
              </div>
              <div class="header-actions">
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
                <SvgIcon class="action-icon" icon="material-symbols--close" @click="handleClose" />
              </div>
            </div>

            <!-- 编辑内容区 -->
            <div 
              ref="editorRef"
              class="mobile-editor-content"
              :class="{ 'readonly': !isEditMode }"
              contenteditable="true"
              @input="handleInput"
              @paste="handlePaste"
              placeholder="请输入笔记内容"
            ></div>

            <!-- 底部信息 -->
            <div class="mobile-editor-footer">
              <span class="footer-text">
                {{ isEditMode ? '编辑模式' : '查看模式' }} | 最后编辑：{{ formatFullDate(currentNote.updateTime) }}
              </span>
            </div>

            <!-- 悬浮提醒按钮 -->
            <div 
              class="remind-float-btn"
              :class="{ active: currentNote.remindTime && currentNote.remindStatus !== 2 }"
              @click="openRemindView"
            >
              <SvgIconOnline :icon="currentNote.remindTime && currentNote.remindStatus !== 2 ? 'mdi:bell-outline' : 'mdi:bell-off-outline'" />
              <span v-if="currentNote.remindTime && currentNote.remindStatus !== 2" class="remind-dot"></span>
            </div>
          </div>

          <!-- 视图3：提醒设置页 -->
          <div v-if="currentView === 'remind'" class="mobile-view remind-view">
            <!-- 顶部导航栏 -->
            <div class="mobile-remind-header">
              <div class="nav-left">
                <SvgIcon class="back-icon" icon="mdi:arrow-left" @click="goBack" />
                <span class="mobile-remind-title">设置提醒</span>
              </div>
              <div class="header-actions">
                <button class="cancel-btn-small" @click="handleCancelRemindSetting">
                  取消
                </button>
                <button class="confirm-btn" @click="handleConfirmRemind">
                  完成
                </button>
              </div>
            </div>

            <!-- 提醒设置内容 -->
            <div class="mobile-remind-content">
              <!-- 日期时间选择器 -->
              <div class="remind-section">
                <div class="section-label">📅 选择日期时间</div>
                <NDatePicker
                  v-model:value="remindTimestamp"
                  type="datetime"
                  format="yyyy-MM-dd HH:mm"
                  :time-picker-props="{ format: 'HH:mm' }"
                  placeholder="请选择提醒日期和时间"
                  clearable
                  :actions="['confirm']"
                  class="mobile-date-picker"
                />
              </div>

              <!-- 重复选项 -->
              <div class="remind-section">
                <div class="section-label">🔄 重复类型</div>
                <div class="radio-group">
                  <label class="radio-item">
                    <input type="radio" v-model="currentRepeatType" value="none" @change="handleRepeatChange" />
                    <span>不重复</span>
                  </label>
                  <label class="radio-item">
                    <input type="radio" v-model="currentRepeatType" value="daily" @change="handleRepeatChange" />
                    <span>每天</span>
                  </label>
                  <label class="radio-item">
                    <input type="radio" v-model="currentRepeatType" value="weekly" @change="handleRepeatChange" />
                    <span>每周</span>
                  </label>
                  <label class="radio-item">
                    <input type="radio" v-model="currentRepeatType" value="monthly" @change="handleRepeatChange" />
                    <span>每月</span>
                  </label>
                  <label class="radio-item">
                    <input type="radio" v-model="currentRepeatType" value="yearly" @change="handleRepeatChange" />
                    <span>每年</span>
                  </label>
                </div>
              </div>

              <!-- 提前提醒选项 -->
              <div v-if="showAdvanceDays" class="remind-section">
                <div class="section-label">⏰ 提前提醒</div>
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

              <!-- 当前提醒信息 -->
              <div v-if="currentNote.remindTime" class="remind-info-box">
                <div class="info-label">💡 提醒信息</div>
                <div class="info-content">
                  原始选择：{{ formatRemindTime(currentNote.remindBaseTime || currentNote.remindTime) }}<br/>
                  下次实际提醒：{{ nextActualRemindTime }}
                  <span v-if="currentNote.remindRepeat && currentNote.remindRepeat !== 'none'" class="repeat-badge">
                    {{ getRepeatText(currentNote.remindRepeat) }}
                  </span>
                  <span v-if="currentNote.remindAdvanceDays && currentNote.remindAdvanceDays > 0" class="advance-badge">
                    提前{{ currentNote.remindAdvanceDays }}天
                  </span>
                </div>
              </div>

              <!-- 取消提醒按钮 -->
              <div v-if="currentNote.remindTime && currentNote.remindStatus !== 2" class="cancel-section">
                <button class="cancel-btn" @click="handleCancelRemind">
                  取消提醒
                </button>
              </div>
            </div>
          </div>

        </div>
      </div>
    </transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { SvgIcon, SvgIconOnline } from '@/components/common'
import { useMessage, useDialog, NDatePicker } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useDebounceFn, useStorage } from '@vueuse/core'
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
  (e: 'remindStatusChanged', noteId: number): void
}>()

const { t: _t } = useI18n()
const message = useMessage()
const dialog = useDialog()
const authStore = useAuthStore()

// 检测深色模式
const isDarkMode = ref(document.documentElement.classList.contains('dark'))

// 监听 dark class 变化
let observer: MutationObserver | null = null
onMounted(async () => {
    if (noteList.value.length === 0) {
        await loadList()
    }
    
    // 监听深色模式变化
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
const editorRef = ref<HTMLDivElement | null>(null)

// 视图状态
const currentView = ref<'list' | 'editor' | 'remind'>('list')
const isEditMode = ref(false) // 是否处于编辑模式

// 数据状态
const currentNote = useStorage<Partial<NotepadInfo>>('sun-panel-notepad-current', { id: 0, title: '', content: '' })
const noteList = useStorage<NotepadInfo[]>('sun-panel-notepad-list', [])
const searchKeyword = ref('')
const showSortMenu = ref(false)
const currentRepeatType = ref<string>('none')
const currentAdvanceDays = ref<number>(0)
const remindTimestamp = ref<number | null>(null)
const isEditingTitle = ref(false) // 是否正在编辑标题
const editingTitleValue = ref('') // 编辑中的标题值

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

const formatRemindTime = (dateStr?: string) => {
    if (!dateStr) return ''
    const date = new Date(dateStr)
    return `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

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
    
    e.preventDefault()
    const text = e.clipboardData?.getData('text/plain') || ''
    document.execCommand('insertText', false, text)
}

// 上传图片
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

// 保存逻辑
const handleSave = async () => {
    // 如果没有编辑器引用或当前便签ID为0且内容为空，不保存
    if (!editorRef.value) return
    
    const content = editorRef.value.innerHTML
    const text = editorRef.value.innerText.trim()
    
    // 如果是新建的空白便签且用户没有输入内容，不保存
    if (!currentNote.value.id && !text && !content) {
        return
    }
    
    try {
        // 使用当前标题，不自动生成
        const title = currentNote.value.title || '无标题'
        const saveId = currentNote.value.id || 0
        
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

const saveContent = useDebounceFn(handleSave, 1000)

// 监听当前便签变化，同步到编辑器
watch(() => currentNote.value, (newNote) => {
    if (editorRef.value && newNote && currentView.value === 'editor') {
        // 只在内容不同时才更新，避免光标跳动
        if (editorRef.value.innerHTML !== newNote.content) {
            editorRef.value.innerHTML = newNote.content || ''
            // 查看模式下禁用编辑
            if (!isEditMode.value) {
                editorRef.value.contentEditable = 'false'
            }
        }
    }
}, { deep: true })

// 监听视图变化，从提醒页返回编辑页时恢复内容
watch(() => currentView.value, (newView, oldView) => {
    // 只在从提醒页返回编辑页时才需要恢复内容
    if (newView === 'editor' && oldView === 'remind') {
        // 使用 nextTick 等待编辑器DOM渲染完成
        nextTick(() => {
            if (editorRef.value && currentNote.value) {
                editorRef.value.innerHTML = currentNote.value.content || ''
                // 根据编辑模式设置可编辑状态
                editorRef.value.contentEditable = isEditMode.value ? 'true' : 'false'
            }
        })
    }
})

// 切换便签 - 进入查看模式
const selectNote = (note: NotepadInfo) => {
    currentNote.value = { ...note }
    currentRepeatType.value = note.remindRepeat || 'none'
    currentAdvanceDays.value = note.remindAdvanceDays || 0
    
    if (currentRepeatType.value === 'none' || currentRepeatType.value === 'daily') {
        currentAdvanceDays.value = 0
        currentNote.value.remindAdvanceDays = 0
    }
    
    // 切换到查看视图(非编辑模式)
    currentView.value = 'editor'
    isEditMode.value = false
    
    // 等待视图渲染后设置内容
    nextTick(() => {
        if (editorRef.value) {
            editorRef.value.innerHTML = note.content || ''
            // 查看模式下禁用编辑
            editorRef.value.contentEditable = 'false'
        }
    })
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
    currentView.value = 'editor'
    isEditMode.value = true // 新建时直接进入编辑模式
    nextTick(() => {
        if (editorRef.value) {
            editorRef.value.innerHTML = ''
            editorRef.value.contentEditable = 'true'
            editorRef.value.focus()
        }
    })
}

// 开始编辑标题
const startMobileEditing = (e?: Event) => {
    if (e) {
        e.stopPropagation()
    }
    isEditingTitle.value = true
    editingTitleValue.value = currentNote.value.title || ''
}

// 保存移动端标题
const saveMobileTitle = async () => {
    const newTitle = editingTitleValue.value.trim() || '无标题'
    
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
    
    isEditingTitle.value = false
    editingTitleValue.value = ''
}

// 关闭
const handleClose = () => {
    if (currentView.value === 'editor') {
        // 在编辑页，如果在编辑模式且有修改，提示保存
        if (isEditMode.value) {
            dialog.warning({
                title: '提示',
                content: '是否保存当前修改？',
                positiveText: '保存',
                negativeText: '不保存',
                onPositiveClick: async () => {
                    await handleSave()
                    currentView.value = 'list'
                    isEditMode.value = false
                },
                onNegativeClick: () => {
                    currentView.value = 'list'
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
            // 查看模式直接返回
            currentView.value = 'list'
        }
    } else if (currentView.value === 'remind') {
        // 在提醒页，直接返回编辑页
        currentView.value = 'editor'
    } else {
        // 在列表页，关闭整个记事本
        emit('update:visible', false)
    }
}

// 返回上一级
const goBack = () => {
    if (currentView.value === 'editor') {
        // 如果在编辑模式，提示保存
        if (isEditMode.value) {
            dialog.warning({
                title: '提示',
                content: '是否保存当前修改？',
                positiveText: '保存',
                negativeText: '不保存',
                onPositiveClick: async () => {
                    await handleSave()
                    currentView.value = 'list'
                    isEditMode.value = false
                },
                onNegativeClick: () => {
                    currentView.value = 'list'
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
            currentView.value = 'list'
        }
    } else if (currentView.value === 'remind') {
        currentView.value = 'editor'
    }
}

// 打开提醒视图
const openRemindView = () => {
    // 进入提醒页前，先保存当前编辑器内容到 currentNote
    if (editorRef.value && currentView.value === 'editor') {
        currentNote.value.content = editorRef.value.innerHTML
    }
    
    currentRepeatType.value = currentNote.value.remindRepeat || 'none'
    
    if (currentNote.value.remindAdvanceDays !== undefined) {
        currentAdvanceDays.value = currentNote.value.remindAdvanceDays
    } else {
        currentAdvanceDays.value = 0
    }
    
    if (currentRepeatType.value === 'none' || currentRepeatType.value === 'daily') {
        currentAdvanceDays.value = 0
    }
    
    const timeStr = currentNote.value.remindBaseTime || currentNote.value.remindTime
    if (timeStr) {
        const date = new Date(timeStr)
        if (!isNaN(date.getTime())) {
            remindTimestamp.value = date.getTime()
        }
    } else {
        remindTimestamp.value = null
    }
    
    currentView.value = 'remind'
}

// 取消提醒设置（不保存直接返回）
const handleCancelRemindSetting = () => {
    // 直接切换视图，由 watch 负责恢复编辑器内容
    currentView.value = 'editor'
}

// 确认提醒
const handleConfirmRemind = () => {
    if (remindTimestamp.value) {
        setRemind(remindTimestamp.value, true)
    } else {
        setRemind(null, true)
    }
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
    if (!currentNote.value.id) {
        message.warning('请先保存便签后再设置提醒')
        return
    }
    try {
        const remindTime = timestamp ? formatLocalDateTime(new Date(timestamp)) : null
        
        // 保存提醒设置时，使用 currentNote.value.content（已在进入提醒页前同步）
        const contentToSave = currentNote.value.content || ''
        
        await saveNotepadContent({
            id: currentNote.value.id,
            title: currentNote.value.title || '',
            content: contentToSave,
            remindTime: remindTime,
            remindStatus: remindTime ? 0 : 2,
            remindRepeat: currentRepeatType.value,
            remindAdvanceDays: currentAdvanceDays.value
        })
        
        // 只更新提醒相关字段，不覆盖整个 currentNote 对象
        currentNote.value.remindTime = remindTime || undefined
        currentNote.value.remindStatus = remindTime ? 0 : 2
        currentNote.value.remindRepeat = currentRepeatType.value
        currentNote.value.remindAdvanceDays = currentAdvanceDays.value
        // 确保 content 不被覆盖
        currentNote.value.content = contentToSave
        
        // 更新本地缓存中的便签内容
        const index = noteList.value.findIndex(n => n.id === currentNote.value.id)
        if (index !== -1) {
            noteList.value[index].remindTime = remindTime || undefined
            noteList.value[index].remindStatus = remindTime ? 0 : 2
            noteList.value[index].remindRepeat = currentRepeatType.value
            noteList.value[index].remindAdvanceDays = currentAdvanceDays.value
            noteList.value[index].content = contentToSave
        }
        
        emit('remindStatusChanged', currentNote.value.id)
        
        if (remindTime) {
            const repeatText = getRepeatText(currentRepeatType.value)
            const advanceText = currentAdvanceDays.value > 0 ? ` [提前${currentAdvanceDays.value}天]` : ''
            message.success(`已设置提醒：${formatRemindTime(remindTime)} ${repeatText !== '不重复' ? '(' + repeatText + ')' : ''}${advanceText}`)
        } else {
            message.success('已取消提醒')
        }
        
        if (autoClose) {
            currentView.value = 'editor'
        }
    } catch (e) {
        console.error('设置提醒失败:', e)
        message.error('设置提醒失败')
    }
}

const handleRepeatChange = () => {
    currentNote.value.remindRepeat = currentRepeatType.value
    
    if (currentRepeatType.value === 'none' || currentRepeatType.value === 'daily') {
        currentAdvanceDays.value = 0
        currentNote.value.remindAdvanceDays = 0
    }
}

const handleAdvanceChange = () => {
    currentNote.value.remindAdvanceDays = currentAdvanceDays.value
}

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

const showAdvanceDays = computed(() => {
    return currentRepeatType.value !== 'none' && currentRepeatType.value !== 'daily'
})

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

const nextActualRemindTime = computed(() => {
    if (currentNote.value.remindTime) {
        const remindTime = new Date(currentNote.value.remindTime)
        return formatLocalDate(remindTime)
    }
    return ''
})

const formatLocalDate = (date: Date): string => {
    const y = date.getFullYear()
    const m = String(date.getMonth() + 1).padStart(2, '0')
    const d = String(date.getDate()).padStart(2, '0')
    return `${y}-${m}-${d}`
}

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
    }
    // 关闭时不再自动保存，由用户手动控制
})

const initData = async () => {
    // 先异步加载列表
    await loadList()
    
    // 移动端：每次打开都先显示列表，不自动恢复上次的便签
    currentView.value = 'list'
    currentNote.value = { id: 0, title: '', content: '' }
}

defineExpose({ refreshData: loadList })
</script>

<style scoped lang="less">
.mobile-notepad-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  width: 100vw;
  height: 100vh;
  height: 100dvh; /* iOS Safari 支持动态视口高度 */
  background: rgba(0, 0, 0, 0.3);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  touch-action: none;
}

.mobile-notepad-container {
  width: 100vw;
  height: 100vh;
  height: 100dvh; /* iOS Safari 支持动态视口高度 */
  background: #fff;
  overflow: hidden;
  touch-action: manipulation;
  -webkit-tap-highlight-color: transparent;
}

.mobile-view {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

// 顶部导航栏
.mobile-header,
.mobile-editor-header,
.mobile-remind-header {
  padding: 12px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #e0e0e0;
  background: #f5f5f7;
  min-height: 56px;
  touch-action: manipulation;
  -webkit-tap-highlight-color: transparent;
}

// 列表页标题居中
.mobile-header {
  position: relative;
  justify-content: center;
}

// 列表页按钮绝对定位
.mobile-header .header-actions {
  position: absolute;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
}

// 编辑页和提醒页按钮正常布局
.mobile-editor-header .header-actions,
.mobile-remind-header .header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  z-index: 1;
}

// 保存按钮
.save-btn {
  padding: 6px 12px;
  background: #007aff;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  touch-action: manipulation;
  
  &:active {
    opacity: 0.7;
    transform: scale(0.95);
  }
}

// 取消编辑按钮
.cancel-edit-btn {
  padding: 6px 12px;
  background: #f5f5f7;
  color: #1d1d1f;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  touch-action: manipulation;
  
  &:active {
    opacity: 0.7;
    transform: scale(0.95);
  }
}

.mobile-editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: #fff;
  border-bottom: 1px solid #e5e5ea;
  position: relative;
}

.nav-left {
  display: flex;
  align-items: center;
  gap: 0;
  flex: 1;
  min-width: 0; // 允许flex子元素收缩
}

.back-icon {
  font-size: 24px;
  color: #007aff;
  cursor: pointer;
  min-width: 32px;
  min-height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  touch-action: manipulation;
  
  &:active {
    opacity: 0.5;
    transform: scale(0.95);
  }
}

.back-btn-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  touch-action: manipulation;
  cursor: pointer;
  
  &:active {
    opacity: 0.5;
    background: rgba(0, 122, 255, 0.1);
    border-radius: 4px;
  }
}

.back-text {
  font-size: 16px;
  font-weight: 500;
  color: #007aff;
  white-space: nowrap;
}

.mobile-title,
.mobile-editor-title,
.mobile-remind-title {
  font-size: 20px;
  font-weight: 700;
  color: #1d1d1f;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
  touch-action: manipulation;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  transition: all 0.2s;
  flex: 1; // 占据剩余空间
  min-width: 0; // 允许收缩
  
  &:active {
    opacity: 0.7;
  }
  
  // 编辑模式下显示可编辑提示
  &.editable {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    
    &::after {
      content: '单击编辑';
      font-size: 11px;
      font-weight: normal;
      color: #86868b;
      opacity: 0.6;
      white-space: nowrap;
    }
    
    &:active::after {
      opacity: 1;
    }
  }
}

.mobile-header .mobile-title {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  text-align: center;
}

.mobile-editor-title-input {
  font-size: 18px;
  font-weight: 600;
  color: #1d1d1f;
  border: none;
  outline: none;
  background: transparent;
  padding: 2px 4px;
  border-radius: 4px;
  width: 100%;
  max-width: 200px;
  
  &:focus {
    background: rgba(0, 122, 255, 0.1);
  }
}

.confirm-btn {
  padding: 6px 16px;
  background: #007aff;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  touch-action: manipulation;
  transition: all 0.1s ease;
  
  &:active {
    opacity: 0.7;
    transform: scale(0.96);
  }
}

.cancel-btn-small {
  padding: 6px 12px;
  background: transparent;
  color: #86868b;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  touch-action: manipulation;
  transition: all 0.1s ease;
  
  &:active {
    background: #f5f5f7;
    transform: scale(0.96);
  }
}

.action-icon {
  cursor: pointer;
  font-size: 20px;
  color: #86868b;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  touch-action: manipulation;
}

// 搜索框
.mobile-search-box {
  padding: 12px 16px;
  position: relative;
  background: #fff;
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
  font-size: 16px;
  outline: none;
  
  &::placeholder {
    color: #86868b;
  }
}

// 列表
.mobile-note-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  -webkit-overflow-scrolling: touch;
}

.mobile-note-item {
  padding: 14px 12px;
  border-radius: 8px;
  cursor: pointer;
  margin-bottom: 4px;
  min-height: 60px;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  touch-action: manipulation;
  transition: background 0.1s ease;
  
  &:active {
    background: #e8e8ed;
    transform: scale(0.98);
  }
}

.mobile-note-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
  gap: 8px;
}

.mobile-note-item-title {
  font-size: 16px;
  font-weight: 600;
  color: #1d1d1f;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.remind-icon {
  color: #86868b;
  font-size: 18px;
  flex-shrink: 0;
  
  &.active {
    color: #007aff;
  }
}

.delete-icon {
  color: #86868b;
  font-size: 20px;
  cursor: pointer;
  flex-shrink: 0;
  min-width: 32px;
  min-height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  touch-action: manipulation;
  transition: all 0.1s ease;
  
  &:active {
    color: #ff3b30;
    transform: scale(0.9);
    background: rgba(255, 59, 48, 0.1);
    border-radius: 4px;
  }
}

.mobile-note-item-time {
  font-size: 13px;
  color: #86868b;
}

.mobile-footer {
  padding: 16px;
  border-top: 1px solid #e0e0e0;
  display: flex;
  justify-content: center;
  background: #f5f5f7;
}

.new-note-btn {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: #f5f5f7; /* 浅色模式：灰色底色 */
  border: none;
  color: #007aff; /* 图标保持蓝色 */
  font-size: 28px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  touch-action: manipulation;
  transition: transform 0.1s ease;
  
  &:active {
    transform: scale(0.9);
  }
}

// 编辑器
.mobile-editor-content {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  outline: none;
  font-size: 16px;
  line-height: 1.6;
  color: #1d1d1f;
  -webkit-overflow-scrolling: touch;
  
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

.mobile-editor-footer {
  padding: 12px 16px;
  border-top: 1px solid #e0e0e0;
  background: #f5f5f7;
}

.footer-text {
  font-size: 12px;
  color: #86868b;
}

// 悬浮提醒按钮
.remind-float-btn {
  position: absolute;
  bottom: 80px;
  right: 16px;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: white; /* 浅色模式：白色底色 */
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 24px;
  color: #007aff; /* 图标保持蓝色 */
  
  &:active {
    transform: scale(0.95);
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
  width: 10px;
  height: 10px;
  background: #ff3b30;
  border-radius: 50%;
  border: 2px solid white;
}

// 提醒设置页
.mobile-remind-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  -webkit-overflow-scrolling: touch;
}

.remind-section {
  margin-bottom: 24px;
}

.section-label {
  font-size: 15px;
  font-weight: 600;
  color: #1d1d1f;
  margin-bottom: 12px;
}

.mobile-date-picker {
  width: 100%;
}

.radio-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.radio-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  background: #f5f5f7;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  touch-action: manipulation;
  transition: background 0.1s ease;
  
  &:active {
    background: #e8e8ed;
    transform: scale(0.98);
  }
  
  input[type="radio"] {
    width: 20px;
    height: 20px;
    accent-color: #007aff;
    pointer-events: none;
  }
  
  span {
    font-size: 16px;
    color: #1d1d1f;
  }
}

.advance-select {
  width: 100%;
  padding: 12px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  background: #f5f5f7;
  font-size: 16px;
  color: #1d1d1f;
  outline: none;
  -webkit-tap-highlight-color: transparent;
  touch-action: manipulation;
}

.remind-info-box {
  padding: 16px;
  background: rgba(0, 122, 255, 0.05);
  border-radius: 8px;
  border: 1px solid rgba(0, 122, 255, 0.2);
  margin-bottom: 24px;
}

.info-label {
  font-size: 14px;
  font-weight: 600;
  color: #007aff;
  margin-bottom: 8px;
}

.info-content {
  font-size: 14px;
  color: #1d1d1f;
  line-height: 1.6;
}

.repeat-badge {
  display: inline-block;
  margin-left: 8px;
  padding: 2px 8px;
  background: #007aff;
  color: white;
  border-radius: 12px;
  font-size: 11px;
}

.advance-badge {
  display: inline-block;
  margin-left: 6px;
  padding: 2px 8px;
  background: #ff9500;
  color: white;
  border-radius: 12px;
  font-size: 11px;
}

.cancel-section {
  margin-top: 24px;
}

.cancel-btn {
  width: 100%;
  padding: 14px;
  background: #fff;
  color: #ff3b30;
  border: 1px solid #ff3b30;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  touch-action: manipulation;
  transition: all 0.1s ease;
  
  &:active {
    background: #ffebee;
    transform: scale(0.98);
  }
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

// 深色模式
.mobile-notepad-container.dark-mode {
  background: #18181c;
}

.mobile-notepad-container.dark-mode .mobile-header,
.mobile-notepad-container.dark-mode .mobile-editor-header,
.mobile-notepad-container.dark-mode .mobile-remind-header,
.mobile-notepad-container.dark-mode .mobile-footer,
.mobile-notepad-container.dark-mode .mobile-editor-footer {
  background: #1c1c20;
  border-color: #2c2c32;
}

.mobile-notepad-container.dark-mode .mobile-title,
.mobile-notepad-container.dark-mode .mobile-editor-title,
.mobile-notepad-container.dark-mode .mobile-remind-title,
.mobile-notepad-container.dark-mode .mobile-note-item-title,
.mobile-notepad-container.dark-mode .section-label {
  color: rgba(255, 255, 255, 0.9);
}

.mobile-notepad-container.dark-mode .action-icon,
.mobile-notepad-container.dark-mode .delete-icon,
.mobile-notepad-container.dark-mode .remind-icon,
.mobile-notepad-container.dark-mode .mobile-note-item-time,
.mobile-notepad-container.dark-mode .footer-text,
.mobile-notepad-container.dark-mode .search-icon,
.mobile-notepad-container.dark-mode .cancel-btn-small {
  color: rgba(255, 255, 255, 0.52);
}

.mobile-notepad-container.dark-mode .search-input {
  background: #2c2c32;
  color: rgba(255, 255, 255, 0.9);
}

.mobile-notepad-container.dark-mode .mobile-search-box {
  background: #18181c;
}

.mobile-notepad-container.dark-mode .search-input::placeholder {
  color: rgba(255, 255, 255, 0.52);
}

.mobile-notepad-container.dark-mode .mobile-note-item:active {
  background: #2c2c32;
}

.mobile-notepad-container.dark-mode .mobile-editor-content {
  color: rgba(255, 255, 255, 0.9);
  background: #18181c;
}

.mobile-notepad-container.dark-mode .mobile-editor-content:empty:before {
  color: rgba(255, 255, 255, 0.52);
}

// 只读模式深色背景
.mobile-notepad-container.dark-mode .mobile-editor-content.readonly {
  background: #2c2c2e;
}

.mobile-notepad-container.dark-mode .remind-float-btn {
  background: #1c1c20;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.mobile-notepad-container.dark-mode .remind-float-btn {
  background: #1c1c20; /* 深色模式：深色底色 */
  color: #0a84ff; /* 图标保持蓝色 */
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.mobile-notepad-container.dark-mode .remind-float-btn.active {
  background: #0a84ff;
}

.mobile-notepad-container.dark-mode .radio-item {
  background: #1c1c20;
}

.mobile-notepad-container.dark-mode .radio-item:active {
  background: #2c2c32;
}

.mobile-notepad-container.dark-mode .radio-item span {
  color: rgba(255, 255, 255, 0.9);
}

.mobile-notepad-container.dark-mode .advance-select {
  background: #2c2c32;
  color: rgba(255, 255, 255, 0.9);
  border-color: #2c2c32;
}

.mobile-notepad-container.dark-mode .info-content {
  color: rgba(255, 255, 255, 0.9);
}

.mobile-notepad-container.dark-mode .cancel-btn {
  background: #2c2c32;
  color: #ff453a;
  border-color: #ff453a;
}

.mobile-notepad-container.dark-mode .cancel-btn:active {
  background: rgba(255, 69, 58, 0.15);
}

// 提醒选择器深色背景
.mobile-notepad-container.dark-mode .remind-picker-overlay {
  background: #18181c;
}

.mobile-notepad-container.dark-mode .remind-picker-box {
  background: #1c1c20;
  border-color: #2c2c32;
}

// 新建按钮：底色随深浅模式变化，但图标保持蓝色
.mobile-notepad-container.dark-mode .new-note-btn {
  background: #2c2c32; /* 深色模式：深色底色 */
  color: #0a84ff; /* 图标保持蓝色 */
}
</style>
