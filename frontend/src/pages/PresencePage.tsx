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
import { attendanceSchema, type AttendanceForm } from "@/schemas/guest";
import { confirmAttendance } from "@/services/public";
import { useRsvpFlow } from "@/contexts/RsvpFlowContext";
import { applyServerErrors } from "@/lib/form-errors";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

// Frame 2: confirmação de presença.
export default function PresencePage() {
  useDocumentTitle("Confirmar presença");
  const navigate = useNavigate();
  const flow = useRsvpFlow();
  const [formError, setFormError] = useState<string | null>(null);

  const form = useForm<AttendanceForm>({
    resolver: zodResolver(attendanceSchema),
    defaultValues: { name: flow.state.guest?.name ?? "", email: flow.state.guest?.email ?? "", website: "" },
  });
  const { errors } = form.formState;

  const mutation = useMutation({ mutationFn: confirmAttendance });

  const onSubmit = form.handleSubmit(async (values) => {
    setFormError(null);
    const data = attendanceSchema.parse(values);
    try {
      await mutation.mutateAsync({ ...data, website: values.website ?? "", contribute: flow.wantsContribution });
      flow.rememberGuest({ name: data.name, email: data.email });
      flow.complete("attend");
      navigate("/presenca/confirmado", { replace: true });
    } catch (err) {
      setFormError(applyServerErrors(err, form.setError, ["name", "email"] as const));
    }
  });

  const busy = form.formState.isSubmitting || mutation.isPending;

  return (
    <PublicPage>
      <Section className="flex-1 py-[clamp(3rem,6.6vw,8rem)]">
        <Column>
          <h1 className="text-title text-balance">Que alegria ter você com a gente! ♡</h1>
          <p className="mt-[1.1em] text-lead text-pretty">Ficamos muito felizes em saber que você deseja participar desse momento tão especial!</p>
          <p className="text-lead text-pretty">
            Para que possamos enviar seu convite com todos os detalhes do nosso Chá de Casa Nova, deixe seu nome e e-mail abaixo.
          </p>

          <form onSubmit={onSubmit} noValidate className="relative mx-auto mt-[clamp(2rem,3vw,3.5rem)] max-w-[49.3rem] space-y-5" aria-busy={busy}>
            <HoneypotField registration={form.register("website")} />
            <FormField label="Nome" error={errors.name?.message}>
              <Input placeholder="nome" autoComplete="name" maxLength={100} disabled={busy} {...form.register("name")} />
            </FormField>
            <FormField label="E-mail" error={errors.email?.message}>
              <Input type="email" inputMode="email" placeholder="e-mail" autoComplete="email" maxLength={254} disabled={busy} {...form.register("email")} />
            </FormField>

            <p className="pt-[clamp(1.5rem,2.4vw,2.8rem)] text-lead text-pretty">Estamos ansiosos para compartilhar essa nova fase com você!</p>

            {formError && (
              <p role="alert" className="rounded-pill bg-white/10 px-5 py-3 text-lg text-sun-soft">
                {formError}
              </p>
            )}

            <div className="pt-2">
              <Button type="submit" disabled={busy} className="min-w-[min(21rem,100%)]">
                {busy ? (
                  <>
                    <Spinner /> confirmando…
                  </>
                ) : (
                  "confirmar"
                )}
              </Button>
            </div>
          </form>
          <PrivacyNotice purpose="enviar o convite e organizar a lista de convidados" />
        </Column>
      </Section>
    </PublicPage>
  );
}
