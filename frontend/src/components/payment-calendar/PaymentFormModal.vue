<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { NDatePicker, NIcon, NSelect } from 'naive-ui'
import { TrashOutline, CheckmarkOutline, PencilOutline } from '@vicons/ionicons5'
import AppModal from '@/components/common/AppModal.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import { PaymentsApiError } from '@/api/payments'
import { usePayments } from '@/composables/usePayments'
import {
  PAYMENT_PAYER_OPTIONS,
  PAYMENT_SHORT_TITLE_MAX_LENGTH,
  type PaymentPayerId,
} from '@/constants/payments'
import type { Payment } from '@/types/payment'

const show = defineModel<boolean>('show', { required: true })

const props = defineProps<{
  initialDate?: number | null
  payment?: Payment | null
}>()

const emit = defineEmits<{
  close: []
}>()

const { payments, addPayment, savePayment, deletePayment, setPaymentClosed } = usePayments()

const paymentDate = ref<number | null>(null)
const remindAt = ref<number | null>(null)
const payerId = ref<PaymentPayerId | null>(null)
const counterparty = ref('')
const amount = ref('')
const shortTitle = ref('')
const comment = ref('')
const isDeleteConfirmOpen = ref(false)
const hasAttemptedSubmit = ref(false)
const isSaving = ref(false)
const isDeleting = ref(false)
const isTogglingClosed = ref(false)
const isEditing = ref(false)
const errorMessage = ref('')
const openedPayment = ref<Payment | null>(null)

const shortTitleLength = computed(() => shortTitle.value.length)
const viewedPayment = computed(() => {
  const opened = openedPayment.value
  if (!opened) return null
  return payments.value.find((item) => item.id === opened.id) ?? opened
})
const isViewMode = computed(() => openedPayment.value != null)
const isFieldsLocked = computed(() => isViewMode.value && !isEditing.value)
const isClosed = computed(() => viewedPayment.value?.isClosed === true)
const isBusy = computed(() => isSaving.value || isDeleting.value || isTogglingClosed.value)
const modalTitle = computed(() => {
  if (!isViewMode.value) return 'Новая оплата'
  return isClosed.value ? 'Оплата проведена' : 'Оплата'
})
const completeButtonLabel = computed(() =>
  isClosed.value ? 'Вернуть оплату в открытые' : 'Отметить оплату как проведённую',
)
const editButtonLabel = computed(() =>
  isEditing.value ? 'Отменить редактирование' : 'Редактировать оплату',
)
const modalCloseLabel = computed(() =>
  isViewMode.value ? 'Закрыть окно оплаты' : 'Закрыть окно добавления оплаты',
)

const datePickerTheme = {
  peers: {
    Input: {
      border: '1px solid #cbd5e1',
      borderHover: '1px solid #cbd5e1',
      borderFocus: '1px solid #93c5fd',
      boxShadowFocus: '0 0 0 3px rgba(147, 197, 253, 0.25)',
      borderRadius: '8px',
      heightMedium: '36px',
      fontSizeMedium: '14px',
    },
  },
}

const selectTheme = {
  peers: {
    InternalSelection: {
      border: '1px solid #cbd5e1',
      borderHover: '1px solid #cbd5e1',
      borderFocus: '1px solid #93c5fd',
      borderActive: '1px solid #93c5fd',
      boxShadowFocus: '0 0 0 3px rgba(147, 197, 253, 0.25)',
      boxShadowActive: '0 0 0 3px rgba(147, 197, 253, 0.25)',
      boxShadowHover: 'none',
      borderRadius: '8px',
      heightMedium: '36px',
      fontSizeMedium: '14px',
      color: '#ffffff',
      placeholderColor: '#94a3b8',
    },
  },
}

const parsedAmount = computed(() => {
  const value = Number(amount.value.replace(/\s/g, ''))
  return Number.isFinite(value) ? value : 0
})

const isShortTitleInvalid = computed(
  () => hasAttemptedSubmit.value && shortTitle.value.trim().length === 0,
)
const isPaymentDateInvalid = computed(() => {
  if (!hasAttemptedSubmit.value) return false
  if (paymentDate.value === null) return true
  return !isAllowedPaymentDate(paymentDate.value)
})
const isRemindAtInvalid = computed(() => {
  if (!hasAttemptedSubmit.value || remindAt.value === null) return false
  return !isAllowedReminderDate(remindAt.value)
})

function startOfDay(timestamp: number) {
  const date = new Date(timestamp)
  date.setHours(0, 0, 0, 0)
  return date.getTime()
}

function todayStart() {
  return startOfDay(Date.now())
}

function isAllowedPaymentDate(timestamp: number) {
  if (startOfDay(timestamp) >= todayStart()) return true
  if (!openedPayment.value) return false
  return startOfDay(timestamp) === startOfDay(openedPayment.value.date)
}

function isAllowedReminderDate(timestamp: number) {
  const day = startOfDay(timestamp)
  if (day < todayStart()) {
    if (!openedPayment.value?.remindAt) return false
    return day === startOfDay(openedPayment.value.remindAt)
  }
  if (paymentDate.value === null) return false
  return day <= startOfDay(paymentDate.value)
}

function isPaymentDateDisabled(timestamp: number) {
  return startOfDay(timestamp) < todayStart()
}

function isReminderDateDisabled(timestamp: number) {
  if (paymentDate.value === null) return true
  const day = startOfDay(timestamp)
  return day < todayStart() || day > startOfDay(paymentDate.value)
}

function fillInitialDate() {
  hasAttemptedSubmit.value = false
  errorMessage.value = ''
  paymentDate.value =
    props.initialDate == null || startOfDay(props.initialDate) < todayStart()
      ? null
      : startOfDay(props.initialDate)
  remindAt.value = null
  payerId.value = null
  counterparty.value = ''
  amount.value = ''
  shortTitle.value = ''
  comment.value = ''
}

function fillFromPayment(payment: Payment) {
  paymentDate.value = startOfDay(payment.date)
  remindAt.value = payment.remindAt
  payerId.value = payment.payerId
  counterparty.value = payment.counterparty
  amount.value = String(payment.amount)
  shortTitle.value = payment.shortTitle
  comment.value = payment.comment
}

function fillForm() {
  openedPayment.value = props.payment ?? null
  isEditing.value = false
  hasAttemptedSubmit.value = false
  errorMessage.value = ''

  if (openedPayment.value) {
    fillFromPayment(openedPayment.value)
    return
  }

  fillInitialDate()
}

watch(show, (isOpen) => {
  if (isOpen) fillForm()
})

watch(paymentDate, (nextDate) => {
  if (remindAt.value === null) return
  if (nextDate === null || startOfDay(remindAt.value) > startOfDay(nextDate)) {
    remindAt.value = null
  }
})

function handleAmountInput(event: Event) {
  const target = event.target as HTMLInputElement | null
  if (!target) return
  const sanitized = target.value.replace(/[^\d]/g, '')
  amount.value = sanitized
  target.value = sanitized
}

function handleShortTitleInput(event: Event) {
  const target = event.target as HTMLInputElement | null
  if (!target) return
  const nextValue = target.value.slice(0, PAYMENT_SHORT_TITLE_MAX_LENGTH)
  shortTitle.value = nextValue
  target.value = nextValue
}

function handleClose() {
  if (isBusy.value) return
  isDeleteConfirmOpen.value = false
  emit('close')
}

function openDeleteConfirm() {
  if (!isViewMode.value || isBusy.value || isEditing.value) return
  errorMessage.value = ''
  isDeleteConfirmOpen.value = true
}

function closeDeleteConfirm() {
  if (isDeleting.value) return
  isDeleteConfirmOpen.value = false
}

function actionError(error: unknown, fallback: string) {
  if (error instanceof PaymentsApiError && error.message) return error.message
  if (error instanceof Error && error.message) return error.message
  return fallback
}

async function confirmDelete() {
  const payment = viewedPayment.value
  if (!payment || isDeleting.value) return

  isDeleting.value = true
  errorMessage.value = ''
  try {
    await deletePayment(payment.id)
    isDeleteConfirmOpen.value = false
    show.value = false
    emit('close')
  } catch (error) {
    errorMessage.value = actionError(error, 'Не удалось удалить оплату')
  } finally {
    isDeleting.value = false
  }
}

async function handleSetClosed() {
  const payment = viewedPayment.value
  if (!payment || isTogglingClosed.value || isEditing.value) return

  const nextClosed = !payment.isClosed
  isTogglingClosed.value = true
  errorMessage.value = ''
  try {
    await setPaymentClosed(payment.id, nextClosed)
    if (nextClosed) {
      show.value = false
      emit('close')
    }
  } catch (error) {
    errorMessage.value = actionError(error, 'Не удалось изменить статус оплаты')
  } finally {
    isTogglingClosed.value = false
  }
}

function handleToggleEdit() {
  if (!isViewMode.value || isBusy.value) return

  if (isEditing.value) {
    const payment = viewedPayment.value
    if (payment) fillFromPayment(payment)
    hasAttemptedSubmit.value = false
    errorMessage.value = ''
    isEditing.value = false
    return
  }

  hasAttemptedSubmit.value = false
  errorMessage.value = ''
  isEditing.value = true
}

function buildPayload() {
  return {
    date: startOfDay(paymentDate.value as number),
    remindAt: remindAt.value === null ? null : startOfDay(remindAt.value),
    payerId: payerId.value,
    counterparty: counterparty.value.trim(),
    amount: parsedAmount.value,
    shortTitle: shortTitle.value.trim(),
    comment: comment.value.trim(),
  }
}

async function handleSubmit() {
  if (isFieldsLocked.value || isSaving.value) return

  hasAttemptedSubmit.value = true
  errorMessage.value = ''

  if (
    paymentDate.value === null ||
    !isAllowedPaymentDate(paymentDate.value) ||
    shortTitle.value.trim().length === 0 ||
    (remindAt.value !== null && !isAllowedReminderDate(remindAt.value))
  ) {
    return
  }

  isSaving.value = true
  try {
    if (isEditing.value) {
      const payment = viewedPayment.value
      if (!payment) return
      const updated = await savePayment(payment.id, buildPayload())
      fillFromPayment(updated)
      isEditing.value = false
      return
    }

    await addPayment(buildPayload())
    show.value = false
    emit('close')
  } catch (error) {
    errorMessage.value = actionError(error, 'Не удалось сохранить оплату')
  } finally {
    isSaving.value = false
  }
}
</script>

<template>
  <AppModal
    v-model:show="show"
    :title="modalTitle"
    width="wide"
    actions-align="center"
    :close-label="modalCloseLabel"
    :trap-focus="false"
    :actions-reserved="isFieldsLocked"
    @close="handleClose"
  >
    <div class="payment-form-modal">
      <div class="payment-form-modal__field">
        <span class="payment-form-modal__label">
          Краткое описание
          <span class="payment-form-modal__required" aria-hidden="true">*</span>
        </span>
        <input
          :value="shortTitle"
          type="text"
          class="payment-form-modal__input"
          :class="{ 'payment-form-modal__input--invalid': isShortTitleInvalid }"
          placeholder="Введите краткое описание"
          autocomplete="off"
          :maxlength="PAYMENT_SHORT_TITLE_MAX_LENGTH"
          :readonly="isFieldsLocked"
          :tabindex="isFieldsLocked ? -1 : undefined"
          :aria-invalid="isShortTitleInvalid"
          aria-label="Краткое описание"
          @input="handleShortTitleInput"
        />
        <div class="payment-form-modal__field-meta">
          <span>Отображается в календаре</span>
          <span>{{ shortTitleLength }} / {{ PAYMENT_SHORT_TITLE_MAX_LENGTH }}</span>
        </div>
      </div>

      <div class="payment-form-modal__row payment-form-modal__row--equal">
        <div class="payment-form-modal__field">
          <span class="payment-form-modal__label">
            Дата платежа
            <span class="payment-form-modal__required" aria-hidden="true">*</span>
          </span>
          <NDatePicker
            v-model:value="paymentDate"
            class="payment-form-modal__date"
            :class="{ 'payment-form-modal__date--invalid': isPaymentDateInvalid }"
            :theme-overrides="datePickerTheme"
            type="date"
            format="dd.MM.yyyy"
            date-format="dd.MM.yyyy"
            placeholder="Выберите дату"
            to="body"
            placement="bottom-start"
            :disabled="isFieldsLocked"
            :is-date-disabled="isPaymentDateDisabled"
            :actions="[]"
          />
        </div>

        <div class="payment-form-modal__field">
          <span class="payment-form-modal__label">Дата напоминания</span>
          <NDatePicker
            v-model:value="remindAt"
            class="payment-form-modal__date"
            :class="{ 'payment-form-modal__date--invalid': isRemindAtInvalid }"
            :theme-overrides="datePickerTheme"
            type="date"
            format="dd.MM.yyyy"
            date-format="dd.MM.yyyy"
            placeholder="Выберите дату"
            to="body"
            placement="bottom-end"
            :disabled="isFieldsLocked || paymentDate === null"
            :is-date-disabled="isReminderDateDisabled"
            clearable
            :actions="[]"
          />
        </div>
      </div>

      <div class="payment-form-modal__field">
        <span class="payment-form-modal__label">Плательщик</span>
        <NSelect
          v-model:value="payerId"
          class="payment-form-modal__select"
          :theme-overrides="selectTheme"
          :options="PAYMENT_PAYER_OPTIONS"
          placeholder="Выберите плательщика"
          to="body"
          clearable
          :disabled="isFieldsLocked"
        />
      </div>

      <div class="payment-form-modal__row payment-form-modal__row--counterparty">
        <div class="payment-form-modal__field">
          <span class="payment-form-modal__label">Контрагент</span>
          <input
            v-model="counterparty"
            type="text"
            class="payment-form-modal__input"
            placeholder="Введите наименование"
            autocomplete="off"
            :readonly="isFieldsLocked"
            :tabindex="isFieldsLocked ? -1 : undefined"
            aria-label="Контрагент"
          />
        </div>

        <div class="payment-form-modal__field">
          <span class="payment-form-modal__label">Сумма оплаты</span>
          <input
            :value="amount"
            type="text"
            inputmode="numeric"
            autocomplete="off"
            class="payment-form-modal__input payment-form-modal__input--amount"
            placeholder="0"
            :readonly="isFieldsLocked"
            :tabindex="isFieldsLocked ? -1 : undefined"
            aria-label="Сумма оплаты"
            @input="handleAmountInput"
          />
        </div>
      </div>

      <div class="payment-form-modal__field">
        <span class="payment-form-modal__label">Комментарий</span>
        <textarea
          v-model="comment"
          class="payment-form-modal__textarea"
          rows="3"
          placeholder="Напишите комментарий"
          :readonly="isFieldsLocked"
          :tabindex="isFieldsLocked ? -1 : undefined"
          aria-label="Комментарий"
        />
      </div>

      <p v-if="errorMessage && !isDeleteConfirmOpen" class="payment-form-modal__error" role="alert">
        {{ errorMessage }}
      </p>
    </div>

    <template v-if="isViewMode" #header-actions>
      <button
        type="button"
        class="payment-form-modal__complete"
        :class="{ 'payment-form-modal__complete--done': isClosed }"
        :aria-label="completeButtonLabel"
        :title="completeButtonLabel"
        :disabled="isBusy || isEditing"
        @click="handleSetClosed"
      >
        <NIcon :size="16">
          <CheckmarkOutline />
        </NIcon>
      </button>
      <button
        type="button"
        class="payment-form-modal__edit"
        :class="{ 'payment-form-modal__edit--active': isEditing }"
        :aria-label="editButtonLabel"
        :title="editButtonLabel"
        :disabled="isBusy"
        @click="handleToggleEdit"
      >
        <NIcon :size="16">
          <PencilOutline />
        </NIcon>
      </button>
      <button
        type="button"
        class="payment-form-modal__delete"
        aria-label="Удалить оплату"
        title="Удалить"
        :disabled="isBusy || isEditing"
        @click="openDeleteConfirm"
      >
        <NIcon :size="16">
          <TrashOutline />
        </NIcon>
      </button>
    </template>

    <template #actions>
      <AppModalButton
        :disabled="isSaving || isFieldsLocked"
        :tabindex="isFieldsLocked ? -1 : undefined"
        @click="handleSubmit"
      >
        {{ isSaving ? 'Сохранение…' : isEditing ? 'Сохранить' : 'Добавить' }}
      </AppModalButton>
    </template>
  </AppModal>

  <AppModal
    v-model:show="isDeleteConfirmOpen"
    title="Удаление оплаты"
    body-variant="center"
    :z-index="2100"
    close-label="Закрыть окно удаления оплаты"
    @close="closeDeleteConfirm"
  >
    <p class="app-modal__message">Вы уверены, что хотите удалить данную оплату?</p>
    <p v-if="errorMessage" class="payment-form-modal__error" role="alert">
      {{ errorMessage }}
    </p>

    <template #actions>
      <div class="payment-form-modal__confirm-actions">
        <AppModalButton :disabled="isDeleting" @click="confirmDelete">
          {{ isDeleting ? 'Удаление…' : 'Да' }}
        </AppModalButton>
        <button
          type="button"
          class="payment-form-modal__confirm-cancel"
          :disabled="isDeleting"
          @click="closeDeleteConfirm"
        >
          Нет
        </button>
      </div>
    </template>
  </AppModal>
</template>

<style scoped>
.payment-form-modal {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.payment-form-modal__row {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(0, 1fr);
  gap: 12px 16px;
}

.payment-form-modal__row--equal {
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
}

.payment-form-modal__row--counterparty {
  grid-template-columns: minmax(0, 1fr) 132px;
}

.payment-form-modal__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.payment-form-modal__label {
  font-size: 13px;
  color: #475569;
}

.payment-form-modal__required {
  color: #dc2626;
  margin-left: 2px;
}

.payment-form-modal__date,
.payment-form-modal__select {
  width: 100%;
}

.payment-form-modal__date :deep(.n-input) {
  height: 36px;
}

.payment-form-modal__date--invalid :deep(.n-input) {
  --n-border: 1px solid #dc2626;
  --n-border-hover: 1px solid #dc2626;
  --n-border-focus: 1px solid #dc2626;
  --n-box-shadow-focus: 0 0 0 3px rgba(220, 38, 38, 0.18);
}

.payment-form-modal__date--invalid :deep(.n-input__border),
.payment-form-modal__date--invalid :deep(.n-input__state-border) {
  border: 1px solid #dc2626 !important;
  box-shadow: none !important;
  transition: none !important;
}

.payment-form-modal__date--invalid :deep(.n-input--focus .n-input__state-border) {
  box-shadow: 0 0 0 3px rgba(220, 38, 38, 0.18) !important;
}

.payment-form-modal__select :deep(.n-base-selection) {
  min-height: 36px;
}

.payment-form-modal__select :deep(.n-base-selection-label) {
  height: 36px;
}

.payment-form-modal__input,
.payment-form-modal__textarea {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  color: #0f172a;
  font-size: 14px;
  font-family: inherit;
}

.payment-form-modal__input {
  height: 36px;
  padding: 0 12px;
}

.payment-form-modal__input--amount {
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.payment-form-modal__field-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 12px;
  line-height: 1.4;
  color: #94a3b8;
}

.payment-form-modal__textarea {
  min-height: 72px;
  padding: 8px 12px;
  line-height: 1.5;
  resize: none;
}

.payment-form-modal__input:focus,
.payment-form-modal__textarea:focus {
  outline: none;
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.25);
}

.payment-form-modal__input--invalid {
  border-color: #dc2626;
  transition: none;
}

.payment-form-modal__input--invalid:focus {
  border-color: #dc2626;
  box-shadow: 0 0 0 3px rgba(220, 38, 38, 0.18);
}

.payment-form-modal__input:read-only,
.payment-form-modal__textarea:read-only {
  background: #f8fafc;
  color: #0f172a;
  cursor: default;
}

.payment-form-modal__input:read-only:focus,
.payment-form-modal__textarea:read-only:focus {
  border-color: #cbd5e1;
  box-shadow: none;
}

.payment-form-modal__error {
  margin: 0;
  padding: 10px 12px;
  border-radius: 8px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #b91c1c;
  font-size: 13px;
  line-height: 1.4;
}

.payment-form-modal__complete {
  width: 28px;
  height: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #64748b;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    color 0.15s ease,
    background-color 0.15s ease;
}

.payment-form-modal__complete:hover:not(:disabled) {
  border-color: #1f883d;
  color: #1f883d;
  background: #f0fdf4;
}

.payment-form-modal__complete--done {
  border-color: #1f883d;
  background: #1f883d;
  color: #ffffff;
}

.payment-form-modal__complete--done:hover:not(:disabled) {
  border-color: #1a7534;
  background: #1a7534;
  color: #ffffff;
}

.payment-form-modal__edit {
  width: 28px;
  height: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #64748b;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    color 0.15s ease,
    background-color 0.15s ease;
}

.payment-form-modal__edit:hover:not(:disabled) {
  border-color: #1f883d;
  color: #1f883d;
  background: #f0fdf4;
}

.payment-form-modal__edit--active {
  border-color: #1f883d;
  background: #1f883d;
  color: #ffffff;
}

.payment-form-modal__edit--active:hover:not(:disabled) {
  border-color: #1a7534;
  background: #1a7534;
  color: #ffffff;
}

.payment-form-modal__delete {
  width: 28px;
  height: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #64748b;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    color 0.15s ease,
    background-color 0.15s ease;
}

.payment-form-modal__delete:hover:not(:disabled) {
  border-color: #cbd5e1;
  color: #dc2626;
  background: #f8fafc;
}

.payment-form-modal__complete:disabled,
.payment-form-modal__edit:disabled,
.payment-form-modal__delete:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.payment-form-modal__confirm-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px;
}

.payment-form-modal__confirm-cancel {
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

.payment-form-modal__confirm-cancel:hover:not(:disabled) {
  background: #f8fafc;
  border-color: #94a3b8;
  color: #334155;
}

.payment-form-modal__confirm-cancel:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

@media (max-width: 560px) {
  .payment-form-modal__row,
  .payment-form-modal__row--equal {
    grid-template-columns: 1fr;
  }
}
</style>
