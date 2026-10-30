<script setup lang="ts">
import { computed, nextTick, onUnmounted, ref, watch } from 'vue'
import { NDropdown } from 'naive-ui'
import { ItemIcon, SvgIconOnline } from '@/components/common'
import AppIcon from '../AppIcon/index.vue'
import { PanelPanelConfigStyleEnum } from '@/enums'
import { t } from '@/locales'

interface MoveTarget {
  label: string
  value: number
}

interface Props {
  visible: boolean
  folder: Panel.ItemInfo | null
  iconStyle: PanelPanelConfigStyleEnum
  iconTextColor?: string
  iconTextInfoHideDescription?: boolean
  iconTextIconHideTitle?: boolean
  editable?: boolean // 登录 + 编辑模式，可增删改（控制头部工具按钮）
  menuEditable?: boolean // 仅登录即可（控制长按/右键菜单的编辑项，与首页菜单一致）
  isMobile?: boolean
  moveTargets?: MoveTarget[] // 可移动到的其他文件夹
}

const props = withDefaults(defineProps<Props>(), {
  editable: false,
  menuEditable: false,
  isMobile: false,
  moveTargets: () => [],
})

const emit = defineEmits<{
  (e: 'update:visible', visible: boolean): void
  (e: 'select', item: Panel.ItemInfo): void
  (e: 'open-url', item: Panel.ItemInfo, target: 'wan' | 'lan'): void
  (e: 'edit', item: Panel.ItemInfo): void
  (e: 'add-item', parentId: number): void
  (e: 'delete', item: Panel.ItemInfo): void
  (e: 'move', item: Panel.ItemInfo, parentId: number): void
}>()

// 面板内的文件夹路径（支持多层）
const path = ref<number[]>([])

// 右键菜单
const menuShow = ref(false)
const menuX = ref(0)
const menuY = ref(0)
const menuItem = ref<Panel.ItemInfo | null>(null)

// 拖拽中的图标
const draggingId = ref<number | null>(null)
const dragOverFolderId = ref<number | null>(null)

function findInTree(node: Panel.ItemInfo | null, id: number): Panel.ItemInfo | null {
  if (!node)
    return null
  if (node.id === id)
    return node

  for (const child of node.children || []) {
    const hit = findInTree(child, id)
    if (hit)
      return hit
  }

  return null
}

// 当前所在文件夹
const current = computed<Panel.ItemInfo | null>(() => {
  if (!props.folder)
    return null
  const lastId = path.value[path.value.length - 1]
  if (!lastId)
    return props.folder

  return findInTree(props.folder, lastId) || props.folder
})

const children = computed<Panel.ItemInfo[]>(() => current.value?.children || [])

// 面包屑：根节点为打开的那一级文件夹
const crumbs = computed(() => {
  const list: { id: number | null, label: string }[] = [{ id: null, label: props.folder?.title || '' }]
  for (const id of path.value) {
    const node = findInTree(props.folder, id)
    if (node)
      list.push({ id: Number(node.id), label: node.title })
  }

  return list
})

watch(() => props.visible, (visible) => {
  if (visible)
    path.value = []
  else
    menuShow.value = false
})

watch(() => props.folder?.id, () => {
  path.value = []
  menuShow.value = false
})

function close() {
  emit('update:visible', false)
}

function navInto(id: number) {
  path.value = [...path.value, id]
}

function navTo(index: number) {
  if (index === 0)
    path.value = []
  else
    path.value = path.value.slice(0, index)
}

function back() {
  if (path.value.length)
    path.value = path.value.slice(0, path.value.length - 1)
  else
    close()
}

function handleItemClick(item: Panel.ItemInfo) {
  // 长按弹菜单后的那次 click 不触发打开
  if (lpFired) {
    lpFired = false
    return
  }
  if (item.isFolder === 1) {
    navInto(item.id as number)
    return
  }
  emit('select', item)
}

// 移动端长按弹出菜单（手机没有右键）
let lpTimerId: number | null = null
let lpFired = false

function onTouchStart(e: TouchEvent, item: Panel.ItemInfo) {
  const touch = e.touches[0]
  lpFired = false
  if (lpTimerId !== null)
    clearTimeout(lpTimerId)

  lpTimerId = window.setTimeout(() => {
    lpFired = true
    // 以按住位置为中心弹出菜单
    openMenu({ clientX: touch.clientX, clientY: touch.clientY, preventDefault: () => {} } as MouseEvent, item)
  }, 500)
}

function onTouchEnd(e?: TouchEvent) {
  if (lpTimerId !== null) {
    clearTimeout(lpTimerId)
    lpTimerId = null
  }
  // 长按菜单已弹出时，阻止抬手后的合成 click 触发 clickoutside 把菜单关掉
  if (e && lpFired)
    e.preventDefault()
}

function openMenu(e: MouseEvent, item: Panel.ItemInfo) {
  e.preventDefault()
  menuItem.value = item
  menuShow.value = false
  nextTick(() => {
    menuX.value = e.clientX
    menuY.value = e.clientY
    menuShow.value = true
  })
}

const menuOptions = computed(() => {
  const item = menuItem.value
  if (!item)
    return []

  const options: any[] = [
    { label: t('iconItem.open'), key: 'open' },
  ]

  // 公网/内网地址入口（与首页右键菜单一致）
  if (item.url)
    options.push({ label: t('panelHome.openWanUrl'), key: 'openWanUrl' })
  if (item.lanUrl)
    options.push({ label: t('panelHome.openLanUrl'), key: 'openLanUrl' })

  // 菜单的编辑项与首页右键菜单一致：登录即可，不依赖编辑模式
  if (props.editable || props.menuEditable) {
    options.push({ label: t('common.edit'), key: 'edit' })

    const moveChildren: any[] = [
      { label: t('iconItem.moveToGroupRoot'), key: 'move:0' },
    ]
    if (props.moveTargets.length) {
      moveChildren.push({ type: 'divider', key: 'move-divider' })
      for (const target of props.moveTargets)
        moveChildren.push({ label: target.label, key: `move:${target.value}` })
    }
    options.push({ label: t('iconItem.moveTo'), key: 'moveTo', children: moveChildren })

    options.push({ type: 'divider', key: 'd1' })
    options.push({ label: t('common.delete'), key: 'delete' })
  }

  return options
})

function handleMenuSelect(key: string) {
  menuShow.value = false
  const item = menuItem.value
  if (!item)
    return

  if (key === 'open') {
    handleItemClick(item)
    return
  }
  if (key === 'openWanUrl' || key === 'openLanUrl') {
    emit('open-url', item, key === 'openWanUrl' ? 'wan' : 'lan')
    return
  }
  if (key === 'edit') {
    emit('edit', item)
    return
  }
  if (key === 'delete') {
    emit('delete', item)
    return
  }
  if (key.startsWith('move:')) {
    const target = Number(key.split(':')[1])
    emit('move', item, target)
  }
}

// 面板内拖拽：把图标拖到子文件夹上
function onDragStart(e: DragEvent, item: Panel.ItemInfo) {
  if (!props.editable)
    return
  draggingId.value = item.id as number
  e.dataTransfer?.setData('text/plain', String(item.id))
  if (e.dataTransfer)
    e.dataTransfer.effectAllowed = 'move'
}

function onDragOver(e: DragEvent, item: Panel.ItemInfo) {
  if (!props.editable || item.isFolder !== 1 || draggingId.value === item.id)
    return
  e.preventDefault()
  dragOverFolderId.value = item.id as number
}

function onDragLeave(item: Panel.ItemInfo) {
  if (dragOverFolderId.value === item.id)
    dragOverFolderId.value = null
}

function onDrop(e: DragEvent, item: Panel.ItemInfo) {
  if (!props.editable || item.isFolder !== 1)
    return
  e.preventDefault()
  const id = draggingId.value
  dragOverFolderId.value = null
  draggingId.value = null
  if (!id || id === item.id)
    return

  const source = findInTree(props.folder, id)
  if (source)
    emit('move', source, item.id as number)
}

function onDragEnd() {
  draggingId.value = null
  dragOverFolderId.value = null
}

// Esc：有子层级时先返回上一层，根层则关闭面板
function onKeydown(e: KeyboardEvent) {
  if (e.key !== 'Escape')
    return
  if (menuShow.value) {
    menuShow.value = false
    return
  }
  e.stopPropagation()
  back()
}

watch(() => props.visible, (visible) => {
  if (visible)
    window.addEventListener('keydown', onKeydown, true)
  else
    window.removeEventListener('keydown', onKeydown, true)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown, true)
})

defineExpose({ back })
</script>

<template>
  <Teleport to="body">
    <Transition name="fp-fade">
      <div v-if="visible" class="fp-backdrop" @click="close" />
    </Transition>

    <Transition name="fp-pop">
      <div v-if="visible" class="fp-wrap" :class="{ 'is-mobile': isMobile }" @click.self="close">
        <section class="fp-panel" :class="{ 'is-mobile': isMobile }" role="dialog" aria-modal="true">
          <!-- 头部 -->
          <header class="fp-head">
            <div class="fp-avatar">
              <ItemIcon :item-icon="current?.icon" :size="34" />
            </div>
            <div class="fp-title">
              <b>{{ current?.title }}</b>
              <span>{{ t('iconItem.itemsCount', { count: children.length }) }}</span>
            </div>
            <div class="fp-tools">
              <button
                v-if="editable" class="fp-tool" :title="t('iconItem.add')"
                @click="emit('add-item', current?.id as number)"
              >
                <SvgIconOnline icon="typcn:plus" style="font-size: 15px;" />
              </button>
              <button class="fp-tool" :title="t('common.close')" @click="close">
                <SvgIconOnline icon="mdi:close" style="font-size: 15px;" />
              </button>
            </div>
          </header>

          <!-- 面包屑：仅在进入子文件夹后显示，根目录时与标题重复 -->
          <nav v-if="crumbs.length > 1" class="fp-crumbs">
            <template v-for="(crumb, index) in crumbs" :key="index">
              <span v-if="index > 0" class="fp-sep">›</span>
              <span
                class="fp-crumb" :class="{ 'is-current': index === crumbs.length - 1 }"
                @click="index < crumbs.length - 1 && navTo(index)"
              >
                {{ crumb.label }}
              </span>
            </template>
          </nav>

          <!-- 内容 -->
          <div class="fp-body">
            <div v-if="children.length === 0" class="fp-empty">
              <div class="fp-empty-grid">
                <div v-for="i in 4" :key="i" />
              </div>
              <p>{{ t('iconItem.folderEmpty') }}</p>
              <button v-if="editable" class="fp-empty-btn" @click="emit('add-item', current?.id as number)">
                {{ t('iconItem.add') }}
              </button>
            </div>

            <div v-else class="fp-grid" :class="{ 'is-info': iconStyle === PanelPanelConfigStyleEnum.info }">
              <div
                v-for="item in children" :key="item.id"
                class="fp-cell"
                :class="{
                  'is-folder': item.isFolder === 1,
                  'is-drop-target': dragOverFolderId === item.id,
                  'is-dragging': draggingId === item.id,
                }"
                :draggable="editable"
                @click="handleItemClick(item)"
                @contextmenu="openMenu($event, item)"
                @touchstart="onTouchStart($event, item)"
                @touchmove="onTouchEnd()"
                @touchend="onTouchEnd($event)"
                @dragstart="onDragStart($event, item)"
                @dragover="onDragOver($event, item)"
                @dragleave="onDragLeave(item)"
                @drop="onDrop($event, item)"
                @dragend="onDragEnd"
              >
                <AppIcon
                  class="cursor-pointer"
                  :item-info="item"
                  :icon-text-color="iconTextColor"
                  :icon-text-info-hide-description="iconTextInfoHideDescription || false"
                  :icon-text-icon-hide-title="iconTextIconHideTitle || false"
                  :panel-style="iconStyle"
                />
              </div>
            </div>
          </div>
        </section>
      </div>
    </Transition>

    <NDropdown
      placement="bottom-start"
      trigger="manual"
      :x="menuX"
      :y="menuY"
      :options="menuOptions"
      :show="menuShow"
      :z-index="10120"
      @select="handleMenuSelect"
      @clickoutside="menuShow = false"
    />
  </Teleport>
</template>

<style scoped>
.fp-backdrop {
  position: fixed;
  inset: 0;
  /* 高于首页页面指示器(z-index:10000)，避免移动端底部抽屉被其遮挡 */
  z-index: 10010;
  background: rgba(6, 8, 12, 0.28);
  backdrop-filter: blur(9px) saturate(0.92);
}

.fp-wrap {
  position: fixed;
  inset: 0;
  z-index: 10011;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  pointer-events: none;
}

.fp-panel {
  width: min(880px, 94vw);
  max-height: min(78vh, 760px);
  display: flex;
  flex-direction: column;
  pointer-events: auto;
  border-radius: 26px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(26, 29, 36, 0.42);
  backdrop-filter: blur(26px) saturate(1.2);
  box-shadow: 0 18px 60px rgba(0, 0, 0, 0.55);
  overflow: hidden;
  color: #f4f6fa;
}

.fp-head {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 18px 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  flex: none;
}

.fp-avatar {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: none;
  background: rgba(255, 255, 255, 0.10);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.12);
  overflow: hidden;
}

.fp-title {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
}

.fp-title b {
  font-size: 16px;
  font-weight: 750;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.fp-title span {
  font-size: 11.5px;
  color: rgba(244, 246, 250, 0.62);
}

.fp-tools {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 6px;
}

.fp-tool {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.06);
  color: #fff;
  cursor: pointer;
  transition: background 0.16s;
}

.fp-tool:hover {
  background: rgba(255, 255, 255, 0.18);
}

.fp-crumbs {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
  padding: 10px 18px 0;
  font-size: 12.5px;
  color: rgba(244, 246, 250, 0.62);
  flex: none;
}

.fp-crumb {
  max-width: 200px;
  padding: 3px 8px;
  border-radius: 7px;
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: background 0.14s, color 0.14s;
}

.fp-crumb:hover {
  background: rgba(255, 255, 255, 0.10);
  color: #fff;
}

.fp-crumb.is-current {
  color: #fff;
  font-weight: 600;
  background: rgba(255, 255, 255, 0.08);
  cursor: default;
}

.fp-sep {
  opacity: 0.4;
}

.fp-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  overscroll-behavior: contain;
  padding: 8px 18px 20px;
}

.fp-grid {
  padding-top: 14px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 22px 6px;
}

.fp-grid.is-info {
  grid-template-columns: repeat(auto-fill, minmax(min(200px, 100%), 1fr));
  gap: 16px;
}

.fp-cell {
  min-width: 0;
  transition: transform 0.16s ease, opacity 0.16s ease;
  /* 移动端：长按弹菜单，禁用系统默认长按行为 */
  -webkit-user-select: none;
  user-select: none;
  -webkit-touch-callout: none;
}

.fp-cell.is-dragging {
  opacity: 0.4;
}

.fp-cell.is-drop-target {
  transform: scale(1.08);
}

.fp-cell.is-drop-target :deep(.folder-plate) {
  box-shadow:
    0 0 0 2px #fff,
    0 0 0 8px rgba(255, 255, 255, 0.18),
    0 12px 28px rgba(255, 255, 255, 0.24);
}

.fp-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  padding: 46px 0 40px;
  color: rgba(244, 246, 250, 0.62);
  font-size: 13px;
}

.fp-empty-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
  width: 76px;
  height: 76px;
}

.fp-empty-grid div {
  border-radius: 8px;
  border: 1px dashed rgba(255, 255, 255, 0.24);
  background: rgba(255, 255, 255, 0.03);
}

.fp-empty-btn {
  padding: 7px 14px;
  border-radius: 9px;
  border: 1px solid rgba(255, 255, 255, 0.18);
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
  font-size: 12.5px;
  cursor: pointer;
}

.fp-empty-btn:hover {
  background: rgba(255, 255, 255, 0.18);
}

/* 移动端：底部抽屉 */
.fp-wrap.is-mobile {
  align-items: flex-end;
  padding: 0;
}

.fp-panel.is-mobile {
  width: 100%;
  max-width: none;
  max-height: 82vh;
  border-radius: 24px 24px 0 0;
}

.fp-panel.is-mobile .fp-head {
  padding-top: 18px;
}

.fp-panel.is-mobile .fp-head::before {
  content: '';
  position: absolute;
  top: 8px;
  left: 50%;
  transform: translateX(-50%);
  width: 36px;
  height: 4px;
  border-radius: 2px;
  background: rgba(255, 255, 255, 0.28);
}

.fp-panel.is-mobile .fp-grid {
  grid-template-columns: repeat(auto-fill, minmax(76px, 1fr));
  gap: 18px 4px;
}

/* 动效 */
.fp-fade-enter-active,
.fp-fade-leave-active {
  transition: opacity 0.24s ease;
}

.fp-fade-enter-from,
.fp-fade-leave-to {
  opacity: 0;
}

.fp-pop-enter-active {
  transition: opacity 0.24s cubic-bezier(0.22, 0.9, 0.3, 1), transform 0.24s cubic-bezier(0.22, 0.9, 0.3, 1);
}

.fp-pop-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.fp-pop-enter-from,
.fp-pop-leave-to {
  opacity: 0;
  transform: scale(0.92) translateY(10px);
}

.fp-wrap.is-mobile.fp-pop-enter-from,
.fp-wrap.is-mobile.fp-pop-leave-to {
  transform: translateY(100%);
  opacity: 0.6;
}

@media (prefers-reduced-motion: reduce) {
  .fp-fade-enter-active,
  .fp-fade-leave-active,
  .fp-pop-enter-active,
  .fp-pop-leave-active {
    transition-duration: 0.01ms;
  }
}
</style>
