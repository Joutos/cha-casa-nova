import { Link } from "react-router-dom";
import { PublicPage } from "@/layouts/PublicLayout";
import { Column, Section } from "@/components/Section";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

// Informação de privacidade (LGPD, art. 9º): o que coletamos, para quê e por quanto tempo.
export default function PrivacyPage() {
  useDocumentTitle("Privacidade");
  return (
    <PublicPage>
      <Section className="flex-1 py-[clamp(3rem,5vw,6rem)]">
        <Column className="text-left">
          <div className="mt-8 space-y-5 text-xl leading-relaxed sm:text-2xl">
            <p>
              Este site existe só para organizar o Chá de Casa Nova da Sofia e do João. Coletamos o mínimo necessário: seu <strong className="font-normal underline underline-offset-4">nome</strong>, seu{" "}
              <strong className="font-normal underline underline-offset-4">e-mail</strong>, as opções que você escolheu e, se quiser, a mensagem que nos escrever.
            </p>
            <p>
              Usamos esses dados para enviar o convite com endereço e horário, saber quantas pessoas virão e guardar as mensagens. O e-mail também evita que a mesma presença seja confirmada duas vezes.
            </p>
            <p>Apenas os anfitriões têm acesso à lista. Nada é vendido nem compartilhado com terceiros, e não usamos cookies de rastreamento.</p>
            <p>
              Os dados serão apagados até 90 dias depois da festa. Se quiser consultar, corrigir ou excluir suas informações antes disso, fale diretamente com a Sofia ou com o João.
            </p>
          </div>
          <p className="mt-10 text-center">
            <Link to="/" className="text-xl underline underline-offset-4">
              Voltar para a página inicial
            </Link>
          </p>
        </Column>
      </Section>
    </PublicPage>
  );
}
