<script setup lang="ts">
import { computed } from 'vue'
import { ItemIcon } from '@/components/common'
import { PanelPanelConfigStyleEnum } from '@/enums'

interface Prop {
  itemInfo?: Panel.ItemInfo
  size?: number // 默认70
  forceBackground?: string // 强制背景色
  iconTextColor?: string
  iconTextInfoHideDescription: boolean
  iconTextIconHideTitle: boolean
  style: PanelPanelConfigStyleEnum
}

const props = withDefaults(defineProps<Prop>(), {
  size: 70,
})

const defaultBackground = '#2a2a2a6b'

const calculateLuminance = (color: string) => {
  const hex = color.replace(/^#/, '')
  const r = parseInt(hex.substring(0, 2), 16)
  const g = parseInt(hex.substring(2, 4), 16)
  const b = parseInt(hex.substring(4, 6), 16)
  return (0.299 * r + 0.587 * g + 0.114 * b) / 255
}

const textColor = computed(() => {
  const luminance = calculateLuminance(props.itemInfo?.icon?.backgroundColor || defaultBackground)
  return luminance > 0.5 ? 'black' : 'white'
})

// 根据面板样式计算 ItemIcon 的尺寸，使图标内容随容器等比缩放
const itemIconSize = computed(() => {
  // 直接传递 props.size，让图标 100% 填充容器
  return props.size
})
</script>

<template>
  <div class="app-icon w-full">
    <!-- 详情图标 -->
    <div
      v-if="style === PanelPanelConfigStyleEnum.info"
      class="app-icon-info w-full rounded-2xl transition-all duration-200 pc-hover-effect flex"
      :style="{ background: itemInfo?.icon?.backgroundColor || defaultBackground }"
      @contextmenu.prevent
      @touchstart.passive
    >
      <!-- 图标 -->
      <div class="app-icon-info-icon">
        <div class="w-full h-full flex items-center justify-center ">
          <ItemIcon :item-icon="itemInfo?.icon" force-background="transparent" :size="itemIconSize" class="overflow-hidden" style="border-radius: 1rem;" />
        </div>
      </div>

      <!-- 文字 -->
      <!-- 如果为纯白色，将自动根据背景的明暗计算字体的黑白色 -->
      <div class="text-white flex items-center" :style="{ color: (iconTextColor === '#ffffff') ? textColor : iconTextColor, maxWidth: 'calc(100% - 20px)', padding: '0 10px' }">
        <div class="app-icon-info-text-box w-full">
          <div class="app-icon-info-text-box-title font-semibold w-full">
            {{ itemInfo?.title }}
          </div>
          <div v-if="!iconTextInfoHideDescription" class="app-icon-info-text-box-description">
            {{ itemInfo?.description }}
          </div>
        </div>
      </div>
    </div>

    <!-- 极简 (小) 图标（APP） -->
    <div v-if="style === PanelPanelConfigStyleEnum.icon" class="app-icon-small-container"
      @contextmenu.prevent
      @touchstart.passive
    >
      <!-- 图标层 -->
      <div class="app-icon-small-icon-wrapper">
        <div
          class="app-icon-small-icon rounded-2xl sunpanel mx-auto rounded-2xl transition-all duration-200 pc-hover-effect"
          :title="itemInfo?.description"
        >
          <ItemIcon :item-icon="itemInfo?.icon" force-background="transparent" :size="itemIconSize" style="border-radius: 1rem;" />
        </div>
      </div>
      <!-- 文字层（绝对定位，不影响图标） -->
      <div class="app-icon-small-text-layer">
        <div
          v-if="!iconTextIconHideTitle"
          class="app-icon-small-title text-center app-icon-text-shadow cursor-pointer mt-[2px]"
          :style="{ color: iconTextColor }"
        >
          <span>{{ itemInfo?.title }}</span>
        </div>
        <div
          v-if="!iconTextInfoHideDescription && itemInfo?.description"
          class="app-icon-small-description text-center text-xs opacity-80 cursor-pointer mt-[1px]"
          :style="{ color: iconTextColor }"
        >
          {{ itemInfo?.description }}
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* PC 端 hover 效果 - 仅在支持 hover 的设备上生效 */
@media (hover: hover) and (pointer: fine) {
  .pc-hover-effect:hover {
    box-shadow: 0 8px 30px rgba(255, 255, 255, 0.3);
    transform: scale(1.1) translateY(-4px);
  }
}

/* 响应式图标块设计 */
.app-icon-info-icon {
  width: min(70px, 100%);
  height: min(70px, 100%);
  min-width: min(70px, 100%);
  aspect-ratio: 1 / 1;
  border-radius: 1rem; /* rounded-2xl = 1rem */
}

.app-icon-small-icon {
  width: min(70px, 100%);
  height: min(70px, 100%);
  aspect-ratio: 1 / 1;
  border-radius: 1rem; /* rounded-2xl = 1rem */
}

/* 小图标容器 */
.app-icon-small-container {
  display: inline-block;
  position: relative;
  margin: 10px 5px; /* 移动端：上下 10px，左右 5px */
}

/* 图标包装层 - 固定宽度，不受文字影响 */
.app-icon-small-icon-wrapper {
  width: min(70px, 100%);
  height: min(70px, 100%);
  margin: 0 auto;
}

/* 图标本身 */
.app-icon-small-icon {
  width: 100%;
  height: 100%;
  aspect-ratio: 1 / 1;
  border-radius: 1rem;
}

/* 文字层 - 绝对定位在图标下方，不影响图标位置 */
.app-icon-small-text-layer {
  position: absolute;
  top: calc(100% + 4px); /* 距离图标底部 4px */
  left: 50%;
  transform: translateX(-50%); /* 居中对齐 */
  width: max-content;
  min-width: min(70px, 100%);
  max-width: 280px; /* 最大宽度限制 */
  text-align: center;
  pointer-events: none; /* 让点击事件穿透 */
}

/* 标题和描述可以超出范围 */
.app-icon-small-title,
.app-icon-small-description {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
  width: 100%;
  margin: 0;
  text-align: center;
}

/* PC 端增加上下间距 */
@media (min-width: 1025px) {
  .app-icon-small-container {
    margin: 14px 5px; /* PC 端：上下 14px，左右 5px */
  }
}

/* 确保item-icon完全填充容器 */
:deep(.item-icon) {
  width: 100% !important;
  height: 100% !important;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 针对不同屏幕尺寸的等比缩小 */
@media (max-width: 1024px) {
  .app-icon-info-icon {
    width: min(60px, 100%);
    height: min(60px, 100%);
    min-width: min(60px, 100%);
    border-radius: 1rem; /* 保持与 PC 端一致的圆角 */
  }
  
  .app-icon-small-icon {
    width: min(60px, 100%);
    height: min(60px, 100%);
    border-radius: 1rem; /* 保持与 PC 端一致的圆角 */
  }
}

/* PC 端标题样式 - 详情模式 */
.app-icon-info-text-box-title {
  font-size: 1rem;
  word-wrap: break-word;
  overflow-wrap: break-word;
  white-space: normal;
  text-align: center;
  line-height: 1.3;
}

/* PC 端小图标标题样式 - 支持超长文本居中对齐 */
.app-icon-small-title {
  font-size: 0.85rem;
  word-wrap: break-word;
  overflow-wrap: break-word;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 150%; /* 允许超出图标宽度，实现以图标中心点对齐 */
  display: block;
  text-align: center;
  position: relative;
  left: 50%;
  transform: translateX(-50%);
}

@media (max-width: 768px) {
  .app-icon-info-icon {
    width: min(60px, 100%);
    height: min(60px, 100%);
    min-width: min(60px, 100%);
    border-radius: 1rem; /* 保持与 PC 端一致的圆角 */
  }
  
  .app-icon-small-icon {
    width: min(60px, 100%);
    height: min(60px, 100%);
    border-radius: 1rem; /* 保持与 PC 端一致的圆角 */
  }
  
  .app-icon-info-text-box-title {
    font-size: 0.85rem !important;
    word-wrap: break-word;
    overflow-wrap: break-word;
    white-space: normal;
    text-align: center;
    line-height: 1.3;
  }
  
  .app-icon-info-text-box-description {
    font-size: 0.7rem !important;
    word-wrap: break-word;
    overflow-wrap: break-word;
    white-space: normal;
  }
  
  .app-icon-small-title {
    font-size: 0.75rem !important;
    word-wrap: break-word;
    overflow-wrap: break-word;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 100%;
    display: block;
    text-align: center;
  }
}

@media (max-width: 480px) {
  .app-icon-info-icon {
    width: min(60px, 100%);
    height: min(60px, 100%);
    min-width: min(60px, 100%);
    border-radius: 1rem; /* 保持与 PC 端一致的圆角 */
  }
  
  .app-icon-small-icon {
    width: min(60px, 100%);
    height: min(60px, 100%);
    border-radius: 1rem; /* 保持与 PC 端一致的圆角 */
  }
  
  .app-icon-info-text-box-title {
    font-size: 0.8rem !important;
    word-wrap: break-word;
    overflow-wrap: break-word;
    white-space: normal;
    text-align: center;
    line-height: 1.3;
  }
  
  .app-icon-info-text-box-description {
    font-size: 0.65rem !important;
  }
  
  .app-icon-small-title {
    font-size: 0.7rem !important;
    word-wrap: break-word;
    overflow-wrap: break-word;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 100%;
    display: block;
    text-align: center;
  }
}

@media (max-width: 360px) {
  .app-icon-info-icon {
    width: min(60px, 100%);
    height: min(60px, 100%);
    min-width: min(60px, 100%);
    border-radius: 1rem; /* 保持与 PC 端一致的圆角 */
  }
  
  .app-icon-small-icon {
    width: min(60px, 100%);
    height: min(60px, 100%);
    border-radius: 1rem; /* 保持与 PC 端一致的圆角 */
  }
  
  .app-icon-info-text-box-title {
    font-size: 0.75rem !important;
    word-wrap: break-word;
    overflow-wrap: break-word;
    white-space: normal;
    text-align: center;
    line-height: 1.3;
  }
  
  .app-icon-info-text-box-description {
    font-size: 0.6rem !important;
  }
  
  .app-icon-small-title {
    font-size: 0.65rem !important;
    word-wrap: break-word;
    overflow-wrap: break-word;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 100%;
    display: block;
    text-align: center;
  }
}

/* 防止 iOS 长按图片显示系统菜单 */
:deep(img) {
  -webkit-touch-callout: none !important;
  -webkit-user-select: none !important;
  user-select: none !important;
  pointer-events: none;
  -webkit-pointer-events: none;
  touch-action: none;
  -webkit-touch-action: none;
  /* iOS Safari 专用：禁止长按菜单 */
  -webkit-hyphens: none;
  hyphens: none;
}

/* 移动端优化 */
@media (max-width: 768px) {
  .app-icon-info,
  .app-icon-small {
    -webkit-touch-callout: none !important;
    -webkit-user-select: none !important;
    user-select: none !important;
    touch-action: manipulation;
    -webkit-touch-action: manipulation;
  }
  
  /* 针对 iOS Safari 的额外防护 */
  .app-icon-info-icon,
  .app-icon-small-icon {
    -webkit-tap-highlight-color: transparent !important;
    -webkit-highlight: none !important;
    /* 关键：阻止 iOS 的图片长按菜单 */
    -webkit-touch-callout: none !important;
    touch-callout: none !important;
  }
  
  /* 图标容器也应用防护 */
  .app-icon-info > div,
  .app-icon-small > div {
    -webkit-touch-callout: none !important;
    -webkit-user-select: none !important;
    user-select: none !important;
  }
}
</style>
