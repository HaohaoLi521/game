<template>
  <main class="stage home-stage" aria-label="这是谐音梗首页">
    <span class="paper paper-xl paper-cream"></span>
    <span class="paper paper-md paper-peach"></span>
    <span class="paper paper-lg paper-yellow"></span>
    <span class="music-wave wave-left"></span>
    <span class="music-wave wave-right"></span>
    <span class="music-note note-one">♪</span>
    <span class="music-note note-two">♫</span>
    <span class="music-note note-three">♪</span>

    <RouterLink class="icon-button settings-button" to="/settings" aria-label="设置">⚙</RouterLink>
    <RouterLink class="home-account" to="/account">账户</RouterLink>

    <section class="home-panel">
      <div class="game-logo" aria-label="这是谐音梗">
        <span>这是</span>
        <span>谐音梗</span>
      </div>
      <p class="logo-subtitle">XIE YIN GAME</p>

      <nav class="home-menu">
        <RouterLink class="menu-button" to="/puzzles">开始游戏</RouterLink>
        <RouterLink class="menu-button" to="/puzzles">选择题目</RouterLink>
        <RouterLink class="menu-button" to="/submit">我要出题</RouterLink>
        <button class="menu-button" type="button" @click="toast = '多人模式会在 M4 接入'">多人游戏</button>
      </nav>
    </section>

    <img class="shoe-hero" :src="shoeUrl" alt="粉色卡通鞋" />

    <RouterLink class="question-card" to="/puzzles" aria-label="进入题库">
      <img :src="birdUrl" alt="" />
      <span>题库</span>
    </RouterLink>

    <p class="home-tagline">双模式答题 · 看图猜梗 · 三分钟热场</p>
    <p v-if="toast" class="toast show">{{ toast }}</p>
  </main>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import birdUrl from "../assets/bird.svg";
import shoeUrl from "../assets/shoe.svg";

const toast = ref("");

watch(toast, (value) => {
  if (!value) return;
  window.setTimeout(() => {
    toast.value = "";
  }, 1800);
});
</script>
