import { cn } from "@/lib/utils";

// Faixa cinza de rodapé presente em todas as telas do Figma. Usamos o espaço vazio
// para o link de privacidade exigido pela LGPD.
export function SiteFooter({ tall = false }: { tall?: boolean }) {
  return (
    <footer
      className={cn(
        "on-light flex items-end justify-center bg-mist px-5 pb-6 text-cocoa",
        tall ? "min-h-[clamp(6rem,8vw,9.4rem)]" : "min-h-[clamp(5rem,6.9vw,8.25rem)]",
      )}
    >
    </footer>
  );
}
