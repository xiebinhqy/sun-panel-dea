<script setup lang="ts">
import { watch } from 'vue'
import { NConfigProvider } from 'naive-ui'
import { NaiveProvider } from '@/components/common'
import { useTheme } from '@/hooks/useTheme'
import { useLanguage } from '@/hooks/useLanguage'
import { usePanelState } from '@/store'

const { theme, themeOverrides } = useTheme()
const { language } = useLanguage()
const panelState = usePanelState()

// 动态设置站点标题和图标
watch(() => panelState.panelConfig.siteTitle, (val) => {
  if (val)
    document.title = val
})
watch(() => panelState.panelConfig.siteFavicon, (val) => {
  if (val) {
    let link = document.querySelector<HTMLLinkElement>('link[rel*="icon"]')
    if (!link) {
      link = document.createElement('link')
      link.rel = 'icon'
      document.head.appendChild(link)
    }
    link.href = val
  }
})
</script>

<template>
  <NConfigProvider
    class="h-full"
    :theme="theme"
    :theme-overrides="themeOverrides"
    :locale="language"
  >
    <NaiveProvider>
      <RouterView />
    </NaiveProvider>
  </NConfigProvider>
</template>
