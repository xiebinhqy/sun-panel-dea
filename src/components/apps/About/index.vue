<script setup lang="ts">
import { NDivider, NGradientText, NTag } from 'naive-ui'
import { onMounted, ref } from 'vue'
import UpdateDialog from './UpdateDialog.vue'
import { get } from '@/api/system/about'
import srcSvglogo from '@/assets/logo.svg'
interface Version {
  versionName: string
  versionCode: number
}

const versionName = ref('')
const updateDialogShow = ref(false)
const frontVersion = import.meta.env.VITE_APP_VERSION || 'unknown'

onMounted(() => {
  get<Version>().then((res) => {
    if (res.code === 0)
      versionName.value = res.data.versionName
  })
})
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
          <span class="font-semibold">v{{ versionName }}</span>
        </NGradientText>
      </div>
      <div class="mt-2">
        <a class="link cursor-pointer" @click="updateDialogShow = true">{{ $t('apps.about.checkUpdate') }}</a>
      </div>
    </div>

    <NDivider style="margin:10px 0">
      •
    </NDivider>
    <div class="flex flex-col items-center justify-center text-base px-4">
      <div class="text-sm text-gray-500 dark:text-gray-400 text-center leading-relaxed mb-4">
        本系统基于 <a href="https://github.com/hslr-s/sun-panel" target="_blank" class="link">Sun-Panel</a> v1.3.0 开源版本进行二次开发与优化
      </div>

      <div class="mt-5">
        <NTag :bordered="false" size="small">
          {{ $t("apps.about.frontVersionText") }}: FV-{{ frontVersion }}
        </NTag>
      </div>

      <UpdateDialog v-model:show="updateDialogShow" @done="updateDialogShow = false" />
    </div>
  </div>
</template>

<style>
.link{
    color:rgb(0, 89, 255)
}
</style>
