import { api, type ApiResponse } from "./client";

export interface UploadResult { id: number; url: string; content_type: string; size: number }
export async function uploadMedia(file: File) { const form = new FormData(); form.append("file", file); return (await api.post<ApiResponse<UploadResult>>("/media/upload", form)).data.data; }
