<script setup lang="ts">
import { ref } from 'vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'

const sidebarOpen = ref(false)

function toggleSidebar() {
  sidebarOpen.value = !sidebarOpen.value
}

function closeSidebar() {
  sidebarOpen.value = false
}
</script>

<template>
  <div class="app-layout">
    <AppHeader @toggle-sidebar="toggleSidebar" />

    <div class="app-layout__body">
      <AppSidebar :open="sidebarOpen" @close="closeSidebar" />

      <div
        v-if="sidebarOpen"
        class="app-layout__overlay"
        @click="closeSidebar"
      />

      <main class="app-layout__content">
        <RouterView class="app-layout__view" />
      </main>
    </div>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.app-layout__body {
  position: relative;
  display: flex;
  flex: 1;
  overflow: hidden;
  background: #ffffff;
}

.app-layout__overlay {
  position: fixed;
  inset: 0;
  top: 64px;
  background: rgba(0, 0, 0, 0.25);
  z-index: 90;
}

.app-layout__content {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  overflow: hidden;
  background: #ffffff;
}

.app-layout__view {
  display: flex;
  flex-direction: column;
  flex: 1;
  width: 100%;
  min-height: 0;
  overflow: hidden;
}
</style>
