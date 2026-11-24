import { useEffect } from "react";

export function useDocumentTitle(title: string) {
  useEffect(() => {
    document.title = title ? `${title} – Chá de Casa Nova` : "Chá de Casa Nova";
  }, [title]);
}
