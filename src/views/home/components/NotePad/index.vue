<template>
  <!-- PC端使用原组件，移动端使用新组件 -->
  <DesktopNotePad v-if="!isMobile" v-bind="$props" @update:visible="$emit('update:visible', $event)" @remind-status-changed="$emit('remindStatusChanged', $event)" />
  <MobileNotePad v-else v-bind="$props" @update:visible="$emit('update:visible', $event)" @remind-status-changed="$emit('remindStatusChanged', $event)" />
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import DesktopNotePad from './DesktopNotePad.vue'
import MobileNotePad from './MobileNotePad.vue'

defineProps<{
  visible: boolean
}>()

defineEmits<{
  (e: 'update:visible', visible: boolean): void
  (e: 'remindStatusChanged', noteId: number): void
}>()

// 检测设备类型（响应式）
const isMobile = ref(false)

const checkMobile = () => {
  if (typeof window !== 'undefined') {
    isMobile.value = window.innerWidth <= 768
  }
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>
