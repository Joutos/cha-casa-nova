import { Button } from "@/components/ui/button";

export function Pagination({ page, pageSize, total, onPage }: { page: number; pageSize: number; total: number; onPage: (p: number) => void }) {
  const pages = Math.max(1, Math.ceil(total / pageSize));
  if (total <= pageSize) return null;
  return (
    <nav aria-label="Paginação" className="mt-6 flex flex-wrap items-center justify-center gap-4 text-white">
      <Button size="md" disabled={page <= 1} onClick={() => onPage(page - 1)}>
        Anterior
      </Button>
      <p className="text-xl tabular-nums" aria-live="polite">
        Página {Math.min(page, pages)} de {pages}
      </p>
      <Button size="md" disabled={page >= pages} onClick={() => onPage(page + 1)}>
        Próxima
      </Button>
    </nav>
  );
}
