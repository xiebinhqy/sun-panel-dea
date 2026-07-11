import { post } from '@/utils/request'

export interface VersionCheckResult {
  hasUpdate: boolean
  currentVersion: string
  latestVersion: string
  releaseNotes: string
  downloadUrl: string
  isDocker: boolean
  isUpdating: boolean
}

export interface UpdateResult {
  message: string
}

export interface UpdateStatus {
  status: 'idle' | 'updating'
}

export function checkUpdate<T>() {
  return post<T>({
    url: '/update/check',
  })
}

export function performUpdate<T>() {
  return post<T>({
    url: '/update/perform',
  })
}

export function getUpdateStatus<T>() {
  return post<T>({
    url: '/update/status',
  })
}