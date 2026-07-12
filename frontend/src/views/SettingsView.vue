<template>
  <main class="settings-shell" aria-label="游戏设置">
    <header class="settings-header">
      <RouterLink to="/" aria-label="返回首页">←</RouterLink>
      <div>
        <p>THIS IS PUN</p>
        <h1>游戏设置</h1>
      </div>
    </header>

    <section class="settings-card">
      <label class="settings-row">
        <span>
          <strong>默认答题方式</strong>
          <small>新题会优先使用此方式</small>
        </span>
        <select :value="settings.answerMode" @change="setAnswerMode">
          <option value="auto">跟随题目</option>
          <option value="manual">手动输入</option>
          <option value="tiles">选字答题</option>
        </select>
      </label>

      <label class="settings-row">
        <span>
          <strong>提示音</strong>
          <small>答对或答错时播放短音效</small>
        </span>
        <input type="checkbox" :checked="settings.soundEnabled" @change="setSoundEnabled" />
      </label>

      <label class="settings-row">
        <span>
          <strong>震动反馈</strong>
          <small>答错时震动，设备不支持时自动忽略</small>
        </span>
        <input type="checkbox" :checked="settings.vibrationEnabled" @change="setVibrationEnabled" />
      </label>
    </section>

    <section class="settings-card danger-card">
      <div>
        <h2>本地进度</h2>
        <p>仅清除当前浏览器保存的通关记录，无法恢复。</p>
      </div>
      <button class="ghost-button" type="button" @click="clearProgress">清空本地进度</button>
      <p v-if="message" class="settings-message">{{ message }}</p>
    </section>
  </main>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useGameStore } from "../stores/game";
import { useSettingsStore, type AnswerModePreference } from "../stores/settings";

const game = useGameStore();
const settings = useSettingsStore();
const message = ref("");

function setAnswerMode(event: Event) {
  settings.setAnswerMode((event.target as HTMLSelectElement).value as AnswerModePreference);
}

function setSoundEnabled(event: Event) {
  settings.setSoundEnabled((event.target as HTMLInputElement).checked);
}

function setVibrationEnabled(event: Event) {
  settings.setVibrationEnabled((event.target as HTMLInputElement).checked);
}

function clearProgress() {
  game.clearProgress();
  message.value = "本地进度已清空";
}
</script>
