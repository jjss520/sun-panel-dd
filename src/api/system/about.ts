import { post } from '@/utils/request'

export function get<T>() {
  return post<T>({
    url: '/about',
  })
}
export function checkUpdate() {
  return post<{
    hasNewVersion: boolean
    latestVersion: string
    updateContent: string
  }>({
    url: '/checkUpdate',
  })
}
