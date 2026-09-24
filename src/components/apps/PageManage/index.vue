<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'
import { NButton, NCard, NForm, NFormItem, NInput, useDialog, useMessage } from 'naive-ui'
import { VueDraggable } from 'vue-draggable-plus'
import { deletes, edit, getList, saveSort } from '@/api/panel/itemPage'
import { RoundCardModal, SvgIcon } from '@/components/common'
import { t } from '@/locales'
import { ss } from '@/utils/storage/local'

interface EditModalArg {
  show: boolean
  editStatus: number // 1.添加 2.编辑
  model: Panel.ItemPage
  rules: FormRules
}

const formRef = ref<FormInst | null>(null)
const ms = useMessage()
const dialog = useDialog()
const sortStatus = ref(false)

const defaultModel = {
  title: '',
  icon: 'material-symbols-ad-group-outline-rounded',
  sort: 9999,
}

const editModalArg = ref<EditModalArg>({
  show: false,
  editStatus: 1,
  model: defaultModel,
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

const pages = ref<Panel.ItemPage[]>([])

function handleAddPage() {
  editModalArg.value.show = true
  editModalArg.value.editStatus = 1
  editModalArg.value.model = { ...defaultModel }
  // Clear page list cache
  ss.remove('pageListCache')
}

function handleEditPage(pageInfo: Panel.ItemPage) {
  editModalArg.value.show = true
  editModalArg.value.model = { ...pageInfo }
  editModalArg.value.editStatus = 2
}

async function handleSaveSort() {
  try {
    const sortItems = pages.value.map((item, index) => ({
      id: item.id as number,
      sort: index + 1,
    }))
    
    const { code, msg } = await saveSort(sortItems)
    if (code === 0) {
      ms.success(t('common.saveSuccess'))
      sortStatus.value = false
      await loadPages()
      // Clear page list cache
      ss.remove('pageListCache')
      // 触发自定义事件通知主页刷新
      window.dispatchEvent(new CustomEvent('pagesUpdated'))
    } else {
      ms.error(`${t('common.saveFail')}:${msg}`)
    }
  } catch (error) {
    ms.error(t('common.saveFail'))
  }
}

async function handleDelete(page: Panel.ItemPage) {
  dialog.warning({
    title: t('common.delete'),
    content: t('common.deleteConfirm'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      try {
        const { code, msg } = await deletes([page.id as number])
        if (code === 0) {
          ms.success(t('common.deleteSuccess'))
          await loadPages()
          // Clear page list cache
          ss.remove('pageListCache')
          // 触发自定义事件通知主页刷新
          window.dispatchEvent(new CustomEvent('pagesUpdated'))
        } else {
          ms.error(msg)
        }
      } catch (error) {
        ms.error(t('common.deleteFail'))
      }
    },
  })
}

async function handleSubmit() {
  formRef.value?.validate(async (errors) => {
    if (!errors) {
      try {
        const { code, msg } = await edit(editModalArg.value.model)
        if (code === 0) {
          ms.success(t('common.saveSuccess'))
          editModalArg.value.show = false
          editModalArg.value.model = { ...defaultModel }
          await loadPages()
          // Clear page list cache
          ss.remove('pageListCache')
          // 触发自定义事件通知主页刷新
          window.dispatchEvent(new CustomEvent('pagesUpdated'))
        } else {
          ms.error(`${t('common.saveFail')}:${msg}`)
        }
      } catch (error) {
        ms.error(t('common.saveFail'))
      }
    }
  })
}

async function loadPages() {
  try {
    const { code, data } = await getList<Common.ListResponse<Panel.ItemPage[]>>()
    if (code === 0) {
      pages.value = data.list
    }
  } catch (error) {
    console.error('Failed to load pages:', error)
  }
}

onMounted(() => {
  loadPages()
})
</script>

<template>
  <div class="h-full">
    <div class="p-2">
      <NButton type="success" size="small" style="margin-right: 10px;" @click="handleAddPage">
        {{ $t('common.add') }}
      </NButton>

      <NButton v-if="!sortStatus" size="small" @click="sortStatus = true">
        {{ $t('common.sort') }}
      </NButton>

      <NButton v-else type="warning" size="small" @click="handleSaveSort">
        {{ $t('common.saveSort') }}
      </NButton>
    </div>

    <div class="overflow-auto w-full mt-[20px] bg-slate-200 dark:bg-zinc-900 rounded-xl" style="height:calc(100% - 65px)">
      <VueDraggable
        v-model="pages"
        item-key="sort"
        :animation="300"
        :style="{ padding: sortStatus ? '20px' : '10px' }"
        :disabled="!sortStatus"
      >
        <div v-for="(item, index) in pages" :key="index" class="w-full">
          <NCard size="small" style="border-radius:10px;margin-bottom: 10px;">
            <div class="flex" :class="sortStatus ? 'cursor-move' : ''">
              <div class="flex items-center">
                <span class="mr-[10px]">
                  <SvgIcon class="text-[20px]" :icon="item.icon || 'material-symbols-ad-group-outline-rounded'" />
                </span>
                <span>{{ item.title }}</span>
              </div>
              <div class="ml-auto">
                <span>
                  <NButton strong secondary type="success" size="small" @click="handleEditPage(item)">
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

    <!-- 编辑弹窗 -->
    <RoundCardModal v-model:show="editModalArg.show" size="small" type="small" :title="editModalArg.editStatus === 1 ? t('common.add') : t('common.edit')" style="width: 400px;">
      <NForm ref="formRef" :model="editModalArg.model" :rules="editModalArg.rules">
        <NFormItem path="title" :label="t('common.title')">
          <NInput v-model:value="editModalArg.model.title" type="text" :maxlength="20" show-count />
        </NFormItem>
      </NForm>
      <template #footer>
        <NButton type="success" size="small" class="float-right" @click="handleSubmit">
          {{ $t('common.confirm') }}
        </NButton>
      </template>
    </RoundCardModal>
  </div>
</template>
