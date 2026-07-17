export const SALARY_SERVICE_OPTIONS = [
  { label: 'Установка чехлов (без подшива)', value: 'covers_without_hem' },
  { label: 'Установка чехлов (с подшивом)', value: 'covers_with_hem' },
  { label: 'Установка накидок', value: 'seat_covers' },
  { label: 'Перетяжка руля', value: 'steering_wheel' },
  { label: 'Ремонт стёкол', value: 'glass_repair' },
  { label: 'Полировка стёкол/фар', value: 'glass_headlight_polish' },
] as const

export type SalaryServiceValue = (typeof SALARY_SERVICE_OPTIONS)[number]['value']
