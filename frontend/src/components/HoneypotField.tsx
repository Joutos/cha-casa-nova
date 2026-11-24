import type { UseFormRegisterReturn } from "react-hook-form";

// Campo-armadilha para robôs: invisível para pessoas e ignorado por leitores de tela.
export function HoneypotField({ registration }: { registration: UseFormRegisterReturn }) {
  return (
    <div aria-hidden className="absolute -left-[10000px] top-auto h-px w-px overflow-hidden">
      <label>
        Deixe este campo em branco
        <input type="text" tabIndex={-1} autoComplete="off" {...registration} />
      </label>
    </div>
  );
}
