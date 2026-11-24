import { useState } from "react";
import { Link } from "react-router-dom";
import { cn } from "@/lib/utils";

interface HeroHeaderProps {
  /** "full": página inicial (791px no Figma); "compact": demais telas (468px). */
  variant?: "full" | "compact" | "slim";
  /** Na home o título do site é o h1; nas outras telas o h1 é o título da seção. */
  titleAs?: "h1" | "p";
}

export function HeroHeader({ variant = "compact", titleAs = "p" }: HeroHeaderProps) {
  const [photoOk, setPhotoOk] = useState(true);
  const Title = titleAs;
  const full = variant === "full";

  return (
    <header
      className={cn(
        "relative isolate flex flex-col items-center overflow-hidden bg-[#6b574d] px-4 text-center text-white",
        full
          ? "min-h-[clamp(24rem,41.2vw,49.45rem)] justify-start pt-[clamp(4.5rem,9.3vw,11.2rem)]"
          : variant === "slim"
            ? "min-h-[clamp(7rem,11vw,12rem)] justify-center"
            : "min-h-[clamp(12.5rem,24.4vw,29.25rem)] justify-center",
      )}
    >
      {photoOk && (
        <img
          src="/images/hero.jpg"
          alt=""
          aria-hidden
          decoding="async"
          {...{ fetchpriority: full ? "high" : "auto" }}
          onError={() => setPhotoOk(false)}
          className={cn("absolute inset-0 -z-20 size-full object-cover", full ? "object-[50%_28%]" : "object-[50%_32%]")}
        />
      )}
      <div className="absolute inset-0 -z-10 bg-cocoa-overlay" aria-hidden />

      {full && <p className="animate-rise text-hero-sub text-balance">Você está convidado para o nosso</p>}
      <Title className={cn("text-hero text-balance tracking-[0.01em]", variant === "slim" && "text-[clamp(2rem,4.2vw,4.75rem)]", full && "mt-[clamp(0.25rem,0.9vw,1.1rem)] animate-rise [animation-delay:120ms]")}>
        {full ? (
          "CHÁ DE CASA NOVA"
        ) : (
          <Link to="/" className="rounded-pill focus-visible:outline-offset-8" aria-label="Chá de Casa Nova, voltar para a página inicial">
            CHÁ DE CASA NOVA
          </Link>
        )}
      </Title>
    </header>
  );
}
