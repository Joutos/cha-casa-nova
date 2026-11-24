import { cn } from "@/lib/utils";

interface Props {
  attending: boolean;
  contribute: boolean;
  messages: number;
  className?: string;
}

// O que a pessoa escolheu na página inicial, em etiquetas no estilo pílula.
export function ChoiceBadges({ attending, contribute, messages, className }: Props) {
  const items = [
    attending && { key: "attend", label: "Presença" },
    contribute && { key: "contribute", label: "Contribuição" },
    messages > 0 && { key: "message", label: messages > 1 ? `${messages} mensagens` : "Mensagem" },
  ].filter(Boolean) as { key: string; label: string }[];

  if (items.length === 0) return <span className="text-brand/60">Nenhuma escolha</span>;
  return (
    <ul className={cn("flex flex-wrap gap-2", className)} aria-label="Escolhas">
      {items.map((i) => (
        <li
          key={i.key}
          className={cn(
            "rounded-full px-3 py-0.5 text-base leading-7",
            i.key === "attend" ? "bg-brand text-white" : i.key === "contribute" ? "bg-sun text-[#4a3500]" : "border border-brand text-brand",
          )}
        >
          {i.label}
        </li>
      ))}
    </ul>
  );
}
