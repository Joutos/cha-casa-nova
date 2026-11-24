import type { HTMLAttributes } from "react";
import { cn } from "@/lib/utils";

// Faixas do Figma: azul (#4365BF, texto branco) e cinza (#D9D9D9, texto marrom).
export function Section({ tone = "blue", className, ...props }: HTMLAttributes<HTMLElement> & { tone?: "blue" | "gray" }) {
  return (
    <section
      className={cn(
        "relative px-5 sm:px-8",
        tone === "blue" ? "bg-brand text-white" : "on-light bg-mist text-cocoa",
        className,
      )}
      {...props}
    />
  );
}

export function Column({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("mx-auto w-full max-w-column text-center", className)} {...props} />;
}
