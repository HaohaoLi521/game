<template>
  <main class="settings-shell" aria-label="我的投稿">
    <header class="settings-header">
      <RouterLink to="/">←</RouterLink>
      <div><p>THIS IS PUN</p><h1>我的投稿</h1></div>
      <RouterLink to="/submit">我要出题</RouterLink>
    </header>
    <section v-if="!player.loggedIn" class="settings-card">
      <p>登录后可查看自己的投稿状态。</p>
      <RouterLink class="primary-button" to="/account">去登录</RouterLink>
    </section>
    <section v-else class="settings-card">
      <p v-if="loading">正在加载投稿…</p>
      <p v-else-if="error" class="submit-error">{{ error }}</p>
      <button v-if="error" class="ghost-button" @click="load">重试</button>
      <p v-else-if="!items.length">你还没有提交过题目。</p>
      <article v-for="item in items" :key="item.id" class="settings-row">
        <span><strong>#{{ item.id }} · {{ item.answer }}</strong><small>{{ item.category }} · 难度 {{ item.difficulty }}</small></span>
        <span><strong>{{ statusText(item.status) }}</strong><small v-if="item.review_note">{{ item.review_note }}</small></span>
      </article>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { listPlayerSubmissions, type PuzzleSubmission } from "../api/submissions";
import { usePlayerStore } from "../stores/player";

const player = usePlayerStore();
const items = ref<PuzzleSubmission[]>([]);
const loading = ref(false);
const error = ref("");

async function load() {
  if (!player.loggedIn) return;
  loading.value = true;
  error.value = "";
  try { items.value = await listPlayerSubmissions(); } catch (err) { error.value = err instanceof Error ? err.message : "加载失败"; } finally { loading.value = false; }
}

function statusText(status: PuzzleSubmission["status"]) { return ({ pending: "审核中", approved: "已通过", rejected: "已驳回" } as const)[status]; }

onMounted(load);
</script>
