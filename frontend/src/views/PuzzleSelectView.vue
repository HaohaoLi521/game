<template>
  <main class="select-shell" aria-label="选择题集">
    <header class="select-header">
      <RouterLink to="/" aria-label="返回首页">←</RouterLink>
      <div>
        <p>THIS IS PUN</p>
        <h1>选择题集</h1>
      </div>
    </header>

    <section v-if="store.loading" class="select-state">题集加载中...</section>
    <section v-else-if="store.error" class="select-state select-error">
      <strong>{{ store.error }}</strong>
      <button type="button" @click="store.loadSets">重新加载</button>
    </section>
    <section v-else-if="!store.sets.length" class="select-state">暂无可玩的题集</section>
    <section v-else class="set-grid" aria-label="题集列表">
      <RouterLink
        v-for="set in store.sets"
        :key="set.id"
        class="set-card"
        :to="{ name: 'game', query: { set: String(set.id) } }"
      >
        <span class="set-category">{{ set.category || "谐音梗" }}</span>
        <h2>{{ set.name }}</h2>
        <p>{{ set.description || "看图猜梗，开始挑战吧。" }}</p>
        <footer>
          <span>{{ set.puzzle_count }} 题</span>
          <strong>开始挑战 →</strong>
        </footer>
      </RouterLink>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted } from "vue";
import { useGameStore } from "../stores/game";

const store = useGameStore();

onMounted(() => {
  void store.loadSets();
});
</script>