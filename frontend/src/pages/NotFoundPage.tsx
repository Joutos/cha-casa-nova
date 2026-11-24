import { Link } from "react-router-dom";
import { PublicPage } from "@/layouts/PublicLayout";
import { Column, Section } from "@/components/Section";
import { buttonVariants } from "@/components/ui/button";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

export default function NotFoundPage() {
  useDocumentTitle("Página não encontrada");
  return (
    <PublicPage>
      <Section className="flex-1 py-[clamp(3rem,6vw,7rem)]">
        <Column>
          <h1 className="text-title">Esta página não existe</h1>
          <p className="mt-4 text-lead">O endereço pode ter sido digitado errado. A festa continua na página inicial.</p>
          <Link to="/" className={buttonVariants({ className: "mt-10" })}>
            Ir para a página inicial
          </Link>
        </Column>
      </Section>
    </PublicPage>
  );
}
