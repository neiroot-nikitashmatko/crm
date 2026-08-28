import type { AnalyticsDateRange, TradeProfit } from '@/types/analytics'

const MS_PER_DAY = 86_400_000

/** Закуп 1 000 ₽, продажа 2 200 ₽ — как в примере с ковриками EVA. */
const EVA_UNIT_COST = 1000
const EVA_UNIT_PRICE = 2200
const EVA_UNITS_PER_DAY = 2

/** Прочие продажи за день, чтобы сумма на плашке выглядела реалистично. */
const OTHER_UNITS_PER_DAY = 3
const OTHER_UNIT_COST = 4500
const OTHER_UNIT_PRICE = 7900

const INVOICES_PER_DAY = 2

export function isTradeProfitMockEnabled(): boolean {
  if (!import.meta.env.DEV) return false
  if (import.meta.env.VITE_MOCK_TRADE_PROFIT === 'false') return false
  return true
}

export function getMockTradeProfit(range: AnalyticsDateRange): TradeProfit {
  const [from, to] = range
  const span = Math.max(0, to - from)
  const days = Math.max(1, Math.floor(span / MS_PER_DAY) + 1)

  const evaQty = EVA_UNITS_PER_DAY * days
  const otherQty = OTHER_UNITS_PER_DAY * days

  const revenue = evaQty * EVA_UNIT_PRICE + otherQty * OTHER_UNIT_PRICE
  const cost = evaQty * EVA_UNIT_COST + otherQty * OTHER_UNIT_COST

  return {
    profit: revenue - cost,
    revenue,
    cost,
    invoicesCount: INVOICES_PER_DAY * days,
  }
}
