<script setup lang="ts">
import { watch } from 'vue'
import { NModal } from 'naive-ui'

const show = defineModel<boolean>('show', { required: true })

withDefaults(
  defineProps<{
    title: string
    maskClosable?: boolean
    width?: 'default' | 'wide' | 'medium' | 'large' | 'xlarge'
    height?: 'auto' | 'tall'
    actionsAlign?: 'center' | 'end'
    bodyVariant?: 'default' | 'center' | 'date'
    closeLabel?: string
    zIndex?: number
    trapFocus?: boolean
    actionsReserved?: boolean
  }>(),
  {
    maskClosable: true,
    width: 'default',
    height: 'auto',
    actionsAlign: 'center',
    bodyVariant: 'default',
    closeLabel: 'Закрыть',
    trapFocus: true,
    actionsReserved: false,
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
  <NModal
    v-model:show="show"
    :mask-closable="maskClosable"
    :z-index="zIndex"
    :trap-focus="trapFocus"
  >
    <div
      class="app-modal"
      :class="[
        `app-modal--width-${width}`,
        `app-modal--height-${height}`,
        `app-modal--body-${bodyVariant}`,
      ]"
    >
      <header class="app-modal__header">
        <h3 class="app-modal__title">{{ title }}</h3>
        <div class="app-modal__header-actions">
          <slot name="header-actions" />
          <button
            type="button"
            class="app-modal__close"
            :aria-label="closeLabel"
            @click="handleClose"
          >
            ×
          </button>
        </div>
      </header>

      <div class="app-modal__body" :class="`app-modal__body--${bodyVariant}`">
        <slot />
      </div>

      <div
        v-if="$slots.actions"
        class="app-modal__actions"
        :class="[
          `app-modal__actions--${actionsAlign}`,
          { 'app-modal__actions--reserved': actionsReserved },
        ]"
      >
        <slot name="actions" />
      </div>
    </div>
  </NModal>
</template>

<style scoped>
.app-modal {
  border: 1px solid #d9e5f2;
  border-radius: 16px;
  background: #ffffff;
  box-shadow: 0 18px 36px rgba(15, 23, 42, 0.18);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.app-modal--width-default {
  width: min(520px, calc(100vw - 32px));
}

.app-modal--width-wide {
  width: min(550px, calc(100vw - 32px));
  max-height: min(90vh, 700px);
}

.app-modal--width-medium {
  width: min(800px, calc(100vw - 32px));
  max-height: min(90vh, 700px);
}

.app-modal--width-large,
.app-modal--width-xlarge {
  max-height: min(96vh, 980px);
}

.app-modal--width-large {
  width: min(880px, calc(100vw - 32px));
}

.app-modal--width-xlarge {
  width: min(1100px, calc(100vw - 32px));
}

.app-modal--width-large .app-modal__body,
.app-modal--width-xlarge .app-modal__body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 20px 24px;
}

.app-modal--width-large .app-modal__actions,
.app-modal--width-xlarge .app-modal__actions {
  padding: 12px 24px 16px;
}

.app-modal--width-large .app-modal__actions--center,
.app-modal--width-xlarge .app-modal__actions--center {
  padding-top: 12px;
  padding-bottom: 16px;
}

.app-modal--height-tall {
  box-sizing: border-box;
  height: min(90vh, 860px);
  max-height: min(90vh, 860px);
}

.app-modal--height-tall .app-modal__body {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.app-modal__header {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 56px;
  padding: 14px 16px;
  border-bottom: 1px solid #e2e8f0;
}

.app-modal__title {
  margin: 0;
  max-width: calc(100% - 240px);
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-modal__header-actions {
  position: absolute;
  top: 50%;
  right: 16px;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
  gap: 8px;
}

.app-modal__close {
  width: 28px;
  height: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #475569;
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    color 0.15s ease,
    background-color 0.15s ease;
}

.app-modal__close:hover {
  border-color: #cbd5e1;
  color: #334155;
  background: #f8fafc;
}

.app-modal__body {
  padding: 16px;
}

.app-modal__body--center {
  padding: 20px 28px 16px;
  text-align: center;
}

.app-modal__body--date {
  width: 100%;
  display: grid;
  place-items: center;
  padding: 24px 32px 19px;
}

.app-modal__actions {
  display: flex;
  padding: 16px;
}

.app-modal--body-default .app-modal__actions,
.app-modal--body-date .app-modal__actions {
  border-top: 1px solid #e2e8f0;
}

.app-modal--width-large .app-modal__actions,
.app-modal--width-xlarge .app-modal__actions {
  border-top: 1px solid #e2e8f0;
}

.app-modal__actions--center {
  justify-content: center;
  padding-top: 16px;
  padding-bottom: 20px;
}

.app-modal__actions--end {
  justify-content: flex-end;
}

.app-modal__actions--reserved {
  visibility: hidden;
  pointer-events: none;
}
</style>

<style>
.app-modal__message {
  margin: 0;
  font-size: 14px;
  line-height: 1.6;
  color: #64748b;
}

.app-modal__textarea {
  width: 100%;
  box-sizing: border-box;
  min-height: 120px;
  padding: 10px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  color: #0f172a;
  font-size: 14px;
  font-family: inherit;
  line-height: 1.5;
  resize: vertical;
}

.app-modal__textarea:focus {
  outline: none;
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.25);
}

.app-modal__date-confirm {
  width: 100%;
  display: flex;
  justify-content: center;
}

.app-modal__date-confirm-btn {
  min-width: 130px;
  border-radius: 10px;
}

.app-modal__date-panel .n-date-panel,
.app-modal__date-panel.n-date-panel {
  width: 292px;
  max-width: 100%;
  margin: 0;
  border: 0;
  box-shadow: none;
}

.app-modal__body--date .n-date-panel-actions {
  justify-content: center;
}

.app-modal__body--date .n-date-panel-actions__prefix {
  display: none;
}

.app-modal__body--date .n-date-panel-actions__suffix {
  width: 100%;
  display: flex;
  justify-content: center;
  align-self: center;
}
</style>
