<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import type { UploadFileInfo } from 'naive-ui'
import { NButton, NCard, NColorPicker, NGrid, NGridItem, NInput, NInputGroup, NPopconfirm, NSelect, NSlider, NSwitch, NUpload, NUploadDragger, useMessage } from 'naive-ui'
import { useAuthStore, usePanelState } from '@/store'
import { set as setUserConfig } from '@/api/panel/userConfig'
import { PanelPanelConfigStyleEnum } from '@/enums/panel'
import { t } from '@/locales'
import { ss } from '@/utils/storage'
import defaultBackground from '@/assets/defaultBackground.webp'

// 用户配置缓存键
const USER_CONFIG_CACHE_KEY = 'USER_CONFIG_CACHE'

const authStore = useAuthStore()
const panelState = usePanelState()
const ms = useMessage()
const showWallpaperInput = ref(false)

const isSaveing = ref(false)

const iconTypeOptions = [
  {
    label: t('apps.baseSettings.detailIcon'),
    value: PanelPanelConfigStyleEnum.info,
  },
  {
    label: t('apps.baseSettings.smallIcon'),
    value: PanelPanelConfigStyleEnum.icon,
  },
]

const maxWidthUnitOption = [
  {
    label: 'px',
    value: 'px',
  },
  {
    label: '%',
    value: '%',
  },
]

// Computed property to handle string conversion for maxWidth
const maxWidthString = computed({
  get: () => String(panelState.panelConfig.maxWidth),
  set: (value) => {
    const numValue = parseInt(value, 10)
    if (!isNaN(numValue)) {
      panelState.panelConfig.maxWidth = numValue
    }
  }
})

watch(panelState.panelConfig, () => {
  if (!isSaveing.value) {
    isSaveing.value = true

    setTimeout(() => {
      panelState.recordState()// 本地记录
      isSaveing.value = false
      uploadCloud()
    }, 1000)
  }
})

function handleUploadBackgroundFinish({
  file,
  event,
}: {
  file: UploadFileInfo
  event?: ProgressEvent
}) {
  const res = JSON.parse((event?.target as XMLHttpRequest).response)
  if (res.code === 0) {
    // 上传成功后，关闭 BING 壁纸和网络壁纸开关
    panelState.panelConfig.useBingWallpaper = false
    panelState.panelConfig.autoNetworkWallpaper = false
    
    // 设置上传的壁纸为当前背景
    panelState.panelConfig.backgroundImageSrc = res.data.imageUrl
    
    ms.success('壁纸上传成功')
  } else {
    ms.error(`上传失败: ${res.msg}`)
  }
  return file
}

function uploadCloud() {
  setUserConfig({ panel: panelState.panelConfig }).then((res) => {
    if (res.code === 0) {
      ms.success(t('apps.baseSettings.configSaved'))
      // 清除用户配置缓存，确保下次加载时获取最新配置
      ss.remove(USER_CONFIG_CACHE_KEY)
    }
    else {
      ms.error(t('apps.baseSettings.configFailed', { message: res.msg }))
    }
  })
}

// BING 壁纸开关变化处理
function handleBingWallpaperChange(value: boolean) {
  if (value) {
    // 关闭"自动获取网络壁纸"开关（互斥）
    panelState.panelConfig.autoNetworkWallpaper = false
    
    // 设置固定路径（与后端下载路径保持一致）
    const bingPath = '/data/uploads/wallpapers/bing/bing_wallpaper.jpg'
    panelState.panelConfig.backgroundImageSrc = bingPath
    
    // 提示用户
    ms.success('已设置为 BING 每日壁纸')
  } else {
    // 关闭时恢复默认背景
    panelState.panelConfig.backgroundImageSrc = defaultBackground
  }
}

// 获取随机壁纸
function fetchRandomWallpaper() {
  const apiUrl = panelState.panelConfig.autoNetworkWallpaperApi || 'https://img.xjh.me/random_img.php?return=302&type=bg&ctype=nature'
  // 使用302跳转，直接将图片地址设置为背景图
  panelState.panelConfig.backgroundImageSrc = apiUrl
  // 保存配置
  uploadCloud()
}

// 自动获取网络壁纸开关变化处理
function handleAutoNetworkWallpaperChange() {
  if (panelState.panelConfig.autoNetworkWallpaper) {
    // 当开关打开时，关闭 BING 壁纸
    panelState.panelConfig.useBingWallpaper = false
    
    // 如果 API 地址为空，使用默认值
    if (!panelState.panelConfig.autoNetworkWallpaperApi) {
      panelState.panelConfig.autoNetworkWallpaperApi = 'https://img.xjh.me/random_img.php?return=302&type=bg&ctype=nature'
    }
    
    // 立即获取一张随机壁纸
    fetchRandomWallpaper()
  } else {
    // 关闭随机壁纸时，恢复默认背景
    panelState.panelConfig.backgroundImageSrc = defaultBackground
  }
}

function resetPanelConfig() {
  panelState.resetPanelConfig()
  uploadCloud()
}
</script>

<template>
  <div class="bg-slate-200 dark:bg-zinc-900 rounded-[10px] p-[8px] overflow-auto">
    <NCard style="border-radius:10px" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        LOGO
      </div>

      <div>
        <div>
          {{ $t('apps.baseSettings.textContent') }}
        </div>
        <div class="flex items-center mt-[5px]">
          <NInput v-model:value="panelState.panelConfig.logoText" type="text" show-count :maxlength="20" placeholder="请输入文字" />
        </div>
      </div>
    </NCard>

    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        {{ $t('apps.baseSettings.clock') }}
      </div>
      <div class="flex items-center mt-[5px]">
        <span class="mr-[10px]">{{ $t('apps.baseSettings.clockSecondShow') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.clockShowSecond" />
      </div>
    </NCard>

    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        {{ $t('apps.baseSettings.searchBar') }}
      </div>
      <div class="flex items-center mt-[5px]">
        <span class="mr-[10px]">{{ $t('common.show') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.searchBoxShow" />
      </div>
      <div v-if="panelState.panelConfig.searchBoxShow" class="flex items-center mt-[5px]">
        <span class="mr-[10px]">{{ $t('apps.baseSettings.searchBarSearchItem') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.searchBoxSearchIcon" />
      </div>
    </NCard>

    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        {{ $t('apps.baseSettings.systemMonitorStatus') }}
      </div>
      <div class="flex items-center mt-[5px]">
        <span class="mr-[10px]">{{ $t('common.show') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.systemMonitorShow" />
      </div>
      <div v-if="panelState.panelConfig.systemMonitorShow" class="flex items-center mt-[5px]">
        <span class="mr-[10px]">{{ $t('apps.baseSettings.showTitle') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.systemMonitorShowTitle" />
      </div>
      <div v-if="panelState.panelConfig.systemMonitorShow" class="flex items-center mt-[5px]">
        <span class="mr-[10px]">{{ $t('apps.baseSettings.publicVisitModeShow') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.systemMonitorPublicVisitModeShow" />
      </div>
    </NCard>

    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        {{ $t('common.icon') }}
      </div>
      <div class="mt-[5px]">
        <div>
          {{ $t('common.style') }}
        </div>
        <div class="flex items-center mt-[5px]">
          <NSelect v-model:value="panelState.panelConfig.iconStyle" :options="iconTypeOptions" />
        </div>
      </div>

      <div v-if="panelState.panelConfig.iconStyle === PanelPanelConfigStyleEnum.info" class="mt-[5px]">
        <div>
          {{ $t('apps.baseSettings.hideDescription') }}
        </div>
        <div class="flex items-center mt-[5px]">
          <NSwitch v-model:value="panelState.panelConfig.iconTextInfoHideDescription" />
        </div>
      </div>

      <div v-if="panelState.panelConfig.iconStyle === PanelPanelConfigStyleEnum.icon" class="mt-[5px]">
        <div>
          {{ $t('apps.baseSettings.hideTitle') }}
        </div>
        <div class="flex items-center mt-[5px]">
          <NSwitch v-model:value="panelState.panelConfig.iconTextIconHideTitle" />
        </div>
      </div>

      <div class="mt-[5px]">
        <div>
          {{ $t('common.textColor') }}
        </div>
        <div class="flex items-center mt-[5px]">
          <NColorPicker
            v-model:value="panelState.panelConfig.iconTextColor"
            :show-alpha="false"
            size="small"
            :modes="['hex']"
            :swatches="[
              '#000000',
              '#ffffff',
              '#18A058',
              '#2080F0',
              '#F0A020',
            ]"
          />
        </div>
      </div>
    </NCard>
    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        {{ $t('apps.baseSettings.wallpaper') }}
      </div>
      <NUpload
        action="/api/file/uploadWallpaper"
        :show-file-list="false"
        name="imgfile"
        :headers="{
          token: authStore.token as string,
        }"
        :directory-dnd="true"
        @finish="handleUploadBackgroundFinish"
      >
        <NUploadDragger style="width: 100%;">
          <div
            class="h-[200px] w-full border bg-slate-100 flex justify-center items-center cursor-pointer rounded-[10px]"
            :style="{ background: `url(${panelState.panelConfig.backgroundImageSrc}) no-repeat`, backgroundSize: 'cover' }"
          >
            <div class="text-shadow text-white">
              {{ $t('apps.baseSettings.uploadOrDragText') }}
            </div>
          </div>
        </NUploadDragger>
      </NUpload>

      <div class="flex items-center mt-[5px]">
        <span class="mr-[10px]">{{ $t('apps.baseSettings.customImageAddress') }}</span>
        <NSwitch v-model:value="showWallpaperInput" />
      </div>
      <div v-if="showWallpaperInput" class="mt-1">
        <NInput v-model:value="panelState.panelConfig.backgroundImageSrc" type="text" size="small" clearable />
      </div>

      <div class="flex items-center mt-[10px]">
        <span class="mr-[10px]">{{ $t('apps.baseSettings.vague') }}</span>
        <NSlider v-model:value="panelState.panelConfig.backgroundBlur" class="max-w-[200px]" :step="2" :max="20" />
      </div>

      <div class="flex items-center mt-[10px]">
        <span class="mr-[10px]">{{ $t('apps.baseSettings.mask') }}</span>
        <NSlider v-model:value="panelState.panelConfig.backgroundMaskNumber" class="max-w-[200px]" :step="0.1" :max="1" />
      </div>

      <div class="flex items-center mt-[10px]">
        <span class="mr-[10px]">BING 每日壁纸</span>
        <NSwitch v-model:value="panelState.panelConfig.useBingWallpaper" @update:value="handleBingWallpaperChange" />
      </div>
      <div v-if="panelState.panelConfig.useBingWallpaper" class="mt-1 text-sm text-slate-500">
        💡 服务器每天 0:05 自动下载最新 BING 壁纸
      </div>

      <div class="flex items-center mt-[10px]">
        <span class="mr-[10px]">{{ $t('apps.baseSettings.autoNetworkWallpaper') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.autoNetworkWallpaper" @click="handleAutoNetworkWallpaperChange" />
      </div>
      <div v-if="panelState.panelConfig.autoNetworkWallpaper" class="mt-1">
        <NInput v-model:value="panelState.panelConfig.autoNetworkWallpaperApi" type="text" size="small" clearable placeholder="请输入网络壁纸API地址" />
      </div>
    </NCard>

    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        {{ $t('apps.baseSettings.contentArea') }}
      </div>

      <NGrid cols="2">
        <NGridItem span="12 400:12">
          <div class="flex items-center mt-[5px]">
            <span class="mr-[10px]">{{ $t('apps.baseSettings.netModeChangeButtonShow') }}</span>
            <NSwitch v-model:value="panelState.panelConfig.netModeChangeButtonShow" />
          </div>
        </NGridItem>

        <NGridItem span="12 400:12">
          <div class="flex items-center mt-[5px]">
            <span class="mr-[10px]">显示编辑模式切换按钮</span>
            <NSwitch v-model:value="panelState.panelConfig.secretModeButtonShow" />
          </div>
        </NGridItem>

        <NGridItem span="12 400:12">
          <div class="flex items-center mt-[10px]">
            <span class="mr-[10px]">{{ $t('apps.baseSettings.maxWidth') }}</span>
            <div class="flex">
              <NInputGroup>
                <NInput v-model:value="maxWidthString" size="small" type="text" :maxlength="10" :style="{ width: '100px' }" placeholder="1200" />
                <NSelect v-model:value="panelState.panelConfig.maxWidthUnit" :style="{ width: '80px' }" :options="maxWidthUnitOption" size="small" />
              </NInputGroup>
            </div>
          </div>
        </NGridItem>
        <NGridItem span="12 400:12">
          <div class="flex items-center mt-[10px]">
            <span class="mr-[10px]">{{ $t('apps.baseSettings.leftRightMargin') }}</span>
            <NSlider v-model:value="panelState.panelConfig.marginX" class="max-w-[200px]" :step="1" :max="100" />
          </div>
        </NGridItem>
        <NGridItem span="12 400:12">
          <div class="flex items-center mt-[10px]">
            <span class="mr-[10px]">{{ $t('apps.baseSettings.topMargin') }} (%)</span>
            <NSlider v-model:value="panelState.panelConfig.marginTop" class="max-w-[200px]" :step="1" :max="50" />
          </div>
        </NGridItem>
        <NGridItem span="12 400:6">
          <div class="flex items-center mt-[10px]">
            <span class="mr-[10px]">{{ $t('apps.baseSettings.bottomMargin') }} (%)</span>
            <NSlider v-model:value="panelState.panelConfig.marginBottom" class="max-w-[200px]" :step="1" :max="50" />
          </div>
        </NGridItem>
      </NGrid>
    </NCard>

    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        {{ $t('apps.baseSettings.customFooter') }}
      </div>

      <NInput
        v-model:value="panelState.panelConfig.footerHtml"
        type="textarea"
        clearable
      />
    </NCard>

    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <NPopconfirm
        @positive-click="resetPanelConfig"
      >
        <template #trigger>
          <NButton size="small" quaternary type="error">
            {{ $t('common.reset') }}
          </NButton>
        </template>
        {{ $t('apps.baseSettings.resetWarnText') }}
      </NPopconfirm>

      <NButton size="small" quaternary type="success" class="ml-[10px]" @click="uploadCloud">
        {{ $t('common.save') }}
      </NButton>
    </NCard>
  </div>
</template>

<style scoped>
.text-shadow{
  text-shadow: 0px 0px 5px gray;
}
</style>
