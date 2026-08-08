import { defineStore } from "pinia";
import { createRoom, joinRoom, openRoomSocket, type Room, type RoomEvent } from "../api/rooms";

export const useRoomStore = defineStore("room", {
  state: () => ({ room: null as Room | null, socket: null as WebSocket | null, connected: false, lastResult: null as any, error: "" }),
  actions: {
    async create(puzzleSetId: number, playerName: string) { this.room = await createRoom(puzzleSetId, playerName); return this.room; },
    async join(roomId: string, playerName: string) { this.room = await joinRoom(roomId, playerName); return this.room; },
    connect(roomId: string) {
      this.disconnect();
      this.socket = openRoomSocket(roomId, (event) => this.receive(event), () => { this.connected = false; });
      this.socket.addEventListener("open", () => { this.connected = true; });
    },
    receive(event: RoomEvent) {
      if (event.room) this.room = event.room;
      if (event.type === "answer_result") this.lastResult = event.payload;
      if (event.type === "error") this.error = String(event.payload || "房间操作失败");
    },
    send(type: string, payload: unknown = {}) { if (this.socket?.readyState === WebSocket.OPEN) this.socket.send(JSON.stringify({ type, payload })); },
    disconnect() { this.socket?.close(); this.socket = null; this.connected = false; },
    clear() { this.disconnect(); this.room = null; this.lastResult = null; this.error = ""; }
  }
});
