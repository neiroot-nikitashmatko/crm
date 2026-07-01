import { computed, ref } from 'vue'
import {
  createCatalogProduct,
  deleteCatalogProduct,
  fetchCatalogProducts,
  updateCatalogProduct,
} from '@/api/catalogProducts'
import type { CatalogProduct, NewCatalogProductInput } from '@/types/productCatalog'

const products = ref<CatalogProduct[]>([])
const isLoaded = ref(false)
const isLoading = ref(false)

function normalizeInput(input: NewCatalogProductInput): NewCatalogProductInput {
  return {
    name: input.name.trim(),
    sku: input.sku.trim(),
    category: input.category.trim(),
    cost: Number.isFinite(input.cost) ? Math.max(0, input.cost) : 0,
  }
}

export function useProductsCatalog() {
  const categoryGroups = computed(() => {
    const groupsMap = new Map<string, CatalogProduct[]>()

    for (const product of products.value) {
      const category = product.category.trim()
      if (!category) continue

      if (!groupsMap.has(category)) {
        groupsMap.set(category, [])
      }
      groupsMap.get(category)?.push(product)
    }

    return Array.from(groupsMap.entries())
      .map(([category, items]) => ({
        category,
        products: [...items].sort((left, right) => left.name.localeCompare(right.name, 'ru')),
      }))
      .sort((left, right) => left.category.localeCompare(right.category, 'ru'))
  })

  const uncategorizedProducts = computed(() =>
    products.value
      .filter((product) => product.category.trim().length === 0)
      .sort((left, right) => left.name.localeCompare(right.name, 'ru')),
  )

  const catalogProductOptions = computed(() =>
    [...products.value]
      .sort((left, right) => left.name.localeCompare(right.name, 'ru'))
      .map((product) => ({
        label: product.name,
        value: product.id,
        sku: product.sku,
      })),
  )

  const hasCatalogProducts = computed(() => products.value.length > 0)

  function getCatalogProductById(id: string) {
    return products.value.find((product) => product.id === id) ?? null
  }

  async function loadCatalog(force = false) {
    if (isLoading.value) return
    if (isLoaded.value && !force) return

    isLoading.value = true
    try {
      products.value = await fetchCatalogProducts()
      isLoaded.value = true
    } finally {
      isLoading.value = false
    }
  }

  async function addProduct(input: NewCatalogProductInput): Promise<CatalogProduct> {
    const createdProduct = await createCatalogProduct(normalizeInput(input))
    products.value = [...products.value, createdProduct].sort((left, right) =>
      left.name.localeCompare(right.name, 'ru'),
    )
    return createdProduct
  }

  async function updateProduct(
    id: string,
    input: NewCatalogProductInput,
  ): Promise<CatalogProduct | null> {
    const updatedProduct = await updateCatalogProduct(id, normalizeInput(input))
    products.value = products.value
      .map((product) => (product.id === id ? updatedProduct : product))
      .sort((left, right) => left.name.localeCompare(right.name, 'ru'))
    return updatedProduct
  }

  async function deleteProduct(id: string): Promise<boolean> {
    await deleteCatalogProduct(id)
    const nextProducts = products.value.filter((product) => product.id !== id)
    if (nextProducts.length === products.value.length) return false
    products.value = nextProducts
    return true
  }

  return {
    products,
    isLoaded,
    isLoading,
    categoryGroups,
    uncategorizedProducts,
    catalogProductOptions,
    hasCatalogProducts,
    getCatalogProductById,
    loadCatalog,
    addProduct,
    updateProduct,
    deleteProduct,
  }
}
