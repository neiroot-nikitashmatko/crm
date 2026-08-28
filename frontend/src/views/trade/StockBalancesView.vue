<script setup lang="ts">
import { onMounted, reactive } from 'vue'
import SectionSubviewHeader from '@/components/common/SectionSubviewHeader.vue'
import { useStockBalances } from '@/composables/useStockBalances'

const UNCATEGORIZED_KEY = '__uncategorized__'
const { balances, categoryGroups, uncategorizedItems, isLoading, loadStockData } = useStockBalances()
const expandedCategories = reactive<Record<string, boolean>>({})

onMounted(() => {
  void loadStockData()
})

function toggleCategory(category: string) {
  expandedCategories[category] = !expandedCategories[category]
}

function isCategoryExpanded(category: string) {
  return expandedCategories[category] === true
}

function formatQuantity(quantity: number) {
  return new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 3 }).format(quantity)
}
</script>

<template>
  <div class="trade-subview">
    <SectionSubviewHeader title="Остатки на складе" />

    <div class="trade-subview__body">
      <section v-if="isLoading && balances.length === 0" class="stock-balances-view__placeholder">
        <p class="stock-balances-view__placeholder-text">Загрузка…</p>
      </section>

      <section v-else-if="balances.length === 0" class="stock-balances-view__placeholder">
        <p class="stock-balances-view__placeholder-text">
          Пока нет остатков на складе. Они появятся после добавления приходной накладной.
        </p>
      </section>

      <section v-else class="stock-balances-view__content">
        <article
          v-for="group in categoryGroups"
          :key="group.category"
          class="stock-balances-view__category"
        >
          <button
            type="button"
            class="stock-balances-view__category-header"
            @click="toggleCategory(group.category)"
          >
            <span class="stock-balances-view__category-icon">
              {{ isCategoryExpanded(group.category) ? '📂' : '📁' }}
            </span>
            <span class="stock-balances-view__category-name">{{ group.category }}</span>
            <span class="stock-balances-view__category-count">{{ group.items.length }}</span>
          </button>

          <div v-if="isCategoryExpanded(group.category)" class="stock-balances-view__table" role="table">
            <div class="stock-balances-view__table-head" role="row">
              <span role="columnheader">Наименование</span>
              <span role="columnheader">Артикул</span>
              <span role="columnheader">Количество</span>
            </div>
            <div
              v-for="item in group.items"
              :key="item.key"
              class="stock-balances-view__table-row"
              role="row"
            >
              <span>{{ item.title }}</span>
              <span>{{ item.sku || '—' }}</span>
              <span
                class="stock-balances-view__qty"
                :class="{ 'stock-balances-view__qty--negative': item.quantity < 0 }"
              >
                {{ formatQuantity(item.quantity) }}
              </span>
            </div>
          </div>
        </article>

        <article
          v-if="uncategorizedItems.length > 0"
          class="stock-balances-view__category"
        >
          <button
            type="button"
            class="stock-balances-view__category-header"
            @click="toggleCategory(UNCATEGORIZED_KEY)"
          >
            <span class="stock-balances-view__category-icon">
              {{ isCategoryExpanded(UNCATEGORIZED_KEY) ? '📂' : '📁' }}
            </span>
            <span class="stock-balances-view__category-name">Товары без категории</span>
            <span class="stock-balances-view__category-count">{{ uncategorizedItems.length }}</span>
          </button>

          <div v-if="isCategoryExpanded(UNCATEGORIZED_KEY)" class="stock-balances-view__table" role="table">
            <div class="stock-balances-view__table-head" role="row">
              <span role="columnheader">Наименование</span>
              <span role="columnheader">Артикул</span>
              <span role="columnheader">Количество</span>
            </div>
            <div
              v-for="item in uncategorizedItems"
              :key="item.key"
              class="stock-balances-view__table-row"
              role="row"
            >
              <span>{{ item.title }}</span>
              <span>{{ item.sku || '—' }}</span>
              <span
                class="stock-balances-view__qty"
                :class="{ 'stock-balances-view__qty--negative': item.quantity < 0 }"
              >
                {{ formatQuantity(item.quantity) }}
              </span>
            </div>
          </div>
        </article>
      </section>
    </div>
  </div>
</template>

<style scoped>
.trade-subview {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.trade-subview__body {
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow: auto;
  padding: 24px;
  box-sizing: border-box;
  scrollbar-gutter: stable;
}

.stock-balances-view__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 240px;
  padding: 32px 24px;
  border: 1px dashed #cbd5e1;
  border-radius: 12px;
  background: #f8fafc;
}

.stock-balances-view__placeholder-text {
  margin: 0;
  max-width: 420px;
  font-size: 15px;
  line-height: 1.5;
  color: #64748b;
  text-align: center;
}

.stock-balances-view__content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.stock-balances-view__category {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #ffffff;
  overflow: hidden;
}

.stock-balances-view__category-header {
  width: 100%;
  border: 0;
  background: #f8fafc;
  padding: 10px 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  text-align: left;
  cursor: pointer;
}

.stock-balances-view__category-icon {
  width: 18px;
}

.stock-balances-view__category-name {
  font-size: 14px;
  font-weight: 600;
  color: #1a202c;
}

.stock-balances-view__category-count {
  margin-left: auto;
  min-width: 24px;
  padding: 0 6px;
  border-radius: 999px;
  background: #e2e8f0;
  color: #4a5568;
  font-size: 12px;
  text-align: center;
}

.stock-balances-view__table {
  display: flex;
  flex-direction: column;
  border-top: 1px solid #e2e8f0;
}

.stock-balances-view__table-head,
.stock-balances-view__table-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 160px 120px;
  gap: 0;
  align-items: stretch;
}

.stock-balances-view__table-head > *,
.stock-balances-view__table-row > * {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  min-height: 52px;
  padding: 10px 12px;
  border-right: 1px solid #e2e8f0;
  border-bottom: 1px solid #e2e8f0;
}

.stock-balances-view__table-head > * {
  background: #f8fafc;
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
}

.stock-balances-view__table-row {
  font-size: 14px;
  color: #1a202c;
}

.stock-balances-view__table-head > *:last-child,
.stock-balances-view__table-row > *:last-child {
  border-right: 0;
}

.stock-balances-view__table-row:last-child > * {
  border-bottom: 0;
}

.stock-balances-view__qty {
  justify-content: flex-end;
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.stock-balances-view__qty--negative {
  color: #dc2626;
}
</style>
