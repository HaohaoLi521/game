import { api, type ApiResponse } from "./client";

export type RoomStatus = "waiting" | "ready" | "playing" | "finished";

export interface RoomPlayer {
  id: string;
  name: string;
  ready: boolean;
  score: number;
  answered: boolean;
  present: boolean;
}

export interface Room {
  id: string;
  host_id: string;
  status: RoomStatus;
  players: RoomPlayer[];
  puzzle_ids: number[];
  current_puzzle: number;
  created_at: string;
  updated_at: string;
}

export interface RoomEvent { type: string; room?: Room; payload?: unknown }

export async function createRoom(puzzleSetId: number, playerName: string) {
  const res = await api.post<ApiResponse<Room>>("/rooms", { puzzle_set_id: puzzleSetId, player_name: playerName });
  return res.data.data;
}

export async function joinRoom(roomId: string, playerName: string) {
  const res = await api.post<ApiResponse<Room>>(`/rooms/${encodeURIComponent(roomId)}/join`, { player_name: playerName });
  return res.data.data;
}

export function openRoomSocket(roomId: string, onEvent: (event: RoomEvent) => void, onClose: () => void) {
  const token = localStorage.getItem("this-is-pun-player-token");
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const socket = new WebSocket(`${protocol}//${window.location.host}/api/v1/rooms/${encodeURIComponent(roomId)}/ws?access_token=${encodeURIComponent(token || "")}`);
  socket.addEventListener("message", (event) => { try { onEvent(JSON.parse(event.data) as RoomEvent); } catch { onEvent({ type: "error", payload: "invalid server event" }); } });
  socket.addEventListener("close", onClose);
  return socket;
}
