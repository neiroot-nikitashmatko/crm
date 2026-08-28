<script setup lang="ts">
import { computed } from 'vue'
import { formatMoney } from '@/utils/money'

const props = withDefaults(
  defineProps<{
    title: string
    amount: number | null
    hint?: string
    loading?: boolean
    format?: 'money' | 'percent'
  }>(),
  {
    format: 'money',
  },
)

const displayAmount = computed(() => {
  if (props.amount === null) return '—'
  if (props.format === 'percent') return formatPercent(props.amount)
  return formatMoney(props.amount)
})

function formatPercent(value: number): string {
  const rounded = Math.round(value * 10) / 10
  if (Number.isInteger(rounded)) return `${rounded}%`
  return `${rounded.toLocaleString('ru-RU', {
    minimumFractionDigits: 1,
    maximumFractionDigits: 1,
  })}%`
}

const amountClass = computed(() => {
  if (props.amount === null) return ''
  if (props.amount < 0) return 'analytics-amount-card__value--negative'
  if (props.amount > 0) return 'analytics-amount-card__value--positive'
  return ''
})
</script>

<template>
  <section class="analytics-amount-card">
    <header class="analytics-amount-card__header">
      <h2 class="analytics-amount-card__title">{{ title }}</h2>
      <div v-if="$slots['header-actions']" class="analytics-amount-card__header-actions">
        <slot name="header-actions" />
      </div>
    </header>

    <div
      class="analytics-amount-card__body"
      :class="{ 'analytics-amount-card__body--loading': loading }"
    >
      <p class="analytics-amount-card__value" :class="amountClass">{{ displayAmount }}</p>
      <p v-if="hint" class="analytics-amount-card__hint">{{ hint }}</p>
    </div>
  </section>
</template>

<style scoped>
.analytics-amount-card {
  display: flex;
  min-width: 0;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #f6f8fa;
  box-sizing: border-box;
}

.analytics-amount-card__header {
  position: relative;
  display: flex;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid #e2e8f0;
}

.analytics-amount-card__header:has(.analytics-amount-card__header-actions)
  .analytics-amount-card__title {
  padding-right: 42px;
}

.analytics-amount-card__title {
  margin: 0;
  min-width: 0;
  flex: 1;
  font-size: 15px;
  font-weight: 700;
  line-height: 1.3;
  color: #1a202c;
}

.analytics-amount-card__header-actions {
  position: absolute;
  top: 50%;
  right: 20px;
  display: flex;
  flex-shrink: 0;
  align-items: center;
  transform: translateY(-50%);
}

.analytics-amount-card__body {
  display: flex;
  flex: 1;
  min-height: 148px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 16px 20px 20px;
  box-sizing: border-box;
}

.analytics-amount-card__body--loading {
  opacity: 0.55;
}

.analytics-amount-card__value {
  margin: 0;
  font-size: 28px;
  font-weight: 700;
  line-height: 1.15;
  letter-spacing: -0.04em;
  color: #1a202c;
  font-variant-numeric: tabular-nums;
  text-align: center;
}

.analytics-amount-card__value--positive {
  color: #1f883d;
}

.analytics-amount-card__value--negative {
  color: #dc2626;
}

.analytics-amount-card__hint {
  margin: 8px 0 0;
  font-size: 11px;
  font-weight: 500;
  line-height: 1.35;
  color: #718096;
  text-align: center;
}
</style>
