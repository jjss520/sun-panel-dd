import { post } from '@/utils/request'

/**
 * 系统设置接口
 * 通用的配置保存和获取接口,支持多个配置项
 */

// 保存系统设置(支持批量保存)
export function setSystemSettings<T>(settings: Record<string, any>) {
  return post<T>({
    url: '/system/setting/set',
    data: { settings },
  })
}

// 获取系统设置(支持获取多个配置项)
export function getSystemSettings<T>(configNames?: string[]) {
  return post<T>({
    url: '/system/setting/get',
    data: { configNames },
  })
}

// 获取单个系统设置
export function getSystemSetting<T>(configName: string) {
  return post<T>({
    url: '/system/setting/getSingle',
    data: { configName },
  })
}
