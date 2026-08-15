import { ApiError, requestJson } from '@/api/httpClient'
import type { AnalyticsDateRange, ClosedDealListItem, EmployeeShareCount, FailedDealShare, FailedLeadShare, LeadToDealConversion, ProductionCategoryCount, TrafficSourceCount } from '@/types/analytics'

interface TrafficSourceResponse {
  items: TrafficSourceCount[]
}

export class AnalyticsApiError extends ApiError {
  constructor(message: string, status: number) {
    super(message, status)
    this.name = 'AnalyticsApiError'
  }
}

async function analyticsRequestJson<T>(path: string, init?: RequestInit): Promise<T> {
  try {
    return await requestJson<T>(path, init)
  } catch (error) {
    if (error instanceof ApiError) {
      throw new AnalyticsApiError(error.message, error.status)
    }
    throw error
  }
}

function normalizeTrafficSourceCounts(items: TrafficSourceCount[] | undefined): TrafficSourceCount[] {
  if (!Array.isArray(items)) return []
  return items.map((item) => ({
    source: String(item.source ?? ''),
    count: Number(item.count ?? 0),
  }))
}

async function fetchTrafficSourceMetrics(
  path: string,
  range: AnalyticsDateRange,
): Promise<TrafficSourceCount[]> {
  const [from, to] = range
  const params = new URLSearchParams({
    from: String(from),
    to: String(to),
  })
  const payload = await analyticsRequestJson<TrafficSourceResponse>(
    `${path}?${params.toString()}`,
    { method: 'GET' },
  )

  return normalizeTrafficSourceCounts(payload.items)
}

export function fetchLeadTrafficMetrics(range: AnalyticsDateRange): Promise<TrafficSourceCount[]> {
  return fetchTrafficSourceMetrics('/api/v1/analytics/leads-traffic', range)
}

export function fetchDealTrafficMetrics(range: AnalyticsDateRange): Promise<TrafficSourceCount[]> {
  return fetchTrafficSourceMetrics('/api/v1/analytics/deals-traffic', range)
}

export async function fetchLeadToDealConversion(
  range: AnalyticsDateRange,
): Promise<LeadToDealConversion> {
  const [from, to] = range
  const params = new URLSearchParams({
    from: String(from),
    to: String(to),
  })
  const payload = await analyticsRequestJson<{ item: LeadToDealConversion }>(
    `/api/v1/analytics/lead-to-deal-conversion?${params.toString()}`,
    { method: 'GET' },
  )

  return {
    leadsCount: Number(payload.item?.leadsCount ?? 0),
    convertedCount: Number(payload.item?.convertedCount ?? 0),
    percent: Number(payload.item?.percent ?? 0),
  }
}

export async function fetchFailedLeadShare(range: AnalyticsDateRange): Promise<FailedLeadShare> {
  const [from, to] = range
  const params = new URLSearchParams({
    from: String(from),
    to: String(to),
  })
  const payload = await analyticsRequestJson<{ item: FailedLeadShare }>(
    `/api/v1/analytics/failed-lead-share?${params.toString()}`,
    { method: 'GET' },
  )

  return {
    leadsCount: Number(payload.item?.leadsCount ?? 0),
    failedCount: Number(payload.item?.failedCount ?? 0),
    percent: Number(payload.item?.percent ?? 0),
  }
}

export async function fetchFailedDealShare(range: AnalyticsDateRange): Promise<FailedDealShare> {
  const [from, to] = range
  const params = new URLSearchParams({
    from: String(from),
    to: String(to),
  })
  const payload = await analyticsRequestJson<{ item: FailedDealShare }>(
    `/api/v1/analytics/failed-deal-share?${params.toString()}`,
    { method: 'GET' },
  )

  return {
    dealsCount: Number(payload.item?.dealsCount ?? 0),
    failedCount: Number(payload.item?.failedCount ?? 0),
    percent: Number(payload.item?.percent ?? 0),
  }
}

export async function fetchClosedDealsProductionShare(
  range: AnalyticsDateRange,
): Promise<ProductionCategoryCount[]> {
  const [from, to] = range
  const params = new URLSearchParams({
    from: String(from),
    to: String(to),
  })
  const payload = await analyticsRequestJson<{ items: ProductionCategoryCount[] }>(
    `/api/v1/analytics/closed-deals-production-share?${params.toString()}`,
    { method: 'GET' },
  )

  if (!Array.isArray(payload.items)) return []
  return payload.items.map((item) => ({
    category: String(item.category ?? ''),
    count: Number(item.count ?? 0),
  }))
}

export async function fetchClosedDealsEmployeeShare(
  range: AnalyticsDateRange,
): Promise<EmployeeShareCount[]> {
  const [from, to] = range
  const params = new URLSearchParams({
    from: String(from),
    to: String(to),
  })
  const payload = await analyticsRequestJson<{ items: EmployeeShareCount[] }>(
    `/api/v1/analytics/closed-deals-employee-share?${params.toString()}`,
    { method: 'GET' },
  )

  if (!Array.isArray(payload.items)) return []
  return payload.items.map((item) => ({
    employee: String(item.employee ?? ''),
    count: Number(item.count ?? 0),
  }))
}

export async function fetchClosedDealsList(
  range: AnalyticsDateRange,
  options?: { requireEmployee?: boolean; requireProduction?: boolean },
): Promise<ClosedDealListItem[]> {
  const [from, to] = range
  const params = new URLSearchParams({
    from: String(from),
    to: String(to),
  })
  if (options?.requireEmployee) {
    params.set('requireEmployee', '1')
  }
  if (options?.requireProduction) {
    params.set('requireProduction', '1')
  }
  const payload = await analyticsRequestJson<{ items: ClosedDealListItem[] }>(
    `/api/v1/analytics/closed-deals?${params.toString()}`,
    { method: 'GET' },
  )

  if (!Array.isArray(payload.items)) return []
  return payload.items.map(mapClosedDealListItem)
}

export async function fetchFailedDealsList(
  range: AnalyticsDateRange,
): Promise<ClosedDealListItem[]> {
  const [from, to] = range
  const params = new URLSearchParams({
    from: String(from),
    to: String(to),
  })
  const payload = await analyticsRequestJson<{ items: ClosedDealListItem[] }>(
    `/api/v1/analytics/failed-deals?${params.toString()}`,
    { method: 'GET' },
  )

  if (!Array.isArray(payload.items)) return []
  return payload.items.map(mapClosedDealListItem)
}

function mapClosedDealListItem(item: ClosedDealListItem): ClosedDealListItem {
  return {
    id: String(item.id ?? ''),
    dealNumber: Number(item.dealNumber ?? 0),
    firstName: String(item.firstName ?? ''),
    patronymic: String(item.patronymic ?? ''),
    phone: String(item.phone ?? ''),
    nomenclature: String(item.nomenclature ?? ''),
    category: String(item.category ?? ''),
    employee: String(item.employee ?? ''),
    failureReason: String(item.failureReason ?? ''),
    createdAt: Number(item.createdAt ?? 0),
    closedAt: Number(item.closedAt ?? 0),
  }
}
