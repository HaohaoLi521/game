<template>
  <main class="settings-shell" aria-label="联机房间">
    <header class="settings-header"><RouterLink to="/rooms">←</RouterLink><div><p>ROOM {{ room.room?.id }}</p><h1>联机房间</h1></div><span>{{ room.connected ? "已连接" : "连接中" }}</span></header>
    <section v-if="room.room" class="settings-card">
      <p>状态：{{ statusText(room.room.status) }} · 题目 {{ room.room.current_puzzle }}</p>
      <article v-for="player in room.room.players" :key="player.id" class="settings-row"><span><strong>{{ player.name }}{{ player.id === room.room?.host_id ? "（房主）" : "" }}</strong></span><span>{{ player.ready ? "已准备" : "未准备" }} · {{ player.score }} 分</span></article>
      <div class="answer-actions"><button class="secondary-button" @click="room.send('ready')">准备</button><button class="primary-button" :disabled="room.room.status !== 'ready'" @click="room.send('start')">开始</button></div>
      <div v-if="room.room.status === 'playing'" class="settings-row"><input v-model="answer" placeholder="输入答案" @keyup.enter="submitAnswer" /><button class="primary-button" @click="submitAnswer">提交</button></div>
      <p v-if="room.lastResult" class="answer-reveal">{{ room.lastResult.correct ? "答对了" : "答案不正确" }}</p>
      <p v-if="room.error" class="submit-error">{{ room.error }}</p>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";
import { useRoute } from "vue-router";
import { useRoomStore } from "../stores/room";

const route = useRoute();
const room = useRoomStore();
const answer = ref("");
function statusText(status: string) { return ({ waiting: "等待玩家", ready: "已准备", playing: "进行中", finished: "已结束" } as Record<string, string>)[status] || status; }
function submitAnswer() { room.send("answer", { answer: answer.value, answer_mode: "manual", elapsed_ms: 0 }); answer.value = ""; }
onMounted(() => room.connect(String(route.params.id)));
onUnmounted(() => room.disconnect());
</script>
