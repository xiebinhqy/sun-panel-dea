<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { NLayout, NLayoutContent, NLayoutSider, NSpace } from 'naive-ui'
import { useAuthStore } from '@/store'
import { AppLoader, RoundCardModal, SvgIcon } from '@/components/common'
import { t } from '@/locales'

interface App {
  name: string
  componentName: string
  icon: string
  auth?: number
}
const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', visible: boolean): void
}>()

const componentName = ref('UserInfo')
const collapsed = ref(false)
const screenWidth = ref(0)
const isSmallScreen = ref(false)
const defaultTitle = t('appLauncher.title')
const title = ref('')
const height = ref('500px')

// 拖拽相关状态
const isDragging = ref(false)
const dragStartPos = ref({ x: 0, y: 0 })
const dragStartModalPos = ref({ x: 0, y: 0 })
const modalPosition = ref({ x: -1, y: -1 }) // -1 表示居中（使用默认居中）
const modalSize = ref({ width: 0, height: 0 })
const modalRef = ref<HTMLElement | null>(null)

const apps = ref<App[]>([
  {
    name: t('apps.userInfo.appName'),
    componentName: 'UserInfo',
    icon: 'material-symbols-person-edit-outline-rounded',
  },
  {
    name: t('apps.baseSettings.appName'),
    componentName: 'Style',
    icon: 'ion-color-palette-outline',
  },
  {
    name: t('apps.itemGroupManage.appName'),
    componentName: 'ItemGroupManage',
    icon: 'material-symbols-ad-group-outline-rounded',
  },
  {
    name: t('apps.uploadsFileManager.appName'),
    componentName: 'UploadFileManager',
    icon: 'tabler:file-upload',
  },
  {
    name: t('apps.exportImport.appName'),
    componentName: 'ImportExport',
    icon: 'icon-park-outline-import-and-export',
  },
  {
    name: t('apps.about.appName'),
    componentName: 'About',
    icon: 'lucide-info',
  },
])

const authStore = useAuthStore()

const show = computed({
  get: () => props.visible,
  set: (visible: boolean) => {
    emit('update:visible', visible)
    // 关闭时重置位置
    if (!visible)
      modalPosition.value = { x: -1, y: -1 }
  },
})

// 获取弹窗尺寸
function getModalSize() {
  const modalEl = document.querySelector('.app-starter-modal-content') as HTMLElement
  if (modalEl) {
    const rect = modalEl.getBoundingClientRect()
    return { width: rect.width, height: rect.height }
  }
  return { width: 900, height: 500 } // 默认尺寸
}

// 计算居中位置
function getCenterPosition() {
  const { width, height } = getModalSize()
  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight
  return {
    x: (viewportWidth - width) / 2,
    y: (viewportHeight - height) / 2,
  }
}

// 更新弹窗位置
function updateModalPosition() {
  if (modalPosition.value.x === -1) {
    // 居中模式，不设置 left/top
    return
  }
  const modalEl = document.querySelector('.app-starter-modal-content') as HTMLElement
  if (modalEl) {
    // 设置固定位置
    modalEl.style.position = 'fixed'
    modalEl.style.left = `${modalPosition.value.x}px`
    modalEl.style.top = `${modalPosition.value.y}px`
    modalEl.style.transform = 'none'
    modalEl.style.margin = '0'
  }
}

// 鼠标按下：准备拖拽
function handleDragStart(e: MouseEvent) {
  // 只响应左键
  if (e.button !== 0)
    return

  // 如果当前是居中模式，先获取实际位置
  if (modalPosition.value.x === -1) {
    const centerPos = getCenterPosition()
    modalPosition.value = centerPos
  }

  // 记录鼠标起始位置
  dragStartPos.value = { x: e.clientX, y: e.clientY }
  // 记录弹窗起始位置
  dragStartModalPos.value = { ...modalPosition.value }
  isDragging.value = true

  // 阻止文本选中
  e.preventDefault()

  document.addEventListener('mousemove', handleDragMove)
  document.addEventListener('mouseup', handleDragEnd)
}

// 鼠标移动：更新位置
function handleDragMove(e: MouseEvent) {
  if (!isDragging.value)
    return

  const deltaX = e.clientX - dragStartPos.value.x
  const deltaY = e.clientY - dragStartPos.value.y

  // 计算新位置
  let newX = dragStartModalPos.value.x + deltaX
  let newY = dragStartModalPos.value.y + deltaY

  // 获取弹窗尺寸
  const { width, height } = getModalSize()
  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight

  // 限制在视口范围内
  newX = Math.max(0, Math.min(newX, viewportWidth - width))
  newY = Math.max(0, Math.min(newY, viewportHeight - height))

  modalPosition.value = { x: newX, y: newY }

  // 应用位置
  updateModalPosition()
}

// 鼠标抬起：结束拖拽
function handleDragEnd() {
  isDragging.value = false
  document.removeEventListener('mousemove', handleDragMove)
  document.removeEventListener('mouseup', handleDragEnd)
}

function handleClickApp(item: App) {
  componentName.value = item.componentName
  if (isSmallScreen.value)
    collapsed.value = true
}

function getScreenWidth() {
  return window.innerWidth
}

function handleResize() {
  screenWidth.value = getScreenWidth()
  if (screenWidth.value < 640) {
    collapsed.value = true
    isSmallScreen.value = true
  }
  else {
    collapsed.value = false
    isSmallScreen.value = false
  }
}

onMounted(() => {
  const adminApp: App = {
    name: t('adminSettingUsers.appName'),
    componentName: 'Users',
    icon: 'lucide-users',
    auth: 1,
  }
  if (authStore.userInfo?.role === 1)
    apps.value.push(adminApp)

  window.addEventListener('resize', handleResize)
  handleResize()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  document.removeEventListener('mousemove', handleDragMove)
  document.removeEventListener('mouseup', handleDragEnd)
})
</script>

<template>
  <div>
    <RoundCardModal
      v-model:show="show"
      style="max-width: 900px;"
      size="small"
    >
      <template #header>
        <div
          class="flex items-center select-none cursor-grab active:cursor-grabbing"
          @mousedown="handleDragStart"
        >
          <div class="text-3xl cursor-pointer" style="color:var(--n-color-target)">
            <SvgIcon class="transition-all duration-500" :icon="collapsed ? 'tabler-layout-sidebar-right-collapse-filled' : 'tabler-layout-sidebar-left-collapse-filled'" />
          </div>
          <div class="ml-1">
            {{ title === '' ? defaultTitle : title }}
          </div>
        </div>
      </template>
      <div
        ref="modalRef"
        class="w-full h-full app-starter-modal-content"
        :class="{ 'app-starter-dragging': isDragging }"
      >
        <NSpace vertical size="large" style="height: 100%;width: 100%;">
          <NLayout has-sider style="border-radius:0.75rem;">
            <NLayoutSider
              v-model:collapsed="collapsed"
              collapse-mode="width"
              :collapsed-width="0"
              :width="isSmallScreen ? '100%' : 240"
              style="height: 100%;"
              content-style="overflow: hidden"
            >
              <div class="w-full h-full dark:bg-[#2c2c32]">
                <div
                  class="p-[5px] bg-slate-200 dark:bg-zinc-900 rounded-xl overflow-auto"
                  :style="{
                    width: isSmallScreen ? '100%' : '220px',
                    minWidth: '200px',
                    height,
                  }"
                >
                  <div
                    v-for="(item, index) in apps"
                    :key="index"
                    :style="{ color: componentName === item.componentName ? 'var(--n-color-target)' : '' }"
                    @click="handleClickApp(item)"
                  >
                    <div
                      class="bg-white dark:bg-zinc-800 p-[10px] rounded-lg mb-[5px] font-bold cursor-pointer flex items-center hover:bg-slate-50 focus:bg-slate-50"
                    >
                      <div class="flex items-center justify-center">
                        <div class="text-lg">
                          <SvgIcon :icon="item.icon" />
                        </div>
                        <span class="ml-2">{{ item.name }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </NLayoutSider>
            <NLayoutContent :content-style="{ height }">
              <div class="rounded-2xl h-full overflow-auto transition-all duration-500 min-w-[300px] h-full" :class="(isSmallScreen && !collapsed) ? 'opacity-0' : 'opacity-100'">
                <AppLoader :component-name="componentName" class="h-full" />
              </div>
            </NLayoutContent>
          </NLayout>
        </NSpace>
      </div>
    </RoundCardModal>
  </div>
</template>

<style scoped>
.text-shadow {
  text-shadow: 0px 0px 5px gray;
}

.cursor-grab {
  cursor: grab;
  user-select: none;
}

.cursor-grab:active {
  cursor: grabbing;
}
</style>

<style>
.dark .app-starter-modal-content .n-layout {
  background-color: #2c2c32;
}

/* 拖拽时禁用 transition，使拖拽更流畅 */
.app-starter-dragging {
  transition: none !important;
}
</style>
