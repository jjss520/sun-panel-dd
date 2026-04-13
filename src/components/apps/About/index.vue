<script setup lang="ts">
import { NDivider, NGradientText, useMessage } from 'naive-ui'
import { onMounted, ref, computed } from 'vue'
import { get, checkUpdate } from '@/api/system/about'
import srcSvglogo from '@/assets/logo.svg'
import { NButton, NSpin, NCard } from 'naive-ui'
import { t } from '@/locales'

interface Version {
  versionName: string
  versionCode: number
}

const versionName = ref('')
const ms = useMessage()

onMounted(() => {
  get<Version>().then((res) => {
    if (res.code === 0)
      versionName.value = res.data.versionName
  })
})

const checking = ref(false)
const hasNewVersion = ref(false)
const latestVersion = ref('')
const updateContent = ref('')
const showUpdateInfo = ref(false)

// 格式化更新内容，处理换行符
const formattedUpdateContent = computed(() => {
  if (!updateContent.value) return ''
  // 将\r\n和\n替换为HTML换行符
  return updateContent.value.replace(/\r\n|\n/g, '<br/>')
})

const handleCheckUpdate = async () => {
  checking.value = true
  showUpdateInfo.value = false
  try {
    const { code, data, msg } = await checkUpdate()
    if (code === 0) {
      hasNewVersion.value = data.hasNewVersion
      latestVersion.value = data.latestVersion
      updateContent.value = data.updateContent
      showUpdateInfo.value = true
      
      // 如果有新版本，提示用户
      if (data.hasNewVersion) {
        ms.warning(t('apps.about.newVersionAvailable') || '发现新版本！')
      }
    } else {
      ms.error(msg || (t('apps.about.checkUpdateFailed') || '检查更新失败'))
    }
  } catch (e) {
      ms.error(t('apps.about.checkUpdateFailed') || '检查更新失败')
  } finally {
    checking.value = false
  }
}
</script>

<template>
  <div class="pt-5">
    <div class="flex flex-col items-center justify-center">
      <img :src="srcSvglogo" width="100" height="100" alt="">
      <div class="text-3xl font-semibold">
        {{ $t('common.appName') }}
      </div>
      <div class="text-xl">
        <NGradientText type="info">
          <a href="https://github.com/jjss520/sun-panel-dd/releases" class="font-semibold" :title="$t('apps.about.viewUpdateLog')" target="_blank">v{{ versionName }}</a>
        </NGradientText>
      </div>
      <div class="mt-4">
        <NSpin :show="checking">
            <template v-if="!showUpdateInfo">
                <a href="javascript:;" @click="handleCheckUpdate" class="link hover:underline">{{ $t('apps.about.checkUpdate') }}</a>
            </template>
            <template v-else>
                <div v-if="!hasNewVersion" class="text-green-500 font-medium">
                    {{ $t('common.currentIsLatestVersion') || '当前已是最新版本' }}
                </div>
                <div v-else class="flex flex-col items-center">
                    <div class="text-orange-500 font-bold mb-2">
                        {{ t('apps.about.newVersion') || '新版本' }}: v{{ latestVersion }}
                    </div>
                    <NCard size="small" class="w-full max-w-md bg-gray-50 dark:bg-gray-800 mb-3 text-left">
                        <div class="text-sm text-gray-700 dark:text-gray-300 max-h-40 overflow-y-auto" v-html="formattedUpdateContent">
                        </div>
                    </NCard>
                    <NButton type="primary" tag="a" href="https://github.com/jjss520/sun-panel-dd/releases" target="_blank">
                        {{ t('apps.about.goToUpdate') || '去更新' }}
                    </NButton>
                </div>
            </template>
        </NSpin>
      </div>
    </div>

    <NDivider style="margin:10px 0">
      •
    </NDivider>
  </div>
</template>

<style>
.link{
    color:rgb(0, 89, 255)
}
</style>
