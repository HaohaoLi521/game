export type AnswerMode = "manual" | "tiles";
export type SubmissionStatus = "pending" | "approved" | "rejected";

export interface AdminUser {
  id: number;
  username: string;
  created_at: string;
}

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

export interface PuzzlePublic {
  id: number;
  puzzle_set_id: number;
  answer_length: number;
  sort_order: number;
}

export interface PuzzleSubmission {
  id: number;
  creator_name: string;
  contact: string;
  status: SubmissionStatus;
  review_note: string;
  reviewed_by: string;
  puzzle_set_id: number;
  author_name: string;
  hint_images: HintImage[];
  hint_text: string;
  answer: string;
  answer_pinyin: string;
  answer_aliases: string[];
  answer_length: number;
  candidate_chars: CandidateChar[];
  default_answer_mode: AnswerMode;
  supported_answer_modes: AnswerMode[];
  blank_template: string;
  category: string;
  difficulty: number;
  explanation: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
  approved_puzzle?: PuzzlePublic;
}

export interface PuzzleInput {
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

export interface AdminPuzzle extends PuzzleInput {
  id: number;
  answer_length: number;
}
