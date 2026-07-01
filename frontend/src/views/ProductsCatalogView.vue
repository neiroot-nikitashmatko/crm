<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { NIcon } from 'naive-ui'
import { PencilOutline, TrashOutline } from '@vicons/ionicons5'
import ProductsCatalogSectionHeader from '@/components/products/ProductsCatalogSectionHeader.vue'
import AppModal from '@/components/common/AppModal.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import { useProductsCatalog } from '@/composables/useProductsCatalog'
import { CatalogProductsApiError } from '@/api/catalogProducts'
import type { CatalogProduct } from '@/types/productCatalog'

const { categoryGroups, uncategorizedProducts, addProduct, updateProduct, deleteProduct, loadCatalog } =
  useProductsCatalog()
const UNCATEGORIZED_KEY = '__uncategorized__'
const isProductModalOpen = ref(false)
const editingProductId = ref<string | null>(null)
const expandedCategories = reactive<Record<string, boolean>>({})
const productForm = reactive({
  name: '',
  sku: '',
  category: '',
  cost: '',
})

const isEditMode = computed(() => editingProductId.value !== null)
const modalTitle = computed(() => (isEditMode.value ? 'Редактировать товар' : 'Добавить товар'))
const submitButtonLabel = computed(() => (isEditMode.value ? 'Сохранить' : 'Добавить'))

const hasAnyProducts = computed(
  () => categoryGroups.value.length > 0 || uncategorizedProducts.value.length > 0,
)
const canSubmitProduct = computed(
  () =>
    productForm.name.trim().length > 0 &&
    productForm.sku.trim().length > 0 &&
    /^[0-9]+$/.test(productForm.cost),
)

function openCreateModal() {
  editingProductId.value = null
  resetProductForm()
  isProductModalOpen.value = true
}

function openEditModal(product: CatalogProduct) {
  editingProductId.value = product.id
  productForm.name = product.name
  productForm.sku = product.sku
  productForm.category = product.category
  productForm.cost = String(product.cost)
  isProductModalOpen.value = true
}

function closeProductModal() {
  isProductModalOpen.value = false
}

function resetProductForm() {
  editingProductId.value = null
  productForm.name = ''
  productForm.sku = ''
  productForm.category = ''
  productForm.cost = ''
}

function handleCostInput(event: Event) {
  const target = event.target as HTMLInputElement | null
  if (!target) return
  const sanitized = target.value.replace(/[^0-9]/g, '')
  productForm.cost = sanitized
  target.value = sanitized
}

async function handleSubmitProduct() {
  if (!canSubmitProduct.value) return

  const payload = {
    name: productForm.name,
    sku: productForm.sku,
    category: productForm.category,
    cost: Number(productForm.cost),
  }

  try {
    if (editingProductId.value) {
      await updateProduct(editingProductId.value, payload)
    } else {
      await addProduct(payload)
    }
    closeProductModal()
    resetProductForm()
  } catch (error) {
    window.alert(getErrorMessage(error))
  }
}

async function handleDeleteProduct(product: CatalogProduct) {
  try {
    await deleteProduct(product.id)
    if (editingProductId.value === product.id) {
      closeProductModal()
      resetProductForm()
    }
  } catch (error) {
    window.alert(getErrorMessage(error))
  }
}

function toggleCategory(category: string) {
  expandedCategories[category] = !expandedCategories[category]
}

function isCategoryExpanded(category: string) {
  return expandedCategories[category] === true
}

function formatCost(cost: number) {
  return `${new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 0 }).format(cost)} ₽`
}

function rowKey(product: CatalogProduct) {
  return product.id
}

function getErrorMessage(error: unknown) {
  if (error instanceof CatalogProductsApiError) return error.message
  if (error instanceof Error) return error.message
  return 'Не удалось выполнить операцию'
}

onMounted(() => {
  void loadCatalog(true)
})
</script>

<template>
  <div class="products-catalog-page">
    <ProductsCatalogSectionHeader @add-product="openCreateModal" />

    <div class="products-catalog-page__body">
      <section v-if="!hasAnyProducts" class="products-catalog-page__placeholder">
        <p class="products-catalog-page__text">
          Товары пока не добавлены. Нажмите «Добавить товар», чтобы создать первую позицию.
        </p>
      </section>

      <section v-else class="products-catalog-page__content">
        <article
          v-for="group in categoryGroups"
          :key="group.category"
          class="products-catalog-page__category"
        >
          <button
            type="button"
            class="products-catalog-page__category-header"
            @click="toggleCategory(group.category)"
          >
            <span class="products-catalog-page__category-icon">
              {{ isCategoryExpanded(group.category) ? '📂' : '📁' }}
            </span>
            <span class="products-catalog-page__category-name">{{ group.category }}</span>
            <span class="products-catalog-page__category-count">{{ group.products.length }}</span>
          </button>

          <div v-if="isCategoryExpanded(group.category)" class="products-catalog-page__table">
            <div class="products-catalog-page__table-head">
              <span>Наименование</span>
              <span>Артикул</span>
              <span>Категория</span>
              <span>Стоимость</span>
              <span class="products-catalog-page__table-head-action" aria-hidden="true" />
            </div>
            <div
              v-for="product in group.products"
              :key="rowKey(product)"
              class="products-catalog-page__table-row"
            >
              <span>{{ product.name }}</span>
              <span>{{ product.sku }}</span>
              <span>{{ product.category || '—' }}</span>
              <span class="products-catalog-page__cost">{{ formatCost(product.cost) }}</span>
              <div class="products-catalog-page__row-action">
                <button
                  type="button"
                  class="products-catalog-page__icon-action"
                  aria-label="Редактировать товар"
                  @click="openEditModal(product)"
                >
                  <NIcon :size="16">
                    <PencilOutline />
                  </NIcon>
                </button>
                <button
                  type="button"
                  class="products-catalog-page__icon-action products-catalog-page__icon-action--danger"
                  aria-label="Удалить товар"
                  @click="handleDeleteProduct(product)"
                >
                  <NIcon :size="16">
                    <TrashOutline />
                  </NIcon>
                </button>
              </div>
            </div>
          </div>
        </article>

        <article
          v-if="uncategorizedProducts.length > 0"
          class="products-catalog-page__category"
        >
          <button
            type="button"
            class="products-catalog-page__category-header"
            @click="toggleCategory(UNCATEGORIZED_KEY)"
          >
            <span class="products-catalog-page__category-icon">
              {{ isCategoryExpanded(UNCATEGORIZED_KEY) ? '📂' : '📁' }}
            </span>
            <span class="products-catalog-page__category-name">Товары без категории</span>
            <span class="products-catalog-page__category-count">{{ uncategorizedProducts.length }}</span>
          </button>

          <div v-if="isCategoryExpanded(UNCATEGORIZED_KEY)" class="products-catalog-page__table">
            <div class="products-catalog-page__table-head">
              <span>Наименование</span>
              <span>Артикул</span>
              <span>Категория</span>
              <span>Стоимость</span>
              <span class="products-catalog-page__table-head-action" aria-hidden="true" />
            </div>
            <div
              v-for="product in uncategorizedProducts"
              :key="rowKey(product)"
              class="products-catalog-page__table-row"
            >
              <span>{{ product.name }}</span>
              <span>{{ product.sku }}</span>
              <span>—</span>
              <span class="products-catalog-page__cost">{{ formatCost(product.cost) }}</span>
              <div class="products-catalog-page__row-action">
                <button
                  type="button"
                  class="products-catalog-page__icon-action"
                  aria-label="Редактировать товар"
                  @click="openEditModal(product)"
                >
                  <NIcon :size="16">
                    <PencilOutline />
                  </NIcon>
                </button>
                <button
                  type="button"
                  class="products-catalog-page__icon-action products-catalog-page__icon-action--danger"
                  aria-label="Удалить товар"
                  @click="handleDeleteProduct(product)"
                >
                  <NIcon :size="16">
                    <TrashOutline />
                  </NIcon>
                </button>
              </div>
            </div>
          </div>
        </article>
      </section>
    </div>
  </div>

  <AppModal
    v-model:show="isProductModalOpen"
    :title="modalTitle"
    width="wide"
    actions-align="end"
    :close-label="isEditMode ? 'Закрыть окно редактирования товара' : 'Закрыть окно добавления товара'"
    @close="resetProductForm"
  >
    <div class="products-catalog-page__modal-fields">
      <label class="products-catalog-page__field">
        <span class="products-catalog-page__label">Наименование</span>
        <input
          v-model="productForm.name"
          type="text"
          class="products-catalog-page__input"
          placeholder="Введите наименование"
        />
      </label>

      <label class="products-catalog-page__field">
        <span class="products-catalog-page__label">Артикул</span>
        <input
          v-model="productForm.sku"
          type="text"
          class="products-catalog-page__input"
          placeholder="Введите артикул"
        />
      </label>

      <label class="products-catalog-page__field">
        <span class="products-catalog-page__label">Категория</span>
        <input
          v-model="productForm.category"
          type="text"
          class="products-catalog-page__input"
          placeholder="Введите категорию (опционально)"
        />
      </label>

      <label class="products-catalog-page__field">
        <span class="products-catalog-page__label">Стоимость</span>
        <input
          v-model="productForm.cost"
          type="text"
          inputmode="numeric"
          autocomplete="off"
          class="products-catalog-page__input"
          placeholder="Введите стоимость"
          @input="handleCostInput"
        />
      </label>
    </div>

    <template #actions>
      <AppModalButton :disabled="!canSubmitProduct" @click="handleSubmitProduct">
        {{ submitButtonLabel }}
      </AppModalButton>
    </template>
  </AppModal>
</template>

<style scoped>
.products-catalog-page {
  display: flex;
  flex-direction: column;
  height: calc(100dvh - 64px);
  max-height: calc(100dvh - 64px);
  overflow: hidden;
  background: #ffffff;
}

.products-catalog-page__body {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  padding: 16px 24px;
  box-sizing: border-box;
}

.products-catalog-page__placeholder {
  width: min(680px, 100%);
  border: 1px dashed #cbd5e1;
  border-radius: 12px;
  padding: 24px;
  background: #f8fafc;
  text-align: center;
  margin: 32px auto 0;
}

.products-catalog-page__text {
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  color: #4a5568;
}

.products-catalog-page__content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.products-catalog-page__category {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #ffffff;
  overflow: hidden;
}

.products-catalog-page__category-header {
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

.products-catalog-page__category-icon {
  width: 18px;
}

.products-catalog-page__category-name {
  font-size: 14px;
  font-weight: 600;
  color: #1a202c;
}

.products-catalog-page__category-count {
  margin-left: auto;
  min-width: 24px;
  padding: 0 6px;
  border-radius: 999px;
  background: #e2e8f0;
  color: #4a5568;
  font-size: 12px;
  text-align: center;
}

.products-catalog-page__table {
  display: flex;
  flex-direction: column;
  border-top: 1px solid #e2e8f0;
}

.products-catalog-page__table-head,
.products-catalog-page__table-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 160px 190px 120px 104px;
  gap: 0;
  align-items: stretch;
}

.products-catalog-page__table-head > *,
.products-catalog-page__table-row > * {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  min-height: 52px;
  padding: 10px 12px;
  border-right: 1px solid #e2e8f0;
  border-bottom: 1px solid #e2e8f0;
}

.products-catalog-page__table-head > * {
  background: #f8fafc;
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
}

.products-catalog-page__table-row {
  font-size: 14px;
  color: #1a202c;
}

.products-catalog-page__table-head > *:last-child,
.products-catalog-page__table-row > *:last-child {
  border-right: 0;
}

.products-catalog-page__cost {
  justify-content: flex-end;
  text-align: right;
}

.products-catalog-page__row-action {
  display: flex;
  justify-content: center;
  gap: 8px;
  flex-shrink: 0;
}

.products-catalog-page__icon-action {
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

.products-catalog-page__icon-action:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #334155;
}

.products-catalog-page__icon-action--danger:hover {
  color: #dc2626;
}

.products-catalog-page__modal-fields {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.products-catalog-page__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.products-catalog-page__label {
  font-size: 13px;
  color: #475569;
}

.products-catalog-page__input {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  color: #0f172a;
  padding: 8px 10px;
  font-size: 14px;
  font-family: inherit;
}

.products-catalog-page__input:focus {
  outline: none;
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.25);
}
</style>
