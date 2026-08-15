<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import AppModal from '@/components/common/AppModal.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'

const VIEWPORT_SIZE = 280
const CIRCLE_SIZE = 240
const OUTPUT_SIZE = 512

const show = defineModel<boolean>('show', { required: true })

const props = withDefaults(
  defineProps<{
    imageUrl: string
    title?: string
    canReplace?: boolean
  }>(),
  {
    title: 'Выберите миниатюру',
    canReplace: false,
  },
)

const emit = defineEmits<{
  confirm: [file: File]
  replace: []
  close: []
}>()

const image = ref<HTMLImageElement | null>(null)
const scale = ref(1)
const minScale = ref(1)
const offsetX = ref(0)
const offsetY = ref(0)
const isDragging = ref(false)
const dragStartX = ref(0)
const dragStartY = ref(0)
const dragOriginX = ref(0)
const dragOriginY = ref(0)

const circleInset = (VIEWPORT_SIZE - CIRCLE_SIZE) / 2

const imageStyle = computed(() => {
  const loaded = image.value
  if (!loaded) return { display: 'none' }

  return {
    width: `${loaded.naturalWidth * scale.value}px`,
    height: `${loaded.naturalHeight * scale.value}px`,
    transform: `translate(${offsetX.value}px, ${offsetY.value}px)`,
  }
})

watch(
  () => [show.value, props.imageUrl] as const,
  ([isOpen, url]) => {
    if (!isOpen || !url) {
      image.value = null
      return
    }
    loadImage(url)
  },
  { immediate: true },
)

function loadImage(url: string) {
  const next = new Image()
  next.onload = () => {
    image.value = next
    const coverScale = CIRCLE_SIZE / Math.min(next.naturalWidth, next.naturalHeight)
    minScale.value = coverScale
    scale.value = coverScale
    offsetX.value = (VIEWPORT_SIZE - next.naturalWidth * coverScale) / 2
    offsetY.value = (VIEWPORT_SIZE - next.naturalHeight * coverScale) / 2
  }
  next.src = url
}

function clampOffset() {
  const loaded = image.value
  if (!loaded) return

  const width = loaded.naturalWidth * scale.value
  const height = loaded.naturalHeight * scale.value
  const minX = circleInset + CIRCLE_SIZE - width
  const minY = circleInset + CIRCLE_SIZE - height

  offsetX.value = Math.min(circleInset, Math.max(minX, offsetX.value))
  offsetY.value = Math.min(circleInset, Math.max(minY, offsetY.value))
}

function handlePointerDown(event: PointerEvent) {
  const target = event.currentTarget as HTMLElement
  target.setPointerCapture(event.pointerId)
  isDragging.value = true
  dragStartX.value = event.clientX
  dragStartY.value = event.clientY
  dragOriginX.value = offsetX.value
  dragOriginY.value = offsetY.value
}

function handlePointerMove(event: PointerEvent) {
  if (!isDragging.value) return
  offsetX.value = dragOriginX.value + (event.clientX - dragStartX.value)
  offsetY.value = dragOriginY.value + (event.clientY - dragStartY.value)
  clampOffset()
}

function handlePointerUp(event: PointerEvent) {
  const target = event.currentTarget as HTMLElement
  if (target.hasPointerCapture(event.pointerId)) {
    target.releasePointerCapture(event.pointerId)
  }
  isDragging.value = false
}

function handleWheel(event: WheelEvent) {
  event.preventDefault()
  const loaded = image.value
  if (!loaded) return

  const factor = event.deltaY < 0 ? 1.08 : 1 / 1.08
  const nextScale = Math.min(minScale.value * 4, Math.max(minScale.value, scale.value * factor))
  const circleCenter = circleInset + CIRCLE_SIZE / 2
  const relativeX = (circleCenter - offsetX.value) / scale.value
  const relativeY = (circleCenter - offsetY.value) / scale.value

  scale.value = nextScale
  offsetX.value = circleCenter - relativeX * nextScale
  offsetY.value = circleCenter - relativeY * nextScale
  clampOffset()
}

function handleScaleInput(event: Event) {
  const loaded = image.value
  if (!loaded) return

  const value = Number((event.target as HTMLInputElement).value)
  const circleCenter = circleInset + CIRCLE_SIZE / 2
  const relativeX = (circleCenter - offsetX.value) / scale.value
  const relativeY = (circleCenter - offsetY.value) / scale.value

  scale.value = value
  offsetX.value = circleCenter - relativeX * scale.value
  offsetY.value = circleCenter - relativeY * scale.value
  clampOffset()
}

async function handleConfirm() {
  const loaded = image.value
  if (!loaded) return

  const sourceSize = CIRCLE_SIZE / scale.value
  const sourceX = (circleInset - offsetX.value) / scale.value
  const sourceY = (circleInset - offsetY.value) / scale.value

  const canvas = document.createElement('canvas')
  canvas.width = OUTPUT_SIZE
  canvas.height = OUTPUT_SIZE
  const context = canvas.getContext('2d')
  if (!context) return

  context.imageSmoothingEnabled = true
  context.imageSmoothingQuality = 'high'
  context.drawImage(loaded, sourceX, sourceY, sourceSize, sourceSize, 0, 0, OUTPUT_SIZE, OUTPUT_SIZE)

  const blob = await new Promise<Blob | null>((resolve) => {
    canvas.toBlob(resolve, 'image/jpeg', 0.9)
  })
  if (!blob) return

  emit('confirm', new File([blob], 'avatar.jpg', { type: 'image/jpeg' }))
  show.value = false
}

function handleReplace() {
  emit('replace')
}

onBeforeUnmount(() => {
  image.value = null
})
</script>

<template>
  <AppModal v-model:show="show" :title="title" width="default" @close="emit('close')">
    <div class="avatar-crop">
      <div
        class="avatar-crop__viewport"
        :class="{ 'avatar-crop__viewport--dragging': isDragging }"
        @pointerdown="handlePointerDown"
        @pointermove="handlePointerMove"
        @pointerup="handlePointerUp"
        @pointercancel="handlePointerUp"
        @wheel.prevent="handleWheel"
      >
        <img v-if="image" class="avatar-crop__image" :src="imageUrl" alt="" :style="imageStyle" />
        <div class="avatar-crop__mask" aria-hidden="true" />
      </div>

      <label class="avatar-crop__zoom">
        <input
          class="avatar-crop__zoom-input"
          type="range"
          :min="minScale"
          :max="minScale * 4"
          :step="minScale / 40"
          :value="scale"
          :disabled="!image"
          aria-label="Масштаб"
          @input="handleScaleInput"
        />
      </label>
    </div>

    <template #actions>
      <div class="avatar-crop__actions">
        <button
          v-if="canReplace"
          type="button"
          class="avatar-crop__replace"
          @click="handleReplace"
        >
          Загрузить другое фото
        </button>
        <AppModalButton :disabled="!image" @click="handleConfirm">Сохранить</AppModalButton>
      </div>
    </template>
  </AppModal>
</template>

<style scoped>
.avatar-crop {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
}

.avatar-crop__viewport {
  position: relative;
  width: 280px;
  height: 280px;
  overflow: hidden;
  border-radius: 16px;
  background: #1a202c;
  touch-action: none;
  cursor: grab;
  user-select: none;
}

.avatar-crop__viewport--dragging {
  cursor: grabbing;
}

.avatar-crop__image {
  position: absolute;
  top: 0;
  left: 0;
  max-width: none;
  pointer-events: none;
}

.avatar-crop__mask {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: rgba(15, 23, 42, 0.55);
  -webkit-mask-image: radial-gradient(circle 120px at center, transparent 119px, #000 120px);
  mask-image: radial-gradient(circle 120px at center, transparent 119px, #000 120px);
}

.avatar-crop__zoom {
  display: flex;
  width: 280px;
}

.avatar-crop__zoom-input {
  width: 100%;
  accent-color: #1f883d;
}

.avatar-crop__actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.avatar-crop__replace {
  min-width: min(100%, 220px);
  padding: 10px 20px;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  background: #ffffff;
  color: #475569;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.35;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.avatar-crop__replace:hover {
  background: #f8fafc;
  border-color: #94a3b8;
  color: #334155;
}
</style>
