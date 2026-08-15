/** Источники трафика, доступные при ручном создании лида. */
export const MANUAL_TRAFFIC_SOURCES = [
  'Знал о производстве',
  'Авито (Автоатрибут)',
  'Авито (AutoFactory)',
  'Авито (Автоателье)',
  'По рекомендации',
  'Увидел вывеску',
  'Оптовый клиент',
  'Яндекс карты',
  'Instagram',
  'Вконтакте',
  '2gis',
  'Визитка (установка чехлов)',
] as const

export type ManualTrafficSource = (typeof MANUAL_TRAFFIC_SOURCES)[number]

export const MANUAL_TRAFFIC_SOURCE_OPTIONS = MANUAL_TRAFFIC_SOURCES.map((value) => ({
  label: value,
  value,
}))

/** Источник лидов из чата Авито (ставится автоматически). */
export const AVITO_CHAT_TRAFFIC_SOURCE = 'Авито Чат'

/** Источник, если звонок Билайн не сопоставился с номером из справочника. */
export const BEELINE_TRAFFIC_SOURCE = 'Билайн'

/** Все источники трафика, которые есть в системе. */
export const ALL_TRAFFIC_SOURCES = [
  ...MANUAL_TRAFFIC_SOURCES,
  AVITO_CHAT_TRAFFIC_SOURCE,
  BEELINE_TRAFFIC_SOURCE,
] as const

export function isManualTrafficSource(value: string): boolean {
  return (MANUAL_TRAFFIC_SOURCES as readonly string[]).includes(value.trim())
}
