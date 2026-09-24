import { post } from '@/utils/request'

// 获取页面列表
export function getList<T>() {
  return post<T>({ url: '/panel/itemPage/getList', data: {} })
}

// 编辑页面
export function edit<T>(data: any) {
  return post<T>({ url: '/panel/itemPage/edit', data })
}

// 删除页面
export function deletes<T>(ids: number[]) {
  return post<T>({ url: '/panel/itemPage/deletes', data: { ids } })
}

// 保存页面排序
export function saveSort<T>(sortItems: Array<{ id: number; sort: number }>) {
  return post<T>({ url: '/panel/itemPage/saveSort', data: { sortItems } })
}
