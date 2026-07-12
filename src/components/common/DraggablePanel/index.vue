<script setup lang="ts">
import { onUnmounted, ref } from 'vue'

defineProps<{
  show: boolean
  title?: string
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
}>()

const panelRef = ref<HTMLElement | null>(null)
const dragging = ref(false)
const dragStartX = ref(0)
const dragStartY = ref(0)
const panelLeft = ref(200)
const panelTop = ref(100)
const panelWidth = ref(900)
const panelHeight = ref(500)

function onMouseDown(e: MouseEvent) {
  // 获取面板实际宽高用于边界计算
  if (panelRef.value) {
    panelWidth.value = panelRef.value.offsetWidth
    panelHeight.value = panelRef.value.offsetHeight
  }
  dragging.value = true
  dragStartX.value = e.clientX - panelLeft.value
  dragStartY.value = e.clientY - panelTop.value
  document.body.style.cursor = 'move'
  document.body.style.userSelect = 'none'
  document.addEventListener('mousemove', onMouseMove)
  document.addEventListener('mouseup', onMouseUp)
}

function onMouseMove(e: MouseEvent) {
  if (!dragging.value)
    return
  panelLeft.value = e.clientX - dragStartX.value
  panelTop.value = e.clientY - dragStartY.value
  // 边界限制：面板完全在视口内
  // 左边界：面板左侧 >= 0
  if (panelLeft.value < 0)
    panelLeft.value = 0
  // 上边界：面板顶部 >= 0
  if (panelTop.value < 0)
    panelTop.value = 0
  // 右边界：面板右侧 <= 视口宽度
  const maxLeft = window.innerWidth - panelWidth.value
  if (panelLeft.value > maxLeft)
    panelLeft.value = maxLeft
  // 下边界：面板底部 <= 视口高度
  const maxTop = window.innerHeight - panelHeight.value
  if (panelTop.value > maxTop)
    panelTop.value = maxTop
}

function onMouseUp() {
  dragging.value = false
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
  document.removeEventListener('mousemove', onMouseMove)
  document.removeEventListener('mouseup', onMouseUp)
}

function handleClose() {
  emit('update:show', false)
}

// 点击遮罩关闭
function handleMaskClick(e: MouseEvent) {
  if (e.target === e.currentTarget)
    handleClose()
}

onUnmounted(() => {
  document.removeEventListener('mousemove', onMouseMove)
  document.removeEventListener('mouseup', onMouseUp)
})
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="draggable-mask" @mousedown="handleMaskClick">
      <div
        ref="panelRef"
        class="draggable-panel"
        :style="{
          left: `${panelLeft}px`,
          top: `${panelTop}px`,
          cursor: dragging ? 'move' : 'default',
        }"
      >
        <!-- 标题栏拖拽手柄 -->
        <div class="draggable-handle" @mousedown="onMouseDown">
          <div class="flex items-center">
            <slot name="header">
              <span class="text-lg font-semibold">{{ title }}</span>
            </slot>
          </div>
          <button class="close-btn" @click="handleClose">
            &times;
          </button>
        </div>
        <!-- 内容区 -->
        <div class="draggable-content">
          <slot />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.draggable-mask {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.3);
  z-index: 1000;
}

.draggable-panel {
  position: fixed;
  width: 900px;
  max-width: 90vw;
  max-height: 80vh;
  background: #fff;
  border-radius: 1rem;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  z-index: 1001;
}

.draggable-handle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: #f5f5f5;
  border-bottom: 1px solid #e5e5e5;
  cursor: move;
  user-select: none;
  border-radius: 1rem 1rem 0 0;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #666;
  padding: 0 8px;
  line-height: 1;
}

.close-btn:hover {
  color: #333;
}

.draggable-content {
  flex: 1;
  overflow: auto;
  padding: 16px;
}

/* 深色模式 */
:root.dark .draggable-panel {
  background: #2c2c32;
  color: #fff;
}

:root.dark .draggable-handle {
  background: #3c3c42;
  border-bottom-color: #4c4c52;
}

:root.dark .close-btn {
  color: #aaa;
}

:root.dark .close-btn:hover {
  color: #fff;
}
</style>
