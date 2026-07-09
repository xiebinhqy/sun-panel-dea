<script setup lang="ts">
import { NButton, NDivider, NGradientText, NModal, NProgress, NSpin, NTag, useDialog, useMessage } from 'naive-ui'
import { onMounted, ref } from 'vue'
import { get } from '@/api/system/about'
import { checkUpdate, performUpdate } from '@/api/system/update'
import { useAppStore } from '@/store'
import srcSvglogo from '@/assets/logo.svg'
import srcGitee from '@/assets/about_image/gitee.png'
import srcGithub from '@/assets/about_image/github.png'
import srcDocker from '@/assets/about_image/docker.png'
import srcBilibili from '@/assets/about_image/bilibili.png'
import srcYoutube from '@/assets/about_image/youtube.png'
import srcQQGroupQR from '@/assets/about_image/qq_group_qr2.png'
import { RoundCardModal } from '@/components/common'
import { t } from '@/locales'

interface Version {
  versionName: string
  versionCode: number
}

interface UpdateCheckResult {
  hasUpdate: boolean
  currentVersion: string
  latestVersion: string
  releaseUrl: string
  downloadUrl: string
  releaseNotes: string
}

const ms = useMessage()
const dialog = useDialog()
const appStore = useAppStore()
const versionName = ref('')
const versionCode = ref(0)
const qqGroupQRShow = ref(false)
const frontVersion = import.meta.env.VITE_APP_VERSION || 'unknown'

// 更新状态
const updating = ref(false)
const updateProgress = ref(0)
const updateStatus = ref('')
const updateModalShow = ref(false)
const updateCheckResult = ref<UpdateCheckResult | null>(null)
const checkingUpdate = ref(false)

onMounted(() => {
  get<Version>().then((res) => {
    if (res.code === 0) {
      versionName.value = res.data.versionName
      versionCode.value = res.data.versionCode
    }
  })
})

// 检查更新
async function handleCheckUpdate() {
  checkingUpdate.value = true
  updateStatus.value = t('apps.about.checkingUpdate')
  updateModalShow.value = true

  try {
    const res = await checkUpdate<UpdateCheckResult>()
    if (res.code === 0) {
      updateCheckResult.value = res.data
      if (res.data.hasUpdate)
        updateStatus.value = t('apps.about.updateAvailable', { version: res.data.latestVersion })

      else
        updateStatus.value = t('apps.about.alreadyLatest')
    }
    else {
      ms.error(res.msg || t('apps.about.checkUpdateFail'))
      updateModalShow.value = false
    }
  }
  catch {
    ms.error(t('apps.about.checkUpdateFail'))
    updateModalShow.value = false
  }
  finally {
    checkingUpdate.value = false
  }
}

// 执行更新
function handleUpdate() {
  dialog.warning({
    title: t('apps.about.confirmUpdate'),
    content: t('apps.about.confirmUpdateContent'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      updating.value = true
      updateProgress.value = 10
      updateStatus.value = t('apps.about.downloadingUpdate')

      try {
        const res = await performUpdate<null>()
        if (res.code === 0) {
          updateProgress.value = 100
          updateStatus.value = t('apps.about.updateSuccess')
          ms.success(t('apps.about.updateSuccess'))
          // 延迟重启
          setTimeout(() => {
            window.location.reload()
          }, 3000)
        }
        else {
          ms.error(res.msg || t('apps.about.updateFail'))
          updateProgress.value = 0
          updating.value = false
        }
      }
      catch {
        ms.error(t('apps.about.updateFail'))
        updateProgress.value = 0
        updating.value = false
      }
    },
  })
}
</script>

<template>
  <div class="pt-5">
    <div class="flex flex-col items-center justify-center">
      <img :src="srcSvglogo" width="100" height="100" alt="">
      <div class="text-3xl font-semibold">
        {{ $t('common.appName') }}
      </div>
      <div class="text-xl">
        <NGradientText type="info">
          v{{ versionName }}
        </NGradientText>
      </div>
      <div class="mt-2 flex gap-2">
        <NButton size="small" @click="handleCheckUpdate">
          <template #icon>
            <span class="i-material-symbols:update" />
          </template>
          {{ $t('apps.about.checkUpdate') }}
        </NButton>
      </div>
    </div>

    <NDivider style="margin:10px 0">
      •
    </NDivider>
    <div class="flex flex-col items-center justify-center text-base">
      <div>
        {{ $t('apps.about.author') }}<a href="https://github.com/xiebinhqy" target="_blank" class="link">红烧猎人</a> | <a href="https://github.com/xiebinhqy/sun-panel-dea/blob/master/doc/donate.md" target="_blank" class="text-red-600 hover:text-red-900">{{ $t('apps.about.donate') }}</a>
      </div>
      <div>
        {{ $t('apps.about.issue') }}<a href="https://github.com/xiebinhqy/sun-panel-dea/issues" target="_blank" class="link">Github Issues</a>
      </div>
      <div>
        {{ $t('apps.about.discussions') }}<a href="https://github.com/xiebinhqy/sun-panel-dea/discussions" target="_blank" class="link">Github Discussions</a>
      </div>
      <div>
        {{ $t('apps.about.QQGroup') }}<a href="http://qm.qq.com/cgi-bin/qm/qr?_wv=1027&k=K6UII6aEPZUeDRIPOEpOSJZH-Vmr_RPu&authKey=jEXhnVekLbDDx5UkQzKtd3bRmhZggkGBxmvW4NT5LLIAFP7toMmqABwvkANGHbLb&noverify=0&group_code=831615449" target="_blank" class="link">{{ $t("apps.about.addQQGroupUrl") }}</a>
        |
        <span class="link cursor-pointer" @click="qqGroupQRShow = !qqGroupQRShow">
          {{ $t('apps.about.QR') }}
        </span>
      </div>

      <div class="flex mt-[10px] flex-wrap justify-center">
        <div class="flex items-center mx-[10px]">
          <img class="w-[20px] h-[20px] mr-[5px]" :src="srcGithub" alt="">
          <a href="https://github.com/xiebinhqy/sun-panel-dea" target="_blank" class="link">Github</a>
        </div>
        <div class="flex items-center mx-[10px]">
          <img class="w-[20px] h-[20px] mr-[5px]" :src="srcGitee" alt="">
          <a href="https://gitee.com/hslr/sun-panel" target="_blank" class="link">Gitee</a>
        </div>
        <div class="flex items-center mx-[10px]">
          <img class="w-[20px] h-[20px] mr-[5px]" :src="srcDocker" alt="">
          <a href="https://hub.docker.com/r/hslr/sun-panel" target="_blank" class="link">Docker</a>
        </div>
        <div class="flex items-center mx-[10px]">
          <img class="w-[20px] h-[20px] mr-[5px]" :src="srcBilibili" alt="">
          <a href="https://space.bilibili.com/27407696/channel/collectiondetail?sid=2023810" target="_blank" class="link">Bilibili</a>
        </div>
        <div v-if="appStore.language !== 'zh-CN'" class="flex items-center mx-[10px]">
          <img class="w-[20px] h-[20px] mr-[5px]" :src="srcYoutube" alt="">
          <a href="https://www.youtube.com/channel/UCKwbFmKU25R602z6P2fgPYg" target="_blank" class="link">YouTube</a>
        </div>
      </div>

      <div class="mt-5">
        <NTag :bordered="false" size="small">
          {{ $t("apps.about.frontVersionText") }}: FV-{{ frontVersion }}
        </NTag>
      </div>

      <RoundCardModal v-model:show="qqGroupQRShow" title="交流群二维码" style="width: 300px;">
        <div class="text-center">
          - 如果失效请返回联系作者 -
        </div>
        <div class="flex justify-center">
          <img :src="srcQQGroupQR" class="h-[260px]">
        </div>
      </RoundCardModal>
    </div>

    <!-- 更新弹窗 -->
    <NModal v-model:show="updateModalShow" preset="card" title="更新" style="max-width: 500px;" :bordered="true" size="small">
      <div class="flex flex-col items-center gap-4 py-4">
        <NSpin v-if="checkingUpdate" size="medium" />
        <div v-else-if="updateCheckResult" class="w-full">
          <div v-if="updateCheckResult.hasUpdate" class="text-center">
            <div class="text-lg font-semibold text-green-500 mb-2">
              {{ $t('apps.about.newVersionFound') }}
            </div>
            <div class="mb-4">
              {{ $t('apps.about.currentVersion') }}: v{{ updateCheckResult.currentVersion }}<br>
              {{ $t('apps.about.latestVersion') }}: v{{ updateCheckResult.latestVersion }}
            </div>
            <div v-if="updateCheckResult.releaseNotes" class="text-left mb-4 p-3 bg-gray-50 dark:bg-gray-800 rounded max-h-[200px] overflow-y-auto text-sm">
              {{ updateCheckResult.releaseNotes }}
            </div>
            <NButton type="primary" :loading="updating" :disabled="updating" @click="handleUpdate">
              {{ updating ? $t('apps.about.updating') : $t('apps.about.updateNow') }}
            </NButton>
          </div>
          <div v-else class="text-center">
            <div class="text-lg text-gray-500">
              {{ $t('apps.about.alreadyLatest') }}
            </div>
          </div>
        </div>
        <div v-if="updating" class="w-full">
          <NProgress :percentage="updateProgress" :height="8" indicator-placement="inside" processing />
          <div class="text-center text-sm text-gray-400 mt-2">
            {{ updateStatus }}
          </div>
        </div>
      </div>
    </NModal>
  </div>
</template>

<style>
.link {
  color: rgb(0, 89, 255);
}
</style>
