import { cn } from "@/lib/utils";
import type { GuestFilter, Stats } from "@/types/api";

const TABS: { value: GuestFilter; label: string; stat: keyof Stats }[] = [
  { value: "", label: "Todos", stat: "total" },
  { value: "attending", label: "Presença", stat: "attending" },
  { value: "contribute", label: "Contribuição", stat: "contribute" },
  { value: "message", label: "Mensagem", stat: "with_message" },
];

// Filtro por tipo de escolha; cada aba já mostra quantas pessoas há nela.
export function FilterTabs({ value, stats, onChange }: { value: GuestFilter; stats?: Stats; onChange: (v: GuestFilter) => void }) {
  return (
    <div role="radiogroup" aria-label="Filtrar por escolha" className="flex flex-wrap gap-2 sm:gap-3">
      {TABS.map((t) => {
        const active = t.value === value;
        return (
          <button
            key={t.value || "all"}
            type="button"
            role="radio"
            aria-checked={active}
            onClick={() => onChange(t.value)}
            className={cn(
              "flex items-baseline gap-2 rounded-pill px-5 py-2 text-xl transition-colors",
              active ? "bg-white text-brand" : "bg-white/10 text-white hover:bg-white/20",
            )}
          >
            {t.label}
            <span className={cn("text-2xl tabular-nums", active ? "text-brand-deep" : "text-white")} aria-label={stats ? `${stats[t.stat]} pessoas` : undefined}>
              {stats ? stats[t.stat] : "–"}
            </span>
          </button>
        );
      })}
    </div>
  );
}
