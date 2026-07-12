<template>
  <main class="submit-shell" aria-label="玩家出题投稿">
    <header class="submit-header">
      <RouterLink to="/" aria-label="返回首页">‹</RouterLink>
      <div>
        <h1>我要出题</h1>
        <p>投稿会进入管理员审核，通过后才会出现在游戏题库里。</p>
      </div>
      <RouterLink to="/game">试玩</RouterLink>
    </header>

    <form class="submit-form" @submit.prevent="submit">
      <section class="submit-card">
        <h2>基本信息</h2>
        <div class="submit-grid">
          <label>
            昵称
            <input v-model.trim="form.creator_name" type="text" placeholder="匿名玩家" />
          </label>
          <label>
            联系方式
            <input v-model.trim="form.contact" type="text" placeholder="微信 / QQ / 邮箱，可选" />
          </label>
          <label>
            答案
            <input v-model.trim="form.answer" required type="text" placeholder="例如：将心比心" />
          </label>
          <label>
            答案拼音
            <input v-model.trim="form.answer_pinyin" type="text" placeholder="可留空，系统自动生成" />
          </label>
          <label>
            分类
            <input v-model.trim="form.category" type="text" placeholder="成语 / 生活 / 动物" />
          </label>
          <label>
            难度
            <input v-model.number="form.difficulty" type="number" min="1" max="5" />
          </label>
        </div>
      </section>

      <section class="submit-card">
        <h2>两组提示</h2>
        <div class="submit-grid two">
          <label>
            上图
            <input v-model.trim="form.hint_one_url" type="text" placeholder="emoji:🐦 或图片 URL" />
            <input type="file" accept="image/*" @change="selectFile($event, 'one')" />
          </label>
          <label>
            上图文案
            <input v-model.trim="form.hint_one_label" type="text" placeholder="这是小鸟" />
          </label>
          <label>
            下图
            <input v-model.trim="form.hint_two_url" type="text" placeholder="emoji:🧍 或图片 URL" />
            <input type="file" accept="image/*" @change="selectFile($event, 'two')" />
          </label>
          <label>
            下图文案
            <input v-model.trim="form.hint_two_label" type="text" placeholder="这是____" />
          </label>
        </div>
      </section>

      <section class="submit-card">
        <div class="submit-title-row">
          <h2>答题方式</h2>
          <button type="button" @click="generateCandidates">按答案生成候选字</button>
        </div>
        <div class="submit-mode-row">
          <label>
            默认模式
            <select v-model="form.default_answer_mode">
              <option value="manual">手动输入</option>
              <option value="tiles">选字答题</option>
            </select>
          </label>
          <label class="submit-check">
            <input v-model="form.support_manual" type="checkbox" />
            支持输入
          </label>
          <label class="submit-check">
            <input v-model="form.support_tiles" type="checkbox" />
            支持选字
          </label>
        </div>
        <label>
          候选字
          <textarea
            v-model="form.candidate_text"
            rows="5"
            placeholder="每行一个候选字，可写：字 拼音。例如：&#10;将 jiang&#10;心 xin&#10;比 bi"
          ></textarea>
        </label>
      </section>

      <section class="submit-card">
        <h2>别名与解释</h2>
        <label>
          答案别名
          <textarea v-model="form.aliases_text" rows="3" placeholder="一行一个，或用逗号分隔"></textarea>
        </label>
        <label>
          解释
          <textarea v-model="form.explanation" rows="4" placeholder="说明这个谐音梗为什么是这个答案"></textarea>
        </label>
      </section>

      <footer class="submit-actions">
        <p v-if="message" class="submit-message">{{ message }}</p>
        <p v-if="error" class="submit-error">{{ error }}</p>
        <button v-if="error" class="ghost-button" type="button" :disabled="saving" @click="submit">再试一次</button>
        <button class="secondary-button" type="button" @click="reset">重填</button>
        <button class="primary-button" type="submit" :disabled="saving">{{ saving ? "提交中" : "提交审核" }}</button>
      </footer>
    </form>
  </main>
</template>

<script setup lang="ts">
import { reactive, ref } from "vue";
import { createPlayerSubmission, createSubmission, type SubmissionInput } from "../api/submissions";
import { uploadMedia } from "../api/media";
import type { AnswerMode, CandidateChar } from "../api/puzzles";
import { usePlayerStore } from "../stores/player";

interface SubmitForm {
  creator_name: string;
  contact: string;
  answer: string;
  answer_pinyin: string;
  aliases_text: string;
  category: string;
  difficulty: number;
  hint_one_url: string;
  hint_one_label: string;
  hint_two_url: string;
  hint_two_label: string;
  default_answer_mode: AnswerMode;
  support_manual: boolean;
  support_tiles: boolean;
  candidate_text: string;
  explanation: string;
}

const saving = ref(false);
const message = ref("");
const error = ref("");
const player = usePlayerStore();
const hintFiles = reactive<{ one: File | null; two: File | null }>({ one: null, two: null });
const form = reactive<SubmitForm>(emptyForm());

function emptyForm(): SubmitForm {
  return {
    creator_name: "",
    contact: "",
    answer: "",
    answer_pinyin: "",
    aliases_text: "",
    category: "成语",
    difficulty: 1,
    hint_one_url: "emoji:❓",
    hint_one_label: "提示图 1",
    hint_two_url: "emoji:❓",
    hint_two_label: "提示图 2",
    default_answer_mode: "tiles",
    support_manual: true,
    support_tiles: true,
    candidate_text: "",
    explanation: ""
  };
}

function reset() {
  Object.assign(form, emptyForm());
  message.value = "";
  error.value = "";
}

function generateCandidates() {
  form.candidate_text = Array.from(form.answer).map((char) => char.trim()).filter(Boolean).join("\n");
}

async function submit() {
  saving.value = true;
  message.value = "";
  error.value = "";
  try {
    if (hintFiles.one) form.hint_one_url = (await uploadMedia(hintFiles.one)).url;
    if (hintFiles.two) form.hint_two_url = (await uploadMedia(hintFiles.two)).url;
    const payload = toPayload();
    const result = player.loggedIn ? await createPlayerSubmission(payload) : await createSubmission(payload);
    message.value = `投稿成功，审核编号 #${result.id}`;
    reset();
    message.value = `投稿成功，审核编号 #${result.id}`;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "投稿失败";
  } finally {
    saving.value = false;
  }
}

function selectFile(event: Event, slot: "one" | "two") {
  hintFiles[slot] = (event.target as HTMLInputElement).files?.[0] || null;
}

function toPayload(): SubmissionInput {
  const modes = supportedModes();
  if (!form.answer || !form.hint_one_url || !form.hint_two_url) {
    throw new Error("请填写答案和两组提示图");
  }
  return {
    creator_name: form.creator_name,
    contact: form.contact,
    puzzle_set_id: 1,
    author_name: form.creator_name || "玩家投稿",
    hint_images: [
      { id: "top", url: form.hint_one_url, label: form.hint_one_label, alt: form.hint_one_label },
      { id: "bottom", url: form.hint_two_url, label: form.hint_two_label, alt: form.hint_two_label }
    ],
    hint_text: "",
    answer: form.answer,
    answer_pinyin: form.answer_pinyin,
    answer_aliases: parseAliases(form.aliases_text),
    candidate_chars: parseCandidates(form.candidate_text),
    default_answer_mode: modes.includes(form.default_answer_mode) ? form.default_answer_mode : modes[0],
    supported_answer_modes: modes,
    blank_template: "",
    category: form.category,
    difficulty: Number(form.difficulty) || 1,
    explanation: form.explanation,
    sort_order: 999
  };
}

function supportedModes(): AnswerMode[] {
  const modes: AnswerMode[] = [];
  if (form.support_manual) modes.push("manual");
  if (form.support_tiles) modes.push("tiles");
  return modes.length ? modes : ["manual", "tiles"];
}

function parseAliases(value: string) {
  return value
    .split(/[\n,，]/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function parseCandidates(value: string): CandidateChar[] {
  return value
    .split("\n")
    .map((line, index) => {
      const parts = line.trim().split(/[\s,，]+/).filter(Boolean);
      return {
        id: `c${index + 1}`,
        char: parts[0] || "",
        pinyin: parts.slice(1).join(" ")
      };
    })
    .filter((candidate) => candidate.char);
}
</script>
