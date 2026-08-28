export interface TrafficSourceMetric {
  source: string
  count: number
  color: string
}

export interface TrafficSourceCount {
  source: string
  count: number
}

export interface LeadToDealConversion {
  leadsCount: number
  convertedCount: number
  percent: number
}

export interface FailedLeadShare {
  leadsCount: number
  failedCount: number
  percent: number
}

export interface FailedDealShare {
  dealsCount: number
  failedCount: number
  percent: number
}

export interface ProductionCategoryCount {
  category: string
  count: number
}

export interface EmployeeShareCount {
  employee: string
  count: number
}

export interface ClosedDealListItem {
  id: string
  dealNumber: number
  firstName: string
  patronymic: string
  phone: string
  nomenclature: string
  category: string
  employee: string
  failureReason: string
  createdAt: number
  closedAt: number
}

export interface FailedLeadListItem {
  id: string
  leadNumber: number
  firstName: string
  patronymic: string
  phone: string
  failureReason: string
  createdAt: number
}

export interface DealTrafficListItem {
  id: string
  dealNumber: number
  firstName: string
  patronymic: string
  phone: string
  trafficSource: string
  createdAt: number
}

export type AnalyticsDateRange = [number, number]

export type AnalyticsPeriodPreset = 'today' | 'month' | 'custom'
