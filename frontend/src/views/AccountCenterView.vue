<template>
  <main class="settings-shell" aria-label="玩家账户中心">
    <header class="settings-header">
      <RouterLink to="/">←</RouterLink>
      <div><p>THIS IS PUN</p><h1>账户中心</h1></div>
      <RouterLink to="/my-submissions">我的投稿</RouterLink>
    </header>

    <section v-if="player.loggedIn" class="settings-card">
      <h2>你好，{{ player.username }}</h2>
      <div class="stats-row">
        <div><span>已通关</span><strong>{{ player.solvedCount }}</strong></div>
        <div><span>成就</span><strong>{{ player.achievements.length }}</strong></div>
      </div>
      <div v-if="player.achievements.length" class="answer-reveal">
        <strong>已解锁成就</strong>
        <span v-for="item in player.achievements" :key="item.id">{{ item.title }}：{{ item.description }}</span>
      </div>
      <button class="ghost-button" type="button" @click="player.logout">退出登录</button>
    </section>

    <section v-else class="settings-card">
      <label class="settings-row"><span><strong>账号</strong></span><input v-model.trim="username" autocomplete="username" /></label>
      <label class="settings-row"><span><strong>密码</strong></span><input v-model="password" type="password" autocomplete="current-password" /></label>
      <p v-if="error" class="submit-error">{{ error }}</p>
      <div class="answer-actions"><button class="secondary-button" type="button" @click="submit(true)">注册</button><button class="primary-button" type="button" @click="submit(false)">登录</button></div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { usePlayerStore } from "../stores/player";

const player = usePlayerStore();
const username = ref("");
const password = ref("");
const error = ref("");

async function submit(register: boolean) {
  error.value = "";
  try { await player.login(username.value, password.value, register); } catch (err) { error.value = err instanceof Error ? err.message : "认证失败"; }
}
</script>
