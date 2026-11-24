// Conteúdo público do evento (vem do Figma). O endereço NÃO fica aqui: ele só
// é entregue pela API para quem tem o link pessoal do convite.
export const EVENT = {
  dateLabel: "23 de Janeiro de 2027",
  hosts: "Sofia e João",
  pixKey: import.meta.env.VITE_PIX_KEY as string | undefined,
  pixQrImage: import.meta.env.VITE_PIX_QR_IMAGE as string | undefined,
} as const;
