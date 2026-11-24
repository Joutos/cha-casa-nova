import { useEffect, useState } from "react";
import { toast } from "sonner";
import { Dialog, DialogContent, DialogDescription, DialogTitle } from "@/components/ui/dialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Spinner } from "@/components/Spinner";
import { useDeleteGuest, useGuestDetail, useUpdateGuest } from "@/hooks/useAdminGuests";
import { formatDateTime } from "@/lib/utils";

interface Props {
  id: string | null;
  onClose: () => void;
}

export function GuestDetailDialog({ id, onClose }: Props) {
  const detail = useGuestDetail(id);
  const update = useUpdateGuest();
  const remove = useDeleteGuest();
  const [confirmOpen, setConfirmOpen] = useState(false);
  const g = detail.data;

  useEffect(() => setConfirmOpen(false), [id]);

  const toggle = (field: "attending" | "wants_to_contribute", value: boolean) => {
    if (!g) return;
    update.mutate({ id: g.id, patch: { [field]: value } });
  };

  const copyInvite = async () => {
    if (!g?.invite_url) return;
    try {
      await navigator.clipboard.writeText(g.invite_url);
      toast.success("Link do convite copiado.");
    } catch {
      toast.error("Não foi possível copiar o link.");
    }
  };

  return (
    <Dialog open={!!id} onOpenChange={(open) => !open && onClose()}>
      <DialogContent aria-describedby={undefined}>
        {detail.isPending && (
          <>
            <DialogTitle className="sr-only">Detalhes do convidado</DialogTitle>
            <p className="flex items-center gap-3 py-10 text-xl" role="status">
              <Spinner /> Carregando…
            </p>
          </>
        )}
        {detail.isError && (
          <>
            <DialogTitle>Não foi possível abrir</DialogTitle>
            <DialogDescription className="mt-3">{detail.error.message}</DialogDescription>
          </>
        )}
        {g && (
          <>
            <DialogTitle className="text-balance">{g.name}</DialogTitle>
            <DialogDescription className="mt-1 break-all text-brand/80">{g.email}</DialogDescription>
            <p className="mt-1 text-base text-brand/70">Inscrição em {formatDateTime(g.created_at)}</p>

            <fieldset className="mt-8 space-y-4" disabled={update.isPending}>
              <legend className="mb-3 text-2xl">Escolhas</legend>
              <label className="flex items-center gap-3 text-xl">
                <Checkbox tone="brand" checked={g.attending} onCheckedChange={(v) => toggle("attending", v === true)} />
                Vai estar presente
                {g.attending && g.attending_confirmed_at && (
                  <span className="text-base text-brand/70">(confirmou em {formatDateTime(g.attending_confirmed_at)})</span>
                )}
              </label>
              <label className="flex items-center gap-3 text-xl">
                <Checkbox tone="brand" checked={g.wants_to_contribute} onCheckedChange={(v) => toggle("wants_to_contribute", v === true)} />
                Quer contribuir via Pix
              </label>
            </fieldset>

            {g.invite_url && (
              <div className="mt-6 rounded-2xl bg-white/60 p-4">
                <p className="text-lg">Link pessoal do convite</p>
                <p className="mt-1 break-all text-sm text-brand/80">{g.invite_url}</p>
                <Button variant="outline" size="sm" className="mt-3" onClick={copyInvite}>
                  Copiar link
                </Button>
              </div>
            )}

            <section className="mt-8" aria-labelledby="msgs-title">
              <h3 id="msgs-title" className="text-2xl">
                Mensagens
              </h3>
              {g.messages.length === 0 ? (
                <p className="mt-2 text-lg text-brand/70">Esta pessoa não deixou mensagem.</p>
              ) : (
                <ul className="mt-3 space-y-3">
                  {g.messages.map((m) => (
                    <li key={m.id} className="rounded-2xl bg-white/70 p-4">
                      {/* Texto puro: o React escapa o conteúdo, sem risco de XSS. */}
                      <p className="whitespace-pre-wrap break-words text-lg leading-snug">{m.body}</p>
                      <p className="mt-2 text-sm text-brand/60">{formatDateTime(m.created_at)}</p>
                    </li>
                  ))}
                </ul>
              )}
            </section>

            <div className="mt-10 flex justify-end">
              <AlertDialog open={confirmOpen} onOpenChange={setConfirmOpen}>
                <AlertDialogTrigger asChild>
                  <Button variant="danger" size="md">
                    Excluir inscrição
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogTitle>Excluir a inscrição de {g.name}?</AlertDialogTitle>
                  <AlertDialogDescription>
                    A presença, a intenção de contribuir e {g.messages.length > 0 ? "as mensagens" : "os dados"} desta pessoa serão apagados. Não dá para desfazer.
                  </AlertDialogDescription>
                  <AlertDialogFooter>
                    <AlertDialogCancel>Manter</AlertDialogCancel>
                    <AlertDialogAction
                      disabled={remove.isPending}
                      onClick={(e) => {
                        e.preventDefault();
                        remove.mutate(g.id, {
                          onSuccess: () => {
                            setConfirmOpen(false);
                            onClose();
                          },
                        });
                      }}
                    >
                      {remove.isPending ? (
                        <>
                          <Spinner /> Excluindo…
                        </>
                      ) : (
                        "Excluir"
                      )}
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
