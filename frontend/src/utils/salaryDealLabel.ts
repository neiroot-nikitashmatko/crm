import { h, type VNode } from 'vue'

export function salaryDealDisplayName(deal: { firstName?: string; patronymic?: string }) {
  const parts = [deal.firstName?.trim() ?? '', deal.patronymic?.trim() ?? ''].filter(Boolean)
  return parts.length > 0 ? parts.join(' ') : 'Без имени'
}

export function salaryDealOptionFullLabel(deal: {
  dealNumber: number | string
  firstName?: string
  patronymic?: string
  phone?: string
}) {
  const parts = [`#${deal.dealNumber}`, salaryDealDisplayName(deal)]
  const phone = deal.phone?.trim()
  if (phone) parts.push(phone)
  return parts.join(' · ')
}

export function salaryDealOptionNumberLabel(dealNumber: number | string) {
  return `#${dealNumber}`
}

export function renderSalaryDealOption(info: {
  node: VNode
  option: {
    label?: string | (() => unknown)
    fullLabel?: string
  }
}) {
  const fullLabel =
    info.option.fullLabel ??
    (typeof info.option.label === 'function' ? String(info.option.label()) : String(info.option.label ?? ''))

  const contentChild = Array.isArray(info.node.children) ? info.node.children[0] : null
  const contentClass =
    contentChild && typeof contentChild === 'object' && 'props' in contentChild
      ? ((contentChild as VNode).props?.class ?? 'n-base-select-option__content')
      : 'n-base-select-option__content'

  return h('div', info.node.props, h('div', { class: contentClass }, fullLabel))
}
