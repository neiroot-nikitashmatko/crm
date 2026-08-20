<script setup lang="ts">
import { computed } from 'vue'
import { getPaymentPayerLabel } from '@/constants/payments'
import type { PaymentCalendarViewMode } from '@/constants/paymentCalendar'
import { usePayments } from '@/composables/usePayments'
import type { Payment } from '@/types/payment'
import { formatMoney } from '@/utils/money'
import {
  buildCurrentWeekGrid,
  buildMonthGrid,
  formatMonthTitle,
  formatWeekTitle,
  getDateKey,
  getWeekdayLabels,
  splitIntoWeeks,
} from '@/utils/calendar'

const props = defineProps<{
  viewMode: PaymentCalendarViewMode
  selectedDate: Date
}>()

const emit = defineEmits<{
  create: [date: Date]
  open: [payment: Payment]
}>()

const { payments } = usePayments()
const today = new Date()
const weekdayLabels = getWeekdayLabels()

const weeks = computed(() => {
  if (props.viewMode === 'week') {
    return [buildCurrentWeekGrid(props.selectedDate, today)]
  }

  return splitIntoWeeks(
    buildMonthGrid(props.selectedDate.getFullYear(), props.selectedDate.getMonth(), today),
  )
})

const calendarTitle = computed(() => {
  if (props.viewMode === 'week') {
    return formatWeekTitle(weeks.value[0] ?? [])
  }

  return formatMonthTitle(props.selectedDate.getFullYear(), props.selectedDate.getMonth())
})

const paymentsByDay = computed(() => {
  const grouped = new Map<string, Payment[]>()

  for (const payment of payments.value) {
    const key = getDateKey(new Date(payment.date))
    const dayPayments = grouped.get(key) ?? []
    dayPayments.push(payment)
    grouped.set(key, dayPayments)
  }

  return grouped
})

function getDayPayments(date: Date) {
  return paymentsByDay.value.get(getDateKey(date)) ?? []
}

function isWeekend(date: Date) {
  const weekday = date.getDay()
  return weekday === 0 || weekday === 6
}

function formatDayLabel(date: Date) {
  return new Intl.DateTimeFormat('ru-RU', {
    day: 'numeric',
    month: 'long',
  }).format(date)
}

function getPaymentTitle(payment: Payment) {
  const parts = [
    payment.shortTitle,
    payment.amount > 0 ? formatMoney(payment.amount) : '',
    payment.counterparty,
    getPaymentPayerLabel(payment.payerId),
  ].filter((part) => part.length > 0)
  if (payment.comment) parts.push(payment.comment)
  if (payment.isClosed) parts.unshift('Проведена')
  return parts.join(' · ')
}

function isPastDay(date: Date) {
  const day = new Date(date)
  day.setHours(0, 0, 0, 0)
  const todayStart = new Date()
  todayStart.setHours(0, 0, 0, 0)
  return day.getTime() < todayStart.getTime()
}

function handleDayClick(date: Date) {
  if (isPastDay(date)) return
  emit('create', date)
}

function handlePaymentClick(payment: Payment) {
  emit('open', payment)
}
</script>

<template>
  <section
    class="payment-calendar"
    :class="{ 'payment-calendar--week': viewMode === 'week' }"
  >
    <header class="payment-calendar__toolbar">
      <h2 class="payment-calendar__month-title">{{ calendarTitle }}</h2>
    </header>

    <div class="payment-calendar__weekdays" aria-hidden="true">
      <span
        v-for="(weekday, index) in weekdayLabels"
        :key="weekday"
        class="payment-calendar__weekday"
        :class="{ 'payment-calendar__weekday--weekend': index >= 5 }"
      >
        {{ weekday }}
      </span>
    </div>

    <div class="payment-calendar__grid" :style="{ '--week-count': weeks.length }">
      <div
        v-for="(week, weekIndex) in weeks"
        :key="`week-${weekIndex}`"
        class="payment-calendar__week"
      >
        <article
          v-for="day in week"
          :key="getDateKey(day.date)"
          class="payment-calendar__day"
          :class="{
            'payment-calendar__day--outside': !day.isCurrentMonth,
            'payment-calendar__day--weekend': isWeekend(day.date),
            'payment-calendar__day--today': day.isToday,
          }"
          role="button"
          tabindex="0"
          :aria-label="`Добавить оплату на ${formatDayLabel(day.date)}`"
          @click="handleDayClick(day.date)"
          @keydown.enter.prevent="handleDayClick(day.date)"
          @keydown.space.prevent="handleDayClick(day.date)"
        >
          <span class="payment-calendar__day-number">{{ day.dayNumber }}</span>
          <div class="payment-calendar__day-content">
            <div
              v-for="payment in getDayPayments(day.date)"
              :key="payment.id"
              class="payment-calendar__entry"
              :class="{ 'payment-calendar__entry--closed': payment.isClosed }"
              :title="getPaymentTitle(payment)"
              role="button"
              tabindex="0"
              :aria-label="`Открыть оплату ${payment.shortTitle}`"
              @click.stop="handlePaymentClick(payment)"
              @keydown.enter.prevent.stop="handlePaymentClick(payment)"
              @keydown.space.prevent.stop="handlePaymentClick(payment)"
            >
              {{ payment.shortTitle }}
            </div>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>

<style scoped>
.payment-calendar {
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  height: 100%;
  min-height: 0;
  --day-inset-x: 8px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
  overflow: hidden;
}

.payment-calendar__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  padding: 6px 16px;
  border-bottom: 1px solid #e2e8f0;
  background: #f8fafc;
}

.payment-calendar__month-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1a202c;
  letter-spacing: -0.02em;
  line-height: 1.25;
}

.payment-calendar__weekdays {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  flex-shrink: 0;
  border-bottom: 1px solid #e2e8f0;
  background: #f8fafc;
}

.payment-calendar__weekday {
  padding: 6px 12px;
  font-size: 11px;
  font-weight: 600;
  line-height: 1.2;
  color: #64748b;
  text-align: center;
  border-right: 1px solid #e2e8f0;
}

.payment-calendar__weekday:last-child {
  border-right: 0;
}

.payment-calendar__weekday--weekend {
  color: #94a3b8;
}

.payment-calendar__grid {
  display: grid;
  grid-template-rows: repeat(var(--week-count), minmax(0, 1fr));
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
}

.payment-calendar__week {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  min-height: 0;
}

.payment-calendar__day {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  min-width: 0;
  min-height: 0;
  padding: clamp(6px, 1.2vh, 10px) var(--day-inset-x);
  border-right: 1px solid #e2e8f0;
  border-bottom: 1px solid #e2e8f0;
  background: #ffffff;
  overflow: hidden;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.payment-calendar__day:hover:not(.payment-calendar__day--today) {
  background: #f8fafc;
}

.payment-calendar__day:focus-visible {
  outline: 2px solid rgba(31, 136, 61, 0.45);
  outline-offset: -2px;
}

.payment-calendar__day:last-child {
  border-right: 0;
}

.payment-calendar__week:last-child .payment-calendar__day {
  border-bottom: 0;
}

.payment-calendar__day--outside {
  background: #f8fafc;
}

.payment-calendar__day--weekend:not(.payment-calendar__day--outside):not(.payment-calendar__day--today) {
  background: #fafbfc;
}

.payment-calendar__day--today {
  background: #eff6ff;
  box-shadow: inset 0 0 0 2px #93c5fd;
}

.payment-calendar__day-number {
  flex-shrink: 0;
  font-size: clamp(13px, 1.6vh, 15px);
  font-weight: 600;
  color: #1a202c;
  line-height: 1.2;
}

.payment-calendar__day--outside .payment-calendar__day-number {
  color: #94a3b8;
  font-weight: 500;
}

.payment-calendar__day--weekend:not(.payment-calendar__day--today) .payment-calendar__day-number {
  color: #94a3b8;
}

.payment-calendar__day--today .payment-calendar__day-number {
  color: #1d4ed8;
}

.payment-calendar__day-content {
  flex: 1 1 auto;
  width: 100%;
  min-height: 0;
  margin-top: 4px;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 3px;
  overflow-y: auto;
}

.payment-calendar--week .payment-calendar__day {
  padding: 12px var(--day-inset-x);
}

.payment-calendar--week .payment-calendar__day-number {
  font-size: 16px;
}

.payment-calendar--week .payment-calendar__day-content {
  gap: 6px;
  margin-top: 8px;
}

.payment-calendar__entry {
  width: 100%;
  padding: 4px 6px;
  border-radius: 4px;
  background: #e8eef4;
  color: #334155;
  font-size: clamp(10px, 1.2vh, 12px);
  line-height: 1.2;
  min-width: 0;
  box-sizing: border-box;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    color 0.15s ease;
}

.payment-calendar__entry:hover {
  background: #dbe4ee;
  color: #1a202c;
}

.payment-calendar__entry:focus-visible {
  outline: 2px solid rgba(31, 136, 61, 0.45);
  outline-offset: 1px;
}

.payment-calendar__entry--closed {
  text-decoration: line-through;
}

.payment-calendar--week .payment-calendar__entry {
  padding: 6px 8px;
  font-size: 13px;
  border-radius: 6px;
}
</style>
