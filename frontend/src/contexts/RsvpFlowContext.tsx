import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";

// O convidado escolhe uma ou mais opções na página inicial e percorre as etapas
// nesta ordem (mesma ordem das opções no Figma).
export const STEPS = ["attend", "contribute", "message"] as const;
export type Step = (typeof STEPS)[number];

export const STEP_PATH: Record<Step, string> = {
  attend: "/presenca",
  contribute: "/contribuir",
  message: "/mensagem",
};

export interface FlowState {
  choices: Step[];
  completed: Step[];
  guest?: { name: string; email: string };
}

const EMPTY: FlowState = { choices: [], completed: [] };
const STORAGE_KEY = "ccn:flow";

/** Próxima etapa escolhida e ainda não concluída, desconsiderando `current`. */
export function nextStep(state: FlowState, current?: Step): Step | null {
  return STEPS.find((s) => s !== current && state.choices.includes(s) && !state.completed.includes(s)) ?? null;
}

function load(): FlowState {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    if (!raw) return EMPTY;
    const parsed = JSON.parse(raw) as FlowState;
    const valid = (xs: unknown): Step[] => (Array.isArray(xs) ? xs.filter((x): x is Step => STEPS.includes(x)) : []);
    return { choices: valid(parsed.choices), completed: valid(parsed.completed), guest: parsed.guest };
  } catch {
    return EMPTY;
  }
}

interface FlowContextValue {
  state: FlowState;
  start: (choices: Step[]) => void;
  complete: (step: Step) => void;
  rememberGuest: (guest: { name: string; email: string }) => void;
  reset: () => void;
  /** Caminho da próxima etapa depois de `current`, ou null se acabou. */
  nextPath: (current: Step) => string | null;
  wantsContribution: boolean;
}

const FlowContext = createContext<FlowContextValue | null>(null);

export function RsvpFlowProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<FlowState>(load);

  useEffect(() => {
    try {
      // sessionStorage: some ao fechar a aba; guarda só nome/e-mail para preencher o próximo formulário.
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    } catch {
      /* navegação privada sem storage: o fluxo segue em memória */
    }
  }, [state]);

  const start = useCallback((choices: Step[]) => setState((s) => ({ ...s, choices, completed: [] })), []);
  const complete = useCallback(
    (step: Step) => setState((s) => (s.completed.includes(step) ? s : { ...s, completed: [...s.completed, step] })),
    [],
  );
  const rememberGuest = useCallback((guest: { name: string; email: string }) => setState((s) => ({ ...s, guest })), []);
  const reset = useCallback(() => setState(EMPTY), []);
  const nextPath = useCallback(
    (current: Step) => {
      const n = nextStep(state, current);
      return n ? STEP_PATH[n] : null;
    },
    [state],
  );

  const value = useMemo(
    () => ({ state, start, complete, rememberGuest, reset, nextPath, wantsContribution: state.choices.includes("contribute") }),
    [state, start, complete, rememberGuest, reset, nextPath],
  );
  return <FlowContext.Provider value={value}>{children}</FlowContext.Provider>;
}

export function useRsvpFlow() {
  const ctx = useContext(FlowContext);
  if (!ctx) throw new Error("useRsvpFlow precisa estar dentro de RsvpFlowProvider");
  return ctx;
}
