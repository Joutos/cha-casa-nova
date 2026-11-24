/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL?: string;
  readonly VITE_PIX_KEY?: string;
  readonly VITE_PIX_QR_IMAGE?: string;
}
interface ImportMeta {
  readonly env: ImportMetaEnv;
}
