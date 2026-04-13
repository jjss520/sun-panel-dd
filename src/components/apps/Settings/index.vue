<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NButton, NCard, NInput, useMessage } from 'naive-ui'
import { t } from '@/locales'
import { setSystemSettings, getSystemSettings } from '@/api/system/systemSetting'

const ms = useMessage()

// 自动判断内外网设置
const pingUrl = ref('http://192.168.1.1:3000/ping')
const testStatus = ref<'idle' | 'testing' | 'success' | 'failed'>('idle')
const isTesting = ref(false)
const isSaving = ref(false)
const isLoading = ref(false)

// 加载设置
async function loadSettings() {
  isLoading.value = true
  try {
    const response = await getSystemSettings<Record<string, string>>(['pingUrl'])
    if (response.code === 0 && response.data) {
      // 如果有保存的值,使用保存的值
      if (response.data.pingUrl) {
        pingUrl.value = response.data.pingUrl
      }
    }
  } catch (error) {
    console.error('加载设置失败:', error)
  } finally {
    isLoading.value = false
  }
}

// 测试连接
async function testConnection() {
  if (!pingUrl.value.trim()) {
    ms.warning(t('apps.settings.pleaseEnterUrl'))
    return
  }

  // 确保URL包含协议
  let url = pingUrl.value.trim()
  if (!url.startsWith('http://') && !url.startsWith('http://')) {
    url = 'http://' + url
  }

  isTesting.value = true
  testStatus.value = 'testing'

  try {
    // 创建AbortController用于超时控制
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 150)

    // 使用 no-cors 模式避免 CORS 错误
    // 这种模式下无法读取响应内容,但可以判断请求是否成功发出
    const response = await fetch(url, {
      method: 'GET',
      // mode: 'no-cors', // 移除 no-cors 以获取状态码
      signal: controller.signal,
    })

    clearTimeout(timeoutId)

    if (response.status === 200) {
      testStatus.value = 'success'
    } else {
      testStatus.value = 'failed'
    }
  } catch (error: any) {
    // 超时或其他错误(包括地址不可达)
    testStatus.value = 'failed'
  } finally {
    isTesting.value = false
  }
}

// 保存设置
async function saveSettings() {
  isSaving.value = true
  try {
    const response = await setSystemSettings({
      pingUrl: pingUrl.value,
    })
    
    if (response.code === 0) {
      ms.success(t('common.saveSuccess'))
    } else {
      ms.error(t('common.saveFail') + ': ' + response.msg)
    }
  } catch (error) {
    console.error('保存设置失败:', error)
    ms.error(t('common.saveFail'))
  } finally {
    isSaving.value = false
  }
}

// 组件挂载时加载设置
onMounted(() => {
  loadSettings()
})
</script>

<template>
  <div class="bg-slate-200 dark:bg-zinc-900 rounded-[10px] p-[8px] overflow-auto">
    <!-- 网络检测设置 -->
    <NCard style="border-radius:10px" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        {{ t('apps.settings.networkDetection') }}
      </div>

      <div class="mt-[15px]">
        <div class="text-sm text-gray-600 dark:text-gray-400 mb-[5px]">
          {{ t('apps.settings.pingUrlHint') }}
        </div>
        <NInput
          v-model:value="pingUrl"
          type="text"
          :placeholder="t('apps.settings.pingUrlPlaceholder')"
          clearable
        />
      </div>

      <!-- 测试连接按钮 -->
      <div class="flex items-center mt-[10px]">
        <NButton
          size="small"
          :loading="isTesting"
          @click="testConnection"
        >
          {{ t('apps.settings.testConnection') }}
        </NButton>

        <!-- 状态显示 -->
        <span
          v-if="testStatus === 'success'"
          class="ml-[10px] text-green-600 dark:text-green-400"
        >
          ✓ {{ t('apps.settings.connectionSuccess') }}
        </span>
        <span
          v-else-if="testStatus === 'failed'"
          class="ml-[10px] text-red-600 dark:text-red-400"
        >
          ✗ {{ t('apps.settings.connectionFailed') }}
        </span>
        <span
          v-else-if="testStatus === 'testing'"
          class="ml-[10px] text-gray-600 dark:text-gray-400"
        >
          {{ t('apps.settings.testing') }}
        </span>
      </div>
    </NCard>

    <!-- 保存按钮 -->
    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <NButton size="small" type="primary" :loading="isSaving" @click="saveSettings">
        {{ $t('common.save') }}
      </NButton>
    </NCard>
  </div>
</template>

<style scoped>
.text-shadow {
  text-shadow: 0px 0px 5px gray;
}
</style>
