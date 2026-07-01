<script setup lang="ts">
import { computed } from 'vue'
import type { DealProduct } from '@/types/deal'
import { formatMoney } from '@/utils/money'
import { calculateProductsTotal } from '@/utils/products'

const props = defineProps<{
  products: DealProduct[]
}>()

const total = computed(() => calculateProductsTotal(props.products))
</script>

<template>
  <div class="products-total-summary">
    <span class="products-total-summary__label">Итого</span>
    <span class="products-total-summary__value">{{ formatMoney(total) }}</span>
  </div>
</template>

<style scoped>
.products-total-summary {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 86px 128px;
  gap: 8px;
  padding-top: 12px;
  margin-top: 2px;
  border-top: 1px solid #e2e8f0;
}

.products-total-summary__label {
  grid-column: 1 / 3;
  padding-right: 4px;
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
  text-align: right;
  align-self: center;
}

.products-total-summary__value {
  font-size: 15px;
  font-weight: 700;
  color: #1a202c;
  text-align: right;
  font-variant-numeric: tabular-nums;
}
</style>
