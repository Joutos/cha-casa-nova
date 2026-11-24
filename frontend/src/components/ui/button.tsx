import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

// Base shadcn/ui, com as variantes do Figma: pílula cinza (#D9D9D9, raio 25px, texto azul).
const buttonVariants = cva(
  "inline-flex select-none items-center justify-center gap-3 whitespace-nowrap rounded-pill font-sans transition-[background-color,color,transform,box-shadow] duration-150 active:translate-y-px disabled:pointer-events-none disabled:opacity-60 [&_svg]:shrink-0",
  {
    variants: {
      variant: {
        pill: "bg-mist text-brand hover:bg-white hover:shadow-[0_6px_0_rgb(0_0_0/0.12)]",
        solid: "bg-brand text-white hover:bg-brand-deep",
        outline: "border-2 border-brand bg-transparent text-brand hover:bg-brand hover:text-white",
        ghost: "bg-transparent text-white underline-offset-4 hover:underline",
        danger: "bg-danger text-white hover:bg-[#8a2f2f]",
      },
      size: {
        hero: "h-[clamp(3.5rem,4.4vw,5.25rem)] px-[clamp(2rem,4vw,5rem)] text-action", // 84px no Figma
        lg: "h-[clamp(3.25rem,4.2vw,5rem)] px-[clamp(1.75rem,3.5vw,4.5rem)] text-action", // 81px
        md: "h-12 px-6 text-xl",
        sm: "h-10 px-4 text-lg",
        icon: "h-11 w-11 text-xl",
      },
    },
    defaultVariants: { variant: "pill", size: "lg" },
  },
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";
    return <Comp className={cn(buttonVariants({ variant, size, className }))} ref={ref} {...props} />;
  },
);
Button.displayName = "Button";

export { Button, buttonVariants };
