<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { AddOutline, PencilOutline, TrashOutline } from '@vicons/ionicons5'
import { NDatePicker, NIcon, NSelect } from 'naive-ui'
import type { SelectOption } from 'naive-ui'
import AppModal from '@/components/common/AppModal.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import SectionSubviewHeader from '@/components/common/SectionSubviewHeader.vue'
import { useIncomingInvoices } from '@/composables/useIncomingInvoices'
import { useProductsCatalog } from '@/composables/useProductsCatalog'
import { useSuppliers } from '@/composables/useSuppliers'
import type { IncomingInvoice } from '@/types/incomingInvoice'
import type { ProductRow } from '@/types/productRow'
import { formatMoney } from '@/utils/money'
import { createEmptyProductRow, normalizeUnitPrice, productsToRows } from '@/utils/products'

const { suppliers, loadSuppliers } = useSuppliers()
const { invoices, isLoading, loadInvoices, addInvoice, updateInvoice, removeInvoice } =
  useIncomingInvoices()
const { catalogProductOptions, hasCatalogProducts, getCatalogProductById, loadCatalog, products } =
  useProductsCatalog()

const isFormModalOpen = ref(false)
const editingInvoiceId = ref<string | null>(null)
const isDeleteModalOpen = ref(false)
const invoiceToDelete = ref<IncomingInvoice | null>(null)
const invoiceDate = ref<number | null>(null)
const supplierId = ref<string | null>(null)
const invoiceComment = ref('')
const rows = ref<ProductRow[]>([createEmptyProductRow()])
const submitError = ref('')
const deleteError = ref('')

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

const supplierOptions = computed(() =>
  suppliers.value.map((supplier) => ({
    label: supplier.name,
    value: supplier.id,
  })),
)

const hasSuppliers = computed(() => suppliers.value.length > 0)

const filledRows = computed(() =>
  rows.value.filter((row) => Boolean(row.catalogProductId) && Number(row.quantity) > 0),
)

const invoiceTotal = computed(() =>
  filledRows.value.reduce((sum, row) => sum + getRowSum(row), 0),
)

const isEditMode = computed(() => editingInvoiceId.value !== null)
const modalTitle = computed(() =>
  isEditMode.value ? 'Редактировать приходную накладную' : 'Новая приходная накладная',
)
const submitButtonLabel = computed(() => (isEditMode.value ? 'Сохранить' : 'Добавить'))
const modalCloseLabel = computed(() =>
  isEditMode.value
    ? 'Закрыть окно редактирования приходной накладной'
    : 'Закрыть окно добавления приходной накладной',
)

const canSubmitInvoice = computed(
  () =>
    invoiceDate.value !== null &&
    Boolean(supplierId.value) &&
    filledRows.value.length > 0,
)

onMounted(() => {
  void loadCatalog()
  void loadSuppliers()
  void loadInvoices()
})

function getRowSum(row: ProductRow) {
  const quantity = Number(row.quantity)
  const safeQuantity = Number.isFinite(quantity) && quantity > 0 ? quantity : 0
  return safeQuantity * normalizeUnitPrice(row.unitPrice)
}

function resetInvoiceForm() {
  editingInvoiceId.value = null
  submitError.value = ''
  invoiceDate.value = Date.now()
  supplierId.value = null
  invoiceComment.value = ''
  rows.value = [createEmptyProductRow()]
}

function getErrorMessage(error: unknown, fallback: string) {
  return error instanceof Error && error.message.trim() ? error.message : fallback
}

function formatInvoiceDate(timestamp: number) {
  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  }).format(timestamp)
}

function getSupplierName(id: string) {
  return suppliers.value.find((supplier) => supplier.id === id)?.name ?? 'Поставщик удалён'
}

function openCreateModal() {
  resetInvoiceForm()
  void loadCatalog()
  void loadSuppliers()
  isFormModalOpen.value = true
}

function openEditModal(invoice: IncomingInvoice) {
  editingInvoiceId.value = invoice.id
  invoiceDate.value = invoice.date
  supplierId.value = invoice.supplierId
  invoiceComment.value = invoice.comment ?? ''
  rows.value =
    invoice.items.length > 0 ? productsToRows(invoice.items, products.value) : [createEmptyProductRow()]
  void loadCatalog()
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

async function handleSubmitInvoice() {
  if (!canSubmitInvoice.value || invoiceDate.value === null || !supplierId.value) return

  const payload = {
    date: invoiceDate.value,
    supplierId: supplierId.value,
    items: filledRows.value.map((row) => ({
      catalogProductId: row.catalogProductId,
      title: row.title,
      quantity: Number(row.quantity),
      unitPrice: normalizeUnitPrice(row.unitPrice),
    })),
    total: invoiceTotal.value,
    comment: invoiceComment.value.trim(),
  }

  submitError.value = ''
  try {
    if (editingInvoiceId.value) {
      await updateInvoice(editingInvoiceId.value, payload)
    } else {
      await addInvoice(payload)
    }
    closeFormModal()
    resetInvoiceForm()
  } catch (error) {
    submitError.value = getErrorMessage(error, 'Не удалось сохранить приходную накладную')
  }
}

function handleDeleteInvoice(invoice: IncomingInvoice) {
  invoiceToDelete.value = invoice
  deleteError.value = ''
  isDeleteModalOpen.value = true
}

function closeDeleteModal() {
  isDeleteModalOpen.value = false
  invoiceToDelete.value = null
  deleteError.value = ''
}

async function confirmDeleteInvoice() {
  if (!invoiceToDelete.value) return
  deleteError.value = ''
  try {
    await removeInvoice(invoiceToDelete.value.id)
    if (editingInvoiceId.value === invoiceToDelete.value.id) {
      closeFormModal()
      resetInvoiceForm()
    }
    closeDeleteModal()
  } catch (error) {
    deleteError.value = getErrorMessage(error, 'Не удалось удалить приходную накладную')
  }
}
</script>

<template>
  <div class="trade-subview">
    <SectionSubviewHeader title="Приходные накладные">
      <template #actions>
        <button
          type="button"
          class="trade-subview__create-btn"
          title="Добавить приходную накладную"
          aria-label="Добавить приходную накладную"
          @click="openCreateModal"
        >
          <NIcon :size="18" :component="AddOutline" />
        </button>
      </template>
    </SectionSubviewHeader>

    <div class="trade-subview__body">
      <section v-if="isLoading && invoices.length === 0" class="incoming-invoices-view__placeholder">
        <p class="incoming-invoices-view__placeholder-text">Загрузка…</p>
      </section>

      <section v-else-if="invoices.length === 0" class="incoming-invoices-view__placeholder">
        <p class="incoming-invoices-view__placeholder-text">
          Пока нет приходных накладных. Добавьте первую через кнопку «+» вверху.
        </p>
      </section>

      <section v-else class="incoming-invoices-view__table-wrap">
        <div class="incoming-invoices-view__table" role="table">
          <div class="incoming-invoices-view__table-row incoming-invoices-view__table-row--head" role="row">
            <span
              class="incoming-invoices-view__cell incoming-invoices-view__cell--head incoming-invoices-view__cell--compact"
              role="columnheader"
            >
              Номер
            </span>
            <span
              class="incoming-invoices-view__cell incoming-invoices-view__cell--head incoming-invoices-view__cell--compact"
              role="columnheader"
            >
              Дата
            </span>
            <span class="incoming-invoices-view__cell incoming-invoices-view__cell--head" role="columnheader">
              Поставщик
            </span>
            <span
              class="incoming-invoices-view__cell incoming-invoices-view__cell--head incoming-invoices-view__cell--compact"
              role="columnheader"
            >
              Общая сумма
            </span>
            <span
              class="incoming-invoices-view__cell incoming-invoices-view__cell--head incoming-invoices-view__cell--actions"
              role="columnheader"
              aria-hidden="true"
            />
          </div>

          <div
            v-for="invoice in invoices"
            :key="invoice.id"
            class="incoming-invoices-view__table-row"
            role="row"
          >
            <span class="incoming-invoices-view__cell incoming-invoices-view__cell--compact incoming-invoices-view__cell--number">
              #{{ invoice.invoiceNumber }}
            </span>
            <span class="incoming-invoices-view__cell incoming-invoices-view__cell--compact incoming-invoices-view__cell--date">
              {{ formatInvoiceDate(invoice.date) }}
            </span>
            <span class="incoming-invoices-view__cell">{{ getSupplierName(invoice.supplierId) }}</span>
            <span class="incoming-invoices-view__cell incoming-invoices-view__cell--compact incoming-invoices-view__cell--sum">
              {{ formatMoney(invoice.total) }}
            </span>
            <div class="incoming-invoices-view__cell incoming-invoices-view__cell--actions">
              <div class="incoming-invoices-view__row-actions">
                <button
                  type="button"
                  class="incoming-invoices-view__icon-action"
                  aria-label="Редактировать приходную накладную"
                  title="Редактировать"
                  @click="openEditModal(invoice)"
                >
                  <NIcon :size="16">
                    <PencilOutline />
                  </NIcon>
                </button>
                <button
                  type="button"
                  class="incoming-invoices-view__icon-action incoming-invoices-view__icon-action--danger"
                  aria-label="Удалить приходную накладную"
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
      <div class="incoming-invoice-modal">
        <section class="incoming-invoice-modal__header-fields">
          <label class="incoming-invoice-modal__field">
            <span class="incoming-invoice-modal__label">
              Дата
              <span class="incoming-invoice-modal__required" aria-hidden="true">*</span>
            </span>
            <NDatePicker
              v-model:value="invoiceDate"
              class="incoming-invoice-modal__date"
              :theme-overrides="datePickerTheme"
              type="date"
              format="dd.MM.yyyy"
              date-format="dd.MM.yyyy"
              placeholder="Выберите дату"
              :actions="[]"
            />
          </label>

          <label class="incoming-invoice-modal__field">
            <span class="incoming-invoice-modal__label">
              Поставщик
              <span class="incoming-invoice-modal__required" aria-hidden="true">*</span>
            </span>
            <NSelect
              v-model:value="supplierId"
              class="incoming-invoice-modal__select"
              :theme-overrides="selectTheme"
              :options="supplierOptions"
              :disabled="!hasSuppliers"
              filterable
              clearable
              placeholder="Выберите поставщика"
            />
          </label>
        </section>

        <p v-if="!hasSuppliers" class="incoming-invoice-modal__note">
          Сначала добавьте поставщика в раздел «Поставщики».
        </p>
        <p v-if="!hasCatalogProducts" class="incoming-invoice-modal__note">
          Сначала добавьте товары в раздел «Товары и услуги», чтобы выбрать их здесь.
        </p>
        <p v-if="submitError" class="incoming-invoice-modal__note incoming-invoice-modal__note--error">
          {{ submitError }}
        </p>

        <section class="incoming-invoice-modal__positions">
          <div class="incoming-invoice-modal__table">
            <div class="incoming-invoice-modal__grid-header">
              <span>Товар / услуга</span>
              <span>Кол-во</span>
              <span>Стоимость</span>
              <span>Сумма</span>
              <span aria-hidden="true" />
            </div>

            <div class="incoming-invoice-modal__rows">
              <div
                v-for="(product, index) in rows"
                :key="product.rowId"
                class="incoming-invoice-modal__row"
              >
                <NSelect
                  :value="product.catalogProductId ?? null"
                  :options="catalogProductOptions"
                  class="incoming-invoice-modal__select"
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
                  class="incoming-invoice-modal__input incoming-invoice-modal__input--numeric"
                  placeholder="1"
                  @input="handleQuantityInput(index, $event)"
                />
                <input
                  :value="product.unitPrice"
                  type="text"
                  inputmode="decimal"
                  autocomplete="off"
                  class="incoming-invoice-modal__input incoming-invoice-modal__input--numeric"
                  placeholder="0"
                  @input="handleCostInput(index, $event)"
                />
                <span class="incoming-invoice-modal__sum">{{ formatMoney(getRowSum(product)) }}</span>
                <button
                  type="button"
                  class="incoming-invoice-modal__icon-action"
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

            <div class="incoming-invoice-modal__table-footer">
              <button
                type="button"
                class="incoming-invoice-modal__add-row-btn"
                :disabled="!hasCatalogProducts"
                @click="addProductRow"
              >
                + Добавить позицию
              </button>

              <div class="incoming-invoice-modal__total">
                <span class="incoming-invoice-modal__total-label">Итого</span>
                <span class="incoming-invoice-modal__total-value">{{ formatMoney(invoiceTotal) }}</span>
              </div>
            </div>
          </div>
        </section>

        <label class="incoming-invoice-modal__field incoming-invoice-modal__field--comment">
          <span class="incoming-invoice-modal__label">Комментарий</span>
          <textarea
            v-model="invoiceComment"
            class="incoming-invoice-modal__textarea"
            rows="2"
            placeholder="Комментарий (необязательно)"
          />
        </label>
      </div>

      <template #actions>
        <AppModalButton
          :disabled="!canSubmitInvoice"
          :title="canSubmitInvoice ? '' : 'Заполните дату, поставщика и хотя бы одну позицию'"
          @click="handleSubmitInvoice"
        >
          {{ submitButtonLabel }}
        </AppModalButton>
      </template>
    </AppModal>

    <AppModal
      v-model:show="isDeleteModalOpen"
      title="Удаление приходной накладной"
      body-variant="center"
      @close="closeDeleteModal"
    >
      <p class="app-modal__message">Вы уверены, что хотите удалить данную приходную накладную?</p>
      <p v-if="deleteError" class="incoming-invoice-modal__note incoming-invoice-modal__note--error">
        {{ deleteError }}
      </p>

      <template #actions>
        <div class="incoming-invoices-view__confirm-actions">
          <AppModalButton @click="confirmDeleteInvoice">Да</AppModalButton>
          <button type="button" class="incoming-invoices-view__confirm-cancel" @click="closeDeleteModal">
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

.incoming-invoices-view__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 240px;
  padding: 32px 24px;
  border: 1px dashed #cbd5e1;
  border-radius: 12px;
  background: #f8fafc;
}

.incoming-invoices-view__placeholder-text {
  margin: 0;
  max-width: 420px;
  font-size: 15px;
  line-height: 1.5;
  color: #64748b;
  text-align: center;
}

.incoming-invoices-view__table-wrap {
  min-width: 0;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
  overflow-x: auto;
}

.incoming-invoices-view__table {
  display: grid;
  width: 100%;
  min-width: 560px;
  grid-template-columns:
    max-content
    max-content
    minmax(160px, 1fr)
    max-content
    max-content;
}

.incoming-invoices-view__table-row {
  display: contents;
}

.incoming-invoices-view__cell {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  min-height: 48px;
  padding: 12px 14px;
  border-right: 1px solid #e2e8f0;
  border-bottom: 1px solid #e2e8f0;
  font-size: 14px;
  line-height: 1.35;
  color: #1a202c;
  background: #ffffff;
}

.incoming-invoices-view__cell--head {
  min-height: 44px;
  padding: 10px 14px;
  background: #f8fafc;
  font-size: 11px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.incoming-invoices-view__table-row .incoming-invoices-view__cell:nth-child(5) {
  border-right: 0;
}

.incoming-invoices-view__table-row:last-child .incoming-invoices-view__cell {
  border-bottom: 0;
}

.incoming-invoices-view__cell--compact {
  padding-left: 10px;
  padding-right: 10px;
  white-space: nowrap;
}

.incoming-invoices-view__cell--number,
.incoming-invoices-view__cell--date {
  font-variant-numeric: tabular-nums;
}

.incoming-invoices-view__cell--sum {
  font-variant-numeric: tabular-nums;
  justify-content: flex-end;
}

.incoming-invoices-view__cell--actions {
  justify-content: center;
  padding: 10px;
  white-space: nowrap;
}

.incoming-invoices-view__row-actions {
  display: inline-flex;
  justify-content: center;
  gap: 6px;
}

.incoming-invoices-view__icon-action {
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

.incoming-invoices-view__icon-action:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #334155;
}

.incoming-invoices-view__icon-action--danger:hover {
  color: #dc2626;
}

.incoming-invoices-view__table-row:not(.incoming-invoices-view__table-row--head):hover .incoming-invoices-view__cell {
  background: #f8fafc;
}

.incoming-invoices-view__confirm-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px;
}

.incoming-invoices-view__confirm-cancel {
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

.incoming-invoices-view__confirm-cancel:hover {
  background: #f8fafc;
  border-color: #94a3b8;
  color: #334155;
}

.incoming-invoice-modal {
  display: flex;
  flex-direction: column;
  gap: 16px;
  flex: 1;
  min-height: 0;
  height: 100%;
}

.incoming-invoice-modal__header-fields {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px 16px;
  flex-shrink: 0;
}

.incoming-invoice-modal__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.incoming-invoice-modal__label {
  font-size: 13px;
  color: #475569;
}

.incoming-invoice-modal__required {
  color: #dc2626;
  margin-left: 2px;
}

.incoming-invoice-modal__date {
  width: 100%;
}

.incoming-invoice-modal__date :deep(.n-input) {
  height: 36px;
}

.incoming-invoice-modal__note {
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

.incoming-invoice-modal__note--error {
  border-color: #fecaca;
  background: #fef2f2;
  color: #b91c1c;
}

.incoming-invoice-modal__positions {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

.incoming-invoice-modal__table {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  background: #ffffff;
}

.incoming-invoice-modal__rows {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  scrollbar-gutter: stable both-edges;
}

.incoming-invoice-modal__grid-header,
.incoming-invoice-modal__table-footer {
  overflow-y: hidden;
  scrollbar-gutter: stable both-edges;
}

.incoming-invoice-modal__row,
.incoming-invoice-modal__grid-header,
.incoming-invoice-modal__table-footer {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 72px 88px 100px 36px;
  gap: 8px;
  align-items: center;
}

.incoming-invoice-modal__grid-header {
  flex-shrink: 0;
  padding: 10px 12px 6px;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  font-size: 13px;
  font-weight: 500;
  color: #475569;
}

.incoming-invoice-modal__row {
  padding: 10px 12px;
  border-bottom: 1px solid #e2e8f0;
}

.incoming-invoice-modal__row:last-child {
  border-bottom: 0;
}

.incoming-invoice-modal__select {
  min-width: 0;
}

.incoming-invoice-modal__select :deep(.n-base-selection) {
  min-height: 36px;
}

.incoming-invoice-modal__select :deep(.n-base-selection-label) {
  height: 36px;
}

.incoming-invoice-modal__input {
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

.incoming-invoice-modal__input:focus {
  outline: none;
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.25);
}

.incoming-invoice-modal__input--numeric {
  text-align: right;
}

.incoming-invoice-modal__sum {
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

.incoming-invoice-modal__icon-action {
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

.incoming-invoice-modal__icon-action:hover:not(:disabled) {
  background: #f8fafc;
  color: #dc2626;
}

.incoming-invoice-modal__icon-action:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.incoming-invoice-modal__table-footer {
  flex-shrink: 0;
  grid-template-columns: minmax(0, 1fr) auto;
  padding: 12px;
  border-top: 1px solid #e2e8f0;
  background: #ffffff;
}

.incoming-invoice-modal__field--comment {
  flex-shrink: 0;
}

.incoming-invoice-modal__textarea {
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

.incoming-invoice-modal__textarea:focus {
  outline: none;
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.25);
}

.incoming-invoice-modal__add-row-btn {
  justify-self: start;
  border: none;
  background: transparent;
  color: #1f883d;
  font-size: 13px;
  font-weight: 600;
  padding: 4px 0;
  cursor: pointer;
}

.incoming-invoice-modal__add-row-btn:hover:not(:disabled) {
  color: #166534;
}

.incoming-invoice-modal__add-row-btn:disabled {
  color: #94a3b8;
  cursor: not-allowed;
}

.incoming-invoice-modal__total {
  display: flex;
  align-items: baseline;
  justify-content: flex-end;
  gap: 14px;
}

.incoming-invoice-modal__total-label {
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
}

.incoming-invoice-modal__total-value {
  font-size: 16px;
  font-weight: 700;
  color: #1a202c;
  text-align: right;
  font-variant-numeric: tabular-nums;
}

@media (max-width: 720px) {
  .incoming-invoice-modal__header-fields {
    grid-template-columns: 1fr;
  }

  .incoming-invoice-modal__row,
  .incoming-invoice-modal__grid-header {
    grid-template-columns: minmax(0, 1fr) 64px 76px 88px 36px;
  }

  .incoming-invoice-modal__table-footer {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .incoming-invoice-modal__total {
    min-width: 0;
    width: 100%;
  }
}
</style>
