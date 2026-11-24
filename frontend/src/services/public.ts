import { request } from "./api";
import type { Invite } from "@/types/api";

export interface AttendancePayload {
  name: string;
  email: string;
  contribute: boolean;
  website: string;
}

export interface MessagePayload extends AttendancePayload {
  message: string;
}

export const confirmAttendance = (p: AttendancePayload) =>
  request<{ message: string }>("/registrations", { method: "POST", body: p });

export const sendMessage = (p: MessagePayload) => request<{ message: string }>("/messages", { method: "POST", body: p });

export const fetchInvite = (token: string, signal?: AbortSignal) =>
  request<Invite>(`/invites/${encodeURIComponent(token)}`, { signal });
