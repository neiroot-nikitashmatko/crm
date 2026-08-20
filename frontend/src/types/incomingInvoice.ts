export interface IncomingInvoiceItem {
  catalogProductId?: string
  title: string
  quantity: number
  unitPrice: number
}

export interface IncomingInvoice {
  id: string
  invoiceNumber: number
  date: number
  supplierId: string
  items: IncomingInvoiceItem[]
  total: number
  comment: string
  createdAt: number
}

export type IncomingInvoiceInput = Omit<IncomingInvoice, 'id' | 'invoiceNumber' | 'createdAt'>
