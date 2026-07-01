import { defineStore } from "pinia";
import {
  checkAnswer,
  getHint,
  getPuzzle,
  getPuzzlesBySet,
  getPuzzleSets,
  type AnswerMode,
  type CandidateChar,
  type CheckAnswerResult,
  type HintResult,
  type Puzzle,
  type PuzzleSet
} from "../api/puzzles";

const progressKey = "this-is-pun-progress";

interface GameState {
  sets: PuzzleSet[];
  puzzles: Puzzle[];
  currentPuzzle: Puzzle | null;
  currentIndex: number;
  answerMode: AnswerMode;
  answerSlots: string[];
  selectedTileIds: Array<string | null>;
  hints: HintResult[];
  hintsUsed: number;
  startTime: number;
  loading: boolean;
  submitting: boolean;
  error: string;
  wrongTick: number;
  lastResult: CheckAnswerResult | null;
  progress: Record<string, boolean>;
}

function readProgress(): Record<string, boolean> {
  try {
    return JSON.parse(localStorage.getItem(progressKey) || "{}");
  } catch {
    return {};
  }
}

function writeProgress(progress: Record<string, boolean>) {
  localStorage.setItem(progressKey, JSON.stringify(progress));
}

function splitChars(value: string) {
  return Array.from(value.trim()).filter(Boolean);
}

function mapCandidateIds(chars: string[], candidates: CandidateChar[]) {
  const used = new Set<string>();
  return chars.map((char) => {
    const candidate = candidates.find((item) => item.char === char && !used.has(item.id));
    if (!candidate) return null;
    used.add(candidate.id);
    return candidate.id;
  });
}

export const useGameStore = defineStore("game", {
  state: (): GameState => ({
    sets: [],
    puzzles: [],
    currentPuzzle: null,
    currentIndex: 0,
    answerMode: "manual",
    answerSlots: [],
    selectedTileIds: [],
    hints: [],
    hintsUsed: 0,
    startTime: Date.now(),
    loading: false,
    submitting: false,
    error: "",
    wrongTick: 0,
    lastResult: null,
    progress: readProgress()
  }),

  getters: {
    answerText(state) {
      return state.answerSlots.join("");
    },
    isAnswerEmpty(state) {
      return state.answerSlots.every((char) => !char);
    },
    solvedCount(state) {
      return state.puzzles.filter((puzzle) => state.progress[String(puzzle.id)]).length;
    }
  },

  actions: {
    async bootstrap() {
      if (this.currentPuzzle) return;
      this.loading = true;
      this.error = "";
      try {
        this.sets = await getPuzzleSets();
        const firstSet = this.sets[0];
        if (!firstSet) throw new Error("暂无题库");
        this.puzzles = await getPuzzlesBySet(firstSet.id);
        if (!this.puzzles.length) throw new Error("题库里还没有题目");
        await this.loadPuzzleByIndex(0);
      } catch (error) {
        this.error = error instanceof Error ? error.message : "加载失败";
      } finally {
        this.loading = false;
      }
    },

    async loadPuzzleByIndex(index: number) {
      if (!this.puzzles[index]) return;
      this.loading = true;
      this.error = "";
      try {
        const puzzle = await getPuzzle(this.puzzles[index].id);
        this.currentPuzzle = puzzle;
        this.currentIndex = index;
        this.answerMode = puzzle.default_answer_mode;
        this.resetAnswer();
        this.hints = [];
        this.hintsUsed = 0;
        this.startTime = Date.now();
        this.lastResult = null;
      } catch (error) {
        this.error = error instanceof Error ? error.message : "加载题目失败";
      } finally {
        this.loading = false;
      }
    },

    resetAnswer() {
      const length = this.currentPuzzle?.answer_length ?? 0;
      this.answerSlots = Array.from({ length }, () => "");
      this.selectedTileIds = Array.from({ length }, () => null);
    },

    setMode(mode: AnswerMode) {
      if (!this.currentPuzzle?.supported_answer_modes.includes(mode)) return;
      this.answerMode = mode;
    },

    setManualAnswer(value: string) {
      const length = this.currentPuzzle?.answer_length ?? 0;
      const chars = splitChars(value).slice(0, Math.max(length, 0));
      this.answerSlots = Array.from({ length }, (_, index) => chars[index] || "");
      const mappedIds = mapCandidateIds(this.answerSlots, this.currentPuzzle?.candidate_chars ?? []);
      this.selectedTileIds = Array.from({ length }, (_, index) => mappedIds[index] ?? null);
    },

    clearAnswer() {
      this.resetAnswer();
    },

    chooseCandidate(candidate: CandidateChar) {
      if (this.isTileUsed(candidate.id)) return;
      const targetIndex = this.answerSlots.findIndex((char) => !char);
      if (targetIndex === -1) return;
      this.placeCandidate(candidate, targetIndex);
    },

    placeCandidate(candidate: CandidateChar, targetIndex: number) {
      if (targetIndex < 0 || targetIndex >= this.answerSlots.length) return;
      if (this.isTileUsed(candidate.id)) return;
      if (this.selectedTileIds[targetIndex]) {
        this.removeSlot(targetIndex);
      }
      this.answerSlots[targetIndex] = candidate.char;
      this.selectedTileIds[targetIndex] = candidate.id;
    },

    removeSlot(index: number) {
      if (index < 0 || index >= this.answerSlots.length) return;
      this.answerSlots[index] = "";
      this.selectedTileIds[index] = null;
    },

    isTileUsed(id: string) {
      return this.selectedTileIds.includes(id);
    },

    candidateById(id: string) {
      return this.currentPuzzle?.candidate_chars.find((candidate) => candidate.id === id);
    },

    async requestHint() {
      if (!this.currentPuzzle || this.hintsUsed >= 3) return;
      const nextLevel = this.hintsUsed + 1;
      const hint = await getHint(this.currentPuzzle.id, this.currentPuzzle.attempt_id, nextLevel);
      this.hintsUsed = nextLevel;
      this.hints.push(hint);
    },

    async submitAnswer() {
      if (!this.currentPuzzle) return null;
      if (this.isAnswerEmpty) {
        this.error = "先填一个答案再提交";
        this.wrongTick += 1;
        return null;
      }
      this.submitting = true;
      this.error = "";
      try {
        const result = await checkAnswer(this.currentPuzzle.id, {
          attempt_id: this.currentPuzzle.attempt_id,
          answer: this.answerText,
          answer_mode: this.answerMode,
          elapsed_ms: Date.now() - this.startTime,
          hints_used: this.hintsUsed
        });
        this.lastResult = result;
        if (result.correct) {
          this.progress[String(this.currentPuzzle.id)] = true;
          writeProgress(this.progress);
        } else {
          this.wrongTick += 1;
        }
        return result;
      } catch (error) {
        this.error = error instanceof Error ? error.message : "提交失败";
        return null;
      } finally {
        this.submitting = false;
      }
    },

    async nextPuzzle() {
      const nextIndex = (this.currentIndex + 1) % this.puzzles.length;
      await this.loadPuzzleByIndex(nextIndex);
    }
  }
});
