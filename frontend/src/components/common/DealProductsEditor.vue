<script setup lang="ts">
import { onMounted } from 'vue'
import { NIcon, NSelect } from 'naive-ui'
import type { SelectOption } from 'naive-ui'
import { TrashOutline } from '@vicons/ionicons5'
import ProductsTotalSummary from '@/components/common/ProductsTotalSummary.vue'
import { useProductsCatalog } from '@/composables/useProductsCatalog'
import type { ProductRow } from '@/types/productRow'
import { createEmptyProductRow, normalizeUnitPrice, rowsToDealProducts } from '@/utils/products'

const rows = defineModel<ProductRow[]>({ required: true })

const props = withDefaults(
  defineProps<{
    disabled?: boolean
    addButtonLabel?: string
  }>(),
  {
    disabled: false,
    addButtonLabel: 'Добавить строку',
  },
)

const emit = defineEmits<{
  persist: []
}>()

const { catalogProductOptions, hasCatalogProducts, getCatalogProductById, loadCatalog } = useProductsCatalog()

const productSelectTheme = {
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
      heightMedium: '34px',
      fontSizeMedium: '14px',
      color: '#ffffff',
      colorDisabled: '#f8fafc',
      textColorDisabled: '#94a3b8',
      placeholderColor: '#94a3b8',
    },
  },
}

onMounted(() => {
  void loadCatalog()
})

function addProductRow() {
  rows.value = [...rows.value, createEmptyProductRow()]
  emit('persist')
}

function removeProductRow(index: number) {
  if (props.disabled) return
  const next = [...rows.value]
  next.splice(index, 1)
  rows.value = next
  emit('persist')
}

function filterCatalogProduct(pattern: string, option: SelectOption) {
  const query = pattern.trim().toLowerCase()
  if (!query) return true

  const label = String(option.label ?? '').toLowerCase()
  const sku = String((option as SelectOption & { sku?: string }).sku ?? '').toLowerCase()

  return label.includes(query) || sku.includes(query)
}

function handleProductSelect(index: number, productId: string | null) {
  const next = [...rows.value]
  const row = next[index]
  if (!row) return

  if (!productId) {
    row.catalogProductId = undefined
    row.title = ''
    row.unitPrice = 0
    rows.value = next
    emit('persist')
    return
  }

  const catalogProduct = getCatalogProductById(productId)
  if (!catalogProduct) return

  row.catalogProductId = catalogProduct.id
  row.title = catalogProduct.name
  row.unitPrice = catalogProduct.cost
  rows.value = next
  emit('persist')
}

function handleQuantityChange() {
  emit('persist')
}

function handlePriceBlur(index: number) {
  const next = [...rows.value]
  const row = next[index]
  if (!row) return

  row.unitPrice = normalizeUnitPrice(row.unitPrice)
  rows.value = next
  emit('persist')
}
</script>

<template>
  <div class="deal-products-editor">
    <p v-if="!hasCatalogProducts" class="deal-products-editor__catalog-note">
      Сначала добавьте товары в раздел «Каталог товаров», чтобы выбрать их здесь.
    </p>

    <div class="deal-products-editor__list">
      <div v-if="rows.length > 0" class="deal-products-editor__grid-header">
        <span>Наименование товара/услуги</span>
        <span>Количество</span>
        <span>Стоимость</span>
        <span class="deal-products-editor__actions-spacer" aria-hidden="true" />
      </div>

      <div
        v-for="(product, index) in rows"
        :key="product.rowId"
        class="deal-products-editor__row"
      >
        <NSelect
          :value="product.catalogProductId ?? null"
          :options="catalogProductOptions"
          class="deal-products-editor__select"
          :theme-overrides="productSelectTheme"
          filterable
          clearable
          placeholder="Начните вводить название"
          :disabled="disabled || !hasCatalogProducts"
          :filter="filterCatalogProduct"
          @update:value="(value) => handleProductSelect(index, value)"
        />
        <input
          v-model.number="product.quantity"
          type="number"
          min="1"
          step="1"
          class="deal-products-editor__input deal-products-editor__quantity-input"
          placeholder="1"
          :disabled="disabled"
          @change="handleQuantityChange"
        />
        <input
          v-model.number="product.unitPrice"
          type="number"
          min="0"
          step="1"
          class="deal-products-editor__input deal-products-editor__price-input"
          placeholder="0"
          :disabled="disabled"
          @blur="handlePriceBlur(index)"
        />
        <button
          type="button"
          class="deal-products-editor__icon-action deal-products-editor__icon-action--danger"
          :disabled="disabled"
          aria-label="Удалить строку"
          @click="removeProductRow(index)"
        >
          <NIcon :size="16">
            <TrashOutline />
          </NIcon>
        </button>
      </div>

      <div class="deal-products-editor__total-wrap">
        <ProductsTotalSummary :products="rowsToDealProducts(rows)" />
      </div>
    </div>

    <button
      type="button"
      class="deal-products-editor__add-row-btn"
      :disabled="disabled || !hasCatalogProducts"
      @click="addProductRow"
    >
      {{ addButtonLabel }}
    </button>
  </div>
</template>

<style scoped>
.deal-products-editor {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.deal-products-editor__catalog-note {
  margin: 0;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #f8fafc;
  color: #475569;
  font-size: 13px;
  line-height: 1.45;
}

.deal-products-editor__list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.deal-products-editor__row,
.deal-products-editor__grid-header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 86px 128px 32px;
  gap: 8px;
  align-items: center;
}

.deal-products-editor__grid-header {
  padding: 0 2px;
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
}

.deal-products-editor__actions-spacer {
  width: 32px;
}

.deal-products-editor__select {
  min-width: 0;
}

.deal-products-editor__select :deep(.n-base-selection) {
  min-height: 34px;
}

.deal-products-editor__select :deep(.n-base-selection-label) {
  height: 34px;
}

.deal-products-editor__select :deep(.n-base-selection__border) {
  border: 1px solid #cbd5e1;
  box-shadow: none;
}

.deal-products-editor__select :deep(.n-base-selection--focus .n-base-selection__state-border),
.deal-products-editor__select :deep(.n-base-selection--active .n-base-selection__state-border) {
  border: 1px solid #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.25);
}

.deal-products-editor__input {
  width: 100%;
  height: 34px;
  min-height: 34px;
  box-sizing: border-box;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  color: #0f172a;
  padding: 0 10px;
  font-size: 14px;
  font-family: inherit;
  line-height: 1.3;
}

.deal-products-editor__input:disabled {
  background: #f8fafc;
  color: #94a3b8;
}

.deal-products-editor__input:focus {
  outline: none;
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.25);
}

.deal-products-editor__quantity-input,
.deal-products-editor__price-input {
  text-align: right;
}

.deal-products-editor__quantity-input::-webkit-outer-spin-button,
.deal-products-editor__quantity-input::-webkit-inner-spin-button,
.deal-products-editor__price-input::-webkit-outer-spin-button,
.deal-products-editor__price-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.deal-products-editor__quantity-input[type='number'],
.deal-products-editor__price-input[type='number'] {
  appearance: textfield;
  -moz-appearance: textfield;
}

.deal-products-editor__icon-action {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #64748b;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.deal-products-editor__icon-action:hover:not(:disabled) {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.deal-products-editor__icon-action--danger:hover:not(:disabled) {
  color: #dc2626;
}

.deal-products-editor__icon-action:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.deal-products-editor__total-wrap {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 86px 128px 32px;
  gap: 8px;
}

.deal-products-editor__total-wrap :deep(.products-total-summary) {
  grid-column: 1 / 4;
}

.deal-products-editor__add-row-btn {
  align-self: flex-start;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  color: #334155;
  font-size: 13px;
  font-weight: 500;
  padding: 7px 12px;
  cursor: pointer;
}

.deal-products-editor__add-row-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
</style>
