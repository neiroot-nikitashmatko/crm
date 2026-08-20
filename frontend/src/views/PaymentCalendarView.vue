<script setup lang="ts">
import { onMounted, ref } from 'vue'
import PaymentCalendarSectionHeader from '@/components/payment-calendar/PaymentCalendarSectionHeader.vue'
import PaymentCalendar from '@/components/payment-calendar/PaymentCalendar.vue'
import PaymentFormModal from '@/components/payment-calendar/PaymentFormModal.vue'
import { PaymentsApiError } from '@/api/payments'
import { usePayments } from '@/composables/usePayments'
import type { PaymentCalendarViewMode } from '@/constants/paymentCalendar'
import type { Payment } from '@/types/payment'

const { loadPayments } = usePayments()
const viewMode = ref<PaymentCalendarViewMode>('month')
const selectedDate = ref(new Date())
const isFormModalOpen = ref(false)
const formInitialDate = ref<number | null>(null)
const selectedPayment = ref<Payment | null>(null)
const loadErrorMessage = ref('')

onMounted(async () => {
  loadErrorMessage.value = ''
  try {
    await loadPayments(true)
  } catch (error) {
    if (error instanceof PaymentsApiError && error.message) {
      loadErrorMessage.value = error.message
      return
    }
    loadErrorMessage.value = 'Не удалось загрузить оплаты'
  }
})

function startOfDay(value: Date | number) {
  const date = new Date(value)
  date.setHours(0, 0, 0, 0)
  return date.getTime()
}

function handleShiftPeriod(direction: -1 | 1) {
  const nextDate = new Date(selectedDate.value)

  if (viewMode.value === 'week') {
    nextDate.setDate(nextDate.getDate() + direction * 7)
  } else {
    nextDate.setMonth(nextDate.getMonth() + direction)
  }

  selectedDate.value = nextDate
}

function handleGoToday() {
  selectedDate.value = new Date()
}

function openCreateModal(date?: Date) {
  selectedPayment.value = null
  const dateValue = date ? startOfDay(date) : null
  formInitialDate.value = dateValue != null && dateValue >= startOfDay(new Date()) ? dateValue : null
  isFormModalOpen.value = true
}

function openPaymentModal(payment: Payment) {
  selectedPayment.value = payment
  isFormModalOpen.value = true
}

function handleFormClose() {
  selectedPayment.value = null
}
</script>

<template>
  <div class="payment-calendar-page">
    <PaymentCalendarSectionHeader
      v-model:view-mode="viewMode"
      @shift-period="handleShiftPeriod"
      @go-today="handleGoToday"
      @create="openCreateModal()"
    />

    <p v-if="loadErrorMessage" class="payment-calendar-page__error" role="alert">
      {{ loadErrorMessage }}
    </p>

    <div class="payment-calendar-page__body">
      <PaymentCalendar
        :view-mode="viewMode"
        :selected-date="selectedDate"
        @create="openCreateModal"
        @open="openPaymentModal"
      />
    </div>

    <PaymentFormModal
      v-model:show="isFormModalOpen"
      :initial-date="formInitialDate"
      :payment="selectedPayment"
      @close="handleFormClose"
    />
  </div>
</template>

<style scoped>
.payment-calendar-page {
  display: flex;
  flex-direction: column;
  height: calc(100dvh - 64px);
  max-height: calc(100dvh - 64px);
  overflow: hidden;
  background: #ffffff;
}

.payment-calendar-page__error {
  margin: 0 24px;
  padding: 10px 12px;
  border-radius: 8px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #b91c1c;
  font-size: 13px;
}

.payment-calendar-page__body {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 0;
  padding: 12px 24px 16px;
  box-sizing: border-box;
  overflow: hidden;
}
</style>
