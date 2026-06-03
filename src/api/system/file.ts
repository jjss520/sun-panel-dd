import { post } from '@/utils/request'

export function getList<T>(path?: string) {
  return post<T>({
    url: '/file/getList',
    data: path ? { path } : {},
  })
}

export function deletes<T>(ids: number[]) {
  return post<T>({
    url: '/file/deletes',
    data: { ids },
  })
}

export function uploadFiles<T>(files: File[]) {
  const formData = new FormData()
  files.forEach(file => {
    formData.append('files[]', file)
  })
  
  return post<T>({
    url: '/file/uploadFiles',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
}
