export const PRODUCTION_NOMENCLATURE_OPTIONS = [
  { label: 'Перетяжка руля', value: 'Перетяжка руля' },
  { label: 'Установка чехлов', value: 'Установка чехлов' },
  { label: 'Установка накидок', value: 'Установка накидок' },
  { label: 'Пошив ковриков', value: 'Пошив ковриков' },
  { label: 'Отрисовка лекал', value: 'Отрисовка лекал' },
  { label: 'Полировка стёкол', value: 'Полировка стёкол' },
  { label: 'Ремонт стёкол', value: 'Ремонт стёкол' },
  { label: 'Полировка фар', value: 'Полировка фар' },
] as const

export const PRODUCTION_SHARE_CATEGORIES = [
  'Перетяжка',
  'Установка',
  'Стёкла',
  'Коврики',
] as const

export const PRODUCTION_SHARE_OTHER_CATEGORY = 'Прочее'

export const PRODUCTION_NOMENCLATURE_TO_CATEGORY: Record<string, string> = {
  'Перетяжка руля': 'Перетяжка',
  'Установка чехлов': 'Установка',
  'Установка накидок': 'Установка',
  'Полировка фар': 'Стёкла',
  'Полировка стёкол': 'Стёкла',
  'Ремонт стёкол': 'Стёкла',
  'Пошив ковриков': 'Коврики',
}

export function productionCategoryForNomenclature(nomenclature: string): string {
  const trimmed = nomenclature.trim()
  return PRODUCTION_NOMENCLATURE_TO_CATEGORY[trimmed] ?? PRODUCTION_SHARE_OTHER_CATEGORY
}

export const PRODUCTION_EMPLOYEE_NAMES = [
  'Никита Хачересов',
  'Сергей Геворкян',
] as const
