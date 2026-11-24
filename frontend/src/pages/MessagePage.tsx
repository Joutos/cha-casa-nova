import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { PublicPage } from "@/layouts/PublicLayout";
import { Column, Section } from "@/components/Section";
import { FormField } from "@/components/FormField";
import { HoneypotField } from "@/components/HoneypotField";
import { PrivacyNotice } from "@/components/PrivacyNotice";
import { Spinner } from "@/components/Spinner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { messageSchema, type MessageForm } from "@/schemas/guest";
import { sendMessage } from "@/services/public";
import { useRsvpFlow } from "@/contexts/RsvpFlowContext";
import { applyServerErrors } from "@/lib/form-errors";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

const MAX = 1000;

// Frame 7: mensagem. Nome e e-mail identificam quem escreveu; se a pessoa já os
// informou na etapa de presença, mostramos só um resumo editável.
export default function MessagePage() {
  useDocumentTitle("Deixar uma mensagem");
  const navigate = useNavigate();
  const flow = useRsvpFlow();
  const known = flow.state.guest;
  const [editingIdentity, setEditingIdentity] = useState(!known);
  const [formError, setFormError] = useState<string | null>(null);

  const form = useForm<MessageForm>({
    resolver: zodResolver(messageSchema),
    defaultValues: { name: known?.name ?? "", email: known?.email ?? "", message: "", website: "" },
  });
  const { errors } = form.formState;
  const length = form.watch("message")?.length ?? 0;
  const mutation = useMutation({ mutationFn: sendMessage });

  const onSubmit = form.handleSubmit(async (values) => {
    setFormError(null);
    const data = messageSchema.parse(values);
    try {
      await mutation.mutateAsync({ ...data, website: values.website ?? "", contribute: flow.wantsContribution });
      flow.rememberGuest({ name: data.name, email: data.email });
      flow.complete("message");
      navigate("/mensagem/enviada", { replace: true });
    } catch (err) {
      const msg = applyServerErrors(err, form.setError, ["name", "email", "message"] as const);
      if (form.getFieldState("name").error || form.getFieldState("email").error) setEditingIdentity(true);
      setFormError(msg);
    }
  });

  const busy = form.formState.isSubmitting || mutation.isPending;
  const showIdentity = editingIdentity || !!errors.name || !!errors.email;

  return (
    <PublicPage>
      <Section className="flex-1 py-[clamp(3rem,6.6vw,8rem)]">
        <Column className="max-w-[58.3rem]">
          <h1 className="text-title text-balance">Seu carinho também faz parte dessa história! ♡</h1>
          <div className="mx-auto max-w-[54rem] space-y-[1em] pt-[1em] text-lead text-pretty">
            <p>Sabemos que nem sempre é possível estar pertinho, mas queremos que você saiba que lembramos de você com muito carinho.</p>
            <p>
              Deixe uma mensagem, um desejo ou algumas palavrinhas para acompanhar essa nova fase da nossa vida. Vamos adorar guardar esse pedacinho de carinho junto com as lembranças do nosso novo lar!
            </p>
            <p>Com carinho, esperamos que você faça parte desse novo capítulo conosco.</p>
          </div>

          <form onSubmit={onSubmit} noValidate className="relative mt-[clamp(2rem,3.5vw,4.2rem)] space-y-5" aria-busy={busy}>
            <HoneypotField registration={form.register("website")} />

            {showIdentity ? (
              <div className="grid gap-5 sm:grid-cols-2">
                <FormField label="Nome" error={errors.name?.message}>
                  <Input placeholder="nome" autoComplete="name" maxLength={100} disabled={busy} {...form.register("name")} />
                </FormField>
                <FormField label="E-mail" error={errors.email?.message}>
                  <Input type="email" inputMode="email" placeholder="e-mail" autoComplete="email" maxLength={254} disabled={busy} {...form.register("email")} />
                </FormField>
              </div>
            ) : (
              <p className="text-left text-note">
                Enviando como <strong className="font-normal underline decoration-white/50 underline-offset-4">{known?.name}</strong> ({known?.email}).{" "}
                <button type="button" className="underline underline-offset-4 hover:text-sun-soft" onClick={() => setEditingIdentity(true)}>
                  Alterar
                </button>
              </p>
            )}

            <FormField label="Mensagem" error={errors.message?.message} hint={`${length}/${MAX} caracteres`}>
              <Textarea placeholder="Mensagem" maxLength={MAX} rows={4} disabled={busy} {...form.register("message")} />
            </FormField>

            {formError && (
              <p role="alert" className="rounded-pill bg-white/10 px-5 py-3 text-lg text-sun-soft">
                {formError}
              </p>
            )}

            <div className="pt-[clamp(1rem,2.4vw,3rem)]">
              <Button type="submit" size="hero" disabled={busy} className="min-w-[min(24.9rem,100%)]">
                {busy ? (
                  <>
                    <Spinner /> Enviando…
                  </>
                ) : (
                  "Enviar"
                )}
              </Button>
            </div>
          </form>
          <PrivacyNotice purpose="identificar quem nos escreveu" />
        </Column>
      </Section>
    </PublicPage>
  );
}
