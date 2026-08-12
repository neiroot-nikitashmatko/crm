<script setup lang="ts">
import { computed, nextTick, reactive, ref } from 'vue'
import { NSelect } from 'naive-ui'
import type { NewLeadForm } from '@/types/lead'
import { MANUAL_TRAFFIC_SOURCE_OPTIONS } from '@/constants/trafficSources'

const PHONE_PREFIX = '+7'

/** Выравниваем селект с соседними input и делаем меню спокойнее / современнее. */
const trafficSourceSelectTheme = {
  peers: {
    InternalSelection: {
      heightMedium: '38px',
      fontSizeMedium: '14px',
      borderRadius: '6px',
      border: '1px solid #e0e0e0',
      borderHover: '1px solid #d0d0d0',
      borderFocus: '1px solid #a8c4e8',
      borderActive: '1px solid #a8c4e8',
      boxShadowFocus: 'none',
      boxShadowActive: 'none',
      boxShadowHover: 'none',
      color: '#ffffff',
      colorActive: '#ffffff',
      textColor: '#1a202c',
      placeholderColor: '#a0aec0',
      caretColor: '#4a5568',
      arrowColor: '#718096',
      paddingSingle: '0 34px 0 10px',
    },
    InternalSelectMenu: {
      borderRadius: '8px',
      color: '#ffffff',
      boxShadow: '0 10px 28px rgba(26, 32, 44, 0.12), 0 2px 8px rgba(26, 32, 44, 0.06)',
      paddingMedium: '6px',
      optionFontSizeMedium: '14px',
      optionHeightMedium: '36px',
      optionPaddingMedium: '0 12px',
      optionTextColor: '#1a202c',
      optionTextColorActive: '#2c5282',
      optionTextColorPressed: '#2c5282',
      optionColorPending: '#f0f6fd',
      optionColorActive: '#e8f1fc',
      optionColorActivePending: '#e8f1fc',
    },
  },
}

withDefaults(defineProps<{
  showTrigger?: boolean
}>(), {
  showTrigger: true,
})

const emit = defineEmits<{
  save: [payload: NewLeadForm]
  layoutChange: []
}>()

const isFormOpen = ref(false)

const form = reactive<NewLeadForm>({
  firstName: '',
  patronymic: '',
  phone: PHONE_PREFIX,
  trafficSource: '',
})

const canSave = computed(
  () =>
    form.firstName.trim().length > 0 &&
    form.trafficSource.trim().length > 0,
)

async function openForm() {
  isFormOpen.value = true
  await nextTick()
  emit('layoutChange')
}

function resetForm() {
  form.firstName = ''
  form.patronymic = ''
  form.phone = PHONE_PREFIX
  form.trafficSource = ''
}

function normalizePhone(rawValue: string): string {
  const digitsOnly = rawValue.replace(/\D/g, '')
  let localPart = digitsOnly

  if (localPart.startsWith('7') || localPart.startsWith('8')) {
    localPart = localPart.slice(1)
  }

  return `${PHONE_PREFIX}${localPart}`
}

function handlePhoneInput(event: Event) {
  const target = event.target as HTMLInputElement | null
  if (!target) return

  const normalized = normalizePhone(target.value)
  form.phone = normalized
  target.value = normalized
}

function cancel() {
  isFormOpen.value = false
  resetForm()
  emit('layoutChange')
}

function save() {
  if (!canSave.value) return
  emit('save', {
    ...form,
    firstName: form.firstName.trim(),
    patronymic: form.patronymic.trim(),
    phone: form.phone.trim(),
    trafficSource: form.trafficSource.trim(),
  })
  cancel()
}

defineExpose({
  openForm,
})
</script>

<template>
  <div class="leads-add-lead-panel">
    <button
      v-if="showTrigger"
      type="button"
      class="leads-add-lead-panel__add-btn"
      @click="openForm"
    >
      Добавить лид
    </button>

    <form v-if="isFormOpen" class="leads-add-lead-panel__form" @submit.prevent="save">
      <label class="leads-add-lead-panel__field">
        <span class="leads-add-lead-panel__label">Имя</span>
        <input
          v-model="form.firstName"
          type="text"
          class="leads-add-lead-panel__input"
          autocomplete="off"
        />
      </label>

      <label class="leads-add-lead-panel__field">
        <span class="leads-add-lead-panel__label">Отчество</span>
        <input
          v-model="form.patronymic"
          type="text"
          class="leads-add-lead-panel__input"
          autocomplete="off"
        />
      </label>

      <label class="leads-add-lead-panel__field">
        <span class="leads-add-lead-panel__label">Телефон</span>
        <input
          :value="form.phone"
          type="tel"
          class="leads-add-lead-panel__input"
          autocomplete="off"
          inputmode="tel"
          @input="handlePhoneInput"
        />
      </label>

      <label class="leads-add-lead-panel__field">
        <span class="leads-add-lead-panel__label">Источник трафика</span>
        <NSelect
          :value="form.trafficSource || null"
          class="leads-add-lead-panel__select"
          :theme-overrides="trafficSourceSelectTheme"
          :options="MANUAL_TRAFFIC_SOURCE_OPTIONS"
          placeholder="Выберите источник"
          :clearable="false"
          :show-checkmark="false"
          :menu-props="{ class: 'leads-add-lead-panel__select-menu' }"
          @update:value="(value: string | null) => { form.trafficSource = value ?? '' }"
        />
      </label>

      <div class="leads-add-lead-panel__actions">
        <button
          type="submit"
          class="leads-add-lead-panel__btn leads-add-lead-panel__btn--primary"
          :disabled="!canSave"
        >
          Сохранить
        </button>
        <button
          type="button"
          class="leads-add-lead-panel__btn leads-add-lead-panel__btn--secondary"
          @click="cancel"
        >
          Отменить
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.leads-add-lead-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}

.leads-add-lead-panel__add-btn {
  display: block;
  width: 100%;
  padding: 7px 12px;
  border: 1px dashed #c5d9f0;
  border-radius: 6px;
  background: #ffffff;
  font-size: 14px;
  font-weight: 500;
  color: #4a5568;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
  -webkit-tap-highlight-color: transparent;
}

.leads-add-lead-panel__add-btn:hover {
  background: #f0f6fd;
  border-color: #a8c4e8;
  color: #2c5282;
}

.leads-add-lead-panel__add-btn:active {
  background: #e8f1fc;
}

.leads-add-lead-panel__form {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}

.leads-add-lead-panel__field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.leads-add-lead-panel__label {
  font-size: 12px;
  font-weight: 500;
  color: #4a5568;
}

.leads-add-lead-panel__input {
  width: 100%;
  height: 38px;
  padding: 0 10px;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  background: #ffffff;
  font-size: 14px;
  line-height: 1.4;
  color: #1a202c;
  font-family: inherit;
  outline: none;
  transition: border-color 0.15s ease;
  box-sizing: border-box;
}

.leads-add-lead-panel__input:focus {
  border-color: #a8c4e8;
}

.leads-add-lead-panel__input::placeholder {
  color: #a0aec0;
}

.leads-add-lead-panel__select {
  width: 100%;
}

.leads-add-lead-panel__select :deep(.n-base-selection) {
  --n-height: 38px;
  height: 38px;
  min-height: 38px;
  border-radius: 6px;
}

.leads-add-lead-panel__select :deep(.n-base-selection-label),
.leads-add-lead-panel__select :deep(.n-base-selection-placeholder),
.leads-add-lead-panel__select :deep(.n-base-selection-input) {
  height: 100%;
  min-height: 100%;
  display: flex;
  align-items: center;
  box-sizing: border-box;
}

.leads-add-lead-panel__select :deep(.n-base-selection-placeholder) {
  padding-left: 10px;
  padding-right: 34px;
  color: #a0aec0;
}

.leads-add-lead-panel__select :deep(.n-base-selection-overlay) {
  inset: 0;
  display: flex;
  align-items: center;
  padding: 0 34px 0 10px;
}

.leads-add-lead-panel__select :deep(.n-base-selection__border),
.leads-add-lead-panel__select :deep(.n-base-selection__state-border) {
  border-radius: 6px;
}

.leads-add-lead-panel__select :deep(.n-base-suffix) {
  right: 10px;
}

.leads-add-lead-panel__select :deep(.n-base-loading),
.leads-add-lead-panel__select :deep(.n-base-clear),
.leads-add-lead-panel__select :deep(.n-base-suffix__arrow) {
  transition: color 0.15s ease, transform 0.15s ease;
}

.leads-add-lead-panel__select :deep(.n-base-selection--active .n-base-suffix__arrow) {
  transform: rotate(180deg);
  color: #4a5568;
}

.leads-add-lead-panel__actions {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 2px;
}

.leads-add-lead-panel__btn {
  width: 100%;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease,
    opacity 0.15s ease;
  -webkit-tap-highlight-color: transparent;
}

.leads-add-lead-panel__btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.leads-add-lead-panel__btn--primary {
  border: 1px solid #4a5568;
  background: #4a5568;
  color: #ffffff;
}

.leads-add-lead-panel__btn--primary:hover:not(:disabled) {
  background: #2d3748;
  border-color: #2d3748;
}

.leads-add-lead-panel__btn--secondary {
  border: 1px solid #e0e0e0;
  background: #ffffff;
  color: #4a5568;
}

.leads-add-lead-panel__btn--secondary:hover {
  background: #f7f7f7;
  border-color: #d0d0d0;
}
</style>

<style>
/* Меню телепортируется в body — scoped-стили на него не действуют. */
.leads-add-lead-panel__select-menu.n-base-select-menu {
  overflow: hidden;
}

.leads-add-lead-panel__select-menu .n-base-select-option {
  border-radius: 6px;
  margin: 1px 0;
  transition: background-color 0.12s ease, color 0.12s ease;
}

.leads-add-lead-panel__select-menu .n-base-select-option .n-base-select-option__content {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.leads-add-lead-panel__select-menu .n-base-select-menu__empty {
  padding: 12px;
  font-size: 13px;
  color: #718096;
}
</style>
