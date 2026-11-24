import { Link } from "react-router-dom";

// Aviso de finalidade da coleta (LGPD), exibido junto aos formulários.
export function PrivacyNotice({ purpose }: { purpose: string }) {
  return (
    <p className="mx-auto mt-6 max-w-[38rem] text-note text-white/90 text-pretty">
      Usamos seu nome e e-mail apenas para {purpose}. Nada é compartilhado com terceiros.{" "}
      <Link to="/privacidade" className="underline underline-offset-4 hover:text-white">
        Saiba mais
      </Link>
      .
    </p>
  );
}
