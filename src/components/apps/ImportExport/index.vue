<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { UploadFileInfo } from 'naive-ui'
import { NAlert, NButton, NCheckbox, NCheckboxGroup, NDivider, NInput, NSpace, NUpload, useMessage } from 'naive-ui'
import { RoundCardModal, SvgIcon } from '@/components/common'
import type { IconGroup, ImportJsonResult } from '@/utils/jsonImportExport'
import { ConfigVersionLowError, FormatError, exportJson, importJsonString } from '@/utils/jsonImportExport'
import { get as getAbout } from '@/api/system/about'
import { edit as addGroup, getList as getGroupList } from '@/api/panel/itemIconGroup'
import { ss } from '@/utils/storage/local'
import { addMultiple as addMultipleIcons, getListByGroupId } from '@/api/panel/itemIcon'
import { getNotepadList, saveNotepadContent, deleteNotepad, type NotepadInfo } from '@/api/panel/notepad'
import { getList as getPageList, edit as addPage } from '@/api/panel/itemPage'

import { t } from '@/locales'

interface ItemGroup extends Panel.ItemIconGroup {
  items?: Panel.ItemInfo[]
}

const ms = useMessage()

const jsonData = ref<string | null>(null)
const importWarning = ref<string[]>([])
const importRoundModalShow = ref(false)
const exportRoundModalShow = ref(false)
const loading = ref(false)
const uploadLoading = ref(false)
const version = ref('') // 当前软件版本
const debug = ref(false)

const importObj = ref<ImportJsonResult | null> (null)

const importItems = ref<string[]>(['icons', 'notepads', 'pages']) // 当前软件版本支持导入导出的项目
const checkedItems = ref<string[]>(['icons', 'notepads', 'pages']) // 当前准备导入的项目

// 导入图标
async function importIcons(): Promise<string | null> {
  const groups = importObj.value?.geticons()
  const batchSize = 50

  if (!groups)
    return null

  try {
    // 获取当前用户的所有页面，建立pageId映射
    const currentPageResponse = await getPageList<Common.ListResponse<Panel.ItemPage[]>>()
    const pageIdMap = new Map<number, number>() // 旧pageId -> 新pageId
    
    if (currentPageResponse.code === 0 && currentPageResponse.data?.list) {
      // 按sort排序，确保顺序一致
      const sortedPages = currentPageResponse.data.list.sort((a, b) => (a.sort || 9999) - (b.sort || 9999))
      
      // 收集导出文件中的所有pageId
      const exportedPageIds = new Set<number>()
      groups.forEach(group => {
        if (group.pageId) {
          exportedPageIds.add(group.pageId)
        }
      })
      
      // 为每个导出的pageId找到对应的页面
      // 策略：如果pageId相同且title相同，则映射；否则按顺序映射
      const exportedPageIdsArray = Array.from(exportedPageIds).sort((a, b) => a - b)
      
      exportedPageIdsArray.forEach((oldPageId, index) => {
        if (index < sortedPages.length && sortedPages[index].id !== undefined) {
          // 按顺序映射到现有页面
          pageIdMap.set(oldPageId, sortedPages[index].id)
        }
        // 如果现有页面不够，后面会创建新页面
      })
    }

    for (let i = 0; i < groups.length; i++) {
      const element = groups[i]

      // 处理pageId映射
      let targetPageId: number | undefined = undefined
      if (element.pageId && pageIdMap.has(element.pageId)) {
        const mappedId = pageIdMap.get(element.pageId)
        if (mappedId !== undefined) {
          targetPageId = mappedId
        }
      }

      // 创建组得到组id，包含pageId字段
      const createGroupResponse = await addGroup<Panel.ItemIconGroup>({
        title: element.title,
        sort: element.sort,
        pageId: targetPageId,  // 使用映射后的pageId
      })

      if (createGroupResponse.code === 0) {
        const groupId = createGroupResponse.data?.id
        // 清除分组列表缓存 (key必须与Home组件一致)
        ss.remove('groupListCache')

        if (groupId) {
          let addIcons: Panel.ItemInfo[] = []


          // 批量添加子项
          for (let iconI = 0; iconI < element.children.length; iconI++) {
            const iconElement = element.children[iconI]

            addIcons.push({
              title: iconElement.title,
              sort: iconElement.sort,
              icon: iconElement.icon,
              url: iconElement.url,
              lanUrl: iconElement.lanUrl,
              description: iconElement.description,
              openMethod: iconElement.openMethod,
              itemIconGroupId: groupId,
            })

            // 每 batchSize 个添加一次
            if (addIcons.length === batchSize || iconI === element.children.length - 1) {
              const response = await addMultipleIcons(addIcons)

              if (response.code !== 0)
                return response.msg

              addIcons = []
            }
          }
        }
      }
      else {
        return createGroupResponse.msg
      }
    }

    // 导入完成后，清除页面缓存并触发刷新事件
    console.log('[导入导出] 图标导入完成，触发页面刷新')
    ss.remove('pageListCache')
    // 清除所有图标列表缓存
    Object.keys(localStorage).forEach(key => {
      if (key.startsWith('itemIconList_')) {
        ss.remove(key)
      }
    })
    // 触发页面更新事件，通知主页刷新
    window.dispatchEvent(new CustomEvent('pagesUpdated'))

    return null
  }
  catch (error) {
    if (error instanceof Error)
      return `${t('common.failed')}: ${error.message}`
    else
      return t('common.unknownError')
  }
}

// 导出图标
async function exportIcons(): Promise<IconGroup[]> {
  const iconGroups: IconGroup[] = []

  // 获取组数据
  const { code, data } = await getGroupList<Common.ListResponse<ItemGroup[]>>()

  if (code === 0) {
    // 使用 Promise.all 等待所有异步操作完成
    await Promise.all(data.list.map(async (element) => {
      const group: IconGroup = {
        title: element.title as string,
        sort: element.sort as 0,
        pageId: element.pageId || undefined,  // 添加pageId字段
        children: [],
      }

      const res = await getListByGroupId<Common.ListResponse<Panel.ItemInfo[]>>(element.id)

      if (res.code === 0) {
        for (const iconElement of res.data.list) {
          group.children.push({
            icon: iconElement.icon,
            sort: iconElement.sort || 99999,
            title: iconElement.title,
            url: iconElement.url,
            lanUrl: iconElement.lanUrl || '',
            description: iconElement.description || '',
            openMethod: iconElement.openMethod || 1,
            lanOnly: iconElement.lanOnly || 0,
          })
        }
      }

      iconGroups.push(group)
    }))

    return iconGroups
  }
  else {
    return []
  }
}

// 导入记事本（包含提醒）- 覆盖模式
async function importNotepads(): Promise<string | null> {
  const notepads = importObj.value?.getNotepads()

  if (!notepads || notepads.length === 0)
    return null

  try {
    // 第一步：获取当前所有记事本并删除（覆盖模式）
    const currentResponse = await getNotepadList()
    if (currentResponse.code === 0 && currentResponse.data && currentResponse.data.length > 0) {
      console.log(`发现 ${currentResponse.data.length} 个现有记事本，开始删除...`)
      for (const existingNotepad of currentResponse.data) {
        const deleteResponse = await deleteNotepad({ id: existingNotepad.id })
        if (deleteResponse.code !== 0) {
          console.warn(`删除记事本 ${existingNotepad.id} 失败:`, deleteResponse.msg)
        }
      }
      console.log('现有记事本已清除')
    }

    // 第二步：导入新的记事本
    console.log(`开始导入 ${notepads.length} 个记事本...`)
    for (const notepad of notepads) {
      // 保存记事本（不包含 id，让后端自动生成）
      const response = await saveNotepadContent({
        id: 0, // 0 表示新建
        title: notepad.title,
        content: notepad.content,
        remindTime: notepad.remindTime || null,
        remindStatus: notepad.remindStatus || 0,
        remindRepeat: notepad.remindRepeat || 'none',
        remindForce: notepad.remindForce || 0,
        remindAdvanceDays: notepad.remindAdvanceDays || 0,
      })

      if (response.code !== 0)
        return `${t('common.failed')}: ${response.msg}`
    }

    console.log('记事本导入成功')
    return null
  }
  catch (error) {
    if (error instanceof Error)
      return `${t('common.failed')}: ${error.message}`
    else
      return t('common.unknownError')
  }
}

// 导出记事本（包含提醒）
async function exportNotepads(): Promise<import('@/utils/jsonImportExport').NotepadItem[]> {
  try {
    const response = await getNotepadList()
    
    if (response.code === 0 && response.data) {
      // 转换数据格式，移除 id 和 userId 等不需要的字段
      return response.data.map((item: NotepadInfo) => ({
        title: item.title,
        content: item.content,
        remindBaseTime: item.remindBaseTime || null,  // 基准时间（用户原始选择的时间）
        remindTime: item.remindTime || null,
        remindStatus: item.remindStatus || 0,
        remindRepeat: item.remindRepeat || 'none',
        remindAdvanceDays: item.remindAdvanceDays || 0,
      }))
    }
    
    return []
  }
  catch (error) {
    console.error('导出记事本失败:', error)
    return []
  }
}

// 导入页面
async function importPages(): Promise<string | null> {
  const pages = importObj.value?.getPages()
  
  if (!pages || pages.length === 0)
    return null
  
  try {
    console.log(`[导入导出] 开始导入 ${pages.length} 个页面...`)
    for (const page of pages) {
      // 创建页面
      const response = await addPage({
        title: page.title,
        icon: page.icon || '',
        sort: page.sort || 9999,
      })
      
      if (response.code !== 0) {
        console.warn(`[导入导出] 创建页面 "${page.title}" 失败:`, response.msg)
      }
    }
    
    console.log('[导入导出] 页面导入完成')
    return null
  }
  catch (error) {
    if (error instanceof Error)
      return `${t('common.failed')}: ${error.message}`
    else
      return t('common.unknownError')
  }
}

// 导出页面
async function exportPages(): Promise<import('@/utils/jsonImportExport').Page[]> {
  try {
    const response = await getPageList<Common.ListResponse<Panel.ItemPage[]>>()
    
    if (response.code === 0 && response.data && response.data.list) {
      return response.data.list.map((page: Panel.ItemPage) => ({
        title: page.title || '',
        icon: page.icon || '',
        sort: page.sort || 9999,
      }))
    }
    
    return []
  }
  catch (error) {
    console.error('导出页面失败:', error)
    return []
  }
}

onMounted(() => {
  interface Version {
    versionName: string
    versionCode: number
  }

  getAbout<Version>().then((res) => {
    if (res.code === 0)
      version.value = res.data.versionName
  })
})

function handleFileChange(options: { file: UploadFileInfo; fileList: Array<UploadFileInfo> }) {
  uploadLoading.value = true
  if (options.file.file) {
    const reader = new FileReader()
    reader.onload = () => {
      if (reader.result) {
        jsonData.value = reader.result as string
        importCheck()
      }
      else {
        ms.error(`${t('common.failed')}: ${t('common.repeatLater')}`)
      }
      uploadLoading.value = false
    }
    reader.readAsText(options.file.file)
  }
}

// 验证导入文件
function importCheck() {
  importWarning.value = []
  if (jsonData.value) {
    try {
      importObj.value = importJsonString(jsonData.value)
      if (importObj.value) {
        if (!importObj.value.isPassCheckMd5())
          importWarning.value.push(t('apps.exportImport.fileModified'))

        if (!importObj.value.isPassCheckConfigVersionOld())
          importWarning.value.push(t('apps.exportImport.warnConfigFileLow'))

        if (!importObj.value.isPassCheckConfigVersionNew())
          importWarning.value.push(t('apps.exportImport.softwareVersionLow'))

        // （暂时不做）此处可以判断，当前的配置文件是否存在的导入项目（不存在隐藏importItems里面的值）操作变量：importItems

        // 通过了验证,打开弹窗
        importRoundModalShow.value = !importRoundModalShow.value

      }
    }
    catch (error) {
      if (error instanceof ConfigVersionLowError) {
        ms.error(t('apps.exportImport.errorConfigFileLow'))
        console.error('The configuration file version is too low to be compatible')
      }
      else if (error instanceof FormatError) {
        ms.error(t('apps.exportImport.errorConfigFileFormat'))
        console.error('The format is incorrect and cannot be imported')
      }
    }
  }
  else {
    ms.error(t('apps.exportImport.errorConfigFileFormat'))
  }
}

// 开始导出
async function handleStartExport() {
  loading.value = true
  // 获取软件版本号
  const exportResult = exportJson(version.value)
  if (checkedItems.value.includes('icons')) {
    const iconGroups = await exportIcons()
    exportResult.addIconsData(iconGroups)
  }

  if (checkedItems.value.includes('notepads')) {
    const notepads = await exportNotepads()
    exportResult.addNotepadsData(notepads)
  }

  if (checkedItems.value.includes('pages')) {
    const pages = await exportPages()
    exportResult.addPagesData(pages)
  }

  jsonData.value = exportResult.string()
  exportResult.exportFile()
  loading.value = false
  exportRoundModalShow.value = false
  // ms.success(t('common.success'))
}

// 开始导入
async function handleStartImport() {
  loading.value = true
  
  // 先导入页面（因为分组需要pageId）
  if (checkedItems.value.includes('pages')) {
    const errMsg = await importPages()
    if (errMsg !== null) {
      ms.error(`${t('common.failed')}:${errMsg}`)
      loading.value = false
      importRoundModalShow.value = false
      return
    }
  }
  
  if (checkedItems.value.includes('icons')) {
    const errMsg = await importIcons()
    if (errMsg !== null) {
      ms.error(`${t('common.failed')}:${errMsg}`) // Use error for failure
      loading.value = false
      importRoundModalShow.value = false
      return
    }
  }

  if (checkedItems.value.includes('notepads')) {
    const errMsg = await importNotepads()
    if (errMsg !== null) {
      ms.error(`${t('common.failed')}:${errMsg}`)
      loading.value = false
      importRoundModalShow.value = false
      return
    }
  }

  // 全部成功
  ms.success(`${t('common.success')}`)
  // 延迟刷新
  setTimeout(() => {
    window.location.reload()
  }, 1000)

  loading.value = false
  importRoundModalShow.value = false
}
</script>

<template>
  <div class="pt-2">
    <NAlert type="info" :bordered="false">
      <p>{{ $t('apps.exportImport.tip') }}</p>
    </NAlert>
    <div class="flex justify-center m-[50px]">
      <div class="m-[10px]">
        <NUpload
          accept=".sun-panel.json,.sunpanel.json"
          directory-dnd
          :default-upload="false"
          :show-file-list="false"
          @change="handleFileChange"
        >
          <NButton type="info" size="large" :loading="uploadLoading">
            <template #icon>
              <SvgIcon icon="fa6:solid-file-import" />
            </template>
            {{ $t('apps.exportImport.import') }}
          </NButton>
        </NUpload>
      </div>
      <div class="m-[10px]">
        <NButton type="info" size="large" @click="exportRoundModalShow = !exportRoundModalShow">
          <template #icon>
            <SvgIcon icon="fa6:solid-file-export" />
          </template>
          {{ $t('apps.exportImport.export') }}
        </NButton>
      </div>
    </div>

    <div class="flex justify-center">
      <a target="_blank">{{ $t('apps.exportImport.transmuteStandard') }}</a>
    </div>

    <!-- 调试模式 -->
    <div v-if="debug">
      <NButton @click="importCheck">
        检查导入
      </NButton>

      <!-- <NButton @click="exportJsonS">
      导出JSON
    </NButton> -->

      <NButton @click="jsonData = ''">
        清空导入数据
      </NButton>

      <NInput
        v-model:value="jsonData"
        type="textarea"
        placeholder="基本的 Textarea"
      />

      <div v-if="jsonData">
        <h2>JSON 数据</h2>
        <pre>{{ jsonData }}</pre>
      </div>
    </div>

    <RoundCardModal v-model:show="importRoundModalShow" style="max-width: 400px;" :title=" $t('apps.exportImport.import')">
      <div v-if="importWarning.length > 0">
        <NAlert :title="$t('common.warning')" type="warning">
          <div v-for="(text, index) in importWarning " :key="index">
            {{ text }}
          </div>
        </NAlert>
      </div>
      <NDivider title-placement="left">
        {{ $t('apps.exportImport.selectImportData') }}
      </NDivider>

      <NSpace justify="center" style="margin-top: 20px;">
        <NCheckboxGroup v-model:value="checkedItems">
          <NCheckbox v-if="importItems.includes('pages')" value="pages" label="页面" />
          <NCheckbox v-if="importItems.includes('icons')" value="icons" :label="$t('apps.exportImport.moduleIcon')" />
          <NCheckbox v-if="importItems.includes('notepads')" value="notepads" :label="$t('apps.exportImport.moduleNotepad')" />
          <NCheckbox v-if="importItems.includes('style')" value="style" :label="$t('apps.exportImport.moduleStyle')" />
        </NCheckboxGroup>
      </NSpace>
      <NSpace justify="center">
        <div class="mt-[50px]">
          <NButton type="success" :disabled="checkedItems.length === 0" :loading="loading" @click="handleStartImport">
            {{ $t('common.continue') }}
          </NButton>
        </div>
      </NSpace>
    </RoundCardModal>

    <RoundCardModal v-model:show="exportRoundModalShow" style="max-width: 400px;" :title=" $t('apps.exportImport.export')">
      <NDivider title-placement="left">
        {{ $t('apps.exportImport.selectExportData') }}
      </NDivider>

      <NSpace justify="center" style="margin-top: 20px;">
        <NCheckboxGroup v-model:value="checkedItems">
          <NCheckbox v-if="importItems.includes('pages')" value="pages" label="页面" />
          <NCheckbox v-if="importItems.includes('icons')" value="icons" :label="$t('apps.exportImport.moduleIcon')" />
          <NCheckbox v-if="importItems.includes('notepads')" value="notepads" :label="$t('apps.exportImport.moduleNotepad')" />
          <NCheckbox v-if="importItems.includes('style')" value="style" :label="$t('apps.exportImport.moduleStyle')" />
        </NCheckboxGroup>
      </NSpace>
      <NSpace justify="center">
        <div class="mt-[50px]">
          <NButton type="success" :disabled="checkedItems.length === 0" :loading="loading" @click="handleStartExport">
            {{ $t('common.continue') }}
          </NButton>
        </div>
      </NSpace>
    </RoundCardModal>
  </div>
</template>
