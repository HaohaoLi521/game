<template>
  <main class="settings-shell" aria-label="联机大厅">
    <header class="settings-header"><RouterLink to="/">←</RouterLink><div><p>THIS IS PUN</p><h1>联机大厅</h1></div></header>
    <section v-if="!player.loggedIn" class="settings-card"><p>登录后才能加入联机房间。</p><RouterLink class="primary-button" to="/account">去登录</RouterLink></section>
    <section v-else class="settings-card">
      <label class="settings-row"><span><strong>昵称</strong></span><input v-model.trim="playerName" placeholder="联机玩家" /></label>
      <label class="settings-row"><span><strong>题集 ID</strong></span><input v-model.number="puzzleSetId" type="number" min="1" /></label>
      <div class="answer-actions"><button class="primary-button" :disabled="loading" @click="create">创建房间</button></div>
      <hr />
      <label class="settings-row"><span><strong>房间号</strong></span><input v-model.trim="roomId" placeholder="输入房间号" /></label>
      <div class="answer-actions"><button class="secondary-button" :disabled="loading" @click="join">加入房间</button></div>
      <p v-if="error" class="submit-error">{{ error }}</p>
    </section>
  </main>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { usePlayerStore } from "../stores/player";
import { useRoomStore } from "../stores/room";

const router = useRouter();
const player = usePlayerStore();
const room = useRoomStore();
const playerName = ref(player.username);
const puzzleSetId = ref(1);
const roomId = ref("");
const error = ref("");
const loading = ref(false);

async function create() { loading.value = true; error.value = ""; try { const created = await room.create(puzzleSetId.value, playerName.value); router.push({ name: "room", params: { id: created.id } }); } catch (err) { error.value = err instanceof Error ? err.message : "创建房间失败"; } finally { loading.value = false; } }
async function join() { loading.value = true; error.value = ""; try { const joined = await room.join(roomId.value, playerName.value); router.push({ name: "room", params: { id: joined.id } }); } catch (err) { error.value = err instanceof Error ? err.message : "加入房间失败"; } finally { loading.value = false; } }
</script>
