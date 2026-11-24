import { useNavigate } from "react-router-dom";
import { PublicPage } from "@/layouts/PublicLayout";
import { Column, Section } from "@/components/Section";
import { Button } from "@/components/ui/button";
import { useRsvpFlow } from "@/contexts/RsvpFlowContext";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

// Frame 8: agradecimento pela mensagem.
export default function MessageSentPage() {
  useDocumentTitle("Mensagem enviada");
  const navigate = useNavigate();
  const { reset } = useRsvpFlow();
  return (
    <PublicPage>
      <Section className="flex-1 py-[clamp(2.5rem,2.4vw,3rem)] pb-[clamp(3rem,5vw,6rem)]">
        <Column>
          <h1 className="text-title text-balance">Muito obrigada por esse carinho! ♡</h1>
          <p className="mt-[1em] text-lead text-pretty">
            Sua mensagem deixou nosso dia ainda mais especial. É muito bom saber que temos pessoas queridas torcendo por nós e fazendo parte dessa nova fase, mesmo que de longe.
          </p>
          <p className="text-lead text-pretty">
            Vamos guardar suas palavras com muito carinho, assim como tantas outras lembranças que queremos construir nesse novo lar.
          </p>
          <div className="mt-[clamp(2rem,4vw,5rem)]">
            <Button
              className="w-full max-w-[45.9rem]"
              onClick={() => {
                reset();
                navigate("/");
              }}
            >
              Retornar para página inicial
            </Button>
          </div>
        </Column>
      </Section>
    </PublicPage>
  );
}
