import { cloneElement, isValidElement, useId, type ReactElement } from "react";
import { cn } from "@/lib/utils";

interface FormFieldProps {
  label: string;
  error?: string;
  hint?: string;
  /** No Figma o rótulo aparece como placeholder; o <label> fica só para leitores de tela. */
  visuallyHiddenLabel?: boolean;
  className?: string;
  children: ReactElement;
}

export function FormField({ label, error, hint, visuallyHiddenLabel = true, className, children }: FormFieldProps) {
  const id = useId();
  const errorId = `${id}-error`;
  const hintId = `${id}-hint`;
  const describedBy = [error && errorId, hint && hintId].filter(Boolean).join(" ") || undefined;

  const control = isValidElement(children)
    ? cloneElement(children as ReactElement<Record<string, unknown>>, {
        id,
        "aria-invalid": error ? true : undefined,
        "aria-describedby": describedBy,
      })
    : children;

  return (
    <div className={cn("text-left", className)}>
      <label htmlFor={id} className={visuallyHiddenLabel ? "sr-only" : "mb-2 block text-xl"}>
        {label}
      </label>
      {control}
      {hint && !error && (
        <p id={hintId} className="mt-2 px-3 text-note opacity-90">
          {hint}
        </p>
      )}
      {error && (
        <p id={errorId} role="alert" className="mt-2 flex items-start gap-2 px-3 text-lg leading-snug text-sun-soft">
          <span aria-hidden className="mt-[0.35em] inline-block size-2.5 shrink-0 rounded-full bg-sun" />
          {error}
        </p>
      )}
    </div>
  );
}
