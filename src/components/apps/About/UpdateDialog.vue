<script setup lang="ts">
import { NButton, NModal, NSpace, NSpin, NText, NTag, NProgress, NDivider } from 'naive-ui'
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { checkUpdate, performUpdate, getUpdateStatus } from '@/api/system/update'
import type { VersionCheckResult } from '@/api/system/update'
import { t } from '@/locales'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'done'): void
}>()

const checking = ref(true)
const updating = ref(false)
const updateResult = ref<VersionCheckResult | null>(null)
const errorMsg = ref('')
const updateMessage = ref('')
const progressPercent = ref(0)
const progressText = ref('')
let statusTimer: ReturnType<typeof setInterval> | null = null

// 模拟进度（实际进度无法实时获取，用轮询状态文件）
function startProgressSimulation() {
  progressPercent.value = 10
  progressText.value = t('apps.about.updateProgress.pulling')
  
  const steps = [
    { percent: 30, textKey: 'apps.about.updateProgress.pulling', delay: 5000 },
    { percent: 60, textKey: 'apps.about.updateProgress.stopping', delay: 10000 },
    { percent: 80, textKey: 'apps.about.updateProgress.restarting', delay: 20000 },
    { percent: 95, textKey: 'apps.about.updateProgress.waiting', delay: 30000 },
  ]

  let stepIndex = 0
  const stepTimer = setInterval(() => {
    if (stepIndex < steps.length) {
      const step = steps[stepIndex]
      progressPercent.value = step.percent
      progressText.value = t(step.textKey)
      stepIndex++
    } else {
      clearInterval(stepTimer)
    }
  }, 3000)

  // 清理
  onUnmounted(() => {
    clearInterval(stepTimer)
  })
}

async function handleCheck() {
  checking.value = true
  errorMsg.value = ''
  try {
    const res = await checkUpdate<VersionCheckResult>()
    if (res.code === 0) {
      updateResult.value = res.data
    } else {
      errorMsg.value = res.msg || t('common.unknownError')
    }
  } catch (e: any) {
    errorMsg.value = e.message || t('common.networkError')
  } finally {
    checking.value = false
  }
}

async function handleUpdate() {
  try {
    updating.value = true
    updateMessage.value = ''
    errorMsg.value = ''
    
    const res = await performUpdate<{ message: string }>()
    if (res.code === 0) {
      const data = res.data as { message?: string }
      updateMessage.value = data?.message || t('apps.about.updateStarted')
      // 开始轮询更新状态
      startProgressSimulation()
      startStatusPolling()
    } else {
      errorMsg.value = res.msg || t('common.unknownError')
      updating.value = false
    }
  } catch (e: any) {
    errorMsg.value = e.message || t('common.networkError')
    updating.value = false
  }
}

function startStatusPolling() {
  statusTimer = setInterval(async () => {
    try {
      const res = await getUpdateStatus<{ status: string }>()
      const statusData = res.data as { status?: string }
      if (res.code === 0 && statusData?.status === 'idle') {
        // 更新完成
        progressPercent.value = 100
        progressText.value = t('apps.about.updateProgress.completed')
        if (statusTimer) {
          clearInterval(statusTimer)
          statusTimer = null
        }
      }
    } catch {
      // 忽略轮询错误
    }
  }, 3000)
}

function handleClose() {
  if (statusTimer) {
    clearInterval(statusTimer)
    statusTimer = null
  }
  emit('update:show', false)
  emit('done')
}

function handleCancel() {
  if (updating.value) return // 更新中不允许关闭
  handleClose()
}

watch(() => props.show, (newVal: boolean) => {
  if (newVal) {
    handleCheck()
  }
})

onMounted(() => {
  if (props.show) {
    handleCheck()
  }
})
</script>

<template>
  <NModal
    :show="props.show"
    @update:show="handleCancel"
    :mask-closable="!updating"
    preset="card"
    style="max-width: 600px; border-radius: 1rem;"
    :title="t('apps.about.checkUpdate')"
    size="small"
    role="dialog"
    aria-modal="true"
  >
    <!-- 检查中 -->
    <div v-if="checking && !updateResult" class="flex flex-col items-center py-8">
      <NSpin size="large" />
      <NText class="mt-4">{{ t('apps.about.checking') }}</NText>
    </div>

    <!-- 错误信息 -->
    <div v-if="errorMsg && !checking" class="flex flex-col items-center py-4">
      <NText type="error">{{ errorMsg }}</NText>
      <NButton class="mt-4" @click="handleCheck">{{ t('common.retry') }}</NButton>
    </div>

    <!-- 检查结果 -->
    <div v-if="updateResult && !checking && !updating" class="py-4">
      <div class="flex items-center justify-between mb-4">
        <NText>{{ t('apps.about.currentVersion') }}: </NText>
        <NTag type="info">{{ updateResult.currentVersion }}</NTag>
      </div>
      <div class="flex items-center justify-between mb-4">
        <NText>{{ t('apps.about.latestVersion') }}: </NText>
        <NTag :type="updateResult.hasUpdate ? 'warning' : 'success'">
          {{ updateResult.latestVersion }}
        </NTag>
      </div>

      <div class="mb-4">
        <NTag v-if="updateResult.hasUpdate" type="warning">
          {{ t('apps.about.newVersionAvailable') }}
        </NTag>
        <NTag v-else type="success">
          {{ t('apps.about.alreadyLatest') }}
        </NTag>
      </div>

      <!-- 更新日志 -->
      <NDivider />
      <NText depth="3" class="text-sm">{{ t('apps.about.releaseNotes') }}</NText>
      <div class="mt-2 max-h-40 overflow-y-auto text-sm bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
        <pre class="whitespace-pre-wrap">{{ updateResult.releaseNotes || t('apps.about.noReleaseNotes') }}</pre>
      </div>
      <NDivider />

      <!-- 更新按钮 -->
      <div class="flex justify-center">
        <NButton
          v-if="updateResult.isDocker && updateResult.hasUpdate"
          type="primary"
          size="large"
          @click="handleUpdate"
        >
          {{ t('apps.about.updateNow') }}
        </NButton>
        <NText v-if="!updateResult.isDocker && updateResult.hasUpdate" type="warning" depth="3">
          {{ t('apps.about.nonDockerUpdateTip') }}
        </NText>
      </div>
    </div>

    <!-- 更新进度 -->
    <div v-if="updating" class="py-4">
      <div class="flex flex-col items-center">
        <NSpin size="large" />
        <NText class="mt-4 text-lg font-bold">{{ t('apps.about.updating') }}</NText>
        
        <div class="w-full mt-6">
          <NProgress
            type="line"
            :percentage="progressPercent"
            :indicator-placement="'inside'"
            :height="24"
            :border-radius="12"
            :fill-bar-color="'#18a058'"
            processing
          />
        </div>
        
        <NText class="mt-2" depth="3">{{ progressText }}</NText>
        
        <div class="mt-6 text-center">
          <NText depth="2" class="text-sm">
            {{ t('apps.about.updateTip') }}
          </NText>
        </div>
      </div>
    </div>

    <!-- 底部按钮 -->
    <template #footer>
      <div class="flex justify-end">
        <NSpace>
          <NButton @click="handleCancel" :disabled="updating">
            {{ updating ? t('common.updating') : t('common.close') }}
          </NButton>
          <NButton v-if="!updateResult?.hasUpdate && !updating" type="primary" @click="handleClose">
            {{ t('common.confirm') }}
          </NButton>
        </NSpace>
      </div>
    </template>
  </NModal>
</template>