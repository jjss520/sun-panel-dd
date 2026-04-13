<script setup lang="ts">
import { NAvatar, NImage } from 'naive-ui'
import { computed, ref, withDefaults } from 'vue'
import { SvgIconOnline } from '@/components/common'

interface Prop {
  itemIcon?: Panel.ItemIcon | null
  size?: number // 默认70
  forceBackground?: string // 强制背景色
}

const props = withDefaults(defineProps<Prop>(), { size: 70 })
const defaultBackground = '#2a2a2a6b'
const defaultStyle = ref({
  width: `${props.size}px`,
  height: `${props.size}px`,
})

// 计算内部元素的大小比例，使图标内容随外部尺寸等比缩放
const innerSize = computed(() => {
  // 内部元素占外部容器的 100%，完全填充
  return Math.round(props.size * 1.0)
})

const iconExt = computed(() => {
  return props.itemIcon?.src?.split('.').pop()
})
</script>

<template>
  <div class="item-icon" :style="defaultStyle">
    <slot>
      <template v-if="itemIcon">
        <template v-if="itemIcon?.itemType === 1">
          <NAvatar :size="props.size" :style="{ backgroundColor: (forceBackground ?? itemIcon?.backgroundColor) || defaultBackground, borderRadius: '12px' }">
            {{ itemIcon.text }}
          </NAvatar>
        </template>

        <template v-else-if="itemIcon?.itemType === 2">
          <div v-if="iconExt === 'svg'" :style="{ backgroundColor: (forceBackground ?? itemIcon?.backgroundColor) || defaultBackground, ...defaultStyle, borderRadius: '12px' }" class="flex justify-center items-center">
            <img :src="itemIcon?.src" :style="{ width: `${innerSize}px`, height: `${innerSize}px` }">
          </div>
          <NImage v-else :style="{ backgroundColor: (forceBackground ?? itemIcon?.backgroundColor) || defaultBackground, ...defaultStyle, borderRadius: '12px' }" :src="itemIcon?.src" preview-disabled />
        </template>

        <template v-else-if="itemIcon?.itemType === 3">
          <NAvatar :size="props.size" :style="{ backgroundColor: (forceBackground ?? itemIcon?.backgroundColor) || defaultBackground, borderRadius: '12px' }">
            <SvgIconOnline :style="{ fontSize: `${innerSize}px` }" :icon="itemIcon.text" />
          </NAvatar>
        </template>
      </template>

      <template v-else>
        <NAvatar :size="props.size" />
      </template>
    </slot>
  </div>
</template>
