<script setup lang="ts">
import { ElConfigProvider } from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import { RouterView } from 'vue-router'
</script>

<template>
  <ElConfigProvider :locale="zhCn">
    <main class="app-root">
      <RouterView />
    </main>
  </ElConfigProvider>
</template>

<style scoped lang="scss">
.app-root {
  min-height: 100dvh;
}
</style>
