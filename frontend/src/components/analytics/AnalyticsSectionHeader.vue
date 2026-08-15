<script setup lang="ts">
import { NDatePicker } from 'naive-ui'
import { useAnalyticsPeriodContext } from '@/composables/useAnalyticsPeriod'

const { selectedPreset, periodStart, periodEnd, selectToday, selectMonth } =
  useAnalyticsPeriodContext()
</script>

<template>
  <header class="analytics-section-header">
    <h1 class="analytics-section-header__title">Аналитика</h1>

    <div class="analytics-section-header__period-controls" aria-label="Выбор периода аналитики">
      <div class="analytics-section-header__quick-actions">
        <button
          type="button"
          class="analytics-section-header__quick-action"
          :class="{ 'analytics-section-header__quick-action--active': selectedPreset === 'today' }"
          @click="selectToday"
        >
          За сегодня
        </button>
        <button
          type="button"
          class="analytics-section-header__quick-action"
          :class="{ 'analytics-section-header__quick-action--active': selectedPreset === 'month' }"
          @click="selectMonth"
        >
          За текущий месяц
        </button>
      </div>

      <div class="analytics-section-header__custom-period">
        <label class="analytics-section-header__date-field">
          <span class="analytics-section-header__date-label">С</span>
          <NDatePicker
            v-model:value="periodStart"
            class="analytics-section-header__date-picker"
            type="date"
            size="small"
            format="dd.MM.yyyy"
          />
        </label>
        <label class="analytics-section-header__date-field">
          <span class="analytics-section-header__date-label">По</span>
          <NDatePicker
            v-model:value="periodEnd"
            class="analytics-section-header__date-picker"
            type="date"
            size="small"
            format="dd.MM.yyyy"
          />
        </label>
      </div>
    </div>
  </header>
</template>

<style scoped>
.analytics-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  min-height: 56px;
  padding: 12px 24px;
  border-bottom: 1px solid #e2e8f0;
  background: #ffffff;
  box-sizing: border-box;
}

.analytics-section-header__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #1a202c;
  letter-spacing: -0.02em;
}

.analytics-section-header__period-controls,
.analytics-section-header__quick-actions,
.analytics-section-header__custom-period {
  display: flex;
  align-items: center;
}

.analytics-section-header__period-controls {
  gap: 10px;
}

.analytics-section-header__quick-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  width: 262px;
  padding: 3px;
  gap: 2px;
  border: 1px solid #d1d9e2;
  border-radius: 9px;
  background: #f8fafc;
  box-sizing: border-box;
}

.analytics-section-header__custom-period {
  gap: 10px;
}

.analytics-section-header__quick-action {
  min-width: 0;
  height: 28px;
  padding: 0 6px;
  border: 1px solid transparent;
  border-radius: 6px;
  background: transparent;
  color: #4a5568;
  font-size: 12px;
  font-weight: 500;
  line-height: 1;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    background-color 0.15s ease,
    color 0.15s ease;
}

.analytics-section-header__quick-action:hover {
  background: #ffffff;
  color: #1a202c;
}

.analytics-section-header__quick-action--active {
  border-color: #b7dfc2;
  background: #ffffff;
  color: #14532d;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.08);
}

.analytics-section-header__date-field {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #718096;
  font-size: 12px;
  white-space: nowrap;
}

.analytics-section-header__date-label {
  flex-shrink: 0;
  color: #718096;
  font-size: 12px;
}

.analytics-section-header__date-picker {
  width: auto;
}

.analytics-section-header__date-picker :deep(.n-input),
.analytics-section-header__date-picker :deep(.n-input-wrapper) {
  width: auto;
}

.analytics-section-header__date-picker :deep(.n-input__input-el) {
  width: 9.2ch;
  min-width: 9.2ch;
  max-width: 9.2ch;
  font-variant-numeric: tabular-nums;
}

@media (max-width: 860px) {
  .analytics-section-header {
    align-items: flex-start;
    flex-direction: column;
    gap: 10px;
  }

  .analytics-section-header__period-controls {
    width: 100%;
    justify-content: space-between;
  }
}

@media (max-width: 600px) {
  .analytics-section-header__period-controls {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
