export type FieldErrors = Record<string, string>;

export interface ApiErrorBody {
  code: string;
  message: string;
  fields?: FieldErrors;
}

export interface Admin {
  id: string;
  name: string;
  email: string;
}

export interface SessionInfo {
  admin: Admin;
  csrf_token: string;
}

export interface GuestSummary {
  id: string;
  name: string;
  email: string;
  attending: boolean;
  wants_to_contribute: boolean;
  message_count: number;
  created_at: string;
  updated_at: string;
}

export interface GuestMessage {
  id: string;
  body: string;
  created_at: string;
}

export interface GuestDetail extends Omit<GuestSummary, "message_count"> {
  attending_confirmed_at: string | null;
  messages: GuestMessage[];
  invite_url?: string;
}

export interface Paginated<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface Stats {
  total: number;
  attending: number;
  contribute: number;
  with_message: number;
}

export type GuestFilter = "" | "attending" | "contribute" | "message";
export type SortKey = "created_at" | "name" | "email";
export type SortOrder = "asc" | "desc";

export interface GuestListQuery {
  q: string;
  type: GuestFilter;
  sort: SortKey;
  order: SortOrder;
  page: number;
  page_size: number;
}

export interface Invite {
  guest_name: string;
  date_label: string;
  time: string;
  address_line1: string;
  address_line2: string;
  maps_query: string;
  signature: string;
}
