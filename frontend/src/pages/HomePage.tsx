import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { PublicPage } from "@/layouts/PublicLayout";
import { Column, Section } from "@/components/Section";
import { Branch, Flower } from "@/components/Decorations";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { EVENT } from "@/config/event";
import { STEP_PATH, STEPS, useRsvpFlow, type Step } from "@/contexts/RsvpFlowContext";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

const OPTIONS: { step: Step; title: string; description: string }[] = [
  { step: "attend", title: "Quero estar presente!", description: "Confirme sua presença para celebrar esse momento especial com a gente." },
  { step: "contribute", title: "Quero contribuir com o nosso lar!", description: "Ajude a construir nosso cantinho com uma contribuição para essa nova fase." },
  { step: "message", title: "Quero deixar uma mensagem!", description: "Mesmo de longe, seu carinho faz parte da nossa história. Deixe uma mensagem para nós! ♡" },
];

export default function HomePage() {
  useDocumentTitle("");
  const navigate = useNavigate();
  const { start } = useRsvpFlow();
  const [selected, setSelected] = useState<Step[]>([]);
  const [error, setError] = useState<string | null>(null);

  const toggle = (step: Step, on: boolean) => {
    setError(null);
    setSelected((cur) => (on ? STEPS.filter((s) => s === step || cur.includes(s)) : cur.filter((s) => s !== step)));
  };

  const onContinue = () => {
    if (selected.length === 0) {
      setError("Escolha pelo menos uma opção para continuar.");
      return;
    }
    start(selected);
    navigate(STEP_PATH[selected[0]]);
  };

  return (
    <PublicPage hero="full" tallFooter>
      {/* Faixa cinza: boas-vindas, data e ilustrações (Frame 1, 3:23) */}
      <Section tone="gray" className="overflow-hidden py-[clamp(3.5rem,4.3vw,5.2rem)]">
        <Flower className="absolute left-[2%] top-3 w-[clamp(4.5rem,13.3vw,16rem)] sm:left-[8.75%] sm:top-[clamp(0.75rem,1.7vw,2rem)]" />
        <Branch className="absolute bottom-2 right-[1%] w-[clamp(5rem,23.4vw,28rem)] sm:bottom-[1.2vw] sm:right-[6.5%]" />
        <Column className="relative max-w-[50rem] px-12 sm:px-0">
          <p className="text-lead text-pretty">
            Estamos preparando nosso cantinho com muito amor e queremos dividir esse momento tão especial com pessoas queridas.
          </p>
          <p className="mt-[1.3em] text-lead text-balance">Sua presença vai deixar nossa casinha ainda mais cheia de carinho!</p>
          <p className="mt-[1.9em] text-lead">Data: {EVENT.dateLabel}</p>
        </Column>
      </Section>

      {/* Faixa azul: escolhas (Frame 1, 3:55 em diante) */}
      <Section className="flex-1 pb-[clamp(4rem,6vw,7rem)] pt-[clamp(3rem,2.85vw,3.5rem)]">
        <Column>
          <h2 className="text-title">Podemos contar com você? ♡</h2>
          <p className="mx-auto mt-[0.8em] max-w-[42rem] text-lead text-balance">
            De pertinho ou de longe, vamos amar compartilhar esse momento com você!
          </p>
        </Column>

        <fieldset className="mx-auto mt-[clamp(2rem,3.4vw,4rem)] w-full max-w-[54rem]" aria-describedby={error ? "choice-error" : undefined}>
          <legend className="sr-only">Como você quer participar? Escolha uma ou mais opções.</legend>
          <ul className="space-y-[clamp(0.9rem,1.6vw,1.9rem)]">
            {OPTIONS.map((o) => {
              const id = `option-${o.step}`;
              return (
                <li key={o.step} className="flex items-start gap-[clamp(0.75rem,0.95vw,1.15rem)]">
                  <Checkbox
                    id={id}
                    checked={selected.includes(o.step)}
                    onCheckedChange={(v) => toggle(o.step, v === true)}
                    aria-describedby={`${id}-desc`}
                    className="mt-[0.15em]"
                  />
                  <label htmlFor={id} className="cursor-pointer select-none text-left">
                    <span className="block text-lead">{o.title}</span>
                    <span id={`${id}-desc`} className="mt-1 block text-note">
                      {o.description}
                    </span>
                  </label>
                </li>
              );
            })}
          </ul>

          {error && (
            <p id="choice-error" role="alert" className="mt-6 flex items-center gap-2 text-lg text-sun-soft">
              <span aria-hidden className="inline-block size-2.5 rounded-full bg-sun" />
              {error}
            </p>
          )}

          <div className="mt-[clamp(2.5rem,3vw,3.6rem)] flex justify-center sm:justify-start sm:pl-[clamp(2.35rem,2.67vw,3.2rem)]">
            <Button size="hero" className="w-full max-w-[37.7rem]" onClick={onContinue}>
              Continuar
            </Button>
          </div>
        </fieldset>
      </Section>
    </PublicPage>
  );
}
