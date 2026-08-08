import { api, type ApiResponse } from "./client";
import type { AdminPuzzle, AdminUser, ArchivedPuzzle, PuzzleInput, PuzzleSubmission, SubmissionStatus } from "./types";

export interface AuthResult {
  token: string;
  user: AdminUser;
}

export async function registerAdmin(username: string, password: string) {
  const res = await api.post<ApiResponse<AuthResult>>("/admin/auth/register", { username, password });
  return res.data.data;
}

export async function loginAdmin(username: string, password: string) {
  const res = await api.post<ApiResponse<AuthResult>>("/admin/auth/login", { username, password });
  return res.data.data;
}

export async function logoutAdmin() {
  const res = await api.post<ApiResponse<{ logged_out: boolean }>>("/admin/auth/logout");
  return res.data.data;
}
export async function listSubmissions(status: SubmissionStatus | "") {
  const res = await api.get<ApiResponse<PuzzleSubmission[]>>("/admin/submissions", {
    params: status ? { status } : undefined
  });
  return res.data.data;
}

export async function approveSubmission(id: number, reviewNote: string) {
  const res = await api.post<ApiResponse<PuzzleSubmission>>(`/admin/submissions/${id}/approve`, {
    review_note: reviewNote
  });
  return res.data.data;
}

export async function rejectSubmission(id: number, reviewNote: string) {
  const res = await api.post<ApiResponse<PuzzleSubmission>>(`/admin/submissions/${id}/reject`, {
    review_note: reviewNote
  });
  return res.data.data;
}

export async function batchReviewSubmissions(ids: number[], action: "approve" | "reject", reviewNote: string) {
  const res = await api.post<ApiResponse<PuzzleSubmission[]>>("/admin/submissions/batch-review", {
    submission_ids: ids,
    action,
    review_note: reviewNote
  });
  return res.data.data;
}

export async function listPuzzles() {
  const res = await api.get<ApiResponse<AdminPuzzle[]>>("/admin/puzzles");
  return res.data.data;
}

export async function createPuzzle(payload: PuzzleInput) {
  const res = await api.post<ApiResponse<AdminPuzzle>>("/admin/puzzles", payload);
  return res.data.data;
}

export async function updatePuzzle(id: number, payload: PuzzleInput) {
  const res = await api.put<ApiResponse<AdminPuzzle>>(`/admin/puzzles/${id}`, payload);
  return res.data.data;
}

export async function deletePuzzle(id: number) {
  const res = await api.delete<ApiResponse<{ deleted: boolean }>>(`/admin/puzzles/${id}`);
  return res.data.data;
}

export async function listArchivedPuzzles() {
  const res = await api.get<ApiResponse<ArchivedPuzzle[]>>("/admin/puzzles/archived");
  return res.data.data;
}

export async function restorePuzzle(id: number) {
  const res = await api.post<ApiResponse<{ restored: boolean; id: number }>>(`/admin/puzzles/${id}/restore`);
  return res.data.data;
}
