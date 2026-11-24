import * as React from "react";
import * as CheckboxPrimitive from "@radix-ui/react-checkbox";
import { cn } from "@/lib/utils";

// Quadrado de 33px com contorno branco (vetor 3:128 do Figma); marcado fica preenchido.
const Checkbox = React.forwardRef<
  React.ElementRef<typeof CheckboxPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof CheckboxPrimitive.Root> & { tone?: "light" | "brand" }
>(({ className, tone = "light", ...props }, ref) => (
  <CheckboxPrimitive.Root
    ref={ref}
    className={cn(
      "peer grid size-[clamp(1.6rem,1.72vw,2.06rem)] shrink-0 place-items-center rounded-[5px] border-2 transition-colors duration-150 disabled:opacity-50",
      tone === "light"
        ? "border-white data-[state=checked]:bg-white data-[state=checked]:text-brand"
        : "border-brand data-[state=checked]:bg-brand data-[state=checked]:text-white",
      className,
    )}
    {...props}
  >
    <CheckboxPrimitive.Indicator forceMount className="data-[state=unchecked]:opacity-0 transition-opacity duration-150">
      <svg viewBox="0 0 24 24" className="size-[70%]" fill="none" stroke="currentColor" strokeWidth="3.2" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
        <path d="M5 12.5l4.5 4.5L19 7.5" />
      </svg>
    </CheckboxPrimitive.Indicator>
  </CheckboxPrimitive.Root>
));
Checkbox.displayName = CheckboxPrimitive.Root.displayName;

export { Checkbox };
