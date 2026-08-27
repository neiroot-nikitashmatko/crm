export interface OutgoingInvoiceItem {
  catalogProductId?: string
  title: string
  quantity: number
  unitPrice: number
}

export interface OutgoingInvoice {
  id: string
  invoiceNumber: number
  date: number
  dealId: string
  items: OutgoingInvoiceItem[]
  total: number
  comment: string
  createdAt: number
}

export type OutgoingInvoiceInput = Omit<OutgoingInvoice, 'id' | 'invoiceNumber' | 'createdAt'>
