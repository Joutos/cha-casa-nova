import type { Config } from "tailwindcss";

// Tokens extraídos do Figma (ver docs/design-analysis.md).
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        brand: {
          DEFAULT: "#4365BF", // azul principal das seções e botões
          deep: "#2F4A99", // texto digitado e foco (contraste AA sobre o cinza)
        },
        mist: "#D9D9D9", // faixas, inputs e botões
        danger: "#A33A3A", // ações destrutivas no painel (não existe no Figma)
        sun: { DEFAULT: "#F5C04A", soft: "#FFE7A8" }, // miolo da flor; usado para erros de validação
        cream: "#E8E2D7", // cartão do convite
        cocoa: {
          DEFAULT: "rgb(71 49 40 / 0.74)", // texto sobre a faixa cinza
          overlay: "rgb(71 49 40 / 0.47)", // película sobre a foto
        },
        // Variáveis do shadcn/ui apontando para a paleta do Figma
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
      },
      fontFamily: {
        sans: ['"Hubballi"', '"Trebuchet MS"', "system-ui", "sans-serif"],
      },
      fontSize: {
        // Escala do Figma (1920px) com mínimo legível em telas pequenas.
        hero: ["clamp(2.6rem, 6.67vw, 8rem)", { lineHeight: "1.05" }], // 128px
        "hero-sub": ["clamp(1.35rem, 3.33vw, 4rem)", { lineHeight: "1.2" }], // 64px
        title: ["clamp(1.65rem, 2.5vw, 3rem)", { lineHeight: "1.25" }], // 48px
        lead: ["clamp(1.2rem, 1.875vw, 2.25rem)", { lineHeight: "1.3" }], // 36px
        action: ["clamp(1.5rem, 2.08vw, 2.5rem)", { lineHeight: "1.2" }], // 40px (≥24px = texto grande)
        note: ["clamp(0.95rem, 1.04vw, 1.25rem)", { lineHeight: "1.35" }], // 20px
      },
      borderRadius: {
        pill: "25px", // botões e campos
        card: "50px", // cartão do convite
      },
      boxShadow: {
        card: "8px 23px 18.5px 0 rgb(0 0 0 / 0.25)",
      },
      maxWidth: {
        column: "54.5rem", // ~870px, coluna de texto das telas internas
      },
      keyframes: {
        rise: { from: { opacity: "0", transform: "translateY(12px)" }, to: { opacity: "1", transform: "none" } },
      },
      animation: { rise: "rise 700ms cubic-bezier(.2,.7,.2,1) both" },
    },
  },
  plugins: [],
} satisfies Config;
