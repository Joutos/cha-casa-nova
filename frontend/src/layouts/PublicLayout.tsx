import type { ReactNode } from "react";
import { HeroHeader } from "@/components/HeroHeader";
import { SiteFooter } from "@/components/SiteFooter";

interface PublicPageProps {
  hero?: "full" | "compact";
  children: ReactNode;
  tallFooter?: boolean;
}

/** Estrutura comum das telas públicas: foto do casal, conteúdo e faixa cinza de rodapé. */
export function PublicPage({ hero = "compact", children, tallFooter }: PublicPageProps) {
  return (
    <div className="flex min-h-dvh flex-col">
      <a href="#conteudo" className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:rounded-pill focus:bg-white focus:px-5 focus:py-2 focus:text-brand">
        Pular para o conteúdo
      </a>
      <HeroHeader variant={hero} titleAs={hero === "full" ? "h1" : "p"} />
      <main id="conteudo" className="flex flex-1 flex-col" tabIndex={-1}>
        {children}
      </main>
      <SiteFooter tall={tallFooter} />
    </div>
  );
}
