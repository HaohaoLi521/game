<template>
  <main class="select-shell" aria-label="创意工坊">
    <header class="select-header">
      <RouterLink to="/" aria-label="返回首页">←</RouterLink>
      <div><p>THIS IS PUN</p><h1>创意工坊</h1></div>
      <RouterLink to="/submit">我要出题</RouterLink>
    </header>
    <section class="filter-row" aria-label="工坊筛选">
      <input v-model.trim="category" placeholder="按分类筛选" @keyup.enter="load(1)" />
      <select v-model.number="difficulty" @change="load(1)">
        <option :value="0">全部难度</option>
        <option v-for="level in 5" :key="level" :value="level">难度 {{ level }}</option>
      </select>
      <button type="button" @click="load(1)">筛选</button>
    </section>
    <section v-if="loading" class="select-state">工坊加载中…</section>
    <section v-else-if="error" class="select-state select-error"><strong>{{ error }}</strong><button type="button" @click="load(page)">重试</button></section>
    <section v-else-if="!items.length" class="select-state">暂无符合条件的题目。</section>
    <section v-else class="set-grid" aria-label="工坊题目列表">
      <RouterLink v-for="item in items" :key="item.id" class="set-card" :to="{ name: 'game', query: { set: String(item.puzzle_set_id), puzzle: String(item.id) } }">
        <span class="set-category">{{ item.category || "未分类" }} · 难度 {{ item.difficulty }}</span>
        <h2>{{ item.hint_images?.[0]?.label || "玩家创作题" }}</h2>
        <p>作者：{{ item.author_name || "匿名玩家" }}</p>
        <footer><strong>开始挑战 →</strong></footer>
      </RouterLink>
    </section>
    <footer v-if="total > pageSize" class="answer-actions">
      <button class="secondary-button" type="button" :disabled="page <= 1" @click="load(page - 1)">上一页</button>
      <span>第 {{ page }} 页</span>
      <button class="secondary-button" type="button" :disabled="page * pageSize >= total" @click="load(page + 1)">下一页</button>
    </footer>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { listWorkshop, type WorkshopItem } from "../api/workshop";

const items = ref<WorkshopItem[]>([]);
const category = ref("");
const difficulty = ref(0);
const page = ref(1);
const pageSize = 12;
const total = ref(0);
const loading = ref(false);
const error = ref("");

async function load(nextPage: number) {
  page.value = Math.max(1, nextPage);
  loading.value = true;
  error.value = "";
  try {
    const result = await listWorkshop({ category: category.value || undefined, difficulty: difficulty.value || undefined, page: page.value, page_size: pageSize });
    items.value = result.items;
    total.value = result.total;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "工坊加载失败";
  } finally {
    loading.value = false;
  }
}

onMounted(() => void load(1));
</script>
