<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { AddOutline, PencilOutline, TrashOutline } from '@vicons/ionicons5'
import { NDatePicker, NIcon, NSelect } from 'naive-ui'
import type { SelectOption } from 'naive-ui'
import AppModal from '@/components/common/AppModal.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import DealDetailsSheet from '@/components/deals/DealDetailsSheet.vue'
import SectionSubviewHeader from '@/components/common/SectionSubviewHeader.vue'
import { useDeals } from '@/composables/useDeals'
import { useOutgoingInvoices } from '@/composables/useOutgoingInvoices'
import { useProductsCatalog } from '@/composables/useProductsCatalog'
import type { OutgoingInvoice } from '@/types/outgoingInvoice'
import type { ProductRow } from '@/types/productRow'
import { formatMoney } from '@/utils/money'
import { createEmptyProductRow, normalizeUnitPrice, productsToRows } from '@/utils/products'
import {
  renderSalaryDealOption,
  salaryDealOptionFullLabel,
  salaryDealOptionNumberLabel,
} from '@/utils/salaryDealLabel'

const { deals, loadDeals } = useDeals()
const { invoices, addInvoice, updateInvoice, removeInvoice } = useOutgoingInvoices()
const { catalogProductOptions, hasCatalogProducts, getCatalogProductById, loadCatalog, products } =
  useProductsCatalog()

const isFormModalOpen = ref(false)
const editingInvoiceId = ref<string | null>(null)
const isDeleteModalOpen = ref(false)
const invoiceToDelete = ref<OutgoingInvoice | null>(null)
const invoiceDate = ref<number | null>(null)
const dealId = ref<string | null>(null)
const dealSearchQuery = ref('')
const isDealsLoading = ref(false)
const invoiceComment = ref('')
const rows = ref<ProductRow[]>([createEmptyProductRow()])
const selectedDealId = ref<string | null>(null)

const datePickerTheme = {
  peers: {
    Input: {
      border: '1px solid #cbd5e1',
      borderHover: '1px solid #cbd5e1',
      borderFocus: '1px solid #93c5fd',
      boxShadowFocus: '0 0 0 3px rgba(147, 197, 253, 0.25)',
      borderRadius: '8px',
      heightMedium: '36px',
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
      heightMedium: '36px',
      fontSizeMedium: '14px',
      color: '#ffffff',
      placeholderColor: '#94a3b8',
    },
  },
}

function isClosedDeal(deal: { columnId?: string; status?: string }) {
  return deal.columnId === 'closed' || String(deal.status ?? '').toLowerCase() === 'closed'
}

const closedDealOptions = computed(() => {
  const query = dealSearchQuery.value.trim().toLowerCase()
  const selectedId = dealId.value

  return deals.value
    .filter((deal) => isClosedDeal(deal) || (selectedId != null && deal.id === selectedId))
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

const hasClosedDeals = computed(() => deals.value.some((deal) => isClosedDeal(deal)))

const filledRows = computed(() =>
  rows.value.filter((row) => Boolean(row.catalogProductId) && Number(row.quantity) > 0),
)

const invoiceTotal = computed(() =>
  filledRows.value.reduce((sum, row) => sum + getRowSum(row), 0),
)

const isEditMode = computed(() => editingInvoiceId.value !== null)
const modalTitle = computed(() =>
  isEditMode.value ? 'Редактировать расходную накладную' : 'Новая расходная накладная',
)
const submitButtonLabel = computed(() => (isEditMode.value ? 'Сохранить' : 'Добавить'))
const modalCloseLabel = computed(() =>
  isEditMode.value
    ? 'Закрыть окно редактирования расходной накладной'
    : 'Закрыть окно добавления расходной накладной',
)

const canSubmitInvoice = computed(
  () =>
    invoiceDate.value !== null &&
    Boolean(dealId.value) &&
    filledRows.value.length > 0,
)

onMounted(() => {
  void loadCatalog()
  void refreshDeals()
})

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

function resetInvoiceForm() {
  editingInvoiceId.value = null
  invoiceDate.value = Date.now()
  dealId.value = null
  dealSearchQuery.value = ''
  invoiceComment.value = ''
  rows.value = [createEmptyProductRow()]
}

function formatInvoiceDate(timestamp: number) {
  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  }).format(timestamp)
}

function getDealById(id: string) {
  return deals.value.find((item) => item.id === id) ?? null
}

function getDealLabel(id: string) {
  const deal = getDealById(id)
  if (!deal) return 'Сделка удалена'
  return salaryDealOptionNumberLabel(deal.dealNumber)
}

function getBuyerName(id: string) {
  const deal = getDealById(id)
  if (!deal) return '—'
  const name = deal.firstName.trim()
  return name || '—'
}

function getBuyerPhone(id: string) {
  const deal = getDealById(id)
  if (!deal) return '—'
  const phone = deal.phone.trim()
  return phone || '—'
}

async function handleOpenDeal(id: string) {
  if (!id) return

  if (!deals.value.some((deal) => deal.id === id)) {
    await loadDeals(true)
  }

  if (deals.value.some((deal) => deal.id === id)) {
    selectedDealId.value = id
  }
}

function handleCloseDealSheet() {
  selectedDealId.value = null
}

function handleDealSearch(query: string) {
  dealSearchQuery.value = query
}

function keepAllFilteredOptions() {
  return true
}

function openCreateModal() {
  resetInvoiceForm()
  void loadCatalog()
  void refreshDeals()
  isFormModalOpen.value = true
}

function openEditModal(invoice: OutgoingInvoice) {
  editingInvoiceId.value = invoice.id
  invoiceDate.value = invoice.date
  dealId.value = invoice.dealId
  dealSearchQuery.value = ''
  invoiceComment.value = invoice.comment ?? ''
  rows.value =
    invoice.items.length > 0 ? productsToRows(invoice.items, products.value) : [createEmptyProductRow()]
  void loadCatalog()
  void refreshDeals()
  isFormModalOpen.value = true
}

function closeFormModal() {
  isFormModalOpen.value = false
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

function handleSubmitInvoice() {
  if (!canSubmitInvoice.value || invoiceDate.value === null || !dealId.value) return

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

  if (editingInvoiceId.value) {
    updateInvoice(editingInvoiceId.value, payload)
  } else {
    addInvoice(payload)
  }

  closeFormModal()
  resetInvoiceForm()
}

function handleDeleteInvoice(invoice: OutgoingInvoice) {
  invoiceToDelete.value = invoice
  isDeleteModalOpen.value = true
}

function closeDeleteModal() {
  isDeleteModalOpen.value = false
  invoiceToDelete.value = null
}

function confirmDeleteInvoice() {
  if (!invoiceToDelete.value) return
  removeInvoice(invoiceToDelete.value.id)
  if (editingInvoiceId.value === invoiceToDelete.value.id) {
    closeFormModal()
    resetInvoiceForm()
  }
  closeDeleteModal()
}
</script>

<template>
  <div class="trade-subview">
    <SectionSubviewHeader title="Расходные накладные">
      <template #actions>
        <button
          type="button"
          class="trade-subview__create-btn"
          title="Добавить расходную накладную"
          aria-label="Добавить расходную накладную"
          @click="openCreateModal"
        >
          <NIcon :size="18" :component="AddOutline" />
        </button>
      </template>
    </SectionSubviewHeader>

    <div class="trade-subview__body">
      <section v-if="invoices.length === 0" class="outgoing-invoices-view__placeholder">
        <p class="outgoing-invoices-view__placeholder-text">
          Пока нет расходных накладных. Добавьте первую через кнопку «+» вверху.
        </p>
      </section>

      <section v-else class="outgoing-invoices-view__table-wrap">
        <div class="outgoing-invoices-view__table" role="table">
          <div class="outgoing-invoices-view__table-row outgoing-invoices-view__table-row--head" role="row">
            <span
              class="outgoing-invoices-view__cell outgoing-invoices-view__cell--head outgoing-invoices-view__cell--compact"
              role="columnheader"
            >
              Номер
            </span>
            <span
              class="outgoing-invoices-view__cell outgoing-invoices-view__cell--head outgoing-invoices-view__cell--compact"
              role="columnheader"
            >
              Дата
            </span>
            <span
              class="outgoing-invoices-view__cell outgoing-invoices-view__cell--head outgoing-invoices-view__cell--compact"
              role="columnheader"
            >
              Номер сделки
            </span>
            <span class="outgoing-invoices-view__cell outgoing-invoices-view__cell--head" role="columnheader">
              Имя покупателя
            </span>
            <span
              class="outgoing-invoices-view__cell outgoing-invoices-view__cell--head outgoing-invoices-view__cell--phone"
              role="columnheader"
            >
              Телефон покупателя
            </span>
            <span
              class="outgoing-invoices-view__cell outgoing-invoices-view__cell--head outgoing-invoices-view__cell--compact"
              role="columnheader"
            >
              Общая сумма
            </span>
            <span
              class="outgoing-invoices-view__cell outgoing-invoices-view__cell--head outgoing-invoices-view__cell--actions"
              role="columnheader"
              aria-hidden="true"
            />
          </div>

          <div
            v-for="invoice in invoices"
            :key="invoice.id"
            class="outgoing-invoices-view__table-row"
            role="row"
          >
            <span class="outgoing-invoices-view__cell outgoing-invoices-view__cell--compact outgoing-invoices-view__cell--number">
              #{{ invoice.invoiceNumber }}
            </span>
            <span class="outgoing-invoices-view__cell outgoing-invoices-view__cell--compact outgoing-invoices-view__cell--date">
              {{ formatInvoiceDate(invoice.date) }}
            </span>
            <span class="outgoing-invoices-view__cell outgoing-invoices-view__cell--compact outgoing-invoices-view__cell--deal">
              <button
                type="button"
                class="outgoing-invoices-view__deal-link"
                :aria-label="`Открыть сделку ${getDealLabel(invoice.dealId)}`"
                @click="handleOpenDeal(invoice.dealId)"
              >
                {{ getDealLabel(invoice.dealId) }}
              </button>
            </span>
            <span class="outgoing-invoices-view__cell outgoing-invoices-view__cell--name">
              {{ getBuyerName(invoice.dealId) }}
            </span>
            <span class="outgoing-invoices-view__cell outgoing-invoices-view__cell--phone">
              {{ getBuyerPhone(invoice.dealId) }}
            </span>
            <span class="outgoing-invoices-view__cell outgoing-invoices-view__cell--compact outgoing-invoices-view__cell--sum">
              {{ formatMoney(invoice.total) }}
            </span>
            <div class="outgoing-invoices-view__cell outgoing-invoices-view__cell--actions">
              <div class="outgoing-invoices-view__row-actions">
                <button
                  type="button"
                  class="outgoing-invoices-view__icon-action"
                  aria-label="Редактировать расходную накладную"
                  title="Редактировать"
                  @click="openEditModal(invoice)"
                >
                  <NIcon :size="16">
                    <PencilOutline />
                  </NIcon>
                </button>
                <button
                  type="button"
                  class="outgoing-invoices-view__icon-action outgoing-invoices-view__icon-action--danger"
                  aria-label="Удалить расходную накладную"
                  title="Удалить"
                  @click="handleDeleteInvoice(invoice)"
                >
                  <NIcon :size="16">
                    <TrashOutline />
                  </NIcon>
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <AppModal
      v-model:show="isFormModalOpen"
      :title="modalTitle"
      width="large"
      height="tall"
      actions-align="center"
      :close-label="modalCloseLabel"
      @close="resetInvoiceForm"
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

          <label class="outgoing-invoice-modal__field">
            <span class="outgoing-invoice-modal__label">
              Сделка
              <span class="outgoing-invoice-modal__required" aria-hidden="true">*</span>
            </span>
            <NSelect
              v-model:value="dealId"
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
            />
          </label>
        </section>

        <p
          v-if="!isDealsLoading && !hasClosedDeals"
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

    <DealDetailsSheet :deal-id="selectedDealId" @close="handleCloseDealSheet" />

    <AppModal
      v-model:show="isDeleteModalOpen"
      title="Удаление расходной накладной"
      body-variant="center"
      @close="closeDeleteModal"
    >
      <p class="app-modal__message">Вы уверены, что хотите удалить данную расходную накладную?</p>

      <template #actions>
        <div class="outgoing-invoices-view__confirm-actions">
          <AppModalButton @click="confirmDeleteInvoice">Да</AppModalButton>
          <button type="button" class="outgoing-invoices-view__confirm-cancel" @click="closeDeleteModal">
            Нет
          </button>
        </div>
      </template>
    </AppModal>
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

.trade-subview__create-btn {
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

.trade-subview__create-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #1f2937;
}

.outgoing-invoices-view__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 240px;
  padding: 32px 24px;
  border: 1px dashed #cbd5e1;
  border-radius: 12px;
  background: #f8fafc;
}

.outgoing-invoices-view__placeholder-text {
  margin: 0;
  max-width: 420px;
  font-size: 15px;
  line-height: 1.5;
  color: #64748b;
  text-align: center;
}

.outgoing-invoices-view__table-wrap {
  min-width: 0;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
  overflow-x: auto;
}

.outgoing-invoices-view__table {
  display: grid;
  width: 100%;
  min-width: 760px;
  grid-template-columns:
    max-content
    max-content
    max-content
    minmax(140px, 1.2fr)
    minmax(150px, 1fr)
    max-content
    max-content;
}

.outgoing-invoices-view__table-row {
  display: contents;
}

.outgoing-invoices-view__cell {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  min-height: 40px;
  padding: 7px 14px;
  border-right: 1px solid #e2e8f0;
  border-bottom: 1px solid #e2e8f0;
  font-size: 14px;
  line-height: 1.35;
  color: #1a202c;
  background: #ffffff;
}

.outgoing-invoices-view__cell--head {
  min-height: 44px;
  padding: 10px 14px;
  background: #f8fafc;
  font-size: 11px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.outgoing-invoices-view__table-row .outgoing-invoices-view__cell:nth-child(7) {
  border-right: 0;
}

.outgoing-invoices-view__table-row:last-child .outgoing-invoices-view__cell {
  border-bottom: 0;
}

.outgoing-invoices-view__cell--compact {
  padding-left: 10px;
  padding-right: 10px;
  white-space: nowrap;
}

.outgoing-invoices-view__cell--number,
.outgoing-invoices-view__cell--date,
.outgoing-invoices-view__cell--deal,
.outgoing-invoices-view__cell--phone {
  font-variant-numeric: tabular-nums;
}

.outgoing-invoices-view__cell--name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.outgoing-invoices-view__cell--phone {
  white-space: nowrap;
}

.outgoing-invoices-view__deal-link {
  margin: 0;
  padding: 0;
  border: none;
  background: transparent;
  color: #1d4ed8;
  font: inherit;
  cursor: pointer;
  text-decoration: none;
}

.outgoing-invoices-view__deal-link:hover {
  color: #1e40af;
  text-decoration: underline;
}

.outgoing-invoices-view__deal-link:focus-visible {
  outline: 2px solid #93c5fd;
  outline-offset: 2px;
  border-radius: 4px;
}

.outgoing-invoices-view__cell--sum {
  font-variant-numeric: tabular-nums;
  justify-content: flex-end;
}

.outgoing-invoices-view__cell--actions {
  justify-content: center;
  padding: 10px;
  white-space: nowrap;
}

.outgoing-invoices-view__row-actions {
  display: inline-flex;
  justify-content: center;
  gap: 6px;
}

.outgoing-invoices-view__icon-action {
  width: 28px;
  height: 28px;
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

.outgoing-invoices-view__icon-action:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #334155;
}

.outgoing-invoices-view__icon-action--danger:hover {
  color: #dc2626;
}

.outgoing-invoices-view__table-row:not(.outgoing-invoices-view__table-row--head):hover .outgoing-invoices-view__cell {
  background: #f8fafc;
}

.outgoing-invoices-view__confirm-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px;
}

.outgoing-invoices-view__confirm-cancel {
  min-width: min(100%, 220px);
  padding: 10px 20px;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  background: #ffffff;
  color: #475569;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.35;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.outgoing-invoices-view__confirm-cancel:hover {
  background: #f8fafc;
  border-color: #94a3b8;
  color: #334155;
}

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
  grid-template-columns: minmax(0, 1fr) 72px 88px 100px 36px;
  gap: 8px;
  align-items: center;
}

.outgoing-invoice-modal__grid-header {
  flex-shrink: 0;
  padding: 10px 12px 6px;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  font-size: 13px;
  font-weight: 500;
  color: #475569;
}

.outgoing-invoice-modal__row {
  padding: 10px 12px;
  border-bottom: 1px solid #e2e8f0;
}

.outgoing-invoice-modal__row:last-child {
  border-bottom: 0;
}

.outgoing-invoice-modal__select {
  min-width: 0;
}

.outgoing-invoice-modal__select :deep(.n-base-selection) {
  min-height: 36px;
}

.outgoing-invoice-modal__select :deep(.n-base-selection-label) {
  height: 36px;
}

.outgoing-invoice-modal__input {
  width: 100%;
  height: 36px;
  min-height: 36px;
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
  height: 36px;
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
  width: 36px;
  height: 36px;
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
    grid-template-columns: minmax(0, 1fr) 64px 76px 88px 36px;
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
