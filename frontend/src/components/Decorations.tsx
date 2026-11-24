import { cn } from "@/lib/utils";

// Ilustrações do Figma (flor 3:42 e ramo 3:25). Puramente decorativas.
export function Flower({ className }: { className?: string }) {
  return <img src="/images/flower.svg" alt="" aria-hidden className={cn("pointer-events-none select-none", className)} />;
}

export function Branch({ className }: { className?: string }) {
  return <img src="/images/branch.svg" alt="" aria-hidden className={cn("pointer-events-none select-none", className)} />;
}

export function HouseIcon({ className }: { className?: string }) {
  return <img src="/images/house.svg" alt="" aria-hidden className={className} />;
}

export function PawIcon({ className }: { className?: string }) {
  return <img src="/images/paw.svg" alt="" aria-hidden className={className} />;
}
