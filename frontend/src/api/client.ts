import axios from "axios";

export const api = axios.create({
  baseURL: "/api/v1",
  timeout: 10000
});

export interface ApiResponse<T> {
  data: T;
  error?: string;
}
