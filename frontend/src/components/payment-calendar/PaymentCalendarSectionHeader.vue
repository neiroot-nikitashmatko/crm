<script setup lang="ts">
import { computed } from 'vue'
import { NIcon } from 'naive-ui'
import { AddOutline, ChevronBackOutline, ChevronForwardOutline } from '@vicons/ionicons5'
import {
  PAYMENT_CALENDAR_VIEW_OPTIONS,
  type PaymentCalendarViewMode,
} from '@/constants/paymentCalendar'

const viewMode = defineModel<PaymentCalendarViewMode>('viewMode', { required: true })

const emit = defineEmits<{
  'shift-period': [direction: -1 | 1]
  'go-today': []
  create: []
}>()

const previousPeriodTitle = computed(() =>
  viewMode.value === 'week' ? 'Предыдущая неделя' : 'Предыдущий месяц',
)

const nextPeriodTitle = computed(() =>
  viewMode.value === 'week' ? 'Следующая неделя' : 'Следующий месяц',
)
</script>

<template>
  <header class="payment-calendar-section-header">
    <h1 class="payment-calendar-section-header__title">Календарь оплат</h1>

    <div class="payment-calendar-section-header__right">
      <button
        type="button"
        class="payment-calendar-section-header__today-btn"
        title="Показать текущий период"
        @click="emit('go-today')"
      >
        Сегодня
      </button>

      <div
        class="payment-calendar-section-header__period-nav"
        role="group"
        aria-label="Переключение периода"
      >
        <button
          type="button"
          class="payment-calendar-section-header__period-btn"
          :title="previousPeriodTitle"
          :aria-label="previousPeriodTitle"
          @click="emit('shift-period', -1)"
        >
          <NIcon :size="18">
            <ChevronBackOutline />
          </NIcon>
        </button>

        <button
          type="button"
          class="payment-calendar-section-header__period-btn"
          :title="nextPeriodTitle"
          :aria-label="nextPeriodTitle"
          @click="emit('shift-period', 1)"
        >
          <NIcon :size="18">
            <ChevronForwardOutline />
          </NIcon>
        </button>
      </div>

      <div
        class="payment-calendar-section-header__view-switch"
        role="group"
        aria-label="Вид календаря"
      >
        <button
          v-for="option in PAYMENT_CALENDAR_VIEW_OPTIONS"
          :key="option.value"
          type="button"
          class="payment-calendar-section-header__view-btn"
          :class="{ 'payment-calendar-section-header__view-btn--active': viewMode === option.value }"
          :aria-pressed="viewMode === option.value"
          @click="viewMode = option.value"
        >
          {{ option.label }}
        </button>
      </div>

      <span class="payment-calendar-section-header__divider" aria-hidden="true" />

      <button
        type="button"
        class="payment-calendar-section-header__create-btn"
        title="Добавить оплату"
        aria-label="Добавить оплату"
        @click="emit('create')"
      >
        <NIcon :size="18">
          <AddOutline />
        </NIcon>
      </button>
    </div>
  </header>
</template>

<style scoped>
.payment-calendar-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  height: 56px;
  padding: 0 16px 0 24px;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
}

.payment-calendar-section-header__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #1a202c;
  letter-spacing: -0.02em;
}

.payment-calendar-section-header__right {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.payment-calendar-section-header__divider {
  flex-shrink: 0;
  align-self: center;
  width: 1px;
  height: 18px;
  margin: 0 2px;
  background: #d1d9e2;
}

.payment-calendar-section-header__create-btn {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #475569;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.payment-calendar-section-header__create-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #1f2937;
}

.payment-calendar-section-header__today-btn {
  height: 34px;
  padding: 0 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #ffffff;
  color: #334155;
  font-size: 13px;
  font-weight: 600;
  line-height: 1;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.payment-calendar-section-header__today-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #1a202c;
}

.payment-calendar-section-header__period-nav {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.payment-calendar-section-header__period-btn {
  width: 34px;
  height: 34px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #ffffff;
  color: #334155;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.payment-calendar-section-header__period-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #1a202c;
}

.payment-calendar-section-header__view-switch {
  display: inline-flex;
  align-items: center;
  padding: 3px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #f8fafc;
}

.payment-calendar-section-header__view-btn {
  padding: 6px 12px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: #64748b;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.2;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    color 0.15s ease,
    box-shadow 0.15s ease;
}

.payment-calendar-section-header__view-btn:hover {
  color: #334155;
}

.payment-calendar-section-header__view-btn--active {
  background: #ffffff;
  color: #1a202c;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.08);
}
</style>
