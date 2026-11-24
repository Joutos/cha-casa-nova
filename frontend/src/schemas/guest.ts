import { z } from "zod";

// Espelha as regras do backend (internal/services/validation.go).
const name = z
  .string()
  .transform((v) => v.trim().replace(/\s+/g, " "))
  .pipe(
    z
      .string()
      .min(1, "Informe seu nome.")
      .min(2, "O nome precisa ter pelo menos 2 letras.")
      .max(100, "O nome pode ter no máximo 100 caracteres.")
      .refine((v) => !/[<>]/.test(v), "O nome contém caracteres não permitidos."),
  );

const email = z
  .string()
  .trim()
  .min(1, "Informe seu e-mail.")
  .max(254, "O e-mail é longo demais.")
  .email("Confira o e-mail: ele parece incompleto.");

export const attendanceSchema = z.object({
  name,
  email,
  website: z.string().max(0).optional().default(""), // honeypot
});

export const messageSchema = attendanceSchema.extend({
  message: z
    .string()
    .trim()
    .min(1, "Escreva sua mensagem antes de enviar.")
    .max(1000, "A mensagem pode ter no máximo 1000 caracteres."),
});

export const loginSchema = z.object({
  email: z.string().trim().min(1, "Informe seu e-mail.").email("Confira o e-mail."),
  password: z.string().min(1, "Informe sua senha."),
});

export type AttendanceForm = z.input<typeof attendanceSchema>;
export type MessageForm = z.input<typeof messageSchema>;
export type LoginForm = z.input<typeof loginSchema>;
