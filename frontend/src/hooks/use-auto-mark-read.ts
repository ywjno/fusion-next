import { useEffect } from "react";
import { useMarkItemsRead } from "@/queries/items";
import type { Item } from "@/lib/api";

export function useAutoMarkRead(article: Item | null, canToggleRead: boolean) {
  const markRead = useMarkItemsRead();

  useEffect(() => {
    if (article && article.unread && canToggleRead) {
      void markRead.mutateAsync([article.id]);
    }
  }, [article, canToggleRead, markRead]);
}
