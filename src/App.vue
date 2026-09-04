<script setup lang="ts">
import { NConfigProvider } from 'naive-ui'
import { NaiveProvider } from '@/components/common'
import { useTheme } from '@/hooks/useTheme'
import { useLanguage } from '@/hooks/useLanguage'

const { theme, themeOverrides } = useTheme()
const { language } = useLanguage()

// 移除一言脚本加载，避免没有hitokoto元素时的错误
// onMounted(() => {
//   const script = document.createElement('script')
//   script.src = 'https://v1.hitokoto.cn/?encode=js&select=%23hitokoto'
//   script.defer = true
//   document.body.appendChild(script)
// })
</script>

<template>
  <NConfigProvider
    class="h-full"
    :theme="theme"
    :theme-overrides="themeOverrides"
    :locale="language"
    :z-index="{
      dialog: 100002,
      dropdown: 100000,
      popover: 100000,
      tooltip: 100000,
      treeSelect: 100003
    }"
  >
    <NaiveProvider>
      <RouterView />
    </NaiveProvider>
  </NConfigProvider>
</template>
