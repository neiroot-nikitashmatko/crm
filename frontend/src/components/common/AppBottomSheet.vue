<script setup lang="ts">
import { watch } from 'vue'

const show = defineModel<boolean>('show', { required: true })

withDefaults(
  defineProps<{
    title: string
    closeLabel?: string
    bodyAlign?: 'start' | 'center'
  }>(),
  {
    closeLabel: 'Закрыть',
    bodyAlign: 'start',
  },
)

const emit = defineEmits<{
  close: []
}>()

function handleClose() {
  show.value = false
}

watch(show, (isOpen, wasOpen) => {
  if (wasOpen && !isOpen) {
    emit('close')
  }
})
</script>

<template>
  <Teleport to="body">
    <Transition name="app-bottom-sheet-backdrop">
      <button
        v-if="show"
        type="button"
        class="app-bottom-sheet__backdrop"
        :aria-label="closeLabel"
        @click="handleClose"
      />
    </Transition>

    <Transition name="app-bottom-sheet">
      <section v-if="show" class="app-bottom-sheet" @click.stop>
        <header class="app-bottom-sheet__header">
          <h2 class="app-bottom-sheet__title">{{ title }}</h2>
          <button
            type="button"
            class="app-bottom-sheet__close-btn"
            :aria-label="closeLabel"
            @click="handleClose"
          >
            ×
          </button>
        </header>

        <div
          class="app-bottom-sheet__body"
          :class="bodyAlign === 'center' ? 'app-bottom-sheet__body--center' : undefined"
        >
          <slot />
        </div>

        <footer v-if="$slots.actions" class="app-bottom-sheet__actions">
          <slot name="actions" />
        </footer>
      </section>
    </Transition>
  </Teleport>
</template>

<style scoped>
.app-bottom-sheet__backdrop {
  position: fixed;
  inset: 0;
  border: 0;
  background: rgba(15, 23, 42, 0.2);
  z-index: 180;
  cursor: default;
}

.app-bottom-sheet {
  position: fixed;
  top: 15px;
  right: 15px;
  bottom: 0;
  left: 15px;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  border-radius: 12px 12px 0 0;
  box-shadow: 0 -8px 24px rgba(15, 23, 42, 0.15);
  z-index: 190;
  overflow: hidden;
}

.app-bottom-sheet__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-shrink: 0;
  padding: 14px 15px;
  border-bottom: 1px solid #e2e8f0;
}

.app-bottom-sheet__title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a202c;
}

.app-bottom-sheet__close-btn {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #475569;
  font-size: 20px;
  line-height: 1;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    color 0.15s ease,
    background-color 0.15s ease;
}

.app-bottom-sheet__close-btn:hover {
  border-color: #cbd5e1;
  color: #334155;
  background: #f8fafc;
}

.app-bottom-sheet__body {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  padding: 20px 15px;
  scrollbar-gutter: stable;
}

.app-bottom-sheet__body--center {
  display: flex;
  flex-direction: column;
}

.app-bottom-sheet__body--center::before,
.app-bottom-sheet__body--center::after {
  content: '';
  flex: 1 1 0;
  min-height: 0;
  pointer-events: none;
}

.app-bottom-sheet__actions {
  flex-shrink: 0;
  display: flex;
  justify-content: center;
  padding: 15px;
  border-top: 1px solid #e2e8f0;
  background: #ffffff;
}

.app-bottom-sheet-enter-active,
.app-bottom-sheet-leave-active {
  transition:
    transform 0.28s ease,
    opacity 0.28s ease;
}

.app-bottom-sheet-enter-from,
.app-bottom-sheet-leave-to {
  transform: translateY(100%);
  opacity: 0.98;
}

.app-bottom-sheet-backdrop-enter-active,
.app-bottom-sheet-backdrop-leave-active {
  transition: opacity 0.2s ease;
}

.app-bottom-sheet-backdrop-enter-from,
.app-bottom-sheet-backdrop-leave-to {
  opacity: 0;
}
</style>
