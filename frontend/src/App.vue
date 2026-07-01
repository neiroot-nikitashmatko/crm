<script setup lang="ts">
import { NConfigProvider, NMessageProvider, ruRU, dateRuRU } from 'naive-ui'
import type { GlobalThemeOverrides } from 'naive-ui'

const themeOverrides: GlobalThemeOverrides = {
  common: {
    bodyColor: '#ffffff',
    cardColor: '#ffffff',
    modalColor: '#ffffff',
    popoverColor: '#ffffff',
    baseColor: '#ffffff',
    primaryColor: '#4a5568',
    primaryColorHover: '#2d3748',
    primaryColorPressed: '#1a202c',
    borderColor: '#e2e8f0',
    dividerColor: '#e2e8f0',
    textColor1: '#1a202c',
    textColor2: '#4a5568',
    textColor3: '#718096',
    hoverColor: '#edf2f7',
    pressedColor: '#e2e8f0',
    fontFamily: 'system-ui, -apple-system, "Segoe UI", Roboto, sans-serif',
  },
}
</script>

<template>
  <NConfigProvider :locale="ruRU" :date-locale="dateRuRU" :theme-overrides="themeOverrides">
    <NMessageProvider>
      <RouterView />
    </NMessageProvider>
  </NConfigProvider>
</template>
