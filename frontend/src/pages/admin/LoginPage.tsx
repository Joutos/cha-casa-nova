import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Navigate, useLocation, useNavigate } from "react-router-dom";
import { HeroHeader } from "@/components/HeroHeader";
import { Column, Section } from "@/components/Section";
import { FormField } from "@/components/FormField";
import { Spinner } from "@/components/Spinner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAuth } from "@/contexts/AuthContext";
import { loginSchema, type LoginForm } from "@/schemas/guest";
import { ApiError } from "@/services/api";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

export default function LoginPage() {
  useDocumentTitle("Entrar");
  const { admin, isLoading, login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as { from?: string } | null)?.from ?? "/admin";
  const [error, setError] = useState<string | null>(null);

  const form = useForm<LoginForm>({ resolver: zodResolver(loginSchema), defaultValues: { email: "", password: "" } });
  const { errors, isSubmitting } = form.formState;

  if (!isLoading && admin) return <Navigate to={from} replace />;

  const onSubmit = form.handleSubmit(async ({ email, password }) => {
    setError(null);
    try {
      await login(email, password);
      navigate(from.startsWith("/admin") ? from : "/admin", { replace: true });
    } catch (err) {
      form.resetField("password");
      setError(err instanceof ApiError ? err.message : "Não foi possível entrar agora.");
    }
  });

  return (
    <div className="flex min-h-dvh flex-col">
      <HeroHeader variant="slim" />
      <main id="conteudo" className="flex flex-1 flex-col">
        <Section className="flex-1 py-[clamp(3rem,6vw,7rem)]">
          <Column className="max-w-[34rem]">
            <h1 className="text-title">Área dos anfitriões</h1>
            <p className="mt-3 text-xl">Entre para ver quem confirmou presença e ler as mensagens.</p>
            <form onSubmit={onSubmit} noValidate className="mt-10 space-y-5" aria-busy={isSubmitting}>
              <FormField label="E-mail" error={errors.email?.message}>
                <Input type="email" placeholder="e-mail" autoComplete="username" disabled={isSubmitting} className="h-16 text-2xl" {...form.register("email")} />
              </FormField>
              <FormField label="Senha" error={errors.password?.message}>
                <Input type="password" placeholder="senha" autoComplete="current-password" disabled={isSubmitting} className="h-16 text-2xl" {...form.register("password")} />
              </FormField>
              {error && (
                <p role="alert" className="rounded-pill bg-white/10 px-5 py-3 text-lg text-sun-soft">
                  {error}
                </p>
              )}
              <Button type="submit" size="md" disabled={isSubmitting} className="mt-2 h-16 w-full text-2xl">
                {isSubmitting ? (
                  <>
                    <Spinner /> Entrando…
                  </>
                ) : (
                  "Entrar"
                )}
              </Button>
            </form>
          </Column>
        </Section>
      </main>
    </div>
  );
}
