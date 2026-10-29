<script setup lang="ts">
import { computed } from 'vue'
import { ItemIcon } from '@/components/common'

interface Prop {
  item?: Panel.ItemInfo | null
  miniSize?: number // 预览小图标尺寸
  showBadge?: boolean // 是否显示数量角标
}

const props = withDefaults(defineProps<Prop>(), {
  miniSize: 22,
  showBadge: true,
})

const children = computed<Panel.ItemInfo[]>(() => props.item?.children || [])

// 固定 4 个格子，不足的用占位格补齐，保证文件夹外形稳定
const cells = computed<(Panel.ItemInfo | null)[]>(() => {
  const list: (Panel.ItemInfo | null)[] = []
  for (let i = 0; i < 4; i++)
    list.push(children.value[i] || null)

  return list
})

const badge = computed(() => children.value.length)
</script>

<template>
  <div class="folder-plate">
    <div class="folder-mini-grid">
      <div
        v-for="(child, index) in cells"
        :key="index"
        class="folder-mini"
        :class="{ 'folder-mini-empty': !child }"
      >
        <ItemIcon v-if="child" :item-icon="child.icon" :size="miniSize" />
      </div>
    </div>
    <span v-if="showBadge" class="folder-badge">{{ badge }}</span>
  </div>
</template>

<style scoped>
.folder-plate {
  position: relative;
  width: 100%;
  height: 100%;
  border-radius: 1rem;
  background: rgba(255, 255, 255, 0.10);
}

.folder-mini-grid {
  position: absolute;
  inset: 9%;
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 1fr 1fr;
  gap: 6%;
}

.folder-mini {
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 24%;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.08);
}

.folder-mini-empty {
  background: rgba(255, 255, 255, 0.04);
  border: 1px dashed rgba(255, 255, 255, 0.18);
  box-shadow: none;
}

.folder-badge {
  position: absolute;
  right: -5px;
  bottom: -5px;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 9px;
  background: rgba(255, 255, 255, 0.94);
  color: #15181f;
  font-size: 11px;
  font-weight: 700;
  line-height: 18px;
  text-align: center;
  z-index: 2;
  pointer-events: none;
}

/* 让小图标填满格子，避免受内部固定尺寸影响 */
.folder-mini :deep(.item-icon) {
  width: 100% !important;
  height: 100% !important;
}

.folder-mini :deep(.n-avatar) {
  width: 100% !important;
  height: 100% !important;
  border-radius: 24% !important;
}
</style>
