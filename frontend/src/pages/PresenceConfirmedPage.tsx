import { useNavigate } from "react-router-dom";
import { PublicPage } from "@/layouts/PublicLayout";
import { Column, Section } from "@/components/Section";
import { Button } from "@/components/ui/button";
import { useRsvpFlow } from "@/contexts/RsvpFlowContext";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

// Frame 4 (há próximas etapas: "continuar") e Frame 3 (fim: "Retornar para página inicial").
export default function PresenceConfirmedPage() {
  useDocumentTitle("Presença confirmada");
  const navigate = useNavigate();
  const flow = useRsvpFlow();
  const next = flow.nextPath("attend");

  const finish = () => {
    flow.reset();
    navigate("/");
  };

  return (
    <PublicPage>
      <Section className="flex-1 py-[clamp(3rem,6vw,7rem)]">
        <Column>
          <h1 className="text-title text-balance">Que alegria ter você com a gente! ♡</h1>
          <p className="mt-[1.1em] text-lead text-pretty">
            Um e-mail será enviado para você com o convite do nosso Chá de Casa Nova e todas as informações necessárias para esse dia especial.
          </p>
          <p className="mt-[1.3em] text-lead text-pretty">Fique de olho na sua caixa de entrada! Estamos felizes em compartilhar esse momento com você. ♡</p>
          <div className="mt-[clamp(2rem,3vw,3.5rem)]">
            {next ? (
              <Button onClick={() => navigate(next)} className="min-w-[min(17.8rem,100%)]">
                continuar
              </Button>
            ) : (
              <Button onClick={finish} className="w-full max-w-[45.9rem]">
                Retornar para página inicial
              </Button>
            )}
          </div>
        </Column>
      </Section>
    </PublicPage>
  );
}
