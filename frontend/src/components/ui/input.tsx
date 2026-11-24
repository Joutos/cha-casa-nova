import * as React from "react";
import { cn } from "@/lib/utils";

export const fieldBase =
  "w-full rounded-pill border-0 bg-mist font-sans text-brand-deep placeholder:text-brand/75 outline-none transition-shadow focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-white aria-[invalid=true]:ring-[3px] aria-[invalid=true]:ring-sun disabled:opacity-60";

const Input = React.forwardRef<HTMLInputElement, React.InputHTMLAttributes<HTMLInputElement>>(
  ({ className, type = "text", ...props }, ref) => (
    <input
      ref={ref}
      type={type}
      className={cn(fieldBase, "h-[clamp(3.5rem,4.4vw,5.25rem)] px-[clamp(1.25rem,1.6vw,2rem)] text-action", className)}
      {...props}
    />
  ),
);
Input.displayName = "Input";

export { Input };
