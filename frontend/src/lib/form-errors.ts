import type { FieldValues, Path, UseFormSetError } from "react-hook-form";
import { ApiError } from "@/services/api";

/**
 * Aplica os erros de campo vindos da API ao formulário e devolve a mensagem
 * geral que deve aparecer acima do botão (ou null se tudo foi atribuído a campos).
 */
export function applyServerErrors<T extends FieldValues>(
  err: unknown,
  setError: UseFormSetError<T>,
  known: readonly Path<T>[],
): string | null {
  if (!(err instanceof ApiError)) return "Algo saiu do esperado. Tente novamente.";
  if (err.fields) {
    let assigned = 0;
    for (const [field, message] of Object.entries(err.fields)) {
      if ((known as readonly string[]).includes(field)) {
        setError(field as Path<T>, { type: "server", message }, { shouldFocus: assigned === 0 });
        assigned++;
      }
    }
    if (assigned === Object.keys(err.fields).length) return null;
  }
  return err.message;
}
