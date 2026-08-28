import type { AnalyticsDateRange, TradeProfit, TradeProfitItem } from '@/types/analytics'

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

function periodDays(range: AnalyticsDateRange): number {
  const [from, to] = range
  const span = Math.max(0, to - from)
  return Math.max(1, Math.floor(span / MS_PER_DAY) + 1)
}

export function isTradeProfitMockEnabled(): boolean {
  if (!import.meta.env.DEV) return false
  if (import.meta.env.VITE_MOCK_TRADE_PROFIT === 'false') return false
  return true
}

export function getMockTradeProfit(range: AnalyticsDateRange): TradeProfit {
  const items = getMockTradeProfitItems(range)
  const revenue = items.reduce((total, item) => total + item.quantity * item.salePrice, 0)
  const cost = items.reduce((total, item) => {
    if (!item.hasCost) return total
    return total + item.quantity * item.costPrice
  }, 0)

  return {
    profit: revenue - cost,
    revenue,
    cost,
    invoicesCount: INVOICES_PER_DAY * periodDays(range),
  }
}

export function getMockTradeProfitItems(range: AnalyticsDateRange): TradeProfitItem[] {
  const days = periodDays(range)
  const evaQty = EVA_UNITS_PER_DAY * days
  const otherQty = OTHER_UNITS_PER_DAY * days

  const items: TradeProfitItem[] = [
    {
      productKey: 'title:авточехлы',
      title: 'Авточехлы для автомобиля Ford Focus -2, 08.2004-06.2011, к.Ghia/Titanium, РЗСиС60/40+подлок., 3Г БАЙРОН ст БАЙРОН "Орегон" Чёр / Чёр / тём-сер',
      quantity: otherQty,
      costPrice: OTHER_UNIT_COST,
      salePrice: OTHER_UNIT_PRICE,
      profit: otherQty * (OTHER_UNIT_PRICE - OTHER_UNIT_COST),
      hasCost: true,
    },
    {
      productKey: 'title:коврики eva',
      title: 'Коврики EVA',
      quantity: evaQty,
      costPrice: EVA_UNIT_COST,
      salePrice: EVA_UNIT_PRICE,
      profit: evaQty * (EVA_UNIT_PRICE - EVA_UNIT_COST),
      hasCost: true,
    },
  ]

  return items.sort((left, right) => right.profit - left.profit || left.title.localeCompare(right.title, 'ru'))
}
