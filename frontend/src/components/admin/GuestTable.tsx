import type { KeyboardEvent } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ChoiceBadges } from "./ChoiceBadges";
import { cn, formatDateTime } from "@/lib/utils";
import type { GuestSummary, SortKey, SortOrder } from "@/types/api";

interface Props {
  items: GuestSummary[];
  sort: SortKey;
  order: SortOrder;
  onSort: (k: SortKey) => void;
  onOpen: (id: string) => void;
}

function SortHeader({ label, k, sort, order, onSort, className }: { label: string; k: SortKey; className?: string } & Pick<Props, "sort" | "order" | "onSort">) {
  const active = sort === k;
  return (
    <TableHead aria-sort={active ? (order === "asc" ? "ascending" : "descending") : "none"} className={className}>
      <button type="button" onClick={() => onSort(k)} className="inline-flex items-center gap-1.5 rounded-md hover:text-brand">
        {label}
        <svg viewBox="0 0 12 12" className={cn("size-3 transition-transform", active ? "opacity-100" : "opacity-30", active && order === "asc" && "rotate-180")} aria-hidden>
          <path d="M2 4l4 4 4-4" stroke="currentColor" strokeWidth="1.6" fill="none" strokeLinecap="round" />
        </svg>
      </button>
    </TableHead>
  );
}

export function GuestTable({ items, sort, order, onSort, onOpen }: Props) {
  const onKey = (e: KeyboardEvent, id: string) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      onOpen(id);
    }
  };

  return (
    <>
      {/* Tabela a partir de 768px */}
      <Table className="hidden md:table">
        <caption className="sr-only">Pessoas inscritas. Selecione uma linha para ver os detalhes.</caption>
        <TableHeader>
          <TableRow>
            <SortHeader label="Nome" k="name" sort={sort} order={order} onSort={onSort} className="pl-8" />
            <SortHeader label="E-mail" k="email" sort={sort} order={order} onSort={onSort} />
            <TableHead>Escolhas</TableHead>
            <SortHeader label="Inscrição" k="created_at" sort={sort} order={order} onSort={onSort} className="pr-8 text-right" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((g) => (
            <TableRow
              key={g.id}
              tabIndex={0}
              onClick={() => onOpen(g.id)}
              onKeyDown={(e) => onKey(e, g.id)}
              className="cursor-pointer hover:bg-white/60 focus-visible:bg-white/60"
              aria-label={`Ver detalhes de ${g.name}`}
            >
              <TableCell className="pl-8 text-xl">{g.name}</TableCell>
              <TableCell className="break-all text-brand/85">{g.email}</TableCell>
              <TableCell>
                <ChoiceBadges attending={g.attending} contribute={g.wants_to_contribute} messages={g.message_count} />
              </TableCell>
              <TableCell className="whitespace-nowrap pr-8 text-right tabular-nums text-brand/85">{formatDateTime(g.created_at)}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>

      {/* Cartões no celular */}
      <ul className="divide-y divide-brand/15 md:hidden">
        {items.map((g) => (
          <li key={g.id}>
            <button type="button" onClick={() => onOpen(g.id)} className="w-full px-5 py-4 text-left hover:bg-white/60">
              <span className="block text-xl">{g.name}</span>
              <span className="block break-all text-base text-brand/80">{g.email}</span>
              <ChoiceBadges className="mt-2" attending={g.attending} contribute={g.wants_to_contribute} messages={g.message_count} />
              <span className="mt-2 block text-sm tabular-nums text-brand/70">Inscrição em {formatDateTime(g.created_at)}</span>
            </button>
          </li>
        ))}
      </ul>
    </>
  );
}
