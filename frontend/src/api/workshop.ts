import { api, type ApiResponse } from "./client";
import type { HintImage } from "./puzzles";

export interface WorkshopItem {
  id: number;
  puzzle_set_id: number;
  author_name: string;
  hint_images: HintImage[];
  category: string;
  difficulty: number;
}

export interface WorkshopPage {
  items: WorkshopItem[];
  page: number;
  page_size: number;
  total: number;
}

export async function listWorkshop(params: { category?: string; difficulty?: number; page?: number; page_size?: number }) {
  const res = await api.get<ApiResponse<WorkshopPage>>("/workshop/submissions", { params });
  return res.data.data;
}
