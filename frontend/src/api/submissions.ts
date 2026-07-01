import { api, type ApiResponse } from "./client";
import type { AnswerMode, CandidateChar, HintImage } from "./puzzles";

export interface SubmissionInput {
  creator_name: string;
  contact: string;
  puzzle_set_id: number;
  author_name: string;
  hint_images: HintImage[];
  hint_text: string;
  answer: string;
  answer_pinyin: string;
  answer_aliases: string[];
  candidate_chars: CandidateChar[];
  default_answer_mode: AnswerMode;
  supported_answer_modes: AnswerMode[];
  blank_template: string;
  category: string;
  difficulty: number;
  explanation: string;
  sort_order: number;
}

export interface PuzzleSubmission {
  id: number;
  creator_name: string;
  contact: string;
  status: "pending" | "approved" | "rejected";
  answer: string;
  created_at: string;
}

export async function createSubmission(payload: SubmissionInput) {
  const res = await api.post<ApiResponse<PuzzleSubmission>>("/submissions", payload);
  return res.data.data;
}
