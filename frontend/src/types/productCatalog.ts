export interface CatalogProduct {
  id: string
  name: string
  sku: string
  category: string
  cost: number
  createdAt: number
}

export interface NewCatalogProductInput {
  name: string
  sku: string
  category: string
  cost: number
}
