<script setup lang="ts">
import ChatsWorkspace from '@/components/chats/ChatsWorkspace.vue'
</script>

<template>
  <div class="chats-page">
    <header class="chats-page__header">
      <h1 class="chats-page__title">Чаты</h1>
    </header>
    <div class="chats-page__body">
      <ChatsWorkspace />
    </div>
  </div>
</template>

<style scoped>
.chats-page {
  display: flex;
  flex-direction: column;
  height: calc(100dvh - 64px);
  max-height: calc(100dvh - 64px);
  overflow: hidden;
  background: #ffffff;
}

.chats-page__header {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  height: 56px;
  padding: 0 16px 0 24px;
  border-bottom: 1px solid #e2e8f0;
  background: #ffffff;
}

.chats-page__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #1a202c;
  letter-spacing: -0.02em;
}

.chats-page__body {
  flex: 1 1 auto;
  min-height: 0;
  padding: 12px 16px 16px;
  display: flex;
  flex-direction: column;
}
</style>
