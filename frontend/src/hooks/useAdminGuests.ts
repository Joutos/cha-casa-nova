import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { deleteGuest, getGuest, getStats, listGuests, updateGuest, type GuestPatch } from "@/services/admin";
import type { GuestListQuery } from "@/types/api";

export const adminKeys = {
  all: ["admin"] as const,
  list: (q: GuestListQuery) => ["admin", "guests", q] as const,
  detail: (id: string) => ["admin", "guest", id] as const,
  stats: ["admin", "stats"] as const,
};

export function useGuestList(q: GuestListQuery) {
  return useQuery({
    queryKey: adminKeys.list(q),
    queryFn: ({ signal }) => listGuests(q, signal),
    placeholderData: keepPreviousData,
  });
}

export function useStats() {
  return useQuery({ queryKey: adminKeys.stats, queryFn: ({ signal }) => getStats(signal) });
}

export function useGuestDetail(id: string | null) {
  return useQuery({
    queryKey: adminKeys.detail(id ?? ""),
    queryFn: ({ signal }) => getGuest(id!, signal),
    enabled: !!id,
  });
}

export function useUpdateGuest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: GuestPatch }) => updateGuest(id, patch),
    onSuccess: (guest) => {
      qc.setQueryData(adminKeys.detail(guest.id), guest);
      qc.invalidateQueries({ queryKey: ["admin", "guests"] });
      qc.invalidateQueries({ queryKey: adminKeys.stats });
      toast.success("Alteração salva.");
    },
    onError: (err) => toast.error(err.message),
  });
}

export function useDeleteGuest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteGuest(id),
    onSuccess: (_, id) => {
      qc.removeQueries({ queryKey: adminKeys.detail(id) });
      qc.invalidateQueries({ queryKey: ["admin", "guests"] });
      qc.invalidateQueries({ queryKey: adminKeys.stats });
      toast.success("Inscrição excluída.");
    },
    onError: (err) => toast.error(err.message),
  });
}
