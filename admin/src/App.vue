<template>
  <main class="admin-app">
    <section v-if="!token" class="auth-shell">
      <div class="auth-panel">
        <div class="brand-block">
          <span>这是谐音梗</span>
          <h1>管理员后台</h1>
          <p>创建账号后即可登录审核玩家投稿。</p>
        </div>

        <div class="auth-tabs">
          <button type="button" :class="{ active: authMode === 'login' }" @click="authMode = 'login'">登录</button>
          <button type="button" :class="{ active: authMode === 'register' }" @click="authMode = 'register'">创建账号</button>
        </div>

        <form class="auth-form" @submit.prevent="submitAuth">
          <label>
            账号
            <input v-model.trim="authForm.username" type="text" autocomplete="username" placeholder="admin" />
          </label>
          <label>
            密码
            <input v-model="authForm.password" type="password" autocomplete="current-password" placeholder="至少 6 位" />
          </label>
          <p v-if="error" class="error-text">{{ error }}</p>
          <button type="submit" :disabled="loading">{{ loading ? "处理中" : authMode === "login" ? "登录后台" : "创建并登录" }}</button>
        </form>
      </div>
    </section>

    <template v-else>
      <header class="topbar">
        <div>
          <strong>这是谐音梗 · 管理后台</strong>
          <span>{{ username }}</span>
        </div>
        <nav>
          <button type="button" :class="{ active: activeView === 'review' }" @click="activeView = 'review'">投稿审核</button>
          <button type="button" :class="{ active: activeView === 'puzzles' }" @click="activeView = 'puzzles'">题库管理</button>
          <a href="http://localhost:5173/submit" target="_blank" rel="noreferrer">玩家投稿页</a>
          <a href="http://localhost:5173/game" target="_blank" rel="noreferrer">试玩</a>
          <button type="button" @click="logout">退出</button>
        </nav>
      </header>

      <section v-if="activeView === 'review'" class="dashboard">
        <aside class="queue-panel">
          <div class="stats-row">
            <div>
              <span>待审核</span>
              <strong>{{ statusCounts.pending }}</strong>
            </div>
            <div>
              <span>已通过</span>
              <strong>{{ statusCounts.approved }}</strong>
            </div>
            <div>
              <span>正式题库</span>
              <strong>{{ puzzleCount }}</strong>
            </div>
          </div>

          <div class="filter-row">
            <button type="button" :class="{ active: statusFilter === 'pending' }" @click="setFilter('pending')">待审核</button>
            <button type="button" :class="{ active: statusFilter === 'approved' }" @click="setFilter('approved')">已通过</button>
            <button type="button" :class="{ active: statusFilter === 'rejected' }" @click="setFilter('rejected')">已拒绝</button>
            <button type="button" :class="{ active: statusFilter === '' }" @click="setFilter('')">全部</button>
          </div>

          <div class="submission-list">
            <button
              v-for="submission in submissions"
              :key="submission.id"
              type="button"
              :class="{ active: selected?.id === submission.id }"
              @click="selected = submission"
            >
              <strong>#{{ submission.id }} {{ submission.answer }}</strong>
              <span>{{ submission.creator_name }} · {{ statusLabel(submission.status) }}</span>
            </button>
            <p v-if="!submissions.length" class="empty-text">当前没有投稿。</p>
          </div>
        </aside>

        <section class="detail-panel">
          <template v-if="selected">
            <div class="detail-head">
              <div>
                <span class="status-pill" :class="selected.status">{{ statusLabel(selected.status) }}</span>
                <h1>{{ selected.answer }}</h1>
                <p>{{ selected.category || "未分类" }} · {{ selected.default_answer_mode === "tiles" ? "选字答题" : "手动输入" }}</p>
              </div>
              <div class="meta-box">
                <span>投稿人</span>
                <strong>{{ selected.creator_name || "匿名玩家" }}</strong>
                <small>{{ selected.contact || "未留联系方式" }}</small>
              </div>
            </div>

            <div class="hint-preview">
              <article v-for="image in selected.hint_images" :key="image.id">
                <span v-if="image.url.startsWith('emoji:')">{{ image.url.replace("emoji:", "") }}</span>
                <img v-else :src="image.url" :alt="image.alt" />
                <strong>{{ image.label }}</strong>
              </article>
            </div>

            <dl class="detail-grid">
              <div>
                <dt>答案拼音</dt>
                <dd>{{ selected.answer_pinyin || "未填写" }}</dd>
              </div>
              <div>
                <dt>答案长度</dt>
                <dd>{{ selected.answer_length }} 个字</dd>
              </div>
              <div>
                <dt>别名</dt>
                <dd>{{ selected.answer_aliases.length ? selected.answer_aliases.join("、") : "无" }}</dd>
              </div>
              <div>
                <dt>候选字</dt>
                <dd>{{ selected.candidate_chars.map((item) => item.char).join(" ") }}</dd>
              </div>
            </dl>

            <section class="explain-block">
              <h2>解释</h2>
              <p>{{ selected.explanation || "玩家没有填写解释。" }}</p>
            </section>

            <section class="review-block">
              <label>
                审核备注
                <textarea v-model="reviewNote" rows="3" placeholder="可选，拒绝时建议写原因"></textarea>
              </label>
              <div v-if="selected.status === 'pending'" class="review-actions">
                <button class="reject-button" type="button" :disabled="loading" @click="review('reject')">拒绝</button>
                <button class="approve-button" type="button" :disabled="loading" @click="review('approve')">通过并入库</button>
              </div>
              <p v-else class="reviewed-text">
                审核人：{{ selected.reviewed_by || "-" }}，备注：{{ selected.review_note || "无" }}
              </p>
            </section>
          </template>
          <p v-else class="empty-detail">请选择一条投稿查看详情。</p>
        </section>
      </section>

      <section v-else class="puzzle-admin">
        <aside class="puzzle-list-panel">
          <div class="puzzle-list-head">
            <div>
              <h1>题库管理</h1>
              <p>直接维护正式题库，玩家端刷新后生效。</p>
            </div>
            <button type="button" @click="resetPuzzleForm">新增题目</button>
          </div>

          <div class="puzzle-list">
            <button
              v-for="puzzle in puzzles"
              :key="puzzle.id"
              type="button"
              :class="{ active: selectedPuzzle?.id === puzzle.id }"
              @click="selectPuzzle(puzzle)"
            >
              <strong>D {{ puzzle.id }} · {{ puzzle.answer }}</strong>
              <span>{{ puzzle.category || "未分类" }} · {{ puzzle.default_answer_mode === "tiles" ? "选字" : "输入" }}</span>
            </button>
            <p v-if="!puzzles.length" class="empty-text">题库暂无题目。</p>
          </div>
        </aside>

        <form class="puzzle-editor" @submit.prevent="savePuzzle">
          <div class="puzzle-editor-head">
            <div>
              <span>{{ selectedPuzzle ? "当前选中" : "新增题目" }}</span>
              <h2>{{ selectedPuzzle?.answer || "填写新题" }}</h2>
            </div>
            <div class="puzzle-editor-actions">
              <button class="secondary-action" type="button" @click="resetPuzzleForm">清空</button>
              <button
                v-if="selectedPuzzle"
                class="reject-button"
                type="button"
                :disabled="loading"
                @click="removePuzzle"
              >
                删除题目
              </button>
              <button class="approve-button" type="submit" :disabled="loading">
                {{ loading ? "保存中" : selectedPuzzle ? "保存修改" : "新增到题库" }}
              </button>
            </div>
          </div>

          <p v-if="puzzleMessage" class="success-text">{{ puzzleMessage }}</p>
          <p v-if="puzzleError" class="error-text">{{ puzzleError }}</p>

          <div class="form-grid">
            <label>
              答案
              <input v-model.trim="puzzleForm.answer" type="text" placeholder="例如：将心比心" />
            </label>
            <label>
              答案拼音
              <input v-model.trim="puzzleForm.answer_pinyin" type="text" placeholder="可留空，后端自动生成" />
            </label>
            <label>
              分类
              <input v-model.trim="puzzleForm.category" type="text" placeholder="成语 / 生活 / 动物" />
            </label>
            <label>
              作者
              <input v-model.trim="puzzleForm.author_name" type="text" placeholder="管理员" />
            </label>
            <label>
              排序
              <input v-model.number="puzzleForm.sort_order" type="number" min="0" />
            </label>
            <label>
              难度
              <input v-model.number="puzzleForm.difficulty" type="number" min="1" max="5" />
            </label>
          </div>

          <div class="mode-editor">
            <label>
              默认模式
              <select v-model="puzzleForm.default_answer_mode">
                <option value="manual">手动输入</option>
                <option value="tiles">选字答题</option>
              </select>
            </label>
            <label class="inline-check">
              <input v-model="puzzleForm.support_manual" type="checkbox" />
              支持输入
            </label>
            <label class="inline-check">
              <input v-model="puzzleForm.support_tiles" type="checkbox" />
              支持选字
            </label>
          </div>

          <section class="editor-section">
            <h3>提示图</h3>
            <div class="form-grid two">
              <label>
                上图地址
                <input v-model.trim="puzzleForm.hint_one_url" type="text" placeholder="emoji:🐦 或图片 URL" />
              </label>
              <label>
                上图文案
                <input v-model.trim="puzzleForm.hint_one_label" type="text" placeholder="这是小鸟" />
              </label>
              <label>
                下图地址
                <input v-model.trim="puzzleForm.hint_two_url" type="text" placeholder="emoji:🧍 或图片 URL" />
              </label>
              <label>
                下图文案
                <input v-model.trim="puzzleForm.hint_two_label" type="text" placeholder="这是____" />
              </label>
            </div>
          </section>

          <section class="editor-section">
            <div class="section-title">
              <h3>候选字</h3>
              <button type="button" @click="generatePuzzleCandidates">按答案生成</button>
            </div>
            <textarea
              v-model="puzzleForm.candidate_text"
              rows="5"
              placeholder="每行一个候选字，可写：字 拼音。例如：&#10;将 jiang&#10;心 xin"
            ></textarea>
          </section>

          <section class="editor-section">
            <h3>别名与解释</h3>
            <label>
              答案别名
              <textarea v-model="puzzleForm.aliases_text" rows="3" placeholder="一行一个，或用逗号分隔"></textarea>
            </label>
            <label>
              解释
              <textarea v-model="puzzleForm.explanation" rows="4" placeholder="答对后展示的解释"></textarea>
            </label>
          </section>
        </form>
      </section>
    </template>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import {
  approveSubmission,
  createPuzzle,
  deletePuzzle,
  listPuzzles,
  listSubmissions,
  loginAdmin,
  logoutAdmin,
  registerAdmin,
  rejectSubmission,
  updatePuzzle
} from "./api/admin";
import type { AdminPuzzle, AnswerMode, CandidateChar, PuzzleInput, PuzzleSubmission, SubmissionStatus } from "./api/types";

interface PuzzleForm {
  answer: string;
  answer_pinyin: string;
  aliases_text: string;
  category: string;
  author_name: string;
  difficulty: number;
  sort_order: number;
  default_answer_mode: AnswerMode;
  support_manual: boolean;
  support_tiles: boolean;
  hint_one_url: string;
  hint_one_label: string;
  hint_two_url: string;
  hint_two_label: string;
  candidate_text: string;
  explanation: string;
}

const authMode = ref<"login" | "register">("login");
const authForm = reactive({ username: "", password: "" });
const token = ref(localStorage.getItem("this-is-pun-admin-token") || "");
const username = ref(localStorage.getItem("this-is-pun-admin-username") || "");
const activeView = ref<"review" | "puzzles">("review");
const submissions = ref<PuzzleSubmission[]>([]);
const allSubmissions = ref<PuzzleSubmission[]>([]);
const selected = ref<PuzzleSubmission | null>(null);
const puzzles = ref<AdminPuzzle[]>([]);
const selectedPuzzle = ref<AdminPuzzle | null>(null);
const statusFilter = ref<SubmissionStatus | "">("pending");
const puzzleCount = ref(0);
const loading = ref(false);
const error = ref("");
const reviewNote = ref("");
const puzzleMessage = ref("");
const puzzleError = ref("");
const puzzleForm = reactive<PuzzleForm>(emptyPuzzleForm());

const statusCounts = computed(() => {
  return allSubmissions.value.reduce(
    (acc, item) => {
      acc[item.status] += 1;
      return acc;
    },
    { pending: 0, approved: 0, rejected: 0 }
  );
});

onMounted(() => {
  if (token.value) {
    refreshData();
  }
});

async function submitAuth() {
  loading.value = true;
  error.value = "";
  try {
    const result =
      authMode.value === "login"
        ? await loginAdmin(authForm.username, authForm.password)
        : await registerAdmin(authForm.username, authForm.password);
    token.value = result.token;
    username.value = result.user.username;
    localStorage.setItem("this-is-pun-admin-token", result.token);
    localStorage.setItem("this-is-pun-admin-username", result.user.username);
    await refreshData();
  } catch (err) {
    error.value = authMode.value === "login" ? "登录失败，请检查账号密码" : "创建失败，账号可能已存在";
  } finally {
    loading.value = false;
  }
}

async function refreshData() {
  error.value = "";
  try {
    const [filtered, all, puzzleList] = await Promise.all([
      listSubmissions(statusFilter.value),
      listSubmissions(""),
      listPuzzles()
    ]);
    submissions.value = filtered;
    allSubmissions.value = all;
    puzzles.value = puzzleList;
    puzzleCount.value = puzzleList.length;
    selected.value = filtered.find((item) => item.id === selected.value?.id) ?? filtered[0] ?? null;
    if (selectedPuzzle.value) {
      const refreshedPuzzle = puzzleList.find((item) => item.id === selectedPuzzle.value?.id) ?? null;
      selectedPuzzle.value = refreshedPuzzle;
      if (!refreshedPuzzle) {
        Object.assign(puzzleForm, {
          ...emptyPuzzleForm(),
          author_name: username.value || "管理员",
          sort_order: puzzleList.reduce((max, puzzle) => Math.max(max, puzzle.sort_order), 0) + 1
        });
        puzzleError.value = "当前题目已被删除或下架，请重新选择。";
      }
    }
    reviewNote.value = "";
  } catch (err) {
    logout();
  }
}

function setFilter(status: SubmissionStatus | "") {
  statusFilter.value = status;
  refreshData();
}

async function review(action: "approve" | "reject") {
  if (!selected.value) return;
  loading.value = true;
  try {
    const next =
      action === "approve"
        ? await approveSubmission(selected.value.id, reviewNote.value)
        : await rejectSubmission(selected.value.id, reviewNote.value);
    selected.value = next;
    await refreshData();
  } finally {
    loading.value = false;
  }
}

function statusLabel(status: SubmissionStatus) {
  if (status === "approved") return "已通过";
  if (status === "rejected") return "已拒绝";
  return "待审核";
}

function emptyPuzzleForm(): PuzzleForm {
  return {
    answer: "",
    answer_pinyin: "",
    aliases_text: "",
    category: "成语",
    author_name: username.value || "管理员",
    difficulty: 1,
    sort_order: 999,
    default_answer_mode: "tiles",
    support_manual: true,
    support_tiles: true,
    hint_one_url: "emoji:❓",
    hint_one_label: "提示图 1",
    hint_two_url: "emoji:❓",
    hint_two_label: "提示图 2",
    candidate_text: "",
    explanation: ""
  };
}

function resetPuzzleForm() {
  Object.assign(puzzleForm, {
    ...emptyPuzzleForm(),
    author_name: username.value || "管理员",
    sort_order: puzzles.value.reduce((max, puzzle) => Math.max(max, puzzle.sort_order), 0) + 1
  });
  selectedPuzzle.value = null;
  puzzleMessage.value = "";
  puzzleError.value = "";
}

function selectPuzzle(puzzle: AdminPuzzle) {
  selectedPuzzle.value = puzzle;
  puzzleMessage.value = "";
  puzzleError.value = "";
  Object.assign(puzzleForm, {
    answer: puzzle.answer,
    answer_pinyin: puzzle.answer_pinyin,
    aliases_text: puzzle.answer_aliases.join("\n"),
    category: puzzle.category,
    author_name: puzzle.author_name,
    difficulty: puzzle.difficulty,
    sort_order: puzzle.sort_order,
    default_answer_mode: puzzle.default_answer_mode,
    support_manual: puzzle.supported_answer_modes.includes("manual"),
    support_tiles: puzzle.supported_answer_modes.includes("tiles"),
    hint_one_url: puzzle.hint_images[0]?.url || "emoji:❓",
    hint_one_label: puzzle.hint_images[0]?.label || "提示图 1",
    hint_two_url: puzzle.hint_images[1]?.url || "emoji:❓",
    hint_two_label: puzzle.hint_images[1]?.label || "提示图 2",
    candidate_text: puzzle.candidate_chars.map((candidate) => `${candidate.char} ${candidate.pinyin}`.trim()).join("\n"),
    explanation: puzzle.explanation
  });
}

function generatePuzzleCandidates() {
  puzzleForm.candidate_text = Array.from(puzzleForm.answer).map((char) => char.trim()).filter(Boolean).join("\n");
}

async function savePuzzle() {
  loading.value = true;
  puzzleMessage.value = "";
  puzzleError.value = "";
  try {
    const saved = selectedPuzzle.value
      ? await updatePuzzle(selectedPuzzle.value.id, toPuzzlePayload())
      : await createPuzzle(toPuzzlePayload());
    puzzleMessage.value = selectedPuzzle.value ? `已保存修改：${saved.answer}` : `已新增题目：${saved.answer}`;
    await refreshData();
    selectPuzzle(saved);
  } catch (err) {
    puzzleError.value = err instanceof Error ? err.message : "新增失败";
  } finally {
    loading.value = false;
  }
}

async function removePuzzle() {
  if (!selectedPuzzle.value) return;
  if (!window.confirm(`确定删除「${selectedPuzzle.value.answer}」吗？删除后玩家端题库也会移除。`)) return;
  loading.value = true;
  puzzleMessage.value = "";
  puzzleError.value = "";
  try {
    await deletePuzzle(selectedPuzzle.value.id);
    puzzleMessage.value = "题目已删除";
    selectedPuzzle.value = null;
    await refreshData();
    resetPuzzleForm();
  } catch (err) {
    puzzleError.value = err instanceof Error ? err.message : "删除失败";
  } finally {
    loading.value = false;
  }
}

function toPuzzlePayload(): PuzzleInput {
  const modes = supportedPuzzleModes();
  return {
    puzzle_set_id: 1,
    author_name: puzzleForm.author_name || username.value || "管理员",
    hint_images: [
      { id: "top", url: puzzleForm.hint_one_url, label: puzzleForm.hint_one_label, alt: puzzleForm.hint_one_label },
      { id: "bottom", url: puzzleForm.hint_two_url, label: puzzleForm.hint_two_label, alt: puzzleForm.hint_two_label }
    ],
    hint_text: "",
    answer: puzzleForm.answer,
    answer_pinyin: puzzleForm.answer_pinyin,
    answer_aliases: parseAliases(puzzleForm.aliases_text),
    candidate_chars: parseCandidates(puzzleForm.candidate_text),
    default_answer_mode: modes.includes(puzzleForm.default_answer_mode) ? puzzleForm.default_answer_mode : modes[0],
    supported_answer_modes: modes,
    blank_template: "",
    category: puzzleForm.category,
    difficulty: Number(puzzleForm.difficulty) || 1,
    explanation: puzzleForm.explanation,
    sort_order: Number(puzzleForm.sort_order) || 999
  };
}

function supportedPuzzleModes(): AnswerMode[] {
  const modes: AnswerMode[] = [];
  if (puzzleForm.support_manual) modes.push("manual");
  if (puzzleForm.support_tiles) modes.push("tiles");
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

async function logout() {
  if (token.value) {
    try {
      await logoutAdmin();
    } catch {
      // Local logout still wins if the server session is already gone.
    }
  }
  clearAdminSession();
}

function clearAdminSession() {
  token.value = "";
  username.value = "";
  localStorage.removeItem("this-is-pun-admin-token");
  localStorage.removeItem("this-is-pun-admin-username");
  submissions.value = [];
  allSubmissions.value = [];
  puzzles.value = [];
  selected.value = null;
  selectedPuzzle.value = null;
}
</script>
