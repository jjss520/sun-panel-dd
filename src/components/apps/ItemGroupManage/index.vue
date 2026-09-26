<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'
import { NButton, NCard, NForm, NFormItem, NInput, NSelect, useDialog, useMessage } from 'naive-ui'
import { VueDraggable } from 'vue-draggable-plus'
import { deletes, edit, getList, saveSort } from '@/api/panel/itemIconGroup'
import { getList as getPageList } from '@/api/panel/itemPage'
import { RoundCardModal, SvgIcon } from '@/components/common'
import { t } from '@/locales'
import { ss } from '@/utils/storage/local'

interface EditModalArg {
  show: boolean
  editStatus: number // 1.添加 2.编辑
  model: Panel.ItemIconGroup
  rules: FormRules
}

const formRef = ref<FormInst | null>(null)
const ms = useMessage()
const dialog = useDialog()
const sortStatus = ref(false)

const defaultMNodal = {
  title: '',
  icon: 'carbon--group-presentation',
  sort: 9999,
  pageId: undefined,
}

const editModalArg = ref<EditModalArg>({
  show: false,
  editStatus: 1,
  model: defaultMNodal,
  rules: {
    title: [
      {
        required: true,
        trigger: 'blur',
        message: t('form.required'),
      },
    ],
  },
})

const groups = ref<Panel.ItemIconGroup[]>([])
const pages = ref<Panel.ItemPage[]>([])
const pageOptions = ref<{ label: string; value: number }[]>([])

function handleAddGroup() {
  editModalArg.value.show = !editModalArg.value.show
  editModalArg.value.editStatus = 1  // 设置为添加模式
  editModalArg.value.model = { ...defaultMNodal }  // 重置表单
  // Clear group list cache
  ss.remove('groupListCache')
}

function handleEditGroup(groupInfo: Panel.ItemIconGroup) {
  editModalArg.value.show = true
  // 深拷贝,确保pageId正确传递
  editModalArg.value.model = {
    id: groupInfo.id,
    title: groupInfo.title || '',
    icon: groupInfo.icon || 'carbon--group-presentation',
    sort: groupInfo.sort || 9999,
    pageId: groupInfo.pageId || undefined,
  }
  editModalArg.value.editStatus = 2
}

// 加载页面列表
async function loadPages() {
  try {
    const res = await getPageList<Common.ListResponse<Panel.ItemPage[]>>()
    if (res.code === 0) {
      pages.value = res.data.list
      pageOptions.value = pages.value.map(page => ({
        label: page.title || '未命名',
        value: page.id as number,
      }))
    }
  } catch (error) {
    console.error('Failed to load pages:', error)
  }
}

// 根据pageId获取页面标题
function getPageTitle(pageId: number): string {
  const page = pages.value.find(p => p.id === pageId)
  return page ? (page.title || '未命名') : ''
}

function handleDragSort() {
  sortStatus.value = true
}

function handleSaveSort() {
  const saveItems: Common.SortItemRequest[] = []
  for (let i = 0; i < groups.value.length; i++) {
    const element = groups.value[i]
    saveItems.push({
      id: element.id as number,
      sort: i + 1,
    })
  }
  saveSort(saveItems).then(({ code, msg }) => {
    if (code === 0) {
      ms.success(t('common.saveSuccess'))
      // 清除分组列表缓存
      ss.remove('groupListCache')
      sortStatus.value = false
      
      // 触发页面更新事件，通知主页和其他组件刷新
      window.dispatchEvent(new CustomEvent('pagesUpdated'))
      
      // 重新加载页面下拉框（确保顺序同步）
      loadPages()
    }
    else {
      ms.error(`${t('common.saveFail')}:${msg}`)
    }
  })
}

function handleDelete(groupInfo: Panel.ItemIconGroup) {
  dialog.warning({
    title: t('common.warning'),
    content: t('apps.itemGroupManage.deleteWarnText', { name: groupInfo.title }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => {
      if (groupInfo.id) {
        deletes([groupInfo.id]).then(({ code, msg }) => {
          if (code !== 0)
            ms.error(t('common.deleteFail'))
          else {
            // 清除分组列表缓存
            ss.remove('groupListCache')
            refreshList()
            
            // 触发页面更新事件，通知主页和其他组件刷新
            window.dispatchEvent(new CustomEvent('pagesUpdated'))
          }
        })
      }
    },

  })
}

function handleSaveGroup() {
  formRef.value?.validate((errors) => {
    if (!errors) {
      console.log('[分组管理] 保存分组', editModalArg.value.model)
      edit(editModalArg.value.model).then(({ code, msg }) => {
        if (code !== 0) {
          ms.error(msg)
        } else {
          ms.success(t('common.saveSuccess'))
          // 清除分组列表缓存
          ss.remove('groupListCache')
          refreshList()
          
          // 触发页面更新事件，通知主页刷新
          window.dispatchEvent(new CustomEvent('pagesUpdated'))
          
          editModalArg.value.show = false
          editModalArg.value.model = { ...defaultMNodal }
        }
      })
    }
    else { console.log(errors) }
  })
}

function refreshList() {
  getList<Common.ListResponse<Panel.ItemIconGroup[]>>().then(({ code, data }) => {
    groups.value = data.list
  })
}

const handlePagesUpdated = () => {
  console.log('[分组管理] 收到页面更新事件，重新加载页面列表')
  loadPages()
}

onMounted(() => {
  refreshList()
  loadPages()
  
  // 监听页面更新事件,重新加载页面列表
  window.addEventListener('pagesUpdated', handlePagesUpdated)
})

onUnmounted(() => {
  window.removeEventListener('pagesUpdated', handlePagesUpdated)
})
</script>

<template>
  <div class="h-full">
    <div class="p-2">
      <NButton type="success" size="small" style="margin-right: 10px;" @click="handleAddGroup">
        {{ $t('common.add') }}
      </NButton>

      <NButton v-if="!sortStatus" size="small" @click="handleDragSort">
        {{ $t('common.sort') }}
      </NButton>

      <NButton v-else type="warning" size="small" @click="handleSaveSort">
        {{ $t('common.saveSort') }}
      </NButton>
    </div>

    <div class=" overflow-auto w-full mt-[20px]  bg-slate-200 dark:bg-zinc-900 rounded-xl" style="height:calc(100% - 65px)">
      <VueDraggable
        v-model="groups"
        item-key="sort" :animation="300"
        :style="{ padding: sortStatus ? '20px' : '10px' }"
        :disabled="!sortStatus"
      >
        <div v-for="(item, index) in groups" :key="index" class="w-full">
          <NCard size="small" style="border-radius:10px;margin-bottom: 10px;">
            <div class="flex" :class="sortStatus ? 'cursor-move' : ''">
              <div class="flex items-center">
                <span class="mr-[10px]">
                  <SvgIcon class="text-[20px]" :icon="item.icon || 'carbon--group-presentation'" />
                </span>
                <div>
                  <div>{{ item.title }}</div>
                  <div v-if="item.pageId" class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    {{ getPageTitle(item.pageId) }}
                  </div>
                </div>
              </div>
              <div class="ml-auto">
                <span>
                  <NButton strong secondary type="success" size="small" @click="handleEditGroup(item)">
                    <template #icon>
                      <SvgIcon icon="basil:edit-solid" />
                    </template>
                  </NButton>
                </span>
                <span class="ml-[10px]">
                  <NButton strong secondary type="error" size="small" class="ml-[10px]" @click="handleDelete(item)">
                    <template #icon>
                      <SvgIcon icon="material-symbols:delete" />
                    </template>
                  </NButton>
                </span>
              </div>
            </div>
          </NCard>
        </div>
      </VueDraggable>
    </div>

    <RoundCardModal v-model:show="editModalArg.show" size="small" type="small" :title="editModalArg.editStatus === 1 ? '添加' : '编辑'" style="width: 400px;">
      <NForm ref="formRef" :model="editModalArg.model" :rules="editModalArg.rules">
        <NFormItem path="title" :label="$t('apps.itemGroupManage.groupName')">
          <NInput v-model:value="editModalArg.model.title" type="text" :maxlength="20" show-count />
        </NFormItem>
        
        <NFormItem path="pageId" label="所属页面">
          <NSelect 
            v-model:value="editModalArg.model.pageId" 
            :options="pageOptions" 
            placeholder="请选择页面"
            clearable
          />
        </NFormItem>
      </NForm>
      <template #footer>
        <NButton type="success" size="small" class="float-right" @click="handleSaveGroup">
          {{ $t('common.confirm') }}
        </NButton>
      </template>
    </RoundCardModal>
  </div>
</template>
