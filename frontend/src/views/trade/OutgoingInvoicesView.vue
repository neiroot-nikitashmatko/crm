<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { AddOutline, PencilOutline, TrashOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import AppModal from '@/components/common/AppModal.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import DealDetailsSheet from '@/components/deals/DealDetailsSheet.vue'
import OutgoingInvoiceFormModal from '@/components/trade/OutgoingInvoiceFormModal.vue'
import SectionSubviewHeader from '@/components/common/SectionSubviewHeader.vue'
import { useDeals } from '@/composables/useDeals'
import { useOutgoingInvoices } from '@/composables/useOutgoingInvoices'
import type { OutgoingInvoice } from '@/types/outgoingInvoice'
import { formatMoney } from '@/utils/money'
import { salaryDealOptionNumberLabel } from '@/utils/salaryDealLabel'

const { deals, loadDeals } = useDeals()
const { invoices, isLoading, loadInvoices, removeInvoice } = useOutgoingInvoices()

const isFormModalOpen = ref(false)
const editingInvoiceId = ref<string | null>(null)
const isDeleteModalOpen = ref(false)
const invoiceToDelete = ref<OutgoingInvoice | null>(null)
const selectedDealId = ref<string | null>(null)
const deleteError = ref('')

onMounted(() => {
  void loadDeals(true)
  void loadInvoices()
})

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

function openCreateModal() {
  editingInvoiceId.value = null
  isFormModalOpen.value = true
}

function openEditModal(invoice: OutgoingInvoice) {
  editingInvoiceId.value = invoice.id
  isFormModalOpen.value = true
}

function closeFormModal() {
  isFormModalOpen.value = false
  editingInvoiceId.value = null
}

function handleInvoiceSaved() {
  closeFormModal()
}

function handleDeleteInvoice(invoice: OutgoingInvoice) {
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
    }
    closeDeleteModal()
  } catch (error) {
    deleteError.value =
      error instanceof Error && error.message.trim()
        ? error.message
        : 'Не удалось удалить расходную накладную'
  }
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
      <section v-if="isLoading && invoices.length === 0" class="outgoing-invoices-view__placeholder">
        <p class="outgoing-invoices-view__placeholder-text">Загрузка…</p>
      </section>

      <section v-else-if="invoices.length === 0" class="outgoing-invoices-view__placeholder">
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

    <OutgoingInvoiceFormModal
      :open="isFormModalOpen"
      :editing-invoice-id="editingInvoiceId"
      @close="closeFormModal"
      @saved="handleInvoiceSaved"
    />

    <DealDetailsSheet :deal-id="selectedDealId" @close="handleCloseDealSheet" />

    <AppModal
      v-model:show="isDeleteModalOpen"
      title="Удаление расходной накладной"
      body-variant="center"
      @close="closeDeleteModal"
    >
      <p class="app-modal__message">Вы уверены, что хотите удалить данную расходную накладную?</p>
      <p v-if="deleteError" class="outgoing-invoices-view__error">{{ deleteError }}</p>

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

.outgoing-invoices-view__error {
  margin: 8px 0 0;
  color: #b91c1c;
  font-size: 13px;
  line-height: 1.4;
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
</style>
