<template>
  <header class="leads-section-header">
    <h1 class="leads-section-header__title">Лиды</h1>
    <div class="leads-section-header__actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<style scoped>
.leads-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  height: 56px;
  padding: 0 24px;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
}

.leads-section-header__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #1a202c;
  letter-spacing: -0.02em;
}

.leads-section-header__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 32px;
}
</style>
