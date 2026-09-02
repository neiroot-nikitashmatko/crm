<script setup lang="ts">
import { computed } from 'vue'
import AppModal from '@/components/common/AppModal.vue'
import { formatMoney } from '@/utils/money'
import type { TradeProfitItem } from '@/types/analytics'

const show = defineModel<boolean>('show', { required: true })

const props = defineProps<{
  items: readonly TradeProfitItem[]
  loading?: boolean
  errorMessage?: string
}>()

const totalProfit = computed(() =>
  props.items.reduce((total, item) => total + item.profit, 0),
)

function formatQuantity(quantity: number): string {
  const rounded = Math.round(quantity * 1000) / 1000
  const value = Number.isInteger(rounded)
    ? String(rounded)
    : rounded.toLocaleString('ru-RU', { maximumFractionDigits: 3 })
  return `${value} шт`
}

function formatPrices(item: TradeProfitItem): string {
  const cost = item.hasCost ? formatMoney(item.costPrice) : '—'
  return `закупка ${cost} / продажа ${formatMoney(item.salePrice)}`
}

function profitClass(amount: number): string {
  if (amount < 0) return 'analytics-trade-profit-modal__profit--negative'
  if (amount > 0) return 'analytics-trade-profit-modal__profit--positive'
  return ''
}
</script>

<template>
  <AppModal v-model:show="show" title="Из чего сложилась прибыль" width="medium">
    <p v-if="errorMessage" class="analytics-trade-profit-modal__error">{{ errorMessage }}</p>
    <p v-else-if="loading" class="analytics-trade-profit-modal__empty">Загрузка…</p>
    <p v-else-if="items.length === 0" class="analytics-trade-profit-modal__empty">
      Нет продаж за выбранный период
    </p>
    <template v-else>
      <ul class="analytics-trade-profit-modal__list">
        <li v-for="item in items" :key="`${item.productKey}:${item.hasCost}:${item.costPrice}`">
          <article class="analytics-trade-profit-modal__item">
            <div class="analytics-trade-profit-modal__content">
              <p class="analytics-trade-profit-modal__title">{{ item.title }}</p>
              <p class="analytics-trade-profit-modal__meta">
                {{ formatQuantity(item.quantity) }} · {{ formatPrices(item) }}
              </p>
            </div>
            <p class="analytics-trade-profit-modal__profit" :class="profitClass(item.profit)">
              {{ formatMoney(item.profit) }}
            </p>
          </article>
        </li>
      </ul>
      <div class="analytics-trade-profit-modal__total">
        <span>Итого</span>
        <span :class="profitClass(totalProfit)">{{ formatMoney(totalProfit) }}</span>
      </div>
    </template>
  </AppModal>
</template>

<style scoped>
.analytics-trade-profit-modal__error,
.analytics-trade-profit-modal__empty {
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  color: #64748b;
}

.analytics-trade-profit-modal__error {
  color: #cf222e;
}

.analytics-trade-profit-modal__list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: min(70vh, 560px);
  margin: 0;
  padding: 0;
  overflow: auto;
  list-style: none;
}

.analytics-trade-profit-modal__item {
  display: flex;
  align-items: center;
  gap: 16px;
  min-width: 0;
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #ffffff;
}

.analytics-trade-profit-modal__content {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: 3px;
}

.analytics-trade-profit-modal__title {
  margin: 0;
  min-width: 0;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.35;
  color: #1a202c;
  overflow-wrap: break-word;
  word-break: break-word;
}

.analytics-trade-profit-modal__profit {
  flex-shrink: 0;
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  line-height: 1.2;
  color: #1a202c;
  font-variant-numeric: tabular-nums;
}

.analytics-trade-profit-modal__profit--positive {
  color: #1f883d;
}

.analytics-trade-profit-modal__profit--negative {
  color: #dc2626;
}

.analytics-trade-profit-modal__meta {
  margin: 0;
  font-size: 12px;
  line-height: 1.3;
  color: #718096;
  font-variant-numeric: tabular-nums;
}

.analytics-trade-profit-modal__total {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 12px;
  padding: 12px 12px 0;
  border-top: 1px solid #e2e8f0;
  font-size: 14px;
  font-weight: 700;
  line-height: 1.3;
  color: #1a202c;
  font-variant-numeric: tabular-nums;
}
</style>
