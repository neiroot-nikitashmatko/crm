# DESIGN

## Цвета

- Текст: `#1a202c` / `#4a5568` / `#718096`
- Border: `#e2e8f0`, inputs: `#cbd5e1`
- CRM primary: `#1f883d` (не путать с Naive primary `#4a5568`)
- Акцент канбана: `countColor` из constants -> CSS var на карточке

## Layout

- Header 64px
- SectionHeader 56px
- страница: `calc(100dvh - 64px)`
- Kanban padding `16px 24px 0`, gap 10px, column bg `#f6f8fa`

## Карточки канбана

Hover: `inset 3px 0 0 var(--accent)`, shadow `0 4px 12px rgba(15,23,42,0.08)`, `translateY(-1px)`

## Bottom sheet

- fixed top/left/right 15px, bottom 0
- radius `12px 12px 0 0`
- backdrop z180/sheet z190

## Даты

- формат `dd.MM.yyyy HH:mm`
- часы 9–19
- шаг 5 минут

## CSS

- Scoped BEM (`block__element--modifier`)
- `:deep()` только для Naive
- scroll areas: `scrollbar-gutter: stable`
