import { api, type ApiResponse } from "./client";

export interface PlayerAuth { access_token: string; refresh_token: string; expires_in: number; player_id: number; username: string }
export interface PlayerProgress { puzzle_id: number; solved_at: string }
export async function registerPlayer(username: string, password: string) { return (await api.post<ApiResponse<PlayerAuth>>("/players/register", { username, password })).data.data; }
export async function loginPlayer(username: string, password: string) { return (await api.post<ApiResponse<PlayerAuth>>("/players/login", { username, password })).data.data; }
export async function getPlayerProgress() { return (await api.get<ApiResponse<PlayerProgress[]>>("/players/me/progress")).data.data; }
export async function savePlayerProgress(puzzleId: number) { return (await api.put<ApiResponse<{ saved: boolean }>>(`/players/me/progress/${puzzleId}`)).data.data; }
