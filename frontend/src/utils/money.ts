export function formatMoney(amount: number): string {
  const value = Number.isFinite(amount) ? amount : 0
  return `${new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 0 }).format(value)} ₽`
}
