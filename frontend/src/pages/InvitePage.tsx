import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { HouseIcon, PawIcon } from "@/components/Decorations";
import { Spinner } from "@/components/Spinner";
import { buttonVariants } from "@/components/ui/button";
import { fetchInvite } from "@/services/public";
import { ApiError } from "@/services/api";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { cn } from "@/lib/utils";

// Frame 6: o convite. Acessível só pelo link pessoal enviado por e-mail, porque
// mostra o endereço da casa.
export default function InvitePage() {
  useDocumentTitle("Seu convite");
  const { token = "" } = useParams();
  const invite = useQuery({
    queryKey: ["invite", token],
    queryFn: ({ signal }) => fetchInvite(token, signal),
    retry: (count, err) => !(err instanceof ApiError && err.status === 404) && count < 2,
  });

  return (
    <div className="grid min-h-dvh place-items-center bg-brand px-3 py-[clamp(1.5rem,7.7vw,5.75rem)]">
      <article className="on-light w-full max-w-[42.25rem] rounded-[clamp(2rem,4vw,3.125rem)] bg-cream px-[clamp(1.25rem,5.4vw,3.6rem)] pb-10 pt-[clamp(2rem,3vw,3rem)] text-brand shadow-card">
        {invite.isPending && (
          <p className="flex items-center justify-center gap-3 py-24 text-2xl" role="status">
            <Spinner /> Abrindo seu convite…
          </p>
        )}

        {invite.isError && (
          <div className="py-16 text-center">
            <HouseIcon className="mx-auto size-[clamp(4.5rem,9vw,7.4rem)]" />
            <h1 className="mt-6 text-4xl">Não encontramos este convite</h1>
            <p className="mt-4 text-xl text-pretty">
              {invite.error instanceof ApiError && invite.error.status === 404
                ? "Confira se o link do e-mail foi copiado inteiro. Se o problema continuar, fale com a gente."
                : invite.error.message}
            </p>
            <Link to="/" className={cn(buttonVariants({ variant: "solid", size: "md" }), "mt-8")}>
              Ir para a página inicial
            </Link>
          </div>
        )}

        {invite.data && (
          <>
            <HouseIcon className="mx-auto size-[clamp(4.5rem,9vw,7.4rem)]" />
            <h1 className="mt-4 text-center text-[clamp(1.9rem,3vw,3rem)] leading-tight text-balance">
              Que bom que você vem{invite.data.guest_name ? `, ${invite.data.guest_name.split(" ")[0]}` : ""}! ♡
            </h1>
            <div className="mx-auto mt-6 max-w-[31.25rem] space-y-[1em] text-[clamp(1.15rem,1.6vw,1.625rem)] leading-snug">
              <p>Estamos muito felizes em poder compartilhar esse momento com você!</p>
              <p>Nossa casinha está ganhando forma e, aos poucos, estamos construindo um cantinho para chamar de lar.</p>
              <p>Ter você pertinho da gente para celebrar esse novo começo torna tudo ainda mais especial.</p>
              <p>Preparamos tudo com muito carinho e mal podemos esperar para receber você!</p>
            </div>
            <div className="mt-8 text-center">
              <p className="text-[clamp(1.2rem,1.8vw,1.75rem)]">{invite.data.date_label}</p>
              <p className="text-[clamp(1.6rem,2.5vw,2.5rem)] leading-tight">{invite.data.time}</p>
              <address className="mt-1 text-[clamp(1.25rem,2vw,2rem)] not-italic leading-tight">
                {invite.data.address_line1}
                <br />
                {invite.data.address_line2}
              </address>
            </div>
            {invite.data.maps_query && (
              <div className="mt-6 overflow-hidden rounded-2xl bg-mist">
                <iframe
                  title={`Mapa: ${invite.data.address_line1}`}
                  src={`https://www.google.com/maps?q=${encodeURIComponent(invite.data.maps_query)}&output=embed`}
                  className="block aspect-[593/190] min-h-56 w-full border-0"
                  loading="lazy"
                  referrerPolicy="no-referrer-when-downgrade"
                  sandbox="allow-scripts allow-same-origin allow-popups"
                />
              </div>
            )}
            <div className="mt-6 flex items-end justify-end gap-2 text-right">
              <div>
                <p className="text-2xl">Com carinho,</p>
                <p className="text-[2rem] leading-tight">{invite.data.signature}</p>
                <p className="text-2xl">e Pepé</p>
              </div>
              <PawIcon className="size-16 -rotate-[34.55deg]" />
            </div>
          </>
        )}
      </article>
    </div>
  );
}
