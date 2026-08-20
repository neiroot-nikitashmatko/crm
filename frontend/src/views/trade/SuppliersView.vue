<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { AddOutline, PencilOutline, TrashOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import AppModal from '@/components/common/AppModal.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import SectionSubviewHeader from '@/components/common/SectionSubviewHeader.vue'
import { useSuppliers } from '@/composables/useSuppliers'
import type { Supplier } from '@/types/supplier'
import { isPhoneFilled, normalizePhone, PHONE_PREFIX } from '@/utils/phone'

const { suppliers, addSupplier, updateSupplier, removeSupplier } = useSuppliers()

const isFormModalOpen = ref(false)
const editingSupplierId = ref<string | null>(null)
const isDeleteModalOpen = ref(false)
const supplierToDelete = ref<Supplier | null>(null)

const supplierForm = reactive({
  name: '',
  contactPerson: '',
  phone: PHONE_PREFIX,
  inn: '',
  kpp: '',
  ogrn: '',
  legalAddress: '',
  actualAddress: '',
  bik: '',
  settlementAccount: '',
  correspondentAccount: '',
})

const isEditMode = computed(() => editingSupplierId.value !== null)
const modalTitle = computed(() => (isEditMode.value ? 'Редактировать поставщика' : 'Новый поставщик'))
const submitButtonLabel = computed(() => (isEditMode.value ? 'Сохранить' : 'Добавить'))
const modalCloseLabel = computed(() =>
  isEditMode.value ? 'Закрыть окно редактирования поставщика' : 'Закрыть окно добавления поставщика',
)

const canSubmitSupplier = computed(
  () =>
    supplierForm.name.trim().length > 0 &&
    supplierForm.contactPerson.trim().length > 0 &&
    isPhoneFilled(supplierForm.phone) &&
    supplierForm.inn.trim().length > 0 &&
    supplierForm.legalAddress.trim().length > 0 &&
    supplierForm.actualAddress.trim().length > 0 &&
    supplierForm.bik.trim().length > 0 &&
    supplierForm.settlementAccount.trim().length > 0 &&
    supplierForm.correspondentAccount.trim().length > 0,
)

function resetSupplierForm() {
  editingSupplierId.value = null
  supplierForm.name = ''
  supplierForm.contactPerson = ''
  supplierForm.phone = PHONE_PREFIX
  supplierForm.inn = ''
  supplierForm.kpp = ''
  supplierForm.ogrn = ''
  supplierForm.legalAddress = ''
  supplierForm.actualAddress = ''
  supplierForm.bik = ''
  supplierForm.settlementAccount = ''
  supplierForm.correspondentAccount = ''
}

function fillSupplierForm(supplier: Supplier) {
  supplierForm.name = supplier.name
  supplierForm.contactPerson = supplier.contactPerson
  supplierForm.phone = supplier.phone
  supplierForm.inn = supplier.inn
  supplierForm.kpp = supplier.kpp
  supplierForm.ogrn = supplier.ogrn
  supplierForm.legalAddress = supplier.legalAddress
  supplierForm.actualAddress = supplier.actualAddress
  supplierForm.bik = supplier.bik
  supplierForm.settlementAccount = supplier.settlementAccount
  supplierForm.correspondentAccount = supplier.correspondentAccount
}

function openCreateModal() {
  resetSupplierForm()
  isFormModalOpen.value = true
}

function openEditModal(supplier: Supplier) {
  editingSupplierId.value = supplier.id
  fillSupplierForm(supplier)
  isFormModalOpen.value = true
}

function closeFormModal() {
  isFormModalOpen.value = false
}

function handlePhoneInput(event: Event) {
  const target = event.target as HTMLInputElement | null
  if (!target) return
  supplierForm.phone = normalizePhone(target.value)
  target.value = supplierForm.phone
}

function handleDigitsInput(
  field: 'inn' | 'kpp' | 'ogrn' | 'bik' | 'settlementAccount' | 'correspondentAccount',
  event: Event,
) {
  const target = event.target as HTMLInputElement | null
  if (!target) return
  const sanitized = target.value.replace(/\D/g, '')
  supplierForm[field] = sanitized
  target.value = sanitized
}

function handleSubmitSupplier() {
  if (!canSubmitSupplier.value) return

  const payload = {
    name: supplierForm.name,
    contactPerson: supplierForm.contactPerson,
    phone: supplierForm.phone,
    inn: supplierForm.inn,
    kpp: supplierForm.kpp,
    ogrn: supplierForm.ogrn,
    legalAddress: supplierForm.legalAddress,
    actualAddress: supplierForm.actualAddress,
    bik: supplierForm.bik,
    settlementAccount: supplierForm.settlementAccount,
    correspondentAccount: supplierForm.correspondentAccount,
  }

  if (editingSupplierId.value) {
    updateSupplier(editingSupplierId.value, payload)
  } else {
    addSupplier(payload)
  }

  closeFormModal()
  resetSupplierForm()
}

function handleDeleteSupplier(supplier: Supplier) {
  supplierToDelete.value = supplier
  isDeleteModalOpen.value = true
}

function closeDeleteModal() {
  isDeleteModalOpen.value = false
  supplierToDelete.value = null
}

function confirmDeleteSupplier() {
  if (!supplierToDelete.value) return
  removeSupplier(supplierToDelete.value.id)
  if (editingSupplierId.value === supplierToDelete.value.id) {
    closeFormModal()
    resetSupplierForm()
  }
  closeDeleteModal()
}
</script>

<template>
  <div class="trade-subview">
    <SectionSubviewHeader title="Поставщики">
      <template #actions>
        <button
          type="button"
          class="trade-subview__create-btn"
          title="Добавить поставщика"
          aria-label="Добавить поставщика"
          @click="openCreateModal"
        >
          <NIcon :size="18" :component="AddOutline" />
        </button>
      </template>
    </SectionSubviewHeader>

    <div class="trade-subview__body">
      <section v-if="suppliers.length === 0" class="suppliers-view__placeholder">
        <p class="suppliers-view__placeholder-text">
          Пока нет поставщиков. Добавьте первого через кнопку «+» вверху.
        </p>
      </section>

      <section v-else class="suppliers-view__table-wrap">
        <div class="suppliers-view__table" role="table">
          <div class="suppliers-view__table-row suppliers-view__table-row--head" role="row">
            <span class="suppliers-view__cell suppliers-view__cell--head" role="columnheader">
              Наименование
            </span>
            <span class="suppliers-view__cell suppliers-view__cell--head" role="columnheader">
              Контактное лицо
            </span>
            <span
              class="suppliers-view__cell suppliers-view__cell--head suppliers-view__cell--compact"
              role="columnheader"
            >
              Телефон
            </span>
            <span class="suppliers-view__cell suppliers-view__cell--head" role="columnheader">
              Фактический адрес
            </span>
            <span
              class="suppliers-view__cell suppliers-view__cell--head suppliers-view__cell--actions"
              role="columnheader"
              aria-hidden="true"
            />
          </div>

          <div v-for="supplier in suppliers" :key="supplier.id" class="suppliers-view__table-row" role="row">
            <span class="suppliers-view__cell">{{ supplier.name }}</span>
            <span class="suppliers-view__cell">{{ supplier.contactPerson }}</span>
            <span class="suppliers-view__cell suppliers-view__cell--compact suppliers-view__cell--phone">
              {{ supplier.phone }}
            </span>
            <span class="suppliers-view__cell suppliers-view__cell--address">{{ supplier.actualAddress }}</span>
            <div class="suppliers-view__cell suppliers-view__cell--actions">
              <div class="suppliers-view__row-actions">
                <button
                  type="button"
                  class="suppliers-view__icon-action"
                  aria-label="Редактировать поставщика"
                  title="Редактировать"
                  @click="openEditModal(supplier)"
                >
                  <NIcon :size="16">
                    <PencilOutline />
                  </NIcon>
                </button>
                <button
                  type="button"
                  class="suppliers-view__icon-action suppliers-view__icon-action--danger"
                  aria-label="Удалить поставщика"
                  title="Удалить"
                  @click="handleDeleteSupplier(supplier)"
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
      width="xlarge"
      actions-align="center"
      :close-label="modalCloseLabel"
      @close="resetSupplierForm"
    >
      <div class="suppliers-modal__fields">
        <label class="suppliers-modal__field">
          <span class="suppliers-modal__label">
            Наименование
            <span class="suppliers-modal__required" aria-hidden="true">*</span>
          </span>
          <input
            v-model="supplierForm.name"
            type="text"
            class="suppliers-modal__input"
            placeholder="Введите наименование"
            autocomplete="off"
          />
        </label>

        <label class="suppliers-modal__field">
          <span class="suppliers-modal__label">
            Контактное лицо
            <span class="suppliers-modal__required" aria-hidden="true">*</span>
          </span>
          <input
            v-model="supplierForm.contactPerson"
            type="text"
            class="suppliers-modal__input"
            placeholder="Введите контактное лицо"
            autocomplete="off"
          />
        </label>

        <label class="suppliers-modal__field">
          <span class="suppliers-modal__label">
            Номер телефона
            <span class="suppliers-modal__required" aria-hidden="true">*</span>
          </span>
          <input
            :value="supplierForm.phone"
            type="tel"
            class="suppliers-modal__input"
            placeholder="+7"
            autocomplete="off"
            @input="handlePhoneInput"
          />
        </label>

        <h3 class="suppliers-modal__section-title">Реквизиты</h3>

        <label class="suppliers-modal__field">
          <span class="suppliers-modal__label">
            ИНН
            <span class="suppliers-modal__required" aria-hidden="true">*</span>
          </span>
          <input
            :value="supplierForm.inn"
            type="text"
            inputmode="numeric"
            class="suppliers-modal__input"
            placeholder="Введите ИНН"
            autocomplete="off"
            @input="handleDigitsInput('inn', $event)"
          />
        </label>

        <label class="suppliers-modal__field">
          <span class="suppliers-modal__label">КПП</span>
          <input
            :value="supplierForm.kpp"
            type="text"
            inputmode="numeric"
            class="suppliers-modal__input"
            placeholder="Введите КПП"
            autocomplete="off"
            @input="handleDigitsInput('kpp', $event)"
          />
        </label>

        <label class="suppliers-modal__field">
          <span class="suppliers-modal__label">ОГРН</span>
          <input
            :value="supplierForm.ogrn"
            type="text"
            inputmode="numeric"
            class="suppliers-modal__input"
            placeholder="Введите ОГРН"
            autocomplete="off"
            @input="handleDigitsInput('ogrn', $event)"
          />
        </label>

        <label class="suppliers-modal__field">
          <span class="suppliers-modal__label">
            БИК
            <span class="suppliers-modal__required" aria-hidden="true">*</span>
          </span>
          <input
            :value="supplierForm.bik"
            type="text"
            inputmode="numeric"
            class="suppliers-modal__input"
            placeholder="Введите БИК"
            autocomplete="off"
            @input="handleDigitsInput('bik', $event)"
          />
        </label>

        <label class="suppliers-modal__field">
          <span class="suppliers-modal__label">
            Расчётный счёт
            <span class="suppliers-modal__required" aria-hidden="true">*</span>
          </span>
          <input
            :value="supplierForm.settlementAccount"
            type="text"
            inputmode="numeric"
            class="suppliers-modal__input"
            placeholder="Введите расчётный счёт"
            autocomplete="off"
            @input="handleDigitsInput('settlementAccount', $event)"
          />
        </label>

        <label class="suppliers-modal__field">
          <span class="suppliers-modal__label">
            Корреспондентский счёт
            <span class="suppliers-modal__required" aria-hidden="true">*</span>
          </span>
          <input
            :value="supplierForm.correspondentAccount"
            type="text"
            inputmode="numeric"
            class="suppliers-modal__input"
            placeholder="Введите корреспондентский счёт"
            autocomplete="off"
            @input="handleDigitsInput('correspondentAccount', $event)"
          />
        </label>

        <label class="suppliers-modal__field suppliers-modal__field--full">
          <span class="suppliers-modal__label">
            Юридический адрес
            <span class="suppliers-modal__required" aria-hidden="true">*</span>
          </span>
          <input
            v-model="supplierForm.legalAddress"
            type="text"
            class="suppliers-modal__input"
            placeholder="Введите юридический адрес"
            autocomplete="off"
          />
        </label>

        <label class="suppliers-modal__field suppliers-modal__field--full">
          <span class="suppliers-modal__label">
            Фактический адрес
            <span class="suppliers-modal__required" aria-hidden="true">*</span>
          </span>
          <input
            v-model="supplierForm.actualAddress"
            type="text"
            class="suppliers-modal__input"
            placeholder="Введите фактический адрес"
            autocomplete="off"
          />
        </label>
      </div>

      <template #actions>
        <AppModalButton
          :disabled="!canSubmitSupplier"
          :title="canSubmitSupplier ? '' : 'Заполните обязательные поля'"
          @click="handleSubmitSupplier"
        >
          {{ submitButtonLabel }}
        </AppModalButton>
      </template>
    </AppModal>

    <AppModal
      v-model:show="isDeleteModalOpen"
      title="Удаление поставщика"
      body-variant="center"
      @close="closeDeleteModal"
    >
      <p class="app-modal__message">Вы уверены, что хотите удалить данного поставщика?</p>

      <template #actions>
        <div class="suppliers-view__confirm-actions">
          <AppModalButton @click="confirmDeleteSupplier">Да</AppModalButton>
          <button type="button" class="suppliers-view__confirm-cancel" @click="closeDeleteModal">
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

.suppliers-view__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 240px;
  padding: 32px 24px;
  border: 1px dashed #cbd5e1;
  border-radius: 12px;
  background: #f8fafc;
}

.suppliers-view__placeholder-text {
  margin: 0;
  max-width: 420px;
  font-size: 15px;
  line-height: 1.5;
  color: #64748b;
  text-align: center;
}

.suppliers-view__table-wrap {
  min-width: 0;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
  overflow-x: auto;
}

.suppliers-view__table {
  display: grid;
  width: 100%;
  min-width: 780px;
  grid-template-columns:
    minmax(140px, 1.2fr)
    minmax(120px, 0.9fr)
    max-content
    minmax(180px, 1.4fr)
    max-content;
}

.suppliers-view__table-row {
  display: contents;
}

.suppliers-view__cell {
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

.suppliers-view__cell--head {
  min-height: 44px;
  padding: 10px 14px;
  background: #f8fafc;
  font-size: 11px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.suppliers-view__table-row .suppliers-view__cell:nth-child(5) {
  border-right: 0;
}

.suppliers-view__table-row:last-child .suppliers-view__cell {
  border-bottom: 0;
}

.suppliers-view__cell--compact {
  padding-left: 10px;
  padding-right: 10px;
  white-space: nowrap;
}

.suppliers-view__cell--head.suppliers-view__cell--compact {
  padding-left: 10px;
  padding-right: 10px;
}

.suppliers-view__cell--phone {
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.01em;
}

.suppliers-view__cell--address {
  min-width: 0;
  overflow-wrap: anywhere;
}

.suppliers-view__cell--actions {
  justify-content: center;
  padding: 10px;
  white-space: nowrap;
}

.suppliers-view__row-actions {
  display: inline-flex;
  justify-content: center;
  gap: 6px;
}

.suppliers-view__icon-action {
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

.suppliers-view__icon-action:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #334155;
}

.suppliers-view__icon-action--danger:hover {
  color: #dc2626;
}

.suppliers-view__table-row:not(.suppliers-view__table-row--head):hover .suppliers-view__cell {
  background: #f8fafc;
}

.suppliers-view__confirm-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px;
}

.suppliers-view__confirm-cancel {
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

.suppliers-view__confirm-cancel:hover {
  background: #f8fafc;
  border-color: #94a3b8;
  color: #334155;
}

.suppliers-modal__fields {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px 16px;
}

.suppliers-modal__section-title {
  grid-column: 1 / -1;
  margin: 2px 0 0;
  font-size: 13px;
  font-weight: 600;
  color: #1a202c;
}

.suppliers-modal__field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.suppliers-modal__field--full {
  grid-column: 1 / -1;
}

.suppliers-modal__label {
  font-size: 13px;
  color: #475569;
}

.suppliers-modal__required {
  color: #dc2626;
  margin-left: 2px;
}

.suppliers-modal__input {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  color: #0f172a;
  padding: 7px 10px;
  font-size: 14px;
  font-family: inherit;
}

.suppliers-modal__input:focus {
  outline: none;
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.25);
}

@media (max-width: 900px) {
  .suppliers-modal__fields {
    grid-template-columns: 1fr;
  }
}
</style>
