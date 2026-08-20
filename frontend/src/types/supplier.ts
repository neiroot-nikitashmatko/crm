export interface Supplier {
  id: string
  name: string
  contactPerson: string
  phone: string
  inn: string
  kpp: string
  ogrn: string
  legalAddress: string
  actualAddress: string
  bik: string
  settlementAccount: string
  correspondentAccount: string
  createdAt: number
}

export type SupplierInput = Omit<Supplier, 'id' | 'createdAt'>
