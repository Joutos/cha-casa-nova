import type { ReactNode } from "react";
import { useNavigate } from "react-router-dom";
import { HeroHeader } from "@/components/HeroHeader";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/contexts/AuthContext";

export function AdminLayout({ children }: { children: ReactNode }) {
  const { admin, logout } = useAuth();
  const navigate = useNavigate();
  return (
    <div className="flex min-h-dvh flex-col">
      <HeroHeader variant="slim" />
      <div className="bg-brand px-5 pt-6 sm:px-8">
        <div className="mx-auto flex max-w-[86rem] items-center justify-between gap-4 text-white">
          <p className="text-xl">Olá, {admin?.name.split(" ")[0]}</p>
          <Button
            variant="ghost"
            size="sm"
            onClick={async () => {
              await logout();
              navigate("/admin/login", { replace: true });
            }}
          >
            Sair
          </Button>
        </div>
      </div>
      <main id="conteudo" className="flex-1 bg-brand px-3 pb-16 sm:px-8">
        {children}
      </main>
    </div>
  );
}
