import axios from "axios";

export const api = axios.create({
  baseURL: "/api/v1",
  timeout: 10000
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("this-is-pun-player-token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

export interface ApiResponse<T> {
  data: T;
  error?: string;
}
