<template>
  <main v-if="hasResult" class="stage result-stage" aria-label="答对反馈">
    <span class="result-spark spark-left">✦</span>
    <span class="result-spark spark-right">✦</span>

    <section class="result-card">
      <div class="check-badge" aria-hidden="true"></div>
      <h1>答对啦！</h1>

      <div class="answer-reveal">
        <strong>{{ result?.answer || store.answerText }}</strong>
        <span>{{ result?.explanation || "这题已经成功通关" }}</span>
      </div>

      <dl class="score-row">
        <div>
          <dt>得分</dt>
          <dd>{{ result?.score ?? 0 }}</dd>
        </div>
        <div>
          <dt>模式</dt>
          <dd>{{ result?.answer_mode === "tiles" ? "选字答题" : "输入答案" }}</dd>
        </div>
      </dl>

      <div class="result-actions">
        <button class="primary-button" type="button" @click="next">下一题</button>
        <RouterLink class="secondary-button" to="/">回首页</RouterLink>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useGameStore } from "../stores/game";

const store = useGameStore();
const router = useRouter();
const result = computed(() => store.lastResult);
const hasResult = computed(() => Boolean(result.value?.correct));

onMounted(() => {
  if (!hasResult.value) {
    router.replace("/game");
  }
});

async function next() {
  await store.nextPuzzle();
  router.push("/game");
}
</script>
