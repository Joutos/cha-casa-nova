import * as React from "react";
import { cn } from "@/lib/utils";
import { fieldBase } from "./input";

const Textarea = React.forwardRef<HTMLTextAreaElement, React.TextareaHTMLAttributes<HTMLTextAreaElement>>(
  ({ className, ...props }, ref) => (
    <textarea
      ref={ref}
      className={cn(
        fieldBase,
        "min-h-[clamp(8rem,7.6vw,9.1rem)] resize-y px-[clamp(1.25rem,1.6vw,2rem)] py-[clamp(0.9rem,1.5vw,1.8rem)] text-lead leading-snug",
        className,
      )}
      {...props}
    />
  ),
);
Textarea.displayName = "Textarea";

export { Textarea };
