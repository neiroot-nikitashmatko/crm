<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { TrashOutline } from '@vicons/ionicons5'
import { NDatePicker, NIcon, NSelect } from 'naive-ui'
import type { SelectOption } from 'naive-ui'
import AppModal from '@/components/common/AppModal.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import { useDeals } from '@/composables/useDeals'
import { useOutgoingInvoices } from '@/composables/useOutgoingInvoices'
import { useProductsCatalog } from '@/composables/useProductsCatalog'
import { useStockBalances } from '@/composables/useStockBalances'
import type { ProductRow } from '@/types/productRow'
import { formatMoney } from '@/utils/money'
import {
  createEmptyProductRow,
  ensureRowIds,
  normalizeUnitPrice,
  productsToRows,
} from '@/utils/products'
import { stockMovementKey } from '@/utils/stockBalances'
import {
  renderSalaryDealOption,
  salaryDealOptionFullLabel,
  salaryDealOptionNumberLabel,
} from '@/utils/salaryDealLabel'

const props = defineProps<{
  open: boolean
  lockedDealId?: string | null
  editingInvoiceId?: string | null
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const { deals, loadDeals } = useDeals()
const { invoices, addInvoice, updateInvoice, hasInvoiceForDeal, loadInvoices } = useOutgoingInvoices()
const { catalogProductOptions, hasCatalogProducts, getCatalogProductById, loadCatalog, products } =
  useProductsCatalog()
const { getOutgoingInvoiceStockIssueKeys, loadStockData } = useStockBalances()

const invoiceDate = ref<number | null>(null)
const dealId = ref<string | null>(null)
const dealSearchQuery = ref('')
const isDealsLoading = ref(false)
const invoiceComment = ref('')
const rows = ref<ProductRow[]>([createEmptyProductRow()])
const submitError = ref('')

const datePickerTheme = {
  peers: {
    Input: {
      border: '1px solid #cbd5e1',
      borderHover: '1px solid #cbd5e1',
      borderFocus: '1px solid #93c5fd',
      boxShadowFocus: '0 0 0 3px rgba(147, 197, 253, 0.25)',
      borderRadius: '8px',
      heightMedium: '32px',
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
      heightMedium: '32px',
      fontSizeMedium: '14px',
      color: '#ffffff',
      placeholderColor: '#94a3b8',
    },
  },
}

const isDealLocked = computed(() => Boolean(props.lockedDealId))
const isEditMode = computed(() => Boolean(props.editingInvoiceId))

const modalTitle = computed(() =>
  isEditMode.value ? 'Редактировать расходную накладную' : 'Новая расходная накладная',
)
const submitButtonLabel = computed(() => (isEditMode.value ? 'Сохранить' : 'Добавить'))
const modalCloseLabel = computed(() =>
  isEditMode.value
    ? 'Закрыть окно редактирования расходной накладной'
    : 'Закрыть окно добавления расходной накладной',
)

function isClosedDeal(deal: { columnId?: string; status?: string }) {
  return deal.columnId === 'closed' || String(deal.status ?? '').toLowerCase() === 'closed'
}

function isDealAvailableForSelection(id: string) {
  if (isEditMode.value && dealId.value === id) return true
  return !hasInvoiceForDeal(id)
}

const closedDealOptions = computed(() => {
  const query = dealSearchQuery.value.trim().toLowerCase()
  const selectedId = dealId.value

  return deals.value
    .filter((deal) => isClosedDeal(deal) || (selectedId != null && deal.id === selectedId))
    .filter((deal) => isDealAvailableForSelection(deal.id) || (selectedId != null && deal.id === selectedId))
    .filter((deal) => {
      if (selectedId && deal.id === selectedId) return true
      if (!query) return true
      return String(deal.dealNumber).includes(query)
    })
    .slice()
    .sort((left, right) => Number(right.dealNumber) - Number(left.dealNumber))
    .map((deal) => ({
      label: salaryDealOptionNumberLabel(deal.dealNumber),
      fullLabel: salaryDealOptionFullLabel(deal),
      value: deal.id,
    }))
})

const hasClosedDeals = computed(() =>
  deals.value.some((deal) => isClosedDeal(deal) && isDealAvailableForSelection(deal.id)),
)

const filledRows = computed(() =>
  rows.value.filter((row) => Boolean(row.catalogProductId) && Number(row.quantity) > 0),
)

const invoiceTotal = computed(() =>
  filledRows.value.reduce((sum, row) => sum + getRowSum(row), 0),
)

const canSubmitInvoice = computed(
  () =>
    invoiceDate.value !== null &&
    Boolean(dealId.value) &&
    filledRows.value.length > 0,
)

const stockIssueKeys = computed(() => {
  const items = rows.value
    .filter((row) => Boolean(row.catalogProductId) && Number(row.quantity) > 0)
    .map((row) => ({
      catalogProductId: row.catalogProductId,
      title: row.title,
      quantity: Number(row.quantity),
    }))

  return getOutgoingInvoiceStockIssueKeys(items, props.editingInvoiceId ?? null)
})

const lockedDealLabel = computed(() => {
  if (!props.lockedDealId) return ''
  const deal = getDealById(props.lockedDealId)
  if (!deal) return 'Сделка удалена'
  return salaryDealOptionNumberLabel(deal.dealNumber)
})

const isModalVisible = computed({
  get: () => props.open,
  set: (value: boolean) => {
    if (!value) handleClose()
  },
})

watch(
  () => props.open,
  (isOpen) => {
    if (!isOpen) return
    submitError.value = ''
    void loadCatalog()
    void loadInvoices()
    void loadStockData()
    void refreshDeals()
    hydrateForm()
  },
)

watch(
  () => products.value.length,
  (length, previousLength) => {
    if (!props.open || props.editingInvoiceId || length === 0) return
    if (previousLength === 0 && length > 0) {
      const targetDealId = props.lockedDealId ?? dealId.value
      if (targetDealId) {
        applyDealProductsToRows(targetDealId)
      }
    }
  },
)

async function refreshDeals() {
  isDealsLoading.value = true
  try {
    await loadDeals(true)
  } finally {
    isDealsLoading.value = false
  }
}

function getRowSum(row: ProductRow) {
  const quantity = Number(row.quantity)
  const safeQuantity = Number.isFinite(quantity) && quantity > 0 ? quantity : 0
  return safeQuantity * normalizeUnitPrice(row.unitPrice)
}

function getDealById(id: string) {
  return deals.value.find((item) => item.id === id) ?? null
}

function resetInvoiceForm() {
  invoiceDate.value = Date.now()
  dealId.value = null
  dealSearchQuery.value = ''
  invoiceComment.value = ''
  rows.value = [createEmptyProductRow()]
  submitError.value = ''
}

function hydrateForm() {
  submitError.value = ''

  if (props.editingInvoiceId) {
    const invoice = invoices.value.find((item) => item.id === props.editingInvoiceId)
    if (!invoice) {
      resetInvoiceForm()
      return
    }

    invoiceDate.value = invoice.date
    dealId.value = invoice.dealId
    dealSearchQuery.value = ''
    invoiceComment.value = invoice.comment ?? ''
    rows.value =
      invoice.items.length > 0 ? productsToRows(invoice.items, products.value) : [createEmptyProductRow()]
    return
  }

  resetInvoiceForm()

  if (props.lockedDealId) {
    dealId.value = props.lockedDealId
    applyDealProductsToRows(props.lockedDealId)
  }
}

function applyDealProductsToRows(selectedDealId: string | null) {
  if (!selectedDealId) {
    rows.value = [createEmptyProductRow()]
    return
  }

  const deal = getDealById(selectedDealId)
  if (!deal) {
    rows.value = [createEmptyProductRow()]
    return
  }

  const dealProducts = deal.products.filter((product) => product.title.trim().length > 0)
  if (dealProducts.length === 0) {
    rows.value = [createEmptyProductRow()]
    return
  }

  rows.value = ensureRowIds(productsToRows(dealProducts, products.value))
}

function handleDealSearch(query: string) {
  dealSearchQuery.value = query
}

function handleDealSelect(value: string | null) {
  dealId.value = value
  submitError.value = ''
  applyDealProductsToRows(value)
}

function keepAllFilteredOptions() {
  return true
}

function addProductRow() {
  rows.value = [...rows.value, createEmptyProductRow()]
}

function removeProductRow(index: number) {
  const next = [...rows.value]
  next.splice(index, 1)
  rows.value = next.length > 0 ? next : [createEmptyProductRow()]
}

function filterCatalogProduct(pattern: string, option: SelectOption) {
  const query = pattern.trim().toLowerCase()
  if (!query) return true

  const label = String(option.label ?? '').toLowerCase()
  const sku = String((option as SelectOption & { sku?: string }).sku ?? '').toLowerCase()

  return label.includes(query) || sku.includes(query)
}

function handleProductSelect(index: number, productId: string | null) {
  submitError.value = ''
  const next = [...rows.value]
  const row = next[index]
  if (!row) return

  if (!productId) {
    row.catalogProductId = undefined
    row.title = ''
    row.unitPrice = 0
    rows.value = next
    return
  }

  const catalogProduct = getCatalogProductById(productId)
  if (!catalogProduct) return

  row.catalogProductId = catalogProduct.id
  row.title = catalogProduct.name
  row.unitPrice = catalogProduct.cost
  rows.value = next
}

function handleQuantityInput(index: number, event: Event) {
  submitError.value = ''
  const target = event.target as HTMLInputElement | null
  const row = rows.value[index]
  if (!target || !row) return

  const parsed = Number(target.value.replace(/[^\d]/g, ''))
  row.quantity = Number.isFinite(parsed) && parsed > 0 ? parsed : 1
}

function handleCostInput(index: number, event: Event) {
  const target = event.target as HTMLInputElement | null
  const row = rows.value[index]
  if (!target || !row) return

  row.unitPrice = normalizeUnitPrice(target.value.replace(/[^\d.]/g, ''))
}

function isRowStockInvalid(row: ProductRow) {
  if (!row.catalogProductId || Number(row.quantity) <= 0) return false

  return stockIssueKeys.value.has(
    stockMovementKey({
      catalogProductId: row.catalogProductId,
      title: row.title,
    }),
  )
}

function handleClose() {
  resetInvoiceForm()
  emit('close')
}

async function handleSubmitInvoice() {
  if (!canSubmitInvoice.value || invoiceDate.value === null || !dealId.value) return

  if (!isEditMode.value && hasInvoiceForDeal(dealId.value)) {
    submitError.value = 'По этой сделке уже есть расходная накладная'
    return
  }

  const payload = {
    date: invoiceDate.value,
    dealId: dealId.value,
    items: filledRows.value.map((row) => ({
      catalogProductId: row.catalogProductId,
      title: row.title,
      quantity: Number(row.quantity),
      unitPrice: normalizeUnitPrice(row.unitPrice),
    })),
    total: invoiceTotal.value,
    comment: invoiceComment.value.trim(),
  }

  const stockError = stockIssueKeys.value.size > 0
  if (stockError) {
    return
  }

  submitError.value = ''
  try {
    if (props.editingInvoiceId) {
      await updateInvoice(props.editingInvoiceId, payload)
    } else {
      await addInvoice(payload)
    }
    resetInvoiceForm()
    emit('saved')
  } catch (error) {
    submitError.value =
      error instanceof Error && error.message.trim()
        ? error.message
        : 'Не удалось сохранить расходную накладную'
  }
}
</script>

<template>
  <AppModal
    v-model:show="isModalVisible"
    :title="modalTitle"
    width="large"
    height="tall"
    actions-align="center"
    :close-label="modalCloseLabel"
    @close="handleClose"
  >
    <div class="outgoing-invoice-modal">
      <section class="outgoing-invoice-modal__header-fields">
        <label class="outgoing-invoice-modal__field">
          <span class="outgoing-invoice-modal__label">
            Дата
            <span class="outgoing-invoice-modal__required" aria-hidden="true">*</span>
          </span>
          <NDatePicker
            v-model:value="invoiceDate"
            class="outgoing-invoice-modal__date"
            :theme-overrides="datePickerTheme"
            type="date"
            format="dd.MM.yyyy"
            date-format="dd.MM.yyyy"
            placeholder="Выберите дату"
            :actions="[]"
          />
        </label>

        <div class="outgoing-invoice-modal__field">
          <span class="outgoing-invoice-modal__label">
            Сделка
            <span class="outgoing-invoice-modal__required" aria-hidden="true">*</span>
          </span>
          <p v-if="isDealLocked" class="outgoing-invoice-modal__deal-readonly">
            {{ lockedDealLabel }}
          </p>
          <NSelect
            v-else
            :value="dealId"
            class="outgoing-invoice-modal__select"
            :theme-overrides="selectTheme"
            :options="closedDealOptions"
            :loading="isDealsLoading"
            :render-option="renderSalaryDealOption"
            filterable
            clearable
            placeholder="Начните вводить номер сделки"
            :filter="keepAllFilteredOptions"
            @search="handleDealSearch"
            @update:value="handleDealSelect"
          />
        </div>
      </section>

      <p
        v-if="submitError"
        class="outgoing-invoice-modal__note outgoing-invoice-modal__note--error"
      >
        {{ submitError }}
      </p>

      <p
        v-if="!isDealLocked && !isDealsLoading && !hasClosedDeals"
        class="outgoing-invoice-modal__note"
      >
        Нет закрытых сделок для выбора
      </p>
      <p v-if="!hasCatalogProducts" class="outgoing-invoice-modal__note">
        Сначала добавьте товары в раздел «Товары и услуги», чтобы выбрать их здесь.
      </p>

      <section class="outgoing-invoice-modal__positions">
        <div class="outgoing-invoice-modal__table">
          <div class="outgoing-invoice-modal__grid-header">
            <span>Товар / услуга</span>
            <span>Кол-во</span>
            <span>Стоимость</span>
            <span>Сумма</span>
            <span aria-hidden="true" />
          </div>

          <div class="outgoing-invoice-modal__rows">
            <div
              v-for="(product, index) in rows"
              :key="product.rowId"
              class="outgoing-invoice-modal__row"
            >
              <div class="outgoing-invoice-modal__product-cell">
                <NSelect
                  :value="product.catalogProductId ?? null"
                  :options="catalogProductOptions"
                  class="outgoing-invoice-modal__select"
                  :theme-overrides="selectTheme"
                  filterable
                  clearable
                  placeholder="Выберите товар"
                  :disabled="!hasCatalogProducts"
                  :filter="filterCatalogProduct"
                  @update:value="(value) => handleProductSelect(index, value)"
                />
                <p
                  class="outgoing-invoice-modal__stock-hint"
                  :class="{ 'outgoing-invoice-modal__stock-hint--visible': isRowStockInvalid(product) }"
                >
                  Нет на складе
                </p>
              </div>
              <input
                :value="product.quantity"
                type="text"
                inputmode="numeric"
                autocomplete="off"
                class="outgoing-invoice-modal__input outgoing-invoice-modal__input--numeric"
                placeholder="1"
                @input="handleQuantityInput(index, $event)"
              />
              <input
                :value="product.unitPrice"
                type="text"
                inputmode="decimal"
                autocomplete="off"
                class="outgoing-invoice-modal__input outgoing-invoice-modal__input--numeric"
                placeholder="0"
                @input="handleCostInput(index, $event)"
              />
              <span class="outgoing-invoice-modal__sum">{{ formatMoney(getRowSum(product)) }}</span>
              <button
                type="button"
                class="outgoing-invoice-modal__icon-action"
                aria-label="Удалить строку"
                :disabled="rows.length === 1"
                @click="removeProductRow(index)"
              >
                <NIcon :size="16">
                  <TrashOutline />
                </NIcon>
              </button>
            </div>
          </div>

          <div class="outgoing-invoice-modal__table-footer">
            <button
              type="button"
              class="outgoing-invoice-modal__add-row-btn"
              :disabled="!hasCatalogProducts"
              @click="addProductRow"
            >
              + Добавить позицию
            </button>

            <div class="outgoing-invoice-modal__total">
              <span class="outgoing-invoice-modal__total-label">Итого</span>
              <span class="outgoing-invoice-modal__total-value">{{ formatMoney(invoiceTotal) }}</span>
            </div>
          </div>
        </div>
      </section>

      <label class="outgoing-invoice-modal__field outgoing-invoice-modal__field--comment">
        <span class="outgoing-invoice-modal__label">Комментарий</span>
        <textarea
          v-model="invoiceComment"
          class="outgoing-invoice-modal__textarea"
          rows="2"
          placeholder="Комментарий (необязательно)"
        />
      </label>
    </div>

    <template #actions>
      <AppModalButton
        :disabled="!canSubmitInvoice"
        :title="canSubmitInvoice ? '' : 'Заполните дату, сделку и хотя бы одну позицию'"
        @click="handleSubmitInvoice"
      >
        {{ submitButtonLabel }}
      </AppModalButton>
    </template>
  </AppModal>
</template>

<style scoped>
.outgoing-invoice-modal {
  display: flex;
  flex-direction: column;
  gap: 16px;
  flex: 1;
  min-height: 0;
  height: 100%;
}

.outgoing-invoice-modal__header-fields {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px 16px;
  flex-shrink: 0;
}

.outgoing-invoice-modal__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.outgoing-invoice-modal__label {
  font-size: 13px;
  color: #475569;
}

.outgoing-invoice-modal__required {
  color: #dc2626;
  margin-left: 2px;
}

.outgoing-invoice-modal__deal-readonly {
  margin: 0;
  min-height: 36px;
  display: flex;
  align-items: center;
  padding: 0 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #f8fafc;
  color: #1a202c;
  font-size: 14px;
  line-height: 1.35;
}

.outgoing-invoice-modal__date {
  width: 100%;
}

.outgoing-invoice-modal__date :deep(.n-input) {
  height: 36px;
}

.outgoing-invoice-modal__note {
  flex-shrink: 0;
  margin: 0;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #f8fafc;
  color: #475569;
  font-size: 13px;
  line-height: 1.45;
}

.outgoing-invoice-modal__note--error {
  border-color: #fecaca;
  background: #fef2f2;
  color: #b91c1c;
}

.outgoing-invoice-modal__positions {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

.outgoing-invoice-modal__table {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  background: #ffffff;
}

.outgoing-invoice-modal__rows {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  scrollbar-gutter: stable both-edges;
}

.outgoing-invoice-modal__grid-header,
.outgoing-invoice-modal__table-footer {
  overflow-y: hidden;
  scrollbar-gutter: stable both-edges;
}

.outgoing-invoice-modal__row,
.outgoing-invoice-modal__grid-header,
.outgoing-invoice-modal__table-footer {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 72px 88px 100px 32px;
  gap: 8px;
  align-items: start;
}

.outgoing-invoice-modal__grid-header {
  flex-shrink: 0;
  padding: 8px 12px 5px;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  font-size: 13px;
  font-weight: 500;
  color: #475569;
}

.outgoing-invoice-modal__row {
  padding: 7px 12px;
  border-bottom: 1px solid #e2e8f0;
}

.outgoing-invoice-modal__row:last-child {
  border-bottom: 0;
}

.outgoing-invoice-modal__product-cell {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.outgoing-invoice-modal__stock-hint {
  margin: 0;
  min-height: 14px;
  font-size: 11px;
  line-height: 14px;
  color: #dc2626;
  visibility: hidden;
}

.outgoing-invoice-modal__stock-hint--visible {
  visibility: visible;
}

.outgoing-invoice-modal__select {
  min-width: 0;
}

.outgoing-invoice-modal__select :deep(.n-base-selection) {
  min-height: 32px;
}

.outgoing-invoice-modal__select :deep(.n-base-selection-label) {
  height: 32px;
}

.outgoing-invoice-modal__input {
  width: 100%;
  height: 32px;
  min-height: 32px;
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

.outgoing-invoice-modal__input:focus {
  outline: none;
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.25);
}

.outgoing-invoice-modal__input--numeric {
  text-align: right;
}

.outgoing-invoice-modal__sum {
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: 0 10px;
  border-radius: 8px;
  background: #f8fafc;
  font-size: 14px;
  color: #334155;
  font-variant-numeric: tabular-nums;
}

.outgoing-invoice-modal__icon-action {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    color 0.15s ease;
}

.outgoing-invoice-modal__icon-action:hover:not(:disabled) {
  background: #f8fafc;
  color: #dc2626;
}

.outgoing-invoice-modal__icon-action:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.outgoing-invoice-modal__table-footer {
  flex-shrink: 0;
  grid-template-columns: minmax(0, 1fr) auto;
  padding: 12px;
  border-top: 1px solid #e2e8f0;
  background: #ffffff;
}

.outgoing-invoice-modal__field--comment {
  flex-shrink: 0;
}

.outgoing-invoice-modal__textarea {
  width: 100%;
  box-sizing: border-box;
  min-height: 52px;
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  color: #0f172a;
  font-size: 14px;
  font-family: inherit;
  line-height: 1.5;
  resize: none;
}

.outgoing-invoice-modal__textarea:focus {
  outline: none;
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.25);
}

.outgoing-invoice-modal__add-row-btn {
  justify-self: start;
  border: none;
  background: transparent;
  color: #1f883d;
  font-size: 13px;
  font-weight: 600;
  padding: 4px 0;
  cursor: pointer;
}

.outgoing-invoice-modal__add-row-btn:hover:not(:disabled) {
  color: #166534;
}

.outgoing-invoice-modal__add-row-btn:disabled {
  color: #94a3b8;
  cursor: not-allowed;
}

.outgoing-invoice-modal__total {
  display: flex;
  align-items: baseline;
  justify-content: flex-end;
  gap: 14px;
}

.outgoing-invoice-modal__total-label {
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
}

.outgoing-invoice-modal__total-value {
  font-size: 16px;
  font-weight: 700;
  color: #1a202c;
  text-align: right;
  font-variant-numeric: tabular-nums;
}

@media (max-width: 720px) {
  .outgoing-invoice-modal__header-fields {
    grid-template-columns: 1fr;
  }

  .outgoing-invoice-modal__row,
  .outgoing-invoice-modal__grid-header {
    grid-template-columns: minmax(0, 1fr) 64px 76px 88px 32px;
  }

  .outgoing-invoice-modal__table-footer {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .outgoing-invoice-modal__total {
    min-width: 0;
    width: 100%;
  }
}
</style>
