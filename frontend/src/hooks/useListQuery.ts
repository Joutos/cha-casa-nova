import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router-dom";
import type { GuestFilter, GuestListQuery, SortKey, SortOrder } from "@/types/api";

const FILTERS: GuestFilter[] = ["", "attending", "contribute", "message"];
const SORTS: SortKey[] = ["created_at", "name", "email"];
export const PAGE_SIZE = 15;

// Filtros, ordenação e página ficam na URL: dá para recarregar, voltar e compartilhar.
export function useListQuery() {
  const [params, setParams] = useSearchParams();

  const query = useMemo<GuestListQuery>(() => {
    const type = params.get("type") as GuestFilter;
    const sort = params.get("sort") as SortKey;
    const order = params.get("order") as SortOrder;
    const page = Number(params.get("page"));
    return {
      q: params.get("q") ?? "",
      type: FILTERS.includes(type) ? type : "",
      sort: SORTS.includes(sort) ? sort : "created_at",
      order: order === "asc" ? "asc" : "desc",
      page: Number.isInteger(page) && page > 0 ? page : 1,
      page_size: PAGE_SIZE,
    };
  }, [params]);

  const update = useCallback(
    (patch: Partial<GuestListQuery>) => {
      setParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          const resetPage = Object.keys(patch).some((k) => k !== "page");
          Object.entries(patch).forEach(([k, v]) => {
            if (v === "" || v === undefined || (k === "page" && v === 1)) next.delete(k);
            else next.set(k, String(v));
          });
          if (resetPage) next.delete("page");
          return next;
        },
        { replace: true },
      );
    },
    [setParams],
  );

  const toggleSort = useCallback(
    (key: SortKey) => {
      if (query.sort === key) update({ sort: key, order: query.order === "asc" ? "desc" : "asc" });
      else update({ sort: key, order: key === "created_at" ? "desc" : "asc" });
    },
    [query.sort, query.order, update],
  );

  return { query, update, toggleSort };
}
