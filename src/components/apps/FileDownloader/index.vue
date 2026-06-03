<script setup lang="ts">
import { NAlert, NEllipsis, NSpin, NSelect, useMessage } from 'naive-ui'
import { computed, onMounted, ref } from 'vue'
import { getList } from '@/api/system/file'
import { t } from '@/locales'
import { useAuthStore } from '@/store'

interface FileInfo {
  id: number
  fileName: string
  src: string
  createTime: string
  fileType?: string
}

const fileList = ref<FileInfo[]>([])
const ms = useMessage()
const loading = ref(false)
const currentPath = ref<string>('') // 当前选中的路径
const authStore = useAuthStore()

// 从文件列表中提取所有唯一的路径前缀
const availablePaths = computed(() => {
  const paths = new Set<string>()
  fileList.value.forEach(file => {
    // 提取路径的前几级目录
    const parts = file.src.split('/')
    if (parts.length > 1) {
      // 取前3部分作为路径（例如：/data/uploads/wallpapers）
      const pathPrefix = parts.slice(0, Math.min(4, parts.length)).join('/')
      paths.add(pathPrefix)
    }
  })
  
  // 转换为选项格式
  const options = Array.from(paths).map(path => ({
    label: path,
    value: path
  }))
  
  // 添加"全部"选项
  return [
    { label: t('apps.fileDownloader.allFiles') || '全部文件', value: '' },
    ...options
  ]
})

async function getFileList() {
  loading.value = true
  try {
    // 如果有路径筛选，添加到请求参数
    const { data } = await getList<Common.ListResponse<FileInfo[]>>(currentPath.value || undefined)
    fileList.value = data.list
  }
  catch (error) {
    ms.error(t('common.failed'))
  }
  finally {
    loading.value = false
  }
}

// 下载文件
async function downloadFile(filePath: string, fileName: string) {
  try {
    const token = authStore.token || ''
    const response = await fetch(`/api/file/download?path=${encodeURIComponent(filePath)}`, {
      method: 'GET',
      headers: {
        'token': token, // 添加token header
      },
    })
    
    if (!response.ok) {
      throw new Error('Download failed')
    }
    
    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = fileName
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    ms.success(t('common.downloading'))
  } catch (error) {
    ms.error(t('common.failed'))
  }
}

// 路径改变时重新加载
function handlePathChange(path: string) {
  currentPath.value = path
  getFileList()
}

onMounted(() => {
  getFileList()
})
</script>

<template>
  <div class="bg-slate-200 dark:bg-zinc-900 p-4" style="height: 100%; display: flex; flex-direction: column;">
    <NSpin v-show="loading" size="small" />
    
    <NAlert type="info" :bordered="false" class="mb-4 flex-shrink-0">
      {{ $t('apps.fileDownloader.alertText') || '点击文件即可下载' }}
    </NAlert>

    <!-- 路径选择器 -->
    <div class="mb-4 flex-shrink-0">
      <NSelect
        v-model:value="currentPath"
        :options="availablePaths"
        :placeholder="$t('apps.fileDownloader.selectPath') || '选择路径'"
        clearable
        @update:value="handlePathChange"
      />
    </div>

    <div v-if="fileList.length === 0 && !loading" class="flex justify-center items-center h-32 text-gray-500 flex-shrink-0">
      {{ $t('apps.fileDownloader.nothingText') || '暂无文件' }}
    </div>

    <div v-else class="space-y-2 flex-1" style="min-height: 0; overflow-y: auto; -webkit-overflow-scrolling: touch;">
      <div
        v-for="item in fileList"
        :key="item.id"
        class="bg-white dark:bg-zinc-800 rounded-lg p-3 flex items-center justify-between hover:bg-gray-50 dark:hover:bg-zinc-700 transition-colors cursor-pointer"
        @click="downloadFile(item.src, item.fileName)"
      >
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2 mb-1">
            <!-- 文件类型图标 -->
            <span v-if="item.fileType === 'wallpaper'" class="text-lg"></span>
            <span v-else class="text-lg">📄</span>
            
            <NEllipsis class="font-medium text-sm">
              {{ item.fileName }}
            </NEllipsis>
          </div>
          
          <div class="text-xs text-gray-500 dark:text-gray-400">
            {{ new Date(item.createTime).toLocaleString('zh-CN') }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 自定义滚动条样式 */
.overflow-auto::-webkit-scrollbar {
  width: 6px;
}

.overflow-auto::-webkit-scrollbar-track {
  background: transparent;
}

.overflow-auto::-webkit-scrollbar-thumb {
  background-color: rgba(0, 0, 0, 0.2);
  border-radius: 3px;
}

.dark .overflow-auto::-webkit-scrollbar-thumb {
  background-color: rgba(255, 255, 255, 0.2);
}
</style>
