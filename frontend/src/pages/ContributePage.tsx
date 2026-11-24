import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { PublicPage } from "@/layouts/PublicLayout";
import { Column, Section } from "@/components/Section";
import { Button } from "@/components/ui/button";
import { EVENT } from "@/config/event";
import { useRsvpFlow } from "@/contexts/RsvpFlowContext";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

// Frame 5: contribuição via Pix. O QR Code ainda não existe; quando existir,
// basta definir VITE_PIX_QR_IMAGE (e VITE_PIX_KEY para o botão de copiar).
export default function ContributePage() {
  useDocumentTitle("Contribuir com o nosso lar");
  const navigate = useNavigate();
  const flow = useRsvpFlow();
  const next = flow.nextPath("contribute");
  const [copied, setCopied] = useState(false);
  const { complete } = flow;

  useEffect(() => complete("contribute"), [complete]);

  const copyKey = async () => {
    if (!EVENT.pixKey) return;
    try {
      await navigator.clipboard.writeText(EVENT.pixKey);
      setCopied(true);
      toast.success("Chave Pix copiada.");
    } catch {
      toast.error("Não foi possível copiar. Selecione a chave e copie manualmente.");
    }
  };

  const onFinish = () => {
    if (next) return navigate(next);
    flow.reset();
    navigate("/");
  };

  return (
    <PublicPage>
      <Section className="py-[clamp(3rem,4.9vw,6rem)]">
        <Column className="max-w-[64rem]">
          <h1 className="text-title text-balance">Que carinho ter você fazendo parte desse novo começo! ♡</h1>
          <p className="mx-auto mt-[0.9em] max-w-[54rem] text-lead text-pretty">
            Estamos construindo nosso cantinho com muito amor, e ficaremos muito felizes com sua contribuição para essa nova fase da nossa vida.
          </p>
          <p className="mx-auto mt-[0.9em] max-w-[54rem] text-lead text-pretty">
            Se desejar nos presentear, você poderá contribuir com qualquer valor via Pix. Cada gesto de carinho nos ajudará a transformar nossa casa em um lar ainda mais especial.
          </p>
        </Column>
      </Section>

      <Section tone="gray" className="py-[clamp(1.5rem,2vw,2.4rem)]">
        <div className="mx-auto grid aspect-square w-[clamp(14rem,18.3vw,22rem)] place-items-center overflow-hidden border border-cocoa/60 bg-mist">
          {EVENT.pixQrImage ? (
            <img src={EVENT.pixQrImage} alt="QR Code do Pix para contribuir com o Chá de Casa Nova" className="size-full object-contain p-3" />
          ) : (
            <p className="px-6 text-center text-xl text-cocoa">O QR Code do Pix aparece aqui em breve.</p>
          )}
        </div>
        {EVENT.pixKey && (
          <div className="mt-5 flex flex-col items-center gap-3 text-center">
            <p className="text-lg">
              Chave Pix: <span className="select-all break-all font-medium text-brand-deep">{EVENT.pixKey}</span>
            </p>
            <Button variant="solid" size="md" onClick={copyKey}>
              {copied ? "Chave copiada" : "Copiar chave Pix"}
            </Button>
          </div>
        )}
      </Section>

      <Section className="flex-1 py-[clamp(2rem,2.9vw,3.5rem)]">
        <Column>
          <p className="mx-auto max-w-[32rem] text-lead text-balance">A gente agradece por fazer parte dessa história!</p>
          <div className="mt-[clamp(1.5rem,2.3vw,2.8rem)]">
            <Button onClick={onFinish} className="w-full max-w-[45.9rem]">
              {next ? "continuar" : "Retornar para página inicial"}
            </Button>
          </div>
        </Column>
      </Section>
    </PublicPage>
  );
}
