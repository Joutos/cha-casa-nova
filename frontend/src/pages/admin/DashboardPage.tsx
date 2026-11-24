import { useEffect, useState } from "react";
import { AdminLayout } from "@/layouts/AdminLayout";
import { FilterTabs } from "@/components/admin/FilterTabs";
import { GuestTable } from "@/components/admin/GuestTable";
import { Pagination } from "@/components/admin/Pagination";
import { GuestDetailDialog } from "@/components/admin/GuestDetailDialog";
import { Spinner } from "@/components/Spinner";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useGuestList, useStats } from "@/hooks/useAdminGuests";
import { useListQuery } from "@/hooks/useListQuery";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { exportUrl } from "@/services/admin";
import { cn } from "@/lib/utils";

export default function DashboardPage() {
  useDocumentTitle("Convidados");
  const { query, update, toggleSort } = useListQuery();
  const [search, setSearch] = useState(query.q);
  const debounced = useDebouncedValue(search, 350);
  const [openId, setOpenId] = useState<string | null>(null);

  useEffect(() => {
    if (debounced.trim() !== query.q) update({ q: debounced.trim() });
  }, [debounced, query.q, update]);

  const list = useGuestList(query);
  const stats = useStats();
  const data = list.data;
  const filtered = query.q !== "" || query.type !== "";

  return (
    <AdminLayout>
      <div className="mx-auto max-w-[86rem]">
        <h1 className="mt-4 text-title">Lista de convidados</h1>

        <div className="mt-6 flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
          <FilterTabs value={query.type} stats={stats.data} onChange={(type) => update({ type })} />
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
            <div className="relative sm:w-80">
              <label htmlFor="busca" className="sr-only">
                Buscar por nome, e-mail ou mensagem
              </label>
              <Input
                id="busca"
                type="search"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="buscar nome, e-mail…"
                maxLength={100}
                className="h-12 px-5 text-xl"
              />
            </div>
            <a
              href={exportUrl(query)}
              download
              className={cn(buttonVariants({ size: "md" }), "whitespace-nowrap")}
            >
              Exportar CSV
            </a>
          </div>
        </div>

        <div className="on-light relative mt-6 overflow-hidden rounded-pill bg-cream text-brand" aria-busy={list.isFetching}>
          {list.isFetching && data && (
            <div className="absolute inset-x-0 top-0 h-1 animate-pulse bg-brand/40" role="status" aria-label="Atualizando lista" />
          )}

          {list.isPending && (
            <p className="flex items-center justify-center gap-3 py-20 text-2xl" role="status">
              <Spinner /> Carregando convidados…
            </p>
          )}

          {list.isError && (
            <div className="py-16 text-center">
              <p className="text-2xl">Não foi possível carregar a lista.</p>
              <p className="mt-2 text-lg text-brand/80">{list.error.message}</p>
              <Button variant="solid" size="md" className="mt-6" onClick={() => list.refetch()}>
                Tentar de novo
              </Button>
            </div>
          )}

          {data && data.items.length === 0 && (
            <div className="px-6 py-16 text-center">
              {filtered ? (
                <>
                  <p className="text-2xl">Ninguém encontrado com esses filtros.</p>
                  <Button
                    variant="outline"
                    size="md"
                    className="mt-6"
                    onClick={() => {
                      setSearch("");
                      update({ q: "", type: "" });
                    }}
                  >
                    Limpar filtros
                  </Button>
                </>
              ) : (
                <>
                  <p className="text-2xl">Ainda não há inscrições.</p>
                  <p className="mt-2 text-lg text-brand/80">Envie o link do site para os convidados; as respostas aparecem aqui.</p>
                </>
              )}
            </div>
          )}

          {data && data.items.length > 0 && (
            <GuestTable items={data.items} sort={query.sort} order={query.order} onSort={toggleSort} onOpen={setOpenId} />
          )}
        </div>

        {data && (
          <>
            <p className="mt-4 text-center text-lg text-white/90" aria-live="polite">
              {data.total === 1 ? "1 pessoa" : `${data.total} pessoas`}
              {filtered ? " neste filtro" : " no total"}
            </p>
            <Pagination page={data.page} pageSize={data.page_size} total={data.total} onPage={(page) => update({ page })} />
          </>
        )}
      </div>

      <GuestDetailDialog id={openId} onClose={() => setOpenId(null)} />
    </AdminLayout>
  );
}
