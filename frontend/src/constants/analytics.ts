import {
  ALL_TRAFFIC_SOURCES,
  AVITO_CHAT_TRAFFIC_SOURCE,
  BEELINE_TRAFFIC_SOURCE,
  MANUAL_TRAFFIC_SOURCES,
} from '@/constants/trafficSources'
import {
  PRODUCTION_EMPLOYEE_NAMES,
  PRODUCTION_SHARE_CATEGORIES,
  PRODUCTION_SHARE_OTHER_CATEGORY,
} from '@/constants/production'
import type {
  EmployeeShareCount,
  ProductionCategoryCount,
  TrafficSourceCount,
  TrafficSourceMetric,
} from '@/types/analytics'

const FALLBACK_TRAFFIC_SOURCE_COLORS = [
  '#2563eb',
  '#0f766e',
  '#d97706',
  '#e11d48',
  '#7c3aed',
  '#0891b2',
  '#65a30d',
  '#ea580c',
  '#4f46e5',
  '#be123c',
  '#059669',
  '#ca8a04',
  '#9333ea',
] as const

export const TRAFFIC_SOURCE_COLORS: Record<string, string> = {
  [AVITO_CHAT_TRAFFIC_SOURCE]: '#2563eb',
  [MANUAL_TRAFFIC_SOURCES[1]]: '#0f766e',
  [MANUAL_TRAFFIC_SOURCES[7]]: '#d97706',
  [MANUAL_TRAFFIC_SOURCES[4]]: '#e11d48',
  [MANUAL_TRAFFIC_SOURCES[8]]: '#7c3aed',
  [MANUAL_TRAFFIC_SOURCES[9]]: '#0891b2',
  [MANUAL_TRAFFIC_SOURCES[5]]: '#65a30d',
  [MANUAL_TRAFFIC_SOURCES[10]]: '#ea580c',
  [MANUAL_TRAFFIC_SOURCES[0]]: '#4f46e5',
  [MANUAL_TRAFFIC_SOURCES[2]]: '#be123c',
  [MANUAL_TRAFFIC_SOURCES[3]]: '#059669',
  [MANUAL_TRAFFIC_SOURCES[6]]: '#ca8a04',
  [MANUAL_TRAFFIC_SOURCES[11]]: '#9333ea',
  [BEELINE_TRAFFIC_SOURCE]: '#475569',
  'Без источника': '#94a3b8',
}

const TRAFFIC_SOURCE_ORDER = new Map<string, number>(
  ALL_TRAFFIC_SOURCES.map((source, index) => [source, index]),
)

function hashSource(source: string): number {
  let hash = 0
  for (let index = 0; index < source.length; index += 1) {
    hash = (hash * 31 + source.charCodeAt(index)) >>> 0
  }
  return hash
}

export function colorForTrafficSource(source: string): string {
  const mapped = TRAFFIC_SOURCE_COLORS[source]
  if (mapped) return mapped
  return FALLBACK_TRAFFIC_SOURCE_COLORS[hashSource(source) % FALLBACK_TRAFFIC_SOURCE_COLORS.length]
}

function compareTrafficMetrics(left: TrafficSourceMetric, right: TrafficSourceMetric): number {
  if (right.count !== left.count) return right.count - left.count

  const leftOrder = TRAFFIC_SOURCE_ORDER.get(left.source) ?? Number.MAX_SAFE_INTEGER
  const rightOrder = TRAFFIC_SOURCE_ORDER.get(right.source) ?? Number.MAX_SAFE_INTEGER
  if (leftOrder !== rightOrder) return leftOrder - rightOrder

  return left.source.localeCompare(right.source, 'ru')
}

export function buildTrafficSourceMetrics(
  items: readonly TrafficSourceCount[],
): TrafficSourceMetric[] {
  const counts = new Map(items.map((item) => [item.source, item.count]))
  const knownSources = new Set<string>(ALL_TRAFFIC_SOURCES)

  const metrics = ALL_TRAFFIC_SOURCES.map((source) => ({
    source,
    count: counts.get(source) ?? 0,
    color: colorForTrafficSource(source),
  }))

  for (const item of items) {
    if (knownSources.has(item.source)) continue
    metrics.push({
      source: item.source,
      count: item.count,
      color: colorForTrafficSource(item.source),
    })
  }

  return metrics.sort(compareTrafficMetrics)
}

export const PRODUCTION_CATEGORY_COLORS: Record<string, string> = {
  Перетяжка: '#d97706',
  'Установка чехлов': '#0f766e',
  Стёкла: '#2563eb',
  Коврики: '#65a30d',
  [PRODUCTION_SHARE_OTHER_CATEGORY]: '#94a3b8',
}

const PRODUCTION_CATEGORY_ORDER = new Map<string, number>(
  PRODUCTION_SHARE_CATEGORIES.map((category, index) => [category, index]),
)

function compareProductionMetrics(left: TrafficSourceMetric, right: TrafficSourceMetric): number {
  if (right.count !== left.count) return right.count - left.count

  const leftOrder = PRODUCTION_CATEGORY_ORDER.get(left.source) ?? Number.MAX_SAFE_INTEGER
  const rightOrder = PRODUCTION_CATEGORY_ORDER.get(right.source) ?? Number.MAX_SAFE_INTEGER
  if (leftOrder !== rightOrder) return leftOrder - rightOrder

  return left.source.localeCompare(right.source, 'ru')
}

export function buildProductionShareMetrics(
  items: readonly ProductionCategoryCount[],
): TrafficSourceMetric[] {
  const counts = new Map(items.map((item) => [item.category, item.count]))
  const knownCategories = new Set<string>(PRODUCTION_SHARE_CATEGORIES)

  const metrics = PRODUCTION_SHARE_CATEGORIES.map((category) => ({
    source: category,
    count: counts.get(category) ?? 0,
    color: PRODUCTION_CATEGORY_COLORS[category] ?? colorForTrafficSource(category),
  }))

  for (const item of items) {
    if (knownCategories.has(item.category)) continue
    metrics.push({
      source: item.category,
      count: item.count,
      color: PRODUCTION_CATEGORY_COLORS[item.category] ?? colorForTrafficSource(item.category),
    })
  }

  return metrics.sort(compareProductionMetrics)
}

export const PRODUCTION_EMPLOYEE_COLORS: Record<string, string> = {
  [PRODUCTION_EMPLOYEE_NAMES[0]]: '#2563eb',
  [PRODUCTION_EMPLOYEE_NAMES[1]]: '#0f766e',
}

const PRODUCTION_EMPLOYEE_ORDER = new Map<string, number>(
  PRODUCTION_EMPLOYEE_NAMES.map((employee, index) => [employee, index]),
)

function compareEmployeeMetrics(left: TrafficSourceMetric, right: TrafficSourceMetric): number {
  if (right.count !== left.count) return right.count - left.count

  const leftOrder = PRODUCTION_EMPLOYEE_ORDER.get(left.source) ?? Number.MAX_SAFE_INTEGER
  const rightOrder = PRODUCTION_EMPLOYEE_ORDER.get(right.source) ?? Number.MAX_SAFE_INTEGER
  if (leftOrder !== rightOrder) return leftOrder - rightOrder

  return left.source.localeCompare(right.source, 'ru')
}

export function buildEmployeeShareMetrics(
  items: readonly EmployeeShareCount[],
): TrafficSourceMetric[] {
  const counts = new Map(items.map((item) => [item.employee, item.count]))
  const knownEmployees = new Set<string>(PRODUCTION_EMPLOYEE_NAMES)

  const metrics = PRODUCTION_EMPLOYEE_NAMES.map((employee) => ({
    source: employee,
    count: counts.get(employee) ?? 0,
    color: PRODUCTION_EMPLOYEE_COLORS[employee] ?? colorForTrafficSource(employee),
  }))

  for (const item of items) {
    if (knownEmployees.has(item.employee)) continue
    metrics.push({
      source: item.employee,
      count: item.count,
      color: colorForTrafficSource(item.employee),
    })
  }

  return metrics.sort(compareEmployeeMetrics)
}
