import { defineStore } from "pinia";
import { getPlayerProgress, loginPlayer, registerPlayer, savePlayerProgress } from "../api/players";

const tokenKey = "this-is-pun-player-token";
const usernameKey = "this-is-pun-player-username";
export const usePlayerStore = defineStore("player", {
  state: () => ({ token: localStorage.getItem(tokenKey) || "", username: localStorage.getItem(usernameKey) || "", progress: {} as Record<string, boolean> }),
  getters: { loggedIn: (state) => Boolean(state.token) },
  actions: {
    async login(username: string, password: string, register = false) { const result = register ? await registerPlayer(username, password) : await loginPlayer(username, password); this.token = result.access_token; this.username = result.username; localStorage.setItem(tokenKey, this.token); localStorage.setItem(usernameKey, this.username); await this.loadProgress(); },
    async loadProgress() { if (!this.token) return; const items = await getPlayerProgress(); this.progress = Object.fromEntries(items.map((item) => [String(item.puzzle_id), true])); },
    async markSolved(puzzleId: number) { if (!this.token) return; await savePlayerProgress(puzzleId); this.progress[String(puzzleId)] = true; },
    logout() { this.token = ""; this.username = ""; this.progress = {}; localStorage.removeItem(tokenKey); localStorage.removeItem(usernameKey); }
  }
});
