import { defineStore } from "pinia";
import { getPlayerProgress, loginPlayer, logoutPlayer, registerPlayer, savePlayerProgress } from "../api/players";

const tokenKey = "this-is-pun-player-token";
const usernameKey = "this-is-pun-player-username";
const guestProgressKey = "this-is-pun-progress";

function readGuestProgress(): Record<string, boolean> {
  try { return JSON.parse(localStorage.getItem(guestProgressKey) || "{}"); } catch { return {}; }
}

export const usePlayerStore = defineStore("player", {
  state: () => ({ token: localStorage.getItem(tokenKey) || "", username: localStorage.getItem(usernameKey) || "", progress: {} as Record<string, boolean> }),
  getters: {
    loggedIn: (state) => Boolean(state.token),
    solvedCount: (state) => Object.values(state.progress).filter(Boolean).length,
    achievements: (state) => [
      ...(Object.values(state.progress).some(Boolean) ? [{ id: "first-solve", title: "首通成就", description: "完成第一道题" }] : []),
      ...(Object.values(state.progress).filter(Boolean).length >= 10 ? [{ id: "ten-solves", title: "十题达成", description: "完成十道题" }] : [])
    ]
  },
  actions: {
    async login(username: string, password: string, register = false) { const result = register ? await registerPlayer(username, password) : await loginPlayer(username, password); this.token = result.access_token; this.username = result.username; localStorage.setItem(tokenKey, this.token); localStorage.setItem(usernameKey, this.username); await this.loadProgress(); },
    async loadProgress() {
      if (!this.token) return;
      const items = await getPlayerProgress();
      this.progress = Object.fromEntries(items.map((item) => [String(item.puzzle_id), true]));
      const guest = readGuestProgress();
      const missing = Object.keys(guest).filter((id) => guest[id] && !this.progress[id] && Number(id) > 0);
      for (const id of missing) { await savePlayerProgress(Number(id)); this.progress[id] = true; }
      if (missing.length || Object.keys(guest).length === 0) localStorage.removeItem(guestProgressKey);
    },
    async markSolved(puzzleId: number) { if (!this.token) return; await savePlayerProgress(puzzleId); this.progress[String(puzzleId)] = true; },
    async logout() { try { await logoutPlayer(); } finally { this.token = ""; this.username = ""; this.progress = {}; localStorage.removeItem(tokenKey); localStorage.removeItem(usernameKey); } }
  }
});
