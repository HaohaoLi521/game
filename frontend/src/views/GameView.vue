<template>
  <main class="stage play-stage" aria-label="这是谐音梗玩法页">
    <header class="play-header">
      <div class="round-meta">D {{ currentPuzzle?.id ?? "--" }}</div>
      <div class="author">作者：{{ currentPuzzle?.author_name ?? "QQ" }}</div>
      <div class="header-actions">
        <button class="icon-button" type="button" @click="requestHint">?</button>
        <RouterLink class="icon-button" to="/">☰</RouterLink>
      </div>
    </header>

    <section v-if="store.loading" class="loading-panel">题目加载中...</section>
    <section v-else-if="store.error && !currentPuzzle" class="loading-panel error-panel">
      <strong>{{ store.error }}</strong>
      <span>请确认后端服务已启动：localhost:8080</span>
    </section>

    <template v-else-if="currentPuzzle">
      <section class="puzzle-board" aria-label="题目图片">
        <article v-for="image in currentPuzzle.hint_images" :key="image.id" class="picture-card">
          <div class="illustration-wrap">
            <span v-if="isEmoji(image.url)" class="emoji-art">{{ toEmoji(image.url) }}</span>
            <img v-else :src="image.url" :alt="image.alt" />
          </div>
          <strong>{{ image.label }}</strong>
        </article>

        <div class="progress-dots" aria-label="关卡进度">
          <button
            v-for="(puzzle, index) in store.puzzles"
            :key="puzzle.id"
            type="button"
            :class="{
              active: puzzle.sort_order <= currentPuzzle.sort_order,
              solved: isPuzzleSolved(puzzle.id),
              current: index === store.currentIndex
            }"
            :disabled="!isPuzzleSolved(puzzle.id) || store.loading"
            :aria-current="index === store.currentIndex ? 'step' : undefined"
            :aria-label="isPuzzleSolved(puzzle.id) ? `第 ${index + 1} 题，已通过，点击切换` : `第 ${index + 1} 题，未通过`"
            :title="isPuzzleSolved(puzzle.id) ? `切换到第 ${index + 1} 题` : `第 ${index + 1} 题未通过`"
            @click="jumpToSolvedPuzzle(index, puzzle.id)"
          ></button>
        </div>
      </section>

      <section ref="answerPanel" class="answer-panel" :class="{ 'is-wrong': wrongAnimating }" aria-labelledby="answer-title">
        <div class="mode-row">
          <div>
            <h1 id="answer-title">猜出这个谐音梗</h1>
            <p>可以手动输入，也可以拖拽候选字</p>
          </div>
          <div class="mode-toggle" role="tablist" aria-label="答题模式">
            <button
              type="button"
              :class="{ active: store.answerMode === 'manual' }"
              @click="store.setMode('manual')"
            >
              输入答案
            </button>
            <button
              type="button"
              :class="{ active: store.answerMode === 'tiles' }"
              @click="store.setMode('tiles')"
            >
              选字答题
            </button>
          </div>
        </div>

        <div class="answer-state">
          <strong>答案长度：{{ currentPuzzle.answer_length }} 个字</strong>
          <span>{{ store.answerText || "当前未作答" }}</span>
        </div>

        <div v-if="store.answerMode === 'manual'" class="manual-card">
          <label for="manual-answer">手动输入答案</label>
          <div class="answer-input-wrap">
            <input
              id="manual-answer"
              v-model="manualAnswer"
              type="text"
              autocomplete="off"
              :maxlength="currentPuzzle.answer_length"
              placeholder="在这里输入答案"
              @keyup.enter="submit"
            />
            <button type="button" @click="store.clearAnswer">清空</button>
          </div>
        </div>

        <div v-else class="tiles-card">
          <div
            class="answer-slots"
            @dragover.prevent
            @drop="dropToPool"
          >
            <button
              v-for="(_, index) in store.answerSlots"
              :key="index"
              class="answer-slot"
              :class="{ filled: store.answerSlots[index] }"
              :data-slot-index="index"
              type="button"
              draggable="true"
              @click="slotClick(index)"
              @mousedown="startMouseSlot($event, index)"
              @dragstart="dragSlot($event, index)"
              @dragover.prevent
              @drop="dropToSlot($event, index)"
            >
              {{ store.answerSlots[index] || " " }}
            </button>
          </div>

          <div class="candidate-grid" data-tile-pool="true" @dragover.prevent @drop="dropToPool">
            <button
              v-for="candidate in currentPuzzle.candidate_chars"
              :key="candidate.id"
              class="candidate-tile"
              :class="{ used: store.isTileUsed(candidate.id) }"
              type="button"
              draggable="true"
              @click="chooseByClick(candidate)"
              @mousedown="startMouseCandidate($event, candidate.id)"
              @dragstart="dragCandidate($event, candidate.id)"
            >
              <span>{{ candidate.pinyin || " " }}</span>
              <strong>{{ candidate.char }}</strong>
            </button>
          </div>
        </div>

        <div v-if="store.hints.length" class="hint-list">
          <p v-for="hint in store.hints" :key="hint.level">{{ hint.text }}</p>
        </div>

        <div class="answer-actions">
          <button class="secondary-button" type="button" :disabled="store.hintsUsed >= 3" @click="requestHint">给点提示</button>
          <button class="ghost-button" type="button" @click="store.clearAnswer">重填</button>
          <button class="primary-button" type="button" :disabled="store.submitting" @click="submit">
            {{ store.submitting ? "提交中" : "提交答案" }}
          </button>
        </div>

        <div class="panel-footer">
          <span>{{ store.error || "答错不扣分，可以继续尝试" }}</span>
          <strong>已通关 {{ store.solvedCount }}/{{ store.puzzles.length }}</strong>
        </div>
      </section>
    </template>
  </main>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { gsap } from "gsap";
import type { CandidateChar } from "../api/puzzles";
import { useGameStore } from "../stores/game";

const store = useGameStore();
const router = useRouter();
const answerPanel = ref<HTMLElement | null>(null);
const wrongAnimating = ref(false);
const suppressClick = ref(false);
const mouseDrag = ref<
  | { type: "candidate"; id: string; startX: number; startY: number }
  | { type: "slot"; index: number; startX: number; startY: number }
  | null
>(null);

const currentPuzzle = computed(() => store.currentPuzzle);
const manualAnswer = computed({
  get: () => store.answerText,
  set: (value: string) => store.setManualAnswer(value)
});

onMounted(() => {
  store.bootstrap();
});

watch(
  () => store.wrongTick,
  async () => {
    if (!answerPanel.value) return;
    wrongAnimating.value = true;
    await nextTick();
    gsap.fromTo(answerPanel.value, { x: -10 }, { x: 0, duration: 0.08, repeat: 5, yoyo: true, onComplete: () => (wrongAnimating.value = false) });
  }
);

function isEmoji(url: string) {
  return url.startsWith("emoji:");
}

function toEmoji(url: string) {
  return url.replace("emoji:", "");
}

async function requestHint() {
  await store.requestHint();
}

function isPuzzleSolved(id: number) {
  return Boolean(store.progress[String(id)]);
}

async function jumpToSolvedPuzzle(index: number, puzzleId: number) {
  if (!isPuzzleSolved(puzzleId) || index === store.currentIndex) return;
  await store.loadPuzzleByIndex(index);
}

async function submit() {
  const result = await store.submitAnswer();
  if (result?.correct) {
    router.push("/result");
  }
}

function dragCandidate(event: DragEvent, id: string) {
  event.dataTransfer?.setData("text/plain", `candidate:${id}`);
}

function dragSlot(event: DragEvent, index: number) {
  if (!store.answerSlots[index]) return;
  event.dataTransfer?.setData("text/plain", `slot:${index}`);
}

function dropToSlot(event: DragEvent, targetIndex: number) {
  event.stopPropagation();
  const data = event.dataTransfer?.getData("text/plain") || "";
  if (data.startsWith("candidate:")) {
    const candidate = store.candidateById(data.replace("candidate:", ""));
    if (candidate) store.placeCandidate(candidate, targetIndex);
  }
  if (data.startsWith("slot:")) {
    const fromIndex = Number(data.replace("slot:", ""));
    const char = store.answerSlots[fromIndex];
    const tileId = store.selectedTileIds[fromIndex];
    store.answerSlots[fromIndex] = store.answerSlots[targetIndex];
    store.selectedTileIds[fromIndex] = store.selectedTileIds[targetIndex];
    store.answerSlots[targetIndex] = char;
    store.selectedTileIds[targetIndex] = tileId;
  }
}

function dropToPool(event: DragEvent) {
  event.stopPropagation();
  const data = event.dataTransfer?.getData("text/plain") || "";
  if (!data.startsWith("slot:")) return;
  store.removeSlot(Number(data.replace("slot:", "")));
}

function chooseByClick(candidate: CandidateChar) {
  if (suppressClick.value) return;
  store.chooseCandidate(candidate);
}

function slotClick(index: number) {
  if (suppressClick.value) return;
  store.removeSlot(index);
}

function startMouseCandidate(event: MouseEvent, id: string) {
  if (store.isTileUsed(id)) return;
  mouseDrag.value = { type: "candidate", id, startX: event.clientX, startY: event.clientY };
  window.addEventListener("mouseup", finishMouseDrag, { once: true });
}

function startMouseSlot(event: MouseEvent, index: number) {
  if (!store.answerSlots[index]) return;
  mouseDrag.value = { type: "slot", index, startX: event.clientX, startY: event.clientY };
  window.addEventListener("mouseup", finishMouseDrag, { once: true });
}

function finishMouseDrag(event: MouseEvent) {
  const drag = mouseDrag.value;
  mouseDrag.value = null;
  if (!drag) return;

  const distance = Math.hypot(event.clientX - drag.startX, event.clientY - drag.startY);
  if (distance < 8) return;

  suppressClick.value = true;
  window.setTimeout(() => {
    suppressClick.value = false;
  }, 0);

  const target = document.elementFromPoint(event.clientX, event.clientY) as HTMLElement | null;
  const slotEl = target?.closest("[data-slot-index]") as HTMLElement | null;
  const poolEl = target?.closest("[data-tile-pool]") as HTMLElement | null;

  if (drag.type === "candidate" && slotEl) {
    const candidate = store.candidateById(drag.id);
    if (candidate) store.placeCandidate(candidate, Number(slotEl.dataset.slotIndex));
    return;
  }

  if (drag.type === "slot" && poolEl) {
    store.removeSlot(drag.index);
    return;
  }

  if (drag.type === "slot" && slotEl) {
    const targetIndex = Number(slotEl.dataset.slotIndex);
    if (targetIndex === drag.index) return;
    const char = store.answerSlots[drag.index];
    const tileId = store.selectedTileIds[drag.index];
    store.answerSlots[drag.index] = store.answerSlots[targetIndex];
    store.selectedTileIds[drag.index] = store.selectedTileIds[targetIndex];
    store.answerSlots[targetIndex] = char;
    store.selectedTileIds[targetIndex] = tileId;
  }
}
</script>
