import { api, type ApiResponse } from "./client";

export type AnswerMode = "manual" | "tiles";

export interface HintImage {
  id: string;
  url: string;
  label: string;
  alt: string;
}

export interface CandidateChar {
  id: string;
  char: string;
  pinyin: string;
}

export interface PuzzleSet {
  id: number;
  name: string;
  description: string;
  category: string;
  domain_type: string;
  cover_url: string;
  puzzle_count: number;
}

export interface Puzzle {
  id: number;
  attempt_id: string;
  puzzle_set_id: number;
  author_name: string;
  hint_images: HintImage[];
  hint_text: string;
  answer_length: number;
  candidate_chars: CandidateChar[];
  default_answer_mode: AnswerMode;
  supported_answer_modes: AnswerMode[];
  blank_template: string;
  category: string;
  difficulty: number;
  sort_order: number;
}

export interface CheckAnswerResult {
  correct: boolean;
  score: number;
  answer?: string;
  answer_mode: AnswerMode;
  normalized: string;
  expected_chars: number;
  elapsed_ms: number;
  explanation?: string;
  message: string;
}

export interface HintResult {
  level: number;
  text: string;
  score_if_correct: number;
}

export async function getPuzzleSets() {
  const res = await api.get<ApiResponse<PuzzleSet[]>>("/puzzle-sets");
  return res.data.data;
}

export async function getPuzzlesBySet(setId: number) {
  const res = await api.get<ApiResponse<Puzzle[]>>(`/puzzle-sets/${setId}/puzzles`);
  return res.data.data;
}

export async function getPuzzle(id: number) {
  const res = await api.get<ApiResponse<Puzzle>>(`/puzzles/${id}`);
  return res.data.data;
}

export async function checkAnswer(puzzleId: number, payload: { attempt_id: string; answer: string; answer_mode: AnswerMode; elapsed_ms: number; hints_used: number }) {
  const res = await api.post<ApiResponse<CheckAnswerResult>>(`/puzzles/${puzzleId}/check`, payload);
  return res.data.data;
}

export async function getHint(puzzleId: number, attemptId: string, level: number) {
  const res = await api.post<ApiResponse<HintResult>>(`/puzzles/${puzzleId}/hint`, { attempt_id: attemptId, level });
  return res.data.data;
}
