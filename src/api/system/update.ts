import { post } from '@/utils/request'

// 检查GitHub最新版本
export function checkUpdate<T>() {
  return post<T>({
    url: '/system/checkUpdate',
  })
}

// 执行更新
export function performUpdate<T>() {
  return post<T>({
    url: '/system/performUpdate',
  })
}
