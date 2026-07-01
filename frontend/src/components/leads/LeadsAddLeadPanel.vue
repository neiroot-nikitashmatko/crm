<script setup lang="ts">
import { nextTick, reactive, ref } from 'vue'
import type { NewLeadForm } from '@/types/lead'

const PHONE_PREFIX = '+7'

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
  emit('save', { ...form })
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
        <input
          v-model="form.trafficSource"
          type="text"
          class="leads-add-lead-panel__input"
          autocomplete="off"
        />
      </label>

      <div class="leads-add-lead-panel__actions">
        <button type="submit" class="leads-add-lead-panel__btn leads-add-lead-panel__btn--primary">
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
  padding: 8px 10px;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  background: #ffffff;
  font-size: 14px;
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
    color 0.15s ease;
  -webkit-tap-highlight-color: transparent;
}

.leads-add-lead-panel__btn--primary {
  border: 1px solid #4a5568;
  background: #4a5568;
  color: #ffffff;
}

.leads-add-lead-panel__btn--primary:hover {
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
