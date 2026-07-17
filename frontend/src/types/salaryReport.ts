export interface SalaryReportEntry {
  id: string
  date: number
  dealId: string
  dealNumberLabel: string
  service: string
  serviceLabel: string
  salary: string
  comment: string
  createdBy: string
}

export type CreateSalaryReportEntryInput = {
  date: number
  dealId: string
  service: string
  salary: string
  comment: string
  employeeId?: string
}
