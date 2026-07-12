import axios from "axios";

export interface ApiResponse<T> {
  data: T;
}

export const api = axios.create({
  baseURL: "/api/v1",
  timeout: 10000
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("this-is-pun-admin-token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const url = error.config?.url || "";
    const isAuthRequest = url.includes("/admin/auth/login") || url.includes("/admin/auth/register");
    if (error.response?.status === 401 && !isAuthRequest) {
      window.dispatchEvent(new Event("this-is-pun-admin-session-expired"));
    }
    return Promise.reject(error);
  }
);
